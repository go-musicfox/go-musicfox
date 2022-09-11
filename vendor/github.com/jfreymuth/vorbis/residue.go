package vorbis

import "errors"

type residue struct {
	residueType     uint16
	begin, end      uint32
	partitionSize   uint32
	classifications uint8
	classbook       uint8
	cascade         []uint8
	books           [][8]int16
}

func (x *residue) ReadFrom(r *bitReader) error {
	x.residueType = r.Read16(16)
	if x.residueType > 2 {
		return errors.New("vorbis: decoding error")
	}
	x.begin = r.Read32(24)
	x.end = r.Read32(24)
	x.partitionSize = r.Read32(24) + 1
	x.classifications = r.Read8(6) + 1
	x.classbook = r.Read8(8)
	x.cascade = make([]uint8, x.classifications)
	for i := range x.cascade {
		highBits := uint8(0)
		lowBits := r.Read8(3)
		if r.ReadBool() {
			highBits = r.Read8(5)
		}
		x.cascade[i] = highBits*8 + lowBits
	}

	x.books = make([][8]int16, x.classifications)
	for i := range x.books {
		for j := 0; j < 8; j++ {
			if x.cascade[i]&(1<<uint(j)) != 0 {
				x.books[i][j] = int16(r.Read8(8))
			} else {
				x.books[i][j] = -1
			}
		}
	}

	return nil
}

func (x *residue) Decode(r *bitReader, doNotDecode []bool, n uint32, books []codebook, out [][]float32) {
	ch := uint32(len(doNotDecode))
	if x.residueType == 2 {
		decode := false
		for _, not := range doNotDecode {
			if !not {
				decode = true
				break
			}
		}
		if !decode {
			return
		}
		n *= ch
		ch = 1
	}
	begin, end := x.begin, x.end
	if begin > n {
		begin = n
	}
	if end > n {
		end = n
	}
	classbook := books[x.classbook]
	classWordsPerCodeword := classbook.dimensions
	nToRead := end - begin
	partitionsToRead := nToRead / x.partitionSize

	if nToRead == 0 {
		return
	}
	cs := (partitionsToRead + classWordsPerCodeword)
	classifications := make([]uint32, ch*cs)
	for pass := 0; pass < 8; pass++ {
		partitionCount := uint32(0)
		for partitionCount < partitionsToRead {
			if pass == 0 {
				for j := uint32(0); j < ch; j++ {
					if !doNotDecode[j] {
						temp := classbook.DecodeScalar(r)
						for i := classWordsPerCodeword; i > 0; i-- {
							classifications[j*cs+(i-1)+partitionCount] = temp % uint32(x.classifications)
							temp /= uint32(x.classifications)
						}
					}
				}
			}
			for classword := uint32(0); classword < classWordsPerCodeword && partitionCount < partitionsToRead; classword++ {
				for j := uint32(0); j < ch; j++ {
					if !doNotDecode[j] {
						vqclass := classifications[j*cs+partitionCount]
						vqbook := x.books[vqclass][pass]
						if vqbook != -1 {
							book := books[vqbook]
							offset := begin + partitionCount*x.partitionSize
							switch x.residueType {
							case 0:
								step := x.partitionSize / book.dimensions
								for i := uint32(0); i < step; i++ {
									tmp := book.DecodeVector(r)
									for k := range tmp {
										out[j][offset+i+uint32(k)*step] += tmp[k]
									}
								}
							case 1:
								var i uint32
								for i < x.partitionSize {
									tmp := book.DecodeVector(r)
									for k := range tmp {
										out[j][offset+i] += tmp[k]
										i++
									}
								}
							case 2:
								var i uint32
								ch := uint32(len(out))
								for i < x.partitionSize {
									tmp := book.DecodeVector(r)
									for k := range tmp {
										out[(offset+i)%ch][(offset+i)/ch] += tmp[k]
										i++
									}
								}
							}
						}
					}
				}
				partitionCount++
			}
		}
	}
}
