// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package crc16 implements the 16-bit cyclic redundancy check, or CRC-16,
// checksum. See http://en.wikipedia.org/wiki/Cyclic_redundancy_check and
// http://www.ross.net/crc/download/crc_v3.txt for information.
package crc16

import "github.com/mewkiz/flac/internal/hashutil"

// Size of a CRC-16 checksum in bytes.
const Size = 2

// Predefined polynomials.
const (
	IBM = 0x8005 // x^16 + x^15 + x^2 + x^0
)

// Table is a 256-word table representing the polynomial for efficient
// processing.
type Table [256]uint16

// IBMTable is the table for the IBM polynomial.
var IBMTable = makeTable(IBM)

// MakeTable returns the Table constructed from the specified polynomial.
func MakeTable(poly uint16) (table *Table) {
	switch poly {
	case IBM:
		return IBMTable
	}
	return makeTable(poly)
}

// makeTable returns the Table constructed from the specified polynomial.
func makeTable(poly uint16) (table *Table) {
	table = new(Table)
	for i := range table {
		crc := uint16(i << 8)
		for j := 0; j < 8; j++ {
			if crc&0x8000 != 0 {
				crc = crc<<1 ^ poly
			} else {
				crc <<= 1
			}
		}
		table[i] = crc
	}
	return table
}

// digest represents the partial evaluation of a checksum.
type digest struct {
	crc   uint16
	table *Table
}

// New creates a new hashutil.Hash16 computing the CRC-16 checksum using the
// polynomial represented by the Table.
func New(table *Table) hashutil.Hash16 {
	return &digest{0, table}
}

// NewIBM creates a new hashutil.Hash16 computing the CRC-16 checksum using the
// IBM polynomial.
func NewIBM() hashutil.Hash16 {
	return New(IBMTable)
}

func (d *digest) Size() int {
	return Size
}

func (d *digest) BlockSize() int {
	return 1
}

func (d *digest) Reset() {
	d.crc = 0
}

// Update returns the result of adding the bytes in p to the crc.
func Update(crc uint16, table *Table, p []byte) uint16 {
	for _, v := range p {
		crc = crc<<8 ^ table[crc>>8^uint16(v)]
	}
	return crc
}

func (d *digest) Write(p []byte) (n int, err error) {
	d.crc = Update(d.crc, d.table, p)
	return len(p), nil
}

// Sum16 returns the 16-bit checksum of the hash.
func (d *digest) Sum16() uint16 {
	return d.crc
}

func (d *digest) Sum(in []byte) []byte {
	s := d.Sum16()
	return append(in, byte(s>>8), byte(s))
}

// Checksum returns the CRC-16 checksum of data, using the polynomial
// represented by the Table.
func Checksum(data []byte, table *Table) uint16 {
	return Update(0, table, data)
}

// ChecksumIBM returns the CRC-16 checksum of data using the IBM polynomial.
func ChecksumIBM(data []byte) uint16 {
	return Update(0, IBMTable, data)
}
