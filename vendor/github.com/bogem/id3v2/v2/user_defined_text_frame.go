package id3v2

import "io"

// UserDefinedTextFrame is used to work with TXXX frames.
// There can be many UserDefinedTextFrames but the Description fields need to be unique.
type UserDefinedTextFrame struct {
	Encoding    Encoding
	Description string
	Value       string
}

func (udtf UserDefinedTextFrame) Size() int {
	return 1 + encodedSize(udtf.Description, udtf.Encoding) + len(udtf.Encoding.TerminationBytes) + encodedSize(udtf.Value, udtf.Encoding)
}

func (udtf UserDefinedTextFrame) UniqueIdentifier() string {
	return udtf.Description
}

func (udtf UserDefinedTextFrame) WriteTo(w io.Writer) (n int64, err error) {
	return useBufWriter(w, func(bw *bufWriter) {
		bw.WriteByte(udtf.Encoding.Key)
		bw.EncodeAndWriteText(udtf.Description, udtf.Encoding)
		bw.Write(udtf.Encoding.TerminationBytes)
		bw.EncodeAndWriteText(udtf.Value, udtf.Encoding)
	})
}

func parseUserDefinedTextFrame(br *bufReader, version byte) (Framer, error) {
	encoding := getEncoding(br.ReadByte())
	description := br.ReadText(encoding)

	if br.Err() != nil {
		return nil, br.Err()
	}

	value := getBytesBuffer()
	defer putBytesBuffer(value)

	if _, err := value.ReadFrom(br); err != nil {
		return nil, err
	}

	udtf := UserDefinedTextFrame{
		Encoding:    encoding,
		Description: decodeText(description, encoding),
		Value:       decodeText(value.Bytes(), encoding),
	}

	return udtf, nil
}
