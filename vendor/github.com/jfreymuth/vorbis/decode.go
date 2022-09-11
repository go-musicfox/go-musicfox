package vorbis

import "errors"

type floorData struct {
	floor     floor
	data      interface{}
	noResidue bool
}

func (d *Decoder) decodePacket(r *bitReader, out []float32) ([]float32, error) {
	if r.ReadBool() {
		return nil, errors.New("vorbis: decoding error")
	}
	modeNumber := r.Read8(ilog(len(d.modes) - 1))
	mode := d.modes[modeNumber]
	// decode window type
	blocktype := mode.blockflag
	longWindow := mode.blockflag == 1
	blocksize := d.blocksize[blocktype]
	spectrumSize := uint32(blocksize / 2)
	windowPrev, windowNext := false, false
	window := windowType{blocksize, blocksize, blocksize}
	if longWindow {
		windowPrev = r.ReadBool()
		windowNext = r.ReadBool()
		if !windowPrev {
			window.prev = d.blocksize[0]
		}
		if !windowNext {
			window.next = d.blocksize[0]
		}
	}

	mapping := &d.mappings[mode.mapping]
	if d.floorBuffer == nil {
		d.floorBuffer = make([]floorData, d.channels)
	}
	for ch := range d.residueBuffer {
		d.residueBuffer[ch] = d.residueBuffer[ch][:spectrumSize]
		for i := range d.residueBuffer[ch] {
			d.residueBuffer[ch][i] = 0
		}
	}

	d.decodeFloors(r, d.floorBuffer, mapping, spectrumSize)
	d.decodeResidue(r, d.residueBuffer, mapping, d.floorBuffer, spectrumSize)
	d.inverseCoupling(mapping, d.residueBuffer)
	d.applyFloor(d.floorBuffer, d.residueBuffer)

	// inverse MDCT
	for ch := range d.rawBuffer {
		d.rawBuffer[ch] = d.rawBuffer[ch][:blocksize]
		imdct(&d.lookup[blocktype], d.residueBuffer[ch], d.rawBuffer[ch])
	}

	// apply window and overlap
	d.applyWindow(&window, d.rawBuffer)
	center := blocksize / 2
	offset := d.blocksize[1]/4 - d.blocksize[0]/4
	n := 0
	if d.hasOverlap {
		n = blocksize / 2
		if longWindow && !windowPrev {
			n -= offset
		}
		if !longWindow && !d.overlapShort {
			n += offset
		}
		if out == nil {
			out = make([]float32, n*d.channels)
		}
	}
	if longWindow {
		start := 0
		if !windowPrev {
			start = offset
		}
		if d.hasOverlap {
			for ch := range d.rawBuffer {
				for i := 0; i < center-start; i++ {
					out[i*d.channels+ch] = d.rawBuffer[ch][start+i] + d.overlap[(start+i)*d.channels+ch]
				}
			}
		}
		d.overlapShort = false
	} else /*short window*/ {
		if d.hasOverlap {
			if d.overlapShort {
				for ch := range d.rawBuffer {
					for i := 0; i < center; i++ {
						out[i*d.channels+ch] = d.rawBuffer[ch][i] + d.overlap[(offset+i)*d.channels+ch]
					}
				}
			} else {
				for i := 0; i < offset*d.channels; i++ {
					out[i] = d.overlap[i]
				}
				for ch := range d.rawBuffer {
					for i := offset; i < offset+center; i++ {
						out[i*d.channels+ch] = d.rawBuffer[ch][i-offset] + d.overlap[i*d.channels+ch]
					}
				}
			}
		}
		d.overlapShort = true
	}

	if !d.hasOverlap {
		n = 0
	}
	overlapCenter := d.blocksize[1] / 4
	oStart := overlapCenter - center/2
	oEnd := overlapCenter + center/2
	for i := 0; i < oStart*d.channels; i++ {
		d.overlap[i] = 0
	}
	for ch := range d.rawBuffer {
		for i := oStart; i < oEnd; i++ {
			d.overlap[i*d.channels+ch] = d.rawBuffer[ch][center+i-oStart]
		}
	}
	for i := oEnd * d.channels; i < len(d.overlap); i++ {
		d.overlap[i] = 0
	}
	d.hasOverlap = true

	return out[:n*d.channels], nil
}

func (d *Decoder) decodeFloors(r *bitReader, floors []floorData, mapping *mapping, n uint32) {
	for ch := range floors {
		floor := d.floors[mapping.submaps[mapping.mux[ch]].floor]
		data := floor.Decode(r, d.codebooks, n)
		floors[ch] = floorData{floor, data, data == nil}
	}

	for i := 0; i < int(mapping.couplingSteps); i++ {
		if !floors[mapping.magnitude[i]].noResidue || !floors[mapping.angle[i]].noResidue {
			floors[mapping.magnitude[i]].noResidue = false
			floors[mapping.angle[i]].noResidue = false
		}
	}
}

func (d *Decoder) decodeResidue(r *bitReader, out [][]float32, mapping *mapping, floors []floorData, n uint32) {
	for i := range mapping.submaps {
		doNotDecode := make([]bool, 0, len(out))
		tmp := make([][]float32, 0, len(out))
		for j := 0; j < d.channels; j++ {
			if mapping.mux[j] == uint8(i) {
				doNotDecode = append(doNotDecode, floors[j].noResidue)
				tmp = append(tmp, out[j])
			}
		}
		d.residues[mapping.submaps[i].residue].Decode(r, doNotDecode, n, d.codebooks, tmp)
	}
}

func (d *Decoder) inverseCoupling(mapping *mapping, residueVectors [][]float32) {
	for i := mapping.couplingSteps; i > 0; i-- {
		magnitudeVector := residueVectors[mapping.magnitude[i-1]]
		angleVector := residueVectors[mapping.angle[i-1]]
		for j := range magnitudeVector {
			m := magnitudeVector[j]
			a := angleVector[j]
			if m > 0 {
				if a > 0 {
					m, a = m, m-a
				} else {
					a, m = m, m+a
				}
			} else {
				if a > 0 {
					m, a = m, m+a
				} else {
					a, m = m, m-a
				}
			}
			magnitudeVector[j] = m
			angleVector[j] = a
		}
	}
}

func (d *Decoder) applyFloor(floors []floorData, residueVectors [][]float32) {
	for ch := range residueVectors {
		if floors[ch].data != nil {
			floors[ch].floor.Apply(residueVectors[ch], floors[ch].data)
		} else {
			for i := range residueVectors[ch] {
				residueVectors[ch][i] = 0
			}
		}
	}
}
