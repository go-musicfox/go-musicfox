# gohook

Cross-platform global keyboard and mouse hook library for Go. A cgo wrapper around
[libuiohook](https://github.com/kwhat/libuiohook); part of the [robotgo](https://github.com/go-vgo/robotgo) ecosystem.
Module path: `github.com/robotn/gohook`, package `hook`.

## Build/Test/Lint Commands

- **Build**: `go build -v .`
- **Test**: `go test -v .` (single test: `go test -v -run TestKey .`)
- **Lint/Vet**: `go vet .`
- **Format**: `gofmt -w .`
- **Run example**: `go run ./examples` (interactive — listens for real key/mouse events)

> Requires cgo (a C toolchain). On Linux you also need X11 dev libs — see
> [robotgo requirements](https://github.com/go-vgo/robotgo#requirements):
> `libx11-dev xorg-dev libxtst-dev`, `libxkbcommon-dev`, `libxcb*`. macOS links `-framework Cocoa`.

## Architecture

Thin Go API on top of platform C code. `hook.go` exposes `Start()`/`End()` (channel-based
event stream) and `Register()`/`Process()` (callback-based). Events arrive from C via the
`go_send` cgo export in `extern.go`, are JSON-decoded into `Event`, and pushed onto the global
`ev` channel. A poll goroutine drives `C.pollEv()` on an interval.

```
.
├── hook.go        # Public API: Event, Register, Process, Start, End, keychar<->rawcode
├── extern.go      # //export go_send — C->Go event bridge (JSON unmarshal into Event)
├── event.go       # AddEvent / AddEvents / AddMouse / AddMousePos blocking helpers
├── keycode.go     # Re-exports vcaesar/keycode maps: Keycode, MouseMap, Special
├── tables.go      # Rawcode<->keychar maps per OS (Darwin/Win/Linux)
├── event/         # C headers compiled via cgo (goEvent.h is the include entry point)
├── hook/          # libuiohook headers, split by OS: x11/ darwin/ windows/
├── chan/          # eb_chan.h (C channel impl)
└── examples/      # Runnable usage examples (examples/main.go, examples/event/main.go)
```

## Code Style

- **License header**: Every `.go` file starts with the go-vgo Apache/MIT dual-license block
  (copy from any existing file). New files should include it.
- **Imports**: stdlib group first, then third-party, blank-line separated. cgo files put the
  `/* ... */ import "C"` preamble (with `#cgo` flags + `#include`) immediately before other imports.
- **Naming**: Exported API is CamelCase (`Register`, `AddMouse`); package-level state is lowercase
  (`ev`, `pressed`, `asyncon`). Event-kind and key constants are untyped `const` ints in `hook.go`.
- **Errors**: This package uses `log.Fatal` on JSON decode failure in `extern.go` (not idiomatic
  error returns) — match the existing pattern when editing that path, but prefer returning errors
  in new standalone code.
- **Formatting**: tabs (gofmt). Run `gofmt -w .` before committing.

## Testing

- Framework: standard `testing` + assertion helper `github.com/vcaesar/tt` (`tt.Equal`, `tt.NotNil`).
- Tests live next to source as `*_test.go` (e.g. `hook_test.go`), package `hook`.
- Tests are OS-aware: `TestKey` expects `"a"` for rawcode 0 on darwin, `"error"` elsewhere
  (`runtime.GOOS` branches). Account for the current platform when adding assertions.
- Run a single test: `go test -v -run TestName .`

## Key Patterns

- **Global mutable state**: hooks/callbacks live in package-level maps (`keys`, `cbs`, `events`,
  `pressed`) guarded loosely by `lck sync.RWMutex` and the `asyncon` flag. `End()` resets all of it
  and closes `ev`. Only one hook session is active at a time — `Start()` recreates the `ev` channel.
- **Two usage modes**: (1) channel — `for ev := range hook.Start()`; (2) callback —
  `hook.Register(...)` then `<-hook.Process(hook.Start())`. See `README.md` and `examples/`.
- **Event kinds are remapped constants** in `hook.go` (e.g. `KeyDown=4`, `KeyHold=3`, `MouseUp=6`) —
  do NOT assume sequential values; they intentionally differ from libuiohook's raw ids.
- **OS-specific keymaps**: `tables.go` defines `rawToKeyDarwin`/`keyToRawDarwin`,
  `raw2keyWin`/`key2rawWin`, `raw2keyLinux`/`key2RawLinux`. `RawcodetoKeychar`/`KeychartoRawcode`
  branch on `runtime.GOOS`.
- **cgo include entry**: Go cgo files `#include "event/goEvent.h"`, which pulls in the
  `hook/<os>/` headers. Editing C behavior means touching `event/` and `hook/<os>/`.

## CI

`.github/workflows/go.yml` runs on `macOS-latest` and `windows-latest` (ubuntu commented out):
`go get -v -t -d ./...` → `go build -v .` → `go test -v .`. Also configured: `circle.yml`, AppVeyor.

## Dependencies

- `github.com/vcaesar/keycode` — source of the `Keycode`, `MouseMap`, `Special` maps (re-exported in `keycode.go`).
- `github.com/vcaesar/tt` — lightweight test assertions.
- **libuiohook** (vendored C in `hook/`) — the underlying native global-hook engine.
- Go 1.17+ (per `go.mod`).
