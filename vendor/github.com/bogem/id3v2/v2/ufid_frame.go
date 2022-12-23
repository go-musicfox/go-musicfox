package id3v2

import "io"

// UFIDFrame is used for "Unique file identifier"
type UFIDFrame struct {
	OwnerIdentifier string
	Identifier      []byte
}

func (ufid UFIDFrame) UniqueIdentifier() string {
	return ufid.OwnerIdentifier
}

func (ufid UFIDFrame) Size() int {
	return encodedSize(ufid.OwnerIdentifier, EncodingISO) + len(EncodingISO.TerminationBytes) + len(ufid.Identifier)
}

func (ufid UFIDFrame) WriteTo(w io.Writer) (n int64, err error) {
	return useBufWriter(w, func(bw *bufWriter) {
		bw.WriteString(ufid.OwnerIdentifier)
		bw.Write(EncodingISO.TerminationBytes)
		bw.Write(ufid.Identifier)
	})
}

func parseUFIDFrame(br *bufReader, version byte) (Framer, error) {
	owner := br.ReadText(EncodingISO)
	ident := br.ReadAll()

	if br.Err() != nil {
		return nil, br.Err()
	}

	ufid := UFIDFrame{
		OwnerIdentifier: decodeText(owner, EncodingISO),
		Identifier:      ident,
	}

	return ufid, nil
}
