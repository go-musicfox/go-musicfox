// Copyright 2016 Albert Nigmatzianov. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package id3v2

import "io"

// TextFrame is used to work with all text frames
// (all T*** frames like TIT2 (title), TALB (album) and so on).
type TextFrame struct {
	Encoding Encoding
	Text     string
}

func (tf TextFrame) Size() int {
	return 1 + encodedSize(tf.Text, tf.Encoding) + len(tf.Encoding.TerminationBytes)
}

func (tf TextFrame) UniqueIdentifier() string {
	return "ID"
}

func (tf TextFrame) WriteTo(w io.Writer) (int64, error) {
	return useBufWriter(w, func(bw *bufWriter) {
		bw.WriteByte(tf.Encoding.Key)
		bw.EncodeAndWriteText(tf.Text, tf.Encoding)

		// https://github.com/bogem/id3v2/pull/52
		// https://github.com/bogem/id3v2/pull/33
		bw.Write(tf.Encoding.TerminationBytes)
	})
}

func parseTextFrame(br *bufReader) (Framer, error) {
	encoding := getEncoding(br.ReadByte())

	if br.Err() != nil {
		return nil, br.Err()
	}

	buf := getBytesBuffer()
	defer putBytesBuffer(buf)
	if _, err := buf.ReadFrom(br); err != nil {
		return nil, err
	}

	tf := TextFrame{
		Encoding: encoding,
		Text:     decodeText(buf.Bytes(), encoding),
	}

	return tf, nil
}
