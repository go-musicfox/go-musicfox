// Package hashutil provides utility interfaces for hash functions.
package hashutil

import "hash"

// Hash8 is the common interface implemented by all 8-bit hash functions.
type Hash8 interface {
	hash.Hash
	// Sum8 returns the 8-bit checksum of the hash.
	Sum8() uint8
}

// Hash16 is the common interface implemented by all 16-bit hash functions.
type Hash16 interface {
	hash.Hash
	// Sum16 returns the 16-bit checksum of the hash.
	Sum16() uint16
}
