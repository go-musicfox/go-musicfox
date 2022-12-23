// Copyright 2016 Albert Nigmatzianov. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package id3v2

import (
	"fmt"
	"io"
)

// PictureFrame structure is used for picture frames (APIC).
// The information about how to add picture frame to tag you can
// see in the docs to tag.AddAttachedPicture function.
//
// Available picture types you can see in constants.
type PictureFrame struct {
	Encoding    Encoding
	MimeType    string
	PictureType byte
	Description string
	Picture     []byte
}

func (pf PictureFrame) UniqueIdentifier() string {
	return fmt.Sprintf("%02X%s", pf.PictureType, pf.Description)
}

func (pf PictureFrame) Size() int {
	return 1 + len(pf.MimeType) + 1 + 1 + encodedSize(pf.Description, pf.Encoding) +
		len(pf.Encoding.TerminationBytes) + len(pf.Picture)
}

func (pf PictureFrame) WriteTo(w io.Writer) (n int64, err error) {
	return useBufWriter(w, func(bw *bufWriter) {
		bw.WriteByte(pf.Encoding.Key)
		bw.WriteString(pf.MimeType)
		bw.WriteByte(0)
		bw.WriteByte(pf.PictureType)
		bw.EncodeAndWriteText(pf.Description, pf.Encoding)
		bw.Write(pf.Encoding.TerminationBytes)
		bw.Write(pf.Picture)
	})
}

func parsePictureFrame(br *bufReader, version byte) (Framer, error) {
	encoding := getEncoding(br.ReadByte())
	mimeType := br.ReadText(EncodingISO)
	pictureType := br.ReadByte()
	description := br.ReadText(encoding)
	picture := br.ReadAll()

	if br.Err() != nil {
		return nil, br.Err()
	}

	pf := PictureFrame{
		Encoding:    encoding,
		MimeType:    string(mimeType),
		PictureType: pictureType,
		Description: decodeText(description, encoding),
		Picture:     picture,
	}

	return pf, nil
}
