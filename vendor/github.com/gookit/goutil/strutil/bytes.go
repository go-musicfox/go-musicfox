package strutil

import (
	"bytes"
	"fmt"
	"strings"
)

// Buffer wrap and extends the bytes.Buffer
type Buffer struct {
	bytes.Buffer
}

// NewBuffer instance
func NewBuffer() *Buffer {
	return &Buffer{}
}

// WriteAny type value to buffer
func (b *Buffer) WriteAny(vs ...any) {
	for _, v := range vs {
		_, _ = b.Buffer.WriteString(fmt.Sprint(v))
	}
}

// QuietWriteByte to buffer
func (b *Buffer) QuietWriteByte(c byte) {
	_ = b.WriteByte(c)
}

// QuietWritef write message to buffer
func (b *Buffer) QuietWritef(tpl string, vs ...any) {
	_, _ = b.WriteString(fmt.Sprintf(tpl, vs...))
}

// Writeln write message to buffer with newline
func (b *Buffer) Writeln(ss ...string) {
	b.QuietWriteln(ss...)
}

// QuietWriteln write message to buffer with newline
func (b *Buffer) QuietWriteln(ss ...string) {
	_, _ = b.WriteString(strings.Join(ss, ""))
	_ = b.WriteByte('\n')
}

// QuietWriteString to buffer
func (b *Buffer) QuietWriteString(ss ...string) {
	_, _ = b.WriteString(strings.Join(ss, ""))
}

// MustWriteString to buffer
func (b *Buffer) MustWriteString(ss ...string) {
	_, err := b.WriteString(strings.Join(ss, ""))
	if err != nil {
		panic(err)
	}
}

// ResetAndGet buffer string.
func (b *Buffer) ResetAndGet() string {
	s := b.String()
	b.Reset()
	return s
}

// ByteChanPool struct
//
// Usage:
//
//	bp := strutil.NewByteChanPool(500, 1024, 1024)
//	buf:=bp.Get()
//	defer bp.Put(buf)
//	// use buf do something ...
//
// refer https://www.flysnow.org/2020/08/21/golang-chan-byte-pool.html
// from https://github.com/minio/minio/blob/master/internal/bpool/bpool.go
type ByteChanPool struct {
	c    chan []byte
	w    int
	wcap int
}

// NewByteChanPool instance
func NewByteChanPool(maxSize int, width int, capWidth int) *ByteChanPool {
	return &ByteChanPool{
		c:    make(chan []byte, maxSize),
		w:    width,
		wcap: capWidth,
	}
}

// Get gets a []byte from the BytePool, or creates a new one if none are
// available in the pool.
func (bp *ByteChanPool) Get() (b []byte) {
	select {
	case b = <-bp.c:
	// reuse existing buffer
	default:
		// create new buffer
		if bp.wcap > 0 {
			b = make([]byte, bp.w, bp.wcap)
		} else {
			b = make([]byte, bp.w)
		}
	}
	return
}

// Put returns the given Buffer to the BytePool.
func (bp *ByteChanPool) Put(b []byte) {
	select {
	case bp.c <- b:
		// buffer went back into pool
	default:
		// buffer didn't go back into pool, just discard
	}
}

// Width returns the width of the byte arrays in this pool.
func (bp *ByteChanPool) Width() (n int) {
	return bp.w
}

// WidthCap returns the cap width of the byte arrays in this pool.
func (bp *ByteChanPool) WidthCap() (n int) {
	return bp.wcap
}
