package vorbis

import "errors"

type floor interface {
	Decode(*bitReader, []codebook, uint32) interface{}
	Apply(out []float32, data interface{})
}

type mapping struct {
	couplingSteps uint16
	angle         []uint8
	magnitude     []uint8
	mux           []uint8
	submaps       []mappingSubmap
}

type mappingSubmap struct {
	floor, residue uint8
}

type mode struct {
	blockflag uint8
	mapping   uint8
}

func (d *Decoder) readSetupHeader(header []byte) error {
	r := newBitReader(header)

	// CODEBOOKS
	d.codebooks = make([]codebook, r.Read16(8)+1)
	for i := range d.codebooks {
		err := d.codebooks[i].ReadFrom(r)
		if err != nil {
			return err
		}
	}

	// TIME DOMAIN TRANSFORMS
	transformCount := r.Read8(6) + 1
	for i := 0; i < int(transformCount); i++ {
		if r.Read16(16) != 0 {
			return errors.New("vorbis: decoding error")
		}
	}

	// FLOORS
	d.floors = make([]floor, r.Read8(6)+1)
	for i := range d.floors {
		var err error
		switch r.Read16(16) {
		case 0:
			f := new(floor0)
			err = f.ReadFrom(r)
			d.floors[i] = f
		case 1:
			f := new(floor1)
			err = f.ReadFrom(r)
			d.floors[i] = f
		default:
			return errors.New("vorbis: decoding error")
		}
		if err != nil {
			return err
		}
	}

	// RESIDUES
	d.residues = make([]residue, r.Read8(6)+1)
	for i := range d.residues {
		err := d.residues[i].ReadFrom(r)
		if err != nil {
			return err
		}
	}

	// MAPPINGS
	d.mappings = make([]mapping, r.Read8(6)+1)
	for i := range d.mappings {
		m := &d.mappings[i]
		if r.Read16(16) != 0 {
			return errors.New("vorbis: decoding error")
		}
		if r.ReadBool() {
			m.submaps = make([]mappingSubmap, r.Read8(4)+1)
		} else {
			m.submaps = make([]mappingSubmap, 1)
		}
		if r.ReadBool() {
			m.couplingSteps = r.Read16(8) + 1
			m.magnitude = make([]uint8, m.couplingSteps)
			m.angle = make([]uint8, m.couplingSteps)
			for i := range m.magnitude {
				m.magnitude[i] = r.Read8(ilog(d.channels - 1))
				m.angle[i] = r.Read8(ilog(d.channels - 1))
			}
		}
		if r.Read8(2) != 0 {
			return errors.New("vorbis: decoding error")
		}
		m.mux = make([]uint8, d.channels)
		if len(m.submaps) > 1 {
			for i := range m.mux {
				m.mux[i] = r.Read8(4)
			}
		}
		for i := range m.submaps {
			r.Read8(8)
			m.submaps[i].floor = r.Read8(8)
			m.submaps[i].residue = r.Read8(8)
		}
	}

	// MODES
	d.modes = make([]mode, r.Read8(6)+1)
	for i := range d.modes {
		m := &d.modes[i]
		m.blockflag = r.Read8(1)
		if r.Read16(16) != 0 {
			return errors.New("vorbis: decoding error")
		}
		if r.Read16(16) != 0 {
			return errors.New("vorbis: decoding error")
		}
		m.mapping = r.Read8(8)
	}

	if !r.ReadBool() {
		return errors.New("vorbis: decoding error")
	}
	d.initLookup()
	return nil
}

func (d *Decoder) initLookup() {
	d.windows[0] = makeWindow(d.blocksize[0])
	d.windows[1] = makeWindow(d.blocksize[1])
	generateIMDCTLookup(d.blocksize[0], &d.lookup[0])
	generateIMDCTLookup(d.blocksize[1], &d.lookup[1])
	d.residueBuffer = make([][]float32, d.channels)
	for i := range d.residueBuffer {
		d.residueBuffer[i] = make([]float32, d.blocksize[1]/2)
	}
	d.rawBuffer = make([][]float32, d.channels)
	for i := range d.rawBuffer {
		d.rawBuffer[i] = make([]float32, d.blocksize[1])
	}
}
