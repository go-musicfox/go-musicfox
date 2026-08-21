# Browser provider: agent-browser

This is the local, self-provisioning `agent-browser` implementation of the
browser-provider contract in `TEMPLATE.md`. It uses native release binaries and
Chrome for Testing. It never requires Node, a project package manager, a
preinstalled browser, or a cloud-browser account.

## Pinned release

Installs are pinned to `agent-browser` **v0.34.0**; every downloaded binary is
verified against the SHA-256 sums recorded in the install scripts below before
it is executed. A binary that fails verification is deleted and the operation
aborts. To bump the pin, update the version and the per-asset sums together —
the sums are the asset `digest` fields the GitHub releases API reports at
`/repos/vercel-labs/agent-browser/releases/tags/<tag>`.

## Prerequisites

- Network access to GitHub Releases and the Chrome-for-Testing download host on
  first install. Warm runs reuse the cached binary and browser.
- Supported release targets: macOS x64/arm64; Linux glibc or musl x64/arm64;
  Windows x64. WSL2 uses the matching Linux target. Windows on ARM may use the
  x64 binary only when the operating system's x64 compatibility layer is active.
- Linux Chrome libraries may require root. The operation below performs the
  install itself when already root or passwordless elevation is available; it
  never delegates commands to the operator.

## Operations

### ensure-installed

Use an existing healthy `agent-browser` from `PATH`; otherwise install the
pinned official release in a per-user cache, verifying the SHA-256 sum of the
downloaded binary before it is ever executed. Do not add binary files to the
repository.

POSIX shell (macOS, Linux, WSL2, Git Bash/MSYS):

```bash
if command -v agent-browser >/dev/null 2>&1; then
  AGENT_BROWSER_BIN=$(command -v agent-browser)
else
  AGENT_BROWSER_VERSION=v0.34.0
  CACHE_ROOT=${XDG_CACHE_HOME:-"$HOME/.cache"}
  TOOL_DIR="$CACHE_ROOT/agent-tools/agent-browser/$AGENT_BROWSER_VERSION"
  mkdir -p "$TOOL_DIR"
  OS=$(uname -s 2>/dev/null || echo unknown)
  ARCH=$(uname -m 2>/dev/null || echo unknown)
  case "$ARCH" in x86_64|amd64) ARCH=x64 ;; arm64|aarch64) ARCH=arm64 ;; *) ARCH=unsupported ;; esac
  case "$OS" in
    Darwin) ASSET="agent-browser-darwin-$ARCH" ;;
    Linux)
      LIBC=linux
      (ldd --version 2>&1 || true) | grep -qi musl && LIBC=linux-musl
      ASSET="agent-browser-$LIBC-$ARCH"
      ;;
    MINGW*|MSYS*|CYGWIN*) ASSET=agent-browser-win32-x64.exe ;;
    *) ASSET=unsupported ;;
  esac
  case "$ASSET" in *unsupported*) echo "Unsupported agent-browser target: $OS/$ARCH" >&2; exit 1 ;; esac
  case "$ASSET" in
    agent-browser-darwin-arm64) ASSET_SHA256=d680a7a96ab86e9ab9d2b571b12919b761e93682ad1de714bbd5ac849c8d7c9c ;;
    agent-browser-darwin-x64) ASSET_SHA256=dad3c9f9e67791a44a768a98847510c61a7b568a0499c602632b8aee411101e7 ;;
    agent-browser-linux-arm64) ASSET_SHA256=ca70bf7c2d269a152b3824cbb65befb7b8258b8aa1cf34767c64ada2abc3d7c8 ;;
    agent-browser-linux-musl-arm64) ASSET_SHA256=c0864fb206e321af48a46fb8331cf08ae60b3fcc1046232c1d1c842db4fc40ca ;;
    agent-browser-linux-musl-x64) ASSET_SHA256=dd4752ba1def81c7443504c284b6559d28dad8ecd02b5faeca6caf4fc1fb948e ;;
    agent-browser-linux-x64) ASSET_SHA256=69eadf5d8d6003a06a5cd2f914ebb261c7754fe1335a9190122c334e91909789 ;;
    agent-browser-win32-x64.exe) ASSET_SHA256=604820a9e86cdb8bba46da737fc0edb31bc92de6691c73dbc61d3673c370a6b5 ;;
  esac
  AGENT_BROWSER_BIN="$TOOL_DIR/$ASSET"
  if [ ! -x "$AGENT_BROWSER_BIN" ]; then
    URL="https://github.com/vercel-labs/agent-browser/releases/download/$AGENT_BROWSER_VERSION/$ASSET"
    TMP="$AGENT_BROWSER_BIN.tmp.$$"
    if command -v curl >/dev/null 2>&1; then curl -fL --retry 3 --connect-timeout 30 --max-time 600 -o "$TMP" "$URL"
    elif command -v wget >/dev/null 2>&1; then wget -T 30 -t 3 -O "$TMP" "$URL"
    else echo "No built-in HTTP downloader is available" >&2; exit 1
    fi
    if command -v sha256sum >/dev/null 2>&1; then GOT_SHA256=$(sha256sum "$TMP" | awk '{print $1}')
    else GOT_SHA256=$(shasum -a 256 "$TMP" | awk '{print $1}')
    fi
    if [ "$GOT_SHA256" != "$ASSET_SHA256" ]; then
      rm -f "$TMP"
      echo "agent-browser $AGENT_BROWSER_VERSION $ASSET checksum mismatch: expected $ASSET_SHA256, got $GOT_SHA256" >&2
      exit 1
    fi
    chmod 755 "$TMP"
    mv "$TMP" "$AGENT_BROWSER_BIN"
  fi
fi

"$AGENT_BROWSER_BIN" install
if ! "$AGENT_BROWSER_BIN" doctor --json >/dev/null 2>&1; then
  if [ "$(uname -s 2>/dev/null || true)" = Linux ]; then
    if [ "$(id -u)" = 0 ]; then
      "$AGENT_BROWSER_BIN" install --with-deps
    elif command -v sudo >/dev/null 2>&1 && sudo -n true 2>/dev/null; then
      sudo -n "$AGENT_BROWSER_BIN" install --with-deps
    fi
  fi
fi
if "$AGENT_BROWSER_BIN" doctor --json >/dev/null; then
  printf 'BROWSER_PROVIDER=agent-browser\nBROWSER_INSTALLED=1\nBROWSER_COMMAND=%s\nBROWSER_VERSION=%s\nBROWSER_NOTES=\n' \
    "$AGENT_BROWSER_BIN" "$("$AGENT_BROWSER_BIN" --version 2>/dev/null || echo unknown)"
else
  printf 'BROWSER_PROVIDER=agent-browser\nBROWSER_INSTALLED=0\nBROWSER_COMMAND=%s\nBROWSER_VERSION=unknown\nBROWSER_NOTES=live browser launch failed after autonomous install\n' "$AGENT_BROWSER_BIN"
  exit 1
fi
```

Native Windows PowerShell:

```powershell
$onPath = Get-Command agent-browser -ErrorAction SilentlyContinue
if ($onPath) { $AgentBrowser = $onPath.Source }
else {
  $arch = [System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture.ToString()
  if ($arch -notin 'X64','Arm64') { throw "Unsupported agent-browser Windows architecture: $arch" }
  $pinnedVersion = 'v0.34.0'
  $expectedSha256 = '604820a9e86cdb8bba46da737fc0edb31bc92de6691c73dbc61d3673c370a6b5'
  $toolDir = Join-Path ([Environment]::GetFolderPath('LocalApplicationData')) "agent-tools/agent-browser/$pinnedVersion"
  New-Item -ItemType Directory -Force -Path $toolDir | Out-Null
  $AgentBrowser = Join-Path $toolDir 'agent-browser-win32-x64.exe'
  if (-not (Test-Path $AgentBrowser)) {
    $url = "https://github.com/vercel-labs/agent-browser/releases/download/$pinnedVersion/agent-browser-win32-x64.exe"
    $tmp = "$AgentBrowser.tmp.$PID"
    if ($PSVersionTable.PSVersion.Major -lt 7) {
      [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12
    }
    Invoke-WebRequest -UseBasicParsing -TimeoutSec 600 -Uri $url -OutFile $tmp
    $gotSha256 = (Get-FileHash -Algorithm SHA256 -Path $tmp).Hash.ToLowerInvariant()
    if ($gotSha256 -ne $expectedSha256) {
      Remove-Item -Force $tmp
      throw "agent-browser $pinnedVersion checksum mismatch: expected $expectedSha256, got $gotSha256"
    }
    Move-Item -Force $tmp $AgentBrowser
  }
}
& $AgentBrowser install
& $AgentBrowser doctor --json | Out-Null
if ($LASTEXITCODE -ne 0) { throw 'agent-browser live browser launch failed after autonomous install' }
$version = (& $AgentBrowser --version 2>$null)
"BROWSER_PROVIDER=agent-browser"
"BROWSER_INSTALLED=1"
"BROWSER_COMMAND=$AgentBrowser"
"BROWSER_VERSION=$version"
"BROWSER_NOTES="
```

### doctor

```bash
"$AGENT_BROWSER_BIN" doctor --json
```

PowerShell: `& $AgentBrowser doctor --json`.

### open

Use a unique, validated session id such as `qa-<runId>`:

```bash
"$AGENT_BROWSER_BIN" --session "$BROWSER_SESSION" open "$BASE_URL" --json
```

### snapshot

```bash
"$AGENT_BROWSER_BIN" --session "$BROWSER_SESSION" snapshot -i --json
```

### interact

Use only a ref returned by the latest snapshot:

```bash
"$AGENT_BROWSER_BIN" --session "$BROWSER_SESSION" click "$ELEMENT_REF" --json
"$AGENT_BROWSER_BIN" --session "$BROWSER_SESSION" fill "$ELEMENT_REF" "$VALUE" --json
```

Other actions use the matching CLI command shown by
`"$AGENT_BROWSER_BIN" --help`; never interpolate untrusted shell fragments.

### assert

Use JSON output and compare the observed value in the current shell. Examples:

```bash
"$AGENT_BROWSER_BIN" --session "$BROWSER_SESSION" get text "$ELEMENT_REF" --json
"$AGENT_BROWSER_BIN" --session "$BROWSER_SESSION" get url --json
"$AGENT_BROWSER_BIN" --session "$BROWSER_SESSION" is visible "$ELEMENT_REF" --json
```

### screenshot

```bash
SCREENSHOT_DIR=$(dirname "$SCREENSHOT_PATH")
mkdir -p "$SCREENSHOT_DIR"
ABS_SCREENSHOT_PATH=$(cd "$SCREENSHOT_DIR" && pwd)/$(basename "$SCREENSHOT_PATH")
"$AGENT_BROWSER_BIN" --session "$BROWSER_SESSION" screenshot --full "$ABS_SCREENSHOT_PATH" --json
test -s "$ABS_SCREENSHOT_PATH"
```

The absolute path avoids the CLI treating a relative multi-segment path as a
selector. PowerShell resolves it with
`[IO.Path]::GetFullPath($ScreenshotPath)`, creates the parent directory, then
checks `Test-Path` and a non-zero file length.

### close

```bash
"$AGENT_BROWSER_BIN" --session "$BROWSER_SESSION" close --json 2>/dev/null || true
```

In PowerShell, run the same arguments with `& $AgentBrowser` in `finally` and
ignore only an already-closed-session error.

## Rules

- Run `agent-browser skills get core` when available before a complex scenario
  so interaction guidance matches the installed CLI version.
- Use a unique session per QA/test run. Never attach to a user's normal browser
  profile unless the operator explicitly requested that profile.
- Keep all operation targets local to the application under test. Do not enable
  a cloud provider or send credentials to a remote browser service.
