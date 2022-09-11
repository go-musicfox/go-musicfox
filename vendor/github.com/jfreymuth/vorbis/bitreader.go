package vorbis

type bitReader struct {
	data      []byte
	position  int
	bitOffset uint
	eof       bool
}

func newBitReader(data []byte) *bitReader {
	return &bitReader{data, 0, 0, false}
}

func (r *bitReader) EOF() bool {
	return r.eof
}

func (r *bitReader) Read1() uint32 {
	if r.position >= len(r.data) {
		r.eof = true
		return 0
	}
	var result uint32
	if r.data[r.position]&(1<<r.bitOffset) != 0 {
		result = 1
	}
	if r.bitOffset < 7 {
		r.bitOffset++
	} else {
		r.bitOffset = 0
		r.position++
	}
	return result
}

func (r *bitReader) read(n uint, bits uint) uint32 {
	if n > bits {
		panic("invalid argument")
	}
	var result uint32
	var written uint
	size := n
	for n > 0 {
		if r.position >= len(r.data) {
			r.eof = true
			return 0
		}
		result |= uint32(r.data[r.position]>>r.bitOffset) << written
		written += 8 - r.bitOffset
		if n < 8-r.bitOffset {
			r.bitOffset += n
			break
		}
		n -= 8 - r.bitOffset
		r.bitOffset = 0
		r.position++
	}
	return result &^ (0xFFFFFFFF << size)
}

func (r *bitReader) Read8(n uint) uint8 {
	return uint8(r.read(n, 8))
}

func (r *bitReader) Read16(n uint) uint16 {
	return uint16(r.read(n, 16))
}

func (r *bitReader) Read32(n uint) uint32 {
	return uint32(r.read(n, 32))
}

func (r *bitReader) ReadBool() bool {
	return r.Read8(1) == 1
}
