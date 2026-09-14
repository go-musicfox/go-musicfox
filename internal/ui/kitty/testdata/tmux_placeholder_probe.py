"""Check actual tmux client output using an isolated server and a PTY.

Invoked by TestUnicodePlaceholderTmuxOutput. No terminal GUI or image transfer
is required: the failure is lost combining marks before reaching the terminal.
"""
import fcntl
import json
import os
from pathlib import Path
import pty
import re
import select
import shlex
import struct
import subprocess
import sys
import termios
import time
import unicodedata

if len(sys.argv) > 2 and sys.argv[1] == "--emit":
    fixture = json.loads(Path(sys.argv[2]).read_text())
    checkpoint = Path(sys.argv[3])
    for index, frame in enumerate(fixture["frames"]):
        os.write(1, ("\x1b[H" + frame.replace("\n", "\r\n")).encode())
        deadline = time.monotonic() + 15
        ack = checkpoint.with_suffix(f".{index}")
        while not ack.exists():
            if time.monotonic() > deadline:
                raise TimeoutError("parent did not acknowledge frame")
            time.sleep(0.01)
    time.sleep(60)
    sys.exit()

fixture_path = Path(sys.argv[1]).resolve()
fixture = json.loads(fixture_path.read_text())
socket = "musicfox-placeholder-" + str(os.getpid())


def tmux(*args):
    return subprocess.check_output(["tmux", "-L", socket, *args], timeout=10)


def cells_from_output(data):
    # Only the VT operations emitted by this fixture are needed. Track cursor
    # motion and overwrites so an earlier correct draw cannot hide later damage.
    cells = {}
    x = y = 0
    tokens = re.findall(r"\x1b\[[0-?]*[ -/]*[@-~]|\x1b[()][0-9A-Za-z]|\x1b.|[^\x1b]", data)
    for token in tokens:
        if token.startswith("\x1b["):
            final, params = token[-1], token[2:-1]
            if params.startswith(("?", ">")):
                continue
            nums = [int(n or 0) for n in params.split(";")] if re.fullmatch(r"[0-9;]*", params) else [0]
            n = nums[0] or 1
            if final in "Hf":
                y, x = n - 1, (nums[1] or 1) - 1 if len(nums) > 1 else 0
            elif final == "d": y = n - 1
            elif final in "G`": x = n - 1
            elif final == "A": y = max(0, y - n)
            elif final == "B": y += n
            elif final == "C": x += n
            elif final == "D": x = max(0, x - n)
            elif final == "J" and nums[0] == 2: cells.clear()
            elif final in "KX":
                for pos in list(cells):
                    if pos[1] == y and (
                        (final == "X" and x <= pos[0] < x + n) or
                        (final == "K" and (nums[0] == 2 or
                         (nums[0] == 0 and pos[0] >= x) or
                         (nums[0] == 1 and pos[0] <= x)))
                    ): del cells[pos]
            continue
        if token.startswith("\x1b"): continue
        if token == "\r": x = 0
        elif token == "\n": y += 1
        elif token == "\b": x = max(0, x - 1)
        elif unicodedata.category(token).startswith("M"):
            cells[x - 1, y] = cells.get((x - 1, y), "") + token
        elif ord(token) >= 32:
            cells[x, y] = token
            x += 1
    return cells


master = None
client = None
try:
    tmux("-f", "/dev/null", "new-session", "-d", "-s", "probe", "-x", "160", "-y", "24", "sleep 60")
    tmux("set-option", "-g", "status", "off")
    right = tmux("split-window", "-h", "-P", "-F", "#{pane_id}", "sleep 60").decode().strip()
    left_pane = tmux("display-message", "-p", "-t", "probe:0.0", "#{pane_id}").decode().strip()
    master, slave = pty.openpty()
    fcntl.ioctl(slave, termios.TIOCSWINSZ, struct.pack("HHHH", 24, 160, 0, 0))

    def setup():
        os.setsid()
        fcntl.ioctl(slave, termios.TIOCSCTTY, 0)

    env = {**os.environ, "TERM": "xterm-256color"}
    client = subprocess.Popen(["tmux", "-L", socket, "attach", "-t", "probe"],
                              stdin=slave, stdout=slave, stderr=slave, preexec_fn=setup, env=env)
    os.close(slave)

    def read_output():
        out = b""
        deadline = time.monotonic() + 0.7
        while time.monotonic() < deadline:
            if select.select([master], [], [], 0.05)[0]:
                out += os.read(master, 65536)
        return out

    transcript = read_output()
    assert client.poll() is None, transcript.decode(errors="replace")
    for side, target in [("right", right), ("left", left_pane)]:
        checkpoint = fixture_path.parent / side
        command = shlex.join([sys.executable, str(Path(__file__).resolve()), "--emit", str(fixture_path), str(checkpoint)])
        tmux("respawn-pane", "-k", "-t", target, command)
        left = int(tmux("display-message", "-p", "-t", target, "#{pane_left}"))
        for frame in range(len(fixture["frames"])):
            transcript += read_output()
            cells = cells_from_output(transcript.decode("utf-8"))
            for row, want in enumerate(fixture["want"]):
                got = "".join(cells.get((left + 15 + col, row), "") for col in range(10))
                assert got == want, f"{side} frame {frame} row {row}: terminal received {got!a}, expected {want!a}"
            print(f"{side} frame {frame}: all 50 cells valid before any subsequent redraw")
            checkpoint.with_suffix(f".{frame}").touch()
    tmux("swap-pane", "-s", right, "-t", left_pane)
    transcript += read_output()
    cells = cells_from_output(transcript.decode("utf-8"))
    for target in [right, left_pane]:
        left = int(tmux("display-message", "-p", "-t", target, "#{pane_left}"))
        for row, want in enumerate(fixture["want"]):
            assert "".join(cells.get((left + 15 + col, row), "") for col in range(10)) == want
    print("swap: placeholder coordinates remain intact")
finally:
    subprocess.run(["tmux", "-L", socket, "kill-server"], stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
    if client is not None: client.wait(timeout=5)
    if master is not None: os.close(master)
