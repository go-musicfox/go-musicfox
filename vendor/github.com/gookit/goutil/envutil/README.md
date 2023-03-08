# Env Util

Provide some commonly ENV util functions.

## Install

```shell
go get github.com/gookit/goutil/envutil
```

## Go docs

- [Go docs](https://pkg.go.dev/github.com/gookit/goutil/envutil)

## Functions API

```go
func Environ() map[string]string
func GetBool(name string, def ...bool) bool
func GetInt(name string, def ...int) int
func Getenv(name string, def ...string) string
func HasShellEnv(shell string) bool
func IsConsole(out io.Writer) bool
func IsGithubActions() bool
func IsLinux() bool
func IsMSys() bool
func IsMac() bool
func IsSupport256Color() bool
func IsSupportColor() bool
func IsSupportTrueColor() bool
func IsTerminal(fd uintptr) bool
func IsWSL() bool
func IsWin() bool
func IsWindows() bool
func ParseEnvValue(val string) string
func ParseValue(val string) (newVal string)
func SetEnvs(mp map[string]string)
func StdIsTerminal() bool
func VarParse(val string) string
func VarReplace(s string) string
```

## Code Check & Testing

```bash
gofmt -w -l ./
golint ./...
```

**Testing**:

```shell
go test -v ./envutil/...
```

**Test limit by regexp**:

```shell
go test -v -run ^TestSetByKeys ./envutil/...
```
