package flacpicture

import (
	"bytes"
	"encoding/binary"
	"io"
)

func encodeUint32(n uint32) []byte {
	buf := bytes.NewBuffer([]byte{})
	if err := binary.Write(buf, binary.BigEndian, n); err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func readUint32(r io.Reader) (res uint32, err error) {
	err = binary.Read(r, binary.BigEndian, &res)
	return
}

func readBytesWith32bitSize(r io.Reader) (res []byte, err error) {
	var size uint32
	size, err = readUint32(r)
	if err != nil {
		return
	}
	bufall := bytes.NewBuffer([]byte{})
	for size > 0 {
		var nn int
		buf := make([]byte, size)
		nn, err = r.Read(buf)
		if err != nil {
			return
		}
		bufall.Write(buf)
		size -= uint32(nn)
	}
	res = bufall.Bytes()
	return
}
