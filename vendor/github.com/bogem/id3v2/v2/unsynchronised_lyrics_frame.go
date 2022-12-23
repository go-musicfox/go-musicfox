// Copyright 2016 Albert Nigmatzianov. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package id3v2

import "io"

// UnsynchronisedLyricsFrame is used to work with USLT frames.
// The information about how to add unsynchronised lyrics/text frame to tag
// you can see in the docs to tag.AddUnsynchronisedLyricsFrame function.
//
// You must choose a three-letter language code from
// ISO 639-2 code list: https://www.loc.gov/standards/iso639-2/php/code_list.php
type UnsynchronisedLyricsFrame struct {
	Encoding          Encoding
	Language          string
	ContentDescriptor string
	Lyrics            string
}

func (uslf UnsynchronisedLyricsFrame) Size() int {
	return 1 + len(uslf.Language) + encodedSize(uslf.ContentDescriptor, uslf.Encoding) +
		+len(uslf.Encoding.TerminationBytes) + encodedSize(uslf.Lyrics, uslf.Encoding)
}

func (uslf UnsynchronisedLyricsFrame) UniqueIdentifier() string {
	return uslf.Language + uslf.ContentDescriptor
}

func (uslf UnsynchronisedLyricsFrame) WriteTo(w io.Writer) (n int64, err error) {
	if len(uslf.Language) != 3 {
		return n, ErrInvalidLanguageLength
	}

	return useBufWriter(w, func(bw *bufWriter) {
		bw.WriteByte(uslf.Encoding.Key)
		bw.WriteString(uslf.Language)
		bw.EncodeAndWriteText(uslf.ContentDescriptor, uslf.Encoding)
		bw.Write(uslf.Encoding.TerminationBytes)
		bw.EncodeAndWriteText(uslf.Lyrics, uslf.Encoding)
	})
}

func parseUnsynchronisedLyricsFrame(br *bufReader, version byte) (Framer, error) {
	encoding := getEncoding(br.ReadByte())
	language := br.Next(3)
	contentDescriptor := br.ReadText(encoding)

	if br.Err() != nil {
		return nil, br.Err()
	}

	lyrics := getBytesBuffer()
	defer putBytesBuffer(lyrics)

	if _, err := lyrics.ReadFrom(br); err != nil {
		return nil, err
	}

	uslf := UnsynchronisedLyricsFrame{
		Encoding:          encoding,
		Language:          string(language),
		ContentDescriptor: decodeText(contentDescriptor, encoding),
		Lyrics:            decodeText(lyrics.Bytes(), encoding),
	}

	return uslf, nil
}
