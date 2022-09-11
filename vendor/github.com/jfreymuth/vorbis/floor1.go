package vorbis

import "sort"

type floor1 struct {
	partitionClassList []uint8
	classes            []floor1Class
	multiplier         uint8
	rangebits          uint8
	xList              []uint32
	sort               []uint32

	step2  []bool
	finalY []uint32
}

type floor1Class struct {
	dimension     uint8
	subclass      uint8
	masterbook    uint8
	subclassBooks []uint8
}

func (f *floor1) ReadFrom(r *bitReader) error {
	f.partitionClassList = make([]uint8, r.Read8(5))
	var maximumClass uint8
	for i := range f.partitionClassList {
		class := r.Read8(4)
		f.partitionClassList[i] = class
		if class > maximumClass {
			maximumClass = class
		}
	}

	f.classes = make([]floor1Class, maximumClass+1)
	for i := range f.classes {
		class := &f.classes[i]
		class.dimension = r.Read8(3) + 1
		class.subclass = r.Read8(2)
		if class.subclass != 0 {
			class.masterbook = r.Read8(8)
		}
		class.subclassBooks = make([]uint8, 1<<class.subclass)
		for i := range class.subclassBooks {
			class.subclassBooks[i] = r.Read8(8) - 1
		}
	}

	f.multiplier = r.Read8(2) + 1
	f.rangebits = r.Read8(4)
	f.xList = append(f.xList, 0, 1<<f.rangebits)
	for _, class := range f.partitionClassList {
		for i := uint8(0); i < f.classes[class].dimension; i++ {
			f.xList = append(f.xList, r.Read32(uint(f.rangebits)))
		}
	}

	f.sort = make([]uint32, len(f.xList))
	for i := range f.sort {
		f.sort[i] = uint32(i)
	}
	sort.Sort(f)

	f.step2 = make([]bool, len(f.xList))
	f.finalY = make([]uint32, len(f.xList))
	return nil
}

func (f *floor1) Decode(r *bitReader, books []codebook, n uint32) interface{} {
	if !r.ReadBool() {
		return nil
	}

	range_ := [4]uint32{256, 128, 86, 64}[f.multiplier-1]
	y := make([]uint32, 0, len(f.xList))
	y = append(y, r.Read32(ilog(int(range_)-1)), r.Read32(ilog(int(range_)-1)))
	for _, classIndex := range f.partitionClassList {
		class := f.classes[classIndex]
		cdim := class.dimension
		cbits := class.subclass
		csub := (uint32(1) << cbits) - 1
		cval := uint32(0)
		if cbits > 0 {
			cval = books[class.masterbook].DecodeScalar(r)
		}
		for j := 0; j < int(cdim); j++ {
			book := class.subclassBooks[cval&csub]
			cval >>= cbits
			if book != 0xFF {
				y = append(y, books[book].DecodeScalar(r))
			} else {
				y = append(y, 0)
			}
		}
	}
	return y
}

func (f *floor1) Apply(out []float32, data interface{}) {
	y := data.([]uint32)
	n := uint32(len(out))
	range_ := [4]uint32{256, 128, 86, 64}[f.multiplier-1]

	f.step2[0], f.step2[1] = true, true
	f.finalY[0], f.finalY[1] = y[0], y[1]

	for i := 2; i < len(f.xList); i++ {
		low := lowNeighbor(f.xList, i)
		high := highNeighbor(f.xList, i)
		predicted := renderPoint(f.xList[low], f.finalY[low], f.xList[high], f.finalY[high], f.xList[i])
		val := y[i]

		highRoom := range_ - predicted
		lowRoom := predicted
		var room uint32
		if highRoom < lowRoom {
			room = highRoom * 2
		} else {
			room = lowRoom * 2
		}

		if val == 0 {
			f.step2[i] = false
			f.finalY[i] = predicted
		} else {
			f.step2[low] = true
			f.step2[high] = true
			f.step2[i] = true
			if val >= room {
				if highRoom > lowRoom {
					f.finalY[i] = val - lowRoom + predicted
				} else {
					f.finalY[i] = predicted - val + highRoom - 1
				}
			} else {
				if val%2 == 1 {
					f.finalY[i] = predicted - (val+1)/2
				} else {
					f.finalY[i] = predicted + val/2
				}
			}
		}
	}

	var hx, lx uint32
	ly := f.finalY[0] * uint32(f.multiplier)

	var hy uint32
	for j := 1; j < len(f.finalY); j++ {
		i := f.sort[j]
		if f.step2[i] {
			hy = f.finalY[i] * uint32(f.multiplier)
			hx = f.xList[i]
			renderLine(lx, ly, hx, hy, out)
			lx = hx
			ly = hy
		}
	}

	if hx < n {
		for i := hx; i < n; i++ {
			out[i] *= inverseDBTable[hy]
		}
	}
}

func (f *floor1) Len() int {
	return len(f.xList)
}
func (f *floor1) Less(i, j int) bool {
	return f.xList[f.sort[i]] < f.xList[f.sort[j]]
}
func (f *floor1) Swap(i, j int) {
	f.sort[i], f.sort[j] = f.sort[j], f.sort[i]
}

func lowNeighbor(v []uint32, index int) int {
	val := v[index]
	best := 0
	max := uint32(0)
	for i := 1; i < index; i++ {
		if v[i] >= val {
			continue
		}
		if v[i] > max {
			best = i
			max = v[i]
		}
	}
	return best
}
func highNeighbor(v []uint32, index int) int {
	val := v[index]
	best := 0
	min := uint32(0xffffffff)
	for i := 1; i < index; i++ {
		if v[i] <= val {
			continue
		}
		if v[i] < min {
			best = i
			min = v[i]
		}
	}
	return best
}

func renderPoint(x0, y0, x1, y1, x uint32) uint32 {
	dy := int(y1) - int(y0)
	adx := x1 - x0
	ady := y1 - y0
	if dy < 0 {
		ady = uint32(-dy)
	}
	err := ady * (x - x0)
	off := err / adx
	if dy < 0 {
		return y0 - off
	}
	return y0 + off
}

func renderLine(x0, y0, x1, y1 uint32, v []float32) {
	dy := int(y1) - int(y0)
	adx := x1 - x0
	ady := y1 - y0
	if dy < 0 {
		ady = uint32(-dy)
	}
	base := dy / int(adx)
	x := x0
	y := y0
	err := uint32(0)
	var sy int
	if dy < 0 {
		sy = base - 1
	} else {
		sy = base + 1
	}

	absBase := uint32(base)
	if base < 0 {
		absBase = uint32(-absBase)
	}
	ady -= absBase * adx

	v[x] *= inverseDBTable[y]
	for x := x0 + 1; x < x1; x++ {
		err += ady
		if err >= adx {
			err -= adx
			y = uint32(int(y) + sy)
		} else {
			y = uint32(int(y) + base)
		}
		v[x] *= inverseDBTable[y]
	}
}
