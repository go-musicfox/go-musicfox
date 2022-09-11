package meta

import (
	"encoding/binary"
	"errors"
	"fmt"
)

// SeekTable contains one or more pre-calculated audio frame seek points.
//
// ref: https://www.xiph.org/flac/format.html#metadata_block_seektable
type SeekTable struct {
	// One or more seek points.
	Points []SeekPoint
}

// parseSeekTable reads and parses the body of a SeekTable metadata block.
func (block *Block) parseSeekTable() error {
	// The number of seek points is derived from the header length, divided by
	// the size of a SeekPoint; which is 18 bytes.
	n := block.Length / 18
	if n < 1 {
		return errors.New("meta.Block.parseSeekTable: at least one seek point is required")
	}
	table := &SeekTable{Points: make([]SeekPoint, n)}
	block.Body = table
	var prev uint64
	for i := range table.Points {
		point := &table.Points[i]
		err := binary.Read(block.lr, binary.BigEndian, point)
		if err != nil {
			return unexpected(err)
		}
		// Seek points within a table must be sorted in ascending order by sample
		// number. Each seek point must have a unique sample number, except for
		// placeholder points.
		sampleNum := point.SampleNum
		if i != 0 && sampleNum != PlaceholderPoint {
			switch {
			case sampleNum < prev:
				return fmt.Errorf("meta.Block.parseSeekTable: invalid seek point order; sample number (%d) < prev (%d)", sampleNum, prev)
			case sampleNum == prev:
				return fmt.Errorf("meta.Block.parseSeekTable: duplicate seek point with sample number (%d)", sampleNum)
			}
		}
	}
	return nil
}

// A SeekPoint specifies the byte offset and initial sample number of a given
// target frame.
//
// ref: https://www.xiph.org/flac/format.html#seekpoint
type SeekPoint struct {
	// Sample number of the first sample in the target frame, or
	// 0xFFFFFFFFFFFFFFFF for a placeholder point.
	SampleNum uint64
	// Offset in bytes from the first byte of the first frame header to the first
	// byte of the target frame's header.
	Offset uint64
	// Number of samples in the target frame.
	NSamples uint16
}

// PlaceholderPoint represent the sample number used to specify placeholder seek
// points.
const PlaceholderPoint = 0xFFFFFFFFFFFFFFFF
