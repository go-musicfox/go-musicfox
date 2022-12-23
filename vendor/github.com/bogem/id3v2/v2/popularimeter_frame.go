package id3v2

import (
	"io"
	"math/big"
)

// PopularimeterFrame structure is used for Popularimeter (POPM).
// https://id3.org/id3v2.3.0#Popularimeter
type PopularimeterFrame struct {
	// Email is the identifier for a POPM frame.
	Email string

	// The rating is 1-255 where 1 is worst and 255 is best. 0 is unknown.
	Rating uint8

	// Counter is the number of times this file has been played by this email.
	Counter *big.Int
}

func (pf PopularimeterFrame) UniqueIdentifier() string {
	return pf.Email
}

func (pf PopularimeterFrame) Size() int {
	ratingSize := 1
	return len(pf.Email) + 1 + ratingSize + len(pf.counterBytes())
}

// counterBytes returns a byte slice that represents the counter.
func (pf PopularimeterFrame) counterBytes() []byte {
	bytes := pf.Counter.Bytes()

	// Specification requires at least 4 bytes for counter, pad if necessary.
	bytesNeeded := 4 - len(bytes)
	if bytesNeeded > 0 {
		padding := make([]byte, bytesNeeded)
		bytes = append(padding, bytes...)
	}

	return bytes
}

func (pf PopularimeterFrame) WriteTo(w io.Writer) (n int64, err error) {
	return useBufWriter(w, func(bw *bufWriter) {
		bw.WriteString(pf.Email)
		bw.WriteByte(0)
		bw.WriteByte(pf.Rating)
		bw.Write(pf.counterBytes())
	})
}

func parsePopularimeterFrame(br *bufReader, version byte) (Framer, error) {
	email := br.ReadText(EncodingISO)
	rating := br.ReadByte()

	counter := big.NewInt(0)
	remainingBytes := br.ReadAll()
	counter = counter.SetBytes(remainingBytes)

	pf := PopularimeterFrame{
		Email:   string(email),
		Rating:  rating,
		Counter: counter,
	}

	return pf, nil
}
