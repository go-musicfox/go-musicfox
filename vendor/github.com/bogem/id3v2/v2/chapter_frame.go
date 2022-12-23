package id3v2

import (
	"encoding/binary"
	"io"
	"time"
)

const (
	nanosInMillis = 1000000
	IgnoredOffset = 0xFFFFFFFF
)

// ChapterFrame is used to work with CHAP frames
// according to spec from http://id3.org/id3v2-chapters-1.0
// This implementation only supports single TIT2 subframe (Title field).
// All other subframes are ignored.
// If StartOffset or EndOffset == id3v2.IgnoredOffset, then it should be ignored
// and StartTime or EndTime should be utilized
type ChapterFrame struct {
	ElementID   string
	StartTime   time.Duration
	EndTime     time.Duration
	StartOffset uint32
	EndOffset   uint32
	Title       *TextFrame
	Description *TextFrame
}

func (cf ChapterFrame) Size() int {
	size := encodedSize(cf.ElementID, EncodingISO) +
		1 + // trailing zero after ElementID
		4 + 4 + 4 + 4 // (Start, End) (Time, Offset)
	if cf.Title != nil {
		size = size +
			frameHeaderSize + // Title frame header size
			cf.Title.Size()
	}
	if cf.Description != nil {
		size = size +
			frameHeaderSize + // Description frame header size
			cf.Description.Size()
	}
	return size
}

func (cf ChapterFrame) UniqueIdentifier() string {
	return cf.ElementID
}

func (cf ChapterFrame) WriteTo(w io.Writer) (n int64, err error) {
	return useBufWriter(w, func(bw *bufWriter) {
		bw.EncodeAndWriteText(cf.ElementID, EncodingISO)
		bw.WriteByte(0)
		binary.Write(bw, binary.BigEndian, int32(cf.StartTime/nanosInMillis))
		binary.Write(bw, binary.BigEndian, int32(cf.EndTime/nanosInMillis))

		binary.Write(bw, binary.BigEndian, cf.StartOffset)
		binary.Write(bw, binary.BigEndian, cf.EndOffset)

		if cf.Title != nil {
			writeFrame(bw, "TIT2", *cf.Title, true)
		}

		if cf.Description != nil {
			writeFrame(bw, "TIT3", *cf.Description, true)
		}
	})
}

func parseChapterFrame(br *bufReader, version byte) (Framer, error) {
	elementID := br.ReadText(EncodingISO)
	synchSafe := version == 4
	var startTime uint32
	var startOffset uint32
	var endTime uint32
	var endOffset uint32

	if err := binary.Read(br, binary.BigEndian, &startTime); err != nil {
		return nil, err
	}
	if err := binary.Read(br, binary.BigEndian, &endTime); err != nil {
		return nil, err
	}
	if err := binary.Read(br, binary.BigEndian, &startOffset); err != nil {
		return nil, err
	}
	if err := binary.Read(br, binary.BigEndian, &endOffset); err != nil {
		return nil, err
	}

	var title TextFrame
	var description TextFrame

	// borrowed from parse.go
	buf := getByteSlice(32 * 1024)
	defer putByteSlice(buf)

	for {
		header, err := parseFrameHeader(buf, br, synchSafe)
		if err == io.EOF || err == errBlankFrame || err == ErrInvalidSizeFormat {
			break
		}
		if err != nil {
			return nil, err
		}
		id, bodySize := header.ID, header.BodySize
		if id == "TIT2" || id == "TIT3" {
			bodyRd := getLimitedReader(br, bodySize)
			br := newBufReader(bodyRd)
			frame, err := parseTextFrame(br)
			if err != nil {
				putLimitedReader(bodyRd)
				return nil, err
			}
			if id == "TIT2" {
				title = frame.(TextFrame)
			} else if id == "TIT3" {
				description = frame.(TextFrame)
			}

			putLimitedReader(bodyRd)
		}
	}

	cf := ChapterFrame{
		ElementID: string(elementID),
		// StartTime is given in milliseconds, so we should convert it to nanoseconds
		// for time.Duration
		StartTime:   time.Duration(int64(startTime) * nanosInMillis),
		EndTime:     time.Duration(int64(endTime) * nanosInMillis),
		StartOffset: startOffset,
		EndOffset:   endOffset,
		Title:       &title,
		Description: &description,
	}
	return cf, nil
}
