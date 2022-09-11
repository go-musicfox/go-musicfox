// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package crc8 implements the 8-bit cyclic redundancy check, or CRC-8,
// checksum. See http://en.wikipedia.org/wiki/Cyclic_redundancy_check and
// http://www.ross.net/crc/download/crc_v3.txt for information.
package crc8

import "github.com/mewkiz/flac/internal/hashutil"

// Size of a CRC-8 checksum in bytes.
const Size = 1

// Predefined polynomials.
const (
	ATM = 0x07 // x^8 + x^2 + x + 1
)

// Table is a 256-word table representing the polynomial for efficient
// processing.
type Table [256]uint8

// ATMTable is the table for the ATM polynomial.
var ATMTable = makeTable(ATM)

// MakeTable returns the Table constructed from the specified polynomial.
func MakeTable(poly uint8) (table *Table) {
	switch poly {
	case ATM:
		return ATMTable
	}
	return makeTable(poly)
}

// makeTable returns the Table constructed from the specified polynomial.
func makeTable(poly uint8) (table *Table) {
	table = new(Table)
	for i := range table {
		crc := uint8(i)
		for j := 0; j < 8; j++ {
			if crc&0x80 != 0 {
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
	crc   uint8
	table *Table
}

// New creates a new hashutil.Hash8 computing the CRC-8 checksum using the
// polynomial represented by the Table.
func New(table *Table) hashutil.Hash8 {
	return &digest{0, table}
}

// NewATM creates a new hashutil.Hash8 computing the CRC-8 checksum using the
// ATM polynomial.
func NewATM() hashutil.Hash8 {
	return New(ATMTable)
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
func Update(crc uint8, table *Table, p []byte) uint8 {
	for _, v := range p {
		crc = table[crc^v]
	}
	return crc
}

func (d *digest) Write(p []byte) (n int, err error) {
	d.crc = Update(d.crc, d.table, p)
	return len(p), nil
}

// Sum8 returns the 8-bit checksum of the hash.
func (d *digest) Sum8() uint8 {
	return d.crc
}

func (d *digest) Sum(in []byte) []byte {
	return append(in, d.crc)
}

// Checksum returns the CRC-8 checksum of data, using the polynomial represented
// by the Table.
func Checksum(data []byte, table *Table) uint8 {
	return Update(0, table, data)
}

// ChecksumATM returns the CRC-8 checksum of data using the ATM polynomial.
func ChecksumATM(data []byte) uint8 {
	return Update(0, ATMTable, data)
}
