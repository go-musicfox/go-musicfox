# Phase 2 — Generate the entrypoint (first run, `--regenerate`, or repair)

Full procedure for the generation phase of `om-prepare-test-env`, steps 2.1–2.6,
plus the fallback for when a reliable script cannot be generated. Entered when no
marker-bearing entrypoint exists, on `--regenerate`, or in repair mode after a
Phase 1 script failure.

### 2.1 Read the repo's own instructions, detect the platform

Read the agent instruction files (`AGENTS.md`, `CLAUDE.md`, README,
`CONTRIBUTING.md`) — most repos document how to run and test the app there.
Record the platform, which also picks the script flavor to generate. From a
POSIX shell:

```bash
UNAME=$(uname -s 2>/dev/null || echo unknown)
case "$UNAME" in
  Linux*)  grep -qiE 'microsoft|wsl' /proc/version 2>/dev/null && PLATFORM=wsl2 || PLATFORM=linux ;;
  Darwin*) PLATFORM=darwin ;;
  MINGW*|MSYS*|CYGWIN*) PLATFORM=win32 ;;   # Git Bash/MSYS: Windows, but sh works -> .sh flavor
  *) PLATFORM=unknown ;;
esac
```

`uname` does not exist in PowerShell — when the working shell is PowerShell,
detect with built-ins instead (PowerShell 5.x only runs on Windows; 7+ sets
the `$Is*` automatics):

```powershell
$Platform = if ($PSVersionTable.PSVersion.Major -lt 6 -or $IsWindows) { 'win32' }
            elseif ($IsMacOS) { 'darwin' } else { 'linux' }
```

Flavor choice: generate `.sh` whenever a POSIX shell is available on the
user's machine (macOS, Linux, WSL2, Git Bash — probe with
`sh -c 'echo ok'`); generate `.ps1` when the user is on native Windows
without one. When in doubt on Windows, `.ps1` is the safe default — PowerShell
is always present there.

Platform notes to honor in everything generated:

- **WSL2** — work inside the WSL filesystem (`~/…`), not `/mnt/c/…`: bind-mount
  I/O across the Windows boundary is an order of magnitude slower and breaks
  file watchers. Docker comes from Docker Desktop's WSL integration — verify
  with `docker info` and, when it fails, say so (the fix is a Docker Desktop
  setting, not a script change). An app bound to `127.0.0.1` inside WSL2 is
  reachable from Windows browsers at `localhost` (auto-forwarded), so the
  descriptor's `baseUrl` works for both sides.
- **Line endings** — generated `.sh` scripts must be written with **LF** line
  endings and committed with a `.gitattributes` rule so a Windows checkout
  (`core.autocrlf=true`) cannot re-write them to CRLF, which breaks `sh` with
  `\r: command not found`:

  ```gitattributes
  *.sh  text eol=lf
  *.ps1 text
  ```

  Add these rules (create `.gitattributes` if missing) whenever the scripts
  are committed. Keep generated `.ps1` content ASCII-only: Windows
  PowerShell 5.x misreads UTF-8 without a BOM, and ASCII sidesteps the whole
  encoding question.
- **Paths** — use forward slashes and relative paths everywhere; they work in
  `sh`, PowerShell, git, and Node alike. Never embed an absolute path or a
  drive letter in a generated script; derive locations from the repo root at
  run time.

Portability rules for everything generated: one entrypoint in the platform's
native flavor (POSIX `sh`, or PowerShell on native Windows) implementing the
same contract; Docker/Compose for services (identical across all four
platforms); bind to `127.0.0.1`; never hardcode a path separator or a port
that might be taken.

### 2.2 Discover how the project runs

Gather everything the script will need. Discover from evidence, never assume:

1. **The repo's own ephemeral/test environment** — package scripts, `Makefile`,
   `Taskfile.yml`, `justfile` targets named like `test:*:ephemeral`, `test-env`,
   `e2e:setup`, `dev:test`, `db:test`; CI workflows (`.github/workflows/*` —
   prefer whatever CI actually runs); `docker-compose*.yml` / `compose*.yml`,
   `Dockerfile`, `.devcontainer/`. When a usable environment exists, **the
   generated script wraps the repo's own up-command** — including its own
   reuse/caching flags when it has them — and never re-implements what the repo
   already does.
2. **The preparation chain** for fresh checkouts and worktrees: install
   dependencies → run code generation → build workspace packages/app, in the
   order the repo's scripts and CI imply (codegen before builds whenever builds
   consume generated artifacts). A discovered env command almost always assumes
   a fully prepared workspace; the script must run this chain — gated by the
   build cache — before the env command.
3. **Backing services**, from evidence: compose service images; ORM/driver
   dependencies in the manifest; connection env vars in `.env.example`
   (`DATABASE_URL`, `MONGO_URL`, `REDIS_URL`, …); migration folders. Produce a
   concrete list with image and port per service (`postgres:16`, `mysql:8`,
   `mongo:7`, `redis:7`), honoring versions the repo pins.
4. **The launch command and port** — in order of preference: Docker/compose
   (parity with production), the dev script (fast startup, best for iterative
   QA), or build + start/serve. Read the port from framework config, dev-server
   output, `EXPOSE`, or compose `ports` — never invent one.
5. **Build inputs and artifacts** for the build cache: source directories,
   lockfiles, build configs (inputs); the build outputs that must exist
   (artifacts); env vars that shape the build output (mode flags, feature
   toggles).

When `auto` is ambiguous — e.g. the app has a database but also a plain dev
server and the user has not said whether they want disposable services — ask
once with `AskUserQuestion` (ephemeral vs. run-against-existing) and encode the
answer in the script. Do not provision containers the user did not ask for.

### 2.3 Write the scripts

Generate `$UP_SCRIPT` and `$DOWN_SCRIPT` implementing the **entrypoint
contract** (`references/entrypoint-contract.md`), with every project-specific
fact discovered in 2.2 baked in as variables at the top of the script
(commands, service images, ports, build inputs, artifacts, env vars). The
scripts are the durable, committed-friendly artifact — plain `sh` (or plain
PowerShell), no agent needed to run them, readable and tweakable by a human.
Write them with LF line endings and add the `.gitattributes` rules from 2.1 in
the same change.

### 2.4 Ensure the configured browser provider (generation-time)

Unless `--browser off`, resolve `BROWSER_PROVIDER` and `BROWSER_FILE` from the
config-loading step (or the `--browser-provider` override):

Validate the final provider name against `^[A-Za-z0-9._-]+$` before
interpolating it into `.ai/browsers/<provider>.md`; reject any other value.

1. If the provider was explicit and `$BROWSER_FILE` is missing, run
   `om-setup-agent-pipeline` to install the selected descriptor, reload, and
   continue. For a config created before browser descriptors existed, the
   implicit Playwright provider may use the legacy Playwright flow below.
2. Read `$BROWSER_FILE`; execute its **ensure-installed** operation, then its
   **doctor** operation. The provider owns platform detection, autonomous CLI
   and browser-engine installation, and live-launch verification. The agent runs
   those operations itself — never hand prerequisite commands to the operator.
3. Independently discover any repository-native committed test runner
   (`playwright.config.*`, `cypress.config.*`, `wdio.conf.*`, or equivalent) and
   record it as `testRunner`. Selecting agent-browser for exploration and
   screenshots never replaces an existing suite.
4. Record the provider operation outputs in the descriptor's `browser` object:
   provider, installed, command, version, descriptor path, and notes. When the
   provider is Playwright, also write the legacy `playwright` object so older
   consumers keep working.
5. Where installation or **doctor** fails after the descriptor exhausts its
   autonomous paths, do not fail the app boot: record `browser.installed: false`
   and the concrete blocker in `notes` so consumers can still run API checks and
   report the browser limitation honestly.

Legacy Playwright fallback: use the repository's existing Playwright setup when
present. Otherwise install it with the package manager selected by the lockfile,
install Chromium and available OS dependencies, and write a minimal shared
config at `$QA_DIR/playwright.config.ts` reading `process.env.BASE_URL`.

Provider provisioning runs once here. The generated entrypoint records the
verified state on warm runs; it never reinstalls a healthy provider.

### 2.5 Verify the script — cold and warm

Generation is not done until the script has proven both halves of its promise:

1. **Cold run**: execute `$UP_SCRIPT` in the real workspace. It must end with a
   healthy environment and a valid descriptor. Note the wall time.
2. **Warm run**: execute `$UP_SCRIPT` again immediately. It must detect the
   running environment via the reuse check and return in seconds — same
   `runId`, no rebuild, no re-provision. If it rebuilds or re-provisions, the
   reuse/cache logic is broken: fix the script and repeat.
3. **Cache-invalidation spot check** (cheap): `touch` one tracked source file,
   confirm the script's fingerprint changes (the script should report it would
   rebuild — or actually rebuild if fast). This proves the cache is not
   permanently stale.

Record both timings in the descriptor `notes`
(e.g. `cold: 184s, warm: 3s (reused)`), so regressions are visible later. Only
after the warm run passes is Phase 2 complete.

### 2.6 Report

Print: the scripts' paths (`$UP_SCRIPT`, `$DOWN_SCRIPT`), the descriptor path,
the base URL, cold/warm timings, and the one-liner for next time, in the form
that runs on **this** platform — `sh .ai/scripts/test-env-up.sh` on
macOS/Linux/WSL2/Git Bash, `pwsh -File .ai/scripts/test-env-up.ps1` on native
Windows (or re-invoking this skill, which now just runs it). Recommend
committing the scripts — plus the `.gitattributes` line-ending rules from 2.1 —
so worktrees and teammates get the fast path for free.

## When a reliable script cannot be generated

Some environments resist scripting — interactive logins, hardware dependencies,
a stack the up-command cannot drive headlessly. If after two repair attempts the
script still cannot pass the cold+warm verification:

1. Record **why** in the repo-local skill (`.ai/skills/om-prepare-test-env/SKILL.md`),
   with whatever partial script did work committed anyway (e.g. services-up
   without the app).
2. Fall back to the agent-driven flow for the unscripted part on each run —
   expensive, but correct — and still write the descriptor so consumers attach
   normally.
3. Re-attempt generation when the blocker changes (the repo-local skill note is
   the reminder). Failing to script is acceptable; failing silently is not —
   the report must say the fast path is unavailable and why.
