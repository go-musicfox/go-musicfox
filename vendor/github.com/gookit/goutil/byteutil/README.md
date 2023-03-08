# Bytes Util

Provide some commonly bytes util functions.

## Install

```shell
go get github.com/gookit/goutil/byteutil
```

## Go docs

- [Go docs](https://pkg.go.dev/github.com/gookit/goutil/byteutil)

## Functions API

```go
var HexEncoder = NewStdEncoder(func(src []byte) []byte { ... }, func(src []byte) ([]byte, error) { ... }) ...
func Md5(src any) []byte
type Buffer struct{ ... }
func NewBuffer() *Buffer
type BytesEncoder interface{ ... }
type ChanPool struct{ ... }
func NewChanPool(maxSize int, width int, capWidth int) *ChanPool
type StdEncoder struct{ ... }
func NewStdEncoder(encFn func(src []byte) []byte, decFn func(src []byte) ([]byte, error)) *StdEncoder
```

## Code Check & Testing

```bash
gofmt -w -l ./
golint ./...
```

**Testing**:

```shell
go test -v ./byteutil/...
```

**Test limit by regexp**:

```shell
go test -v -run ^TestSetByKeys ./byteutil/...
```
