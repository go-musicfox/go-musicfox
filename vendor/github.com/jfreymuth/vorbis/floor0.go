package vorbis

import (
	"math"
)

type floor0 struct {
	order           uint8
	rate            uint16
	barkMapSize     uint16
	amplitudeBits   uint8
	amplitudeOffset uint8
	bookList        []uint8
}

type floor0Data struct {
	amplitude    uint32
	coefficients []float32
}

func (f *floor0) ReadFrom(r *bitReader) error {
	f.order = r.Read8(8)
	f.rate = r.Read16(16)
	f.barkMapSize = r.Read16(16)
	f.amplitudeBits = r.Read8(6)
	f.amplitudeOffset = r.Read8(8)
	f.bookList = make([]uint8, r.Read8(4)+1)
	for i := range f.bookList {
		f.bookList[i] = r.Read8(8)
	}
	return nil
}

func (f *floor0) Decode(r *bitReader, books []codebook, n uint32) interface{} {
	amplitude := r.Read32(uint(f.amplitudeBits))
	if amplitude == 0 {
		return nil
	}
	bookNumber := r.Read8(ilog(len(f.bookList)))
	book := books[f.bookList[bookNumber]]
	coefficients := make([]float32, f.order)
	i := 0
	last := float32(0)
	for {
		tempVector := book.DecodeVector(r)
		for _, c := range tempVector {
			coefficients[i] = c + last
			i++
			if i >= len(coefficients) {
				return floor0Data{amplitude, coefficients}
			}
		}
		last = tempVector[len(tempVector)-1]
	}
}

func (f *floor0) Apply(out []float32, data interface{}) {
	d := data.(floor0Data)
	n := uint32(len(out))
	i := uint32(0)
	for i < n {
		mapi := f.mapResult(i, n)
		w := math.Pi * float64(mapi) / float64(f.barkMapSize)
		cosw := math.Cos(w)
		var p, q float64
		if f.order%2 == 1 {
			p = 1 - cosw*cosw
			for j := 0; j <= int(f.order-3)/2; j++ {
				tmp := math.Cos(float64(d.coefficients[2*j+1])) - cosw
				p *= 4 * tmp * tmp
			}
			q = 1 / 4
			for j := 0; j <= int(f.order-1)/2; j++ {
				tmp := math.Cos(float64(d.coefficients[2*j])) - cosw
				q *= 4 * tmp * tmp
			}
		} else {
			p = (1 - cosw*cosw) / 2
			for j := 0; j <= int(f.order-2)/2; j++ {
				tmp := math.Cos(float64(d.coefficients[2*j+1])) - cosw
				p *= 4 * tmp * tmp
			}
			q = (1 + cosw*cosw) / 2
			for j := 0; j <= int(f.order-2)/2; j++ {
				tmp := math.Cos(float64(d.coefficients[2*j])) - cosw
				q *= 4 * tmp * tmp
			}
		}
		linearFloorValue := math.Exp(.11512925 * (float64(d.amplitude)*float64(f.amplitudeOffset)/(float64(uint64(1)<<f.amplitudeBits-1)*math.Sqrt(p+q)) - float64(f.amplitudeOffset)))
		for f.mapResult(i, n) == mapi {
			out[i] *= float32(linearFloorValue)
			i++
		}
	}
}

func (f *floor0) mapResult(i, n uint32) int {
	if i >= n {
		return -1
	}
	b := int(math.Floor(bark(float64(f.rate)*float64(i)/2*float64(n)) * float64(f.barkMapSize) / bark(.5*float64(f.rate))))
	if b > int(f.barkMapSize)-1 {
		return int(f.barkMapSize) - 1
	}
	return b
}

func bark(x float64) float64 {
	return 13.1*math.Atan(.00074*x) + 2.24*math.Atan(.0000000185*x*x) + .0001*x
}
