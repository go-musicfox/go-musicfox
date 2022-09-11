# bitio

[![Build Status](https://travis-ci.org/icza/bitio.svg?branch=master)](https://travis-ci.org/icza/bitio)
[![GoDoc](https://godoc.org/github.com/icza/bitio?status.svg)](https://godoc.org/github.com/icza/bitio)
[![Go Report Card](https://goreportcard.com/badge/github.com/icza/bitio)](https://goreportcard.com/report/github.com/icza/bitio)
[![codecov](https://codecov.io/gh/icza/bitio/branch/master/graph/badge.svg)](https://codecov.io/gh/icza/bitio)

Package `bitio` provides an optimized bit-level `Reader` and `Writer` for Go.

You can use `Reader.ReadBits()` to read arbitrary number of bits from an `io.Reader` and return it as an `uint64`,
and `Writer.WriteBits()` to write arbitrary number of bits of an `uint64` value to an `io.Writer`.

Both `Reader` and `Writer` also provide optimized methods for reading / writing
1 bit of information in the form of a `bool` value: `Reader.ReadBool()` and `Writer.WriteBool()`.
These make this package ideal for compression algorithms that use [Huffman coding](https://en.wikipedia.org/wiki/Huffman_coding) for example,
where decision whether to step left or right in the Huffman tree is the most frequent operation.

`Reader` and `Writer` give a _bit-level_ view  of the underlying `io.Reader` and `io.Writer`, but they also
provide a _byte-level_ view (`io.Reader` and `io.Writer`) at the same time. This means you can also use
the `Reader.Read()` and `Writer.Write()` methods to read and write slices of bytes. These will give
you best performance if the underlying `io.Reader` and `io.Writer` are aligned to a byte boundary
(else all the individual bytes are assembled from / spread to multiple bytes). You can ensure
byte boundary alignment by calling the `Align()` method of `Reader` and `Writer`. As an extra,
`io.ByteReader` and `io.ByteWriter` are also implemented.

### Bit order

The more general highest-bits-first order is used. So for example if the input provides the bytes `0x8f` and `0x55`:

    HEXA    8    f     5    5
    BINARY  1100 1111  0101 0101
            aaaa bbbc  ccdd dddd

Then ReadBits will return the following values:

    r := NewReader(bytes.NewBuffer([]byte{0x8f, 0x55}))
    a, err := r.ReadBits(4) //   1100 = 0x08
    b, err := r.ReadBits(3) //    111 = 0x07
    c, err := r.ReadBits(3) //    101 = 0x05
    d, err := r.ReadBits(6) // 010101 = 0x15

Writing the above values would result in the same sequence of bytes:

    b := &bytes.Buffer{}
    w := NewWriter(b)
    err := w.WriteBits(0x08, 4)
    err = w.WriteBits(0x07, 3)
    err = w.WriteBits(0x05, 3)
    err = w.WriteBits(0x15, 6)
    err = w.Close()
    // b will hold the bytes: 0x8f and 0x55

### Error handling

All `ReadXXX()` and `WriteXXX()` methods return an error which you are expected to handle.
For convenience, there are also matching `TryReadXXX()` and `TryWriteXXX()` methods
which do not return an error. Instead they store the (first) error in the
`Reader.TryError` / `Writer.TryError` field which you can inspect later.
These `TryXXX()` methods are a no-op if a `TryError` has been encountered before,
so it's safe to call multiple `TryXXX()` methods and defer the error checking.

For example:

    r := NewReader(bytes.NewBuffer([]byte{0x8f, 0x55}))
    a := r.TryReadBits(4) //   1100 = 0x08
    b := r.TryReadBits(3) //    111 = 0x07
    c := r.TryReadBits(3) //    101 = 0x05
    d := r.TryReadBits(6) // 010101 = 0x15
    if r.TryError != nil {
        // Handle error
    }

This allows you to easily convert the result of individual `ReadBits()`, like this:

    r := NewReader(bytes.NewBuffer([]byte{0x8f, 0x55}))
    a := byte(r.TryReadBits(4))   //   1100 = 0x08
    b := int32(r.TryReadBits(3))  //    111 = 0x07
    c := int64(r.TryReadBits(3))  //    101 = 0x05
    d := uint16(r.TryReadBits(6)) // 010101 = 0x15
    if r.TryError != nil {
        // Handle error
    }

And similarly:

    b := &bytes.Buffer{}
    w := NewWriter(b)
    w.TryWriteBits(0x08, 4)
    w.TryWriteBits(0x07, 3)
    w.TryWriteBits(0x05, 3)
    w.TryWriteBits(0x15, 6)
    if w.TryError != nil {
        // Handle error
    }
    err = w.Close()
    // b will hold the bytes: 0x8f and 0x55

### Number of processed bits

For performance reasons, `Reader` and `Writer` do not keep track of the number of read or written bits.
If you happen to need the total number of processed bits, you may use the `CountReader` and `CountWriter` types
which have identical API to that of `Reader` and `Writer`, but they also maintain the number of processed bits
which you can query using the `BitsCount` field.
