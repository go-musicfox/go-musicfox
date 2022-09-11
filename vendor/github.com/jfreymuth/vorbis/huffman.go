package vorbis

type huffmanCode []uint32

func (h huffmanCode) Lookup(r *bitReader) uint32 {
	i := uint32(0)
	for i&1 == 0 {
		i = h[i+r.Read1()]
	}
	return i >> 1
}

type huffmanBuilder struct {
	code      huffmanCode
	minLength []uint8
}

func newHuffmanBuilder(size uint32) *huffmanBuilder {
	return &huffmanBuilder{
		code:      make(huffmanCode, size),
		minLength: make([]uint8, size/2),
	}
}

func (t *huffmanBuilder) Put(entry uint32, length uint8) {
	t.put(0, entry, length-1)
}

func (t *huffmanBuilder) put(index, entry uint32, length uint8) bool {
	if length < t.minLength[index/2] {
		return false
	}
	if length == 0 {
		if t.code[index] == 0 {
			t.code[index] = entry*2 + 1
			return true
		}
		if t.code[index+1] == 0 {
			t.code[index+1] = entry*2 + 1
			t.minLength[index/2] = 1
			return true
		}
		t.minLength[index/2] = 1
		return false
	}
	if t.code[index]&1 == 0 {
		if t.code[index] == 0 {
			t.code[index] = t.findEmpty(index + 2)
		}
		if t.put(t.code[index], entry, length-1) {
			return true
		}
	}
	if t.code[index+1]&1 == 0 {
		if t.code[index+1] == 0 {
			t.code[index+1] = t.findEmpty(index + 2)
		}
		if t.put(t.code[index+1], entry, length-1) {
			return true
		}
	}
	t.minLength[index/2] = length + 1
	return false
}

func (t *huffmanBuilder) findEmpty(index uint32) uint32 {
	for t.code[index] != 0 {
		index += 2
	}
	return index
}
