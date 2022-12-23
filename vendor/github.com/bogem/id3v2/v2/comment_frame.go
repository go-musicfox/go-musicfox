// Copyright 2016 Albert Nigmatzianov. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package id3v2

import "io"

// CommentFrame is used to work with COMM frames.
// The information about how to add comment frame to tag you can
// see in the docs to tag.AddCommentFrame function.
//
// You must choose a three-letter language code from
// ISO 639-2 code list: https://www.loc.gov/standards/iso639-2/php/code_list.php
type CommentFrame struct {
	Encoding    Encoding
	Language    string
	Description string
	Text        string
}

func (cf CommentFrame) Size() int {
	return 1 + len(cf.Language) + encodedSize(cf.Description, cf.Encoding) +
		+len(cf.Encoding.TerminationBytes) + encodedSize(cf.Text, cf.Encoding)
}

func (cf CommentFrame) UniqueIdentifier() string {
	return cf.Language + cf.Description
}

func (cf CommentFrame) WriteTo(w io.Writer) (n int64, err error) {
	if len(cf.Language) != 3 {
		return n, ErrInvalidLanguageLength
	}

	return useBufWriter(w, func(bw *bufWriter) {
		bw.WriteByte(cf.Encoding.Key)
		bw.WriteString(cf.Language)
		bw.EncodeAndWriteText(cf.Description, cf.Encoding)
		bw.Write(cf.Encoding.TerminationBytes)
		bw.EncodeAndWriteText(cf.Text, cf.Encoding)
	})
}

func parseCommentFrame(br *bufReader, version byte) (Framer, error) {
	encoding := getEncoding(br.ReadByte())
	language := br.Next(3)
	description := br.ReadText(encoding)

	if br.Err() != nil {
		return nil, br.Err()
	}

	text := getBytesBuffer()
	defer putBytesBuffer(text)
	if _, err := text.ReadFrom(br); err != nil {
		return nil, err
	}

	cf := CommentFrame{
		Encoding:    encoding,
		Language:    string(language),
		Description: decodeText(description, encoding),
		Text:        decodeText(text.Bytes(), encoding),
	}

	return cf, nil
}
