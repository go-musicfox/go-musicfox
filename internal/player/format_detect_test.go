package player

import (
	"bytes"
	"io"
	"testing"
)

func TestSniffFormatMagicNumbers(t *testing.T) {
	for _, test := range []struct {
		name string
		data []byte
		want SongType
	}{
		{name: "flac", data: []byte("fLaC\x00\x00\x00\x22...STREAMINFO"), want: Flac},
		{name: "ogg", data: []byte("OggS\x00\x02junk"), want: Ogg},
		{name: "wav", data: append([]byte("RIFFxxxxWAVE"), bytes.Repeat([]byte{0}, 64)...), want: Wav},
		{name: "mp3 with id3", data: append([]byte("ID3\x04\x00\x00\x00\x00\x00\x00"), bytes.Repeat([]byte{0}, 64)...), want: Mp3},
		{name: "mp3 frame sync", data: []byte{0xFF, 0xFB, 0x90, 0x00, 0x01, 0x02}, want: Mp3},
	} {
		t.Run(test.name, func(t *testing.T) {
			test.data = append(test.data, bytes.Repeat([]byte{0xAA}, 64)...)
			got, ok := sniffFormat(bytes.NewReader(test.data))
			if !ok || got != test.want {
				t.Fatalf("type=%v ok=%v, want %v true", got, ok, test.want)
			}
		})
	}
}

func TestSniffFormatUnknownKeepsDeclaration(t *testing.T) {
	r := bytes.NewReader([]byte("\x00\x00\x00\x18ftypM4A \x00\x00\x00\x00"))
	if tp, ok := sniffFormat(r); ok {
		t.Fatalf("sniffed type=%v, want unknown", tp)
	}
	if pos, _ := r.Seek(0, io.SeekCurrent); pos != 0 {
		t.Fatalf("reader position=%d, want reset to 0", pos)
	}
}

func TestSniffFormatEmpty(t *testing.T) {
	if tp, ok := sniffFormat(bytes.NewReader(nil)); ok {
		t.Fatalf("sniffed type=%v on empty reader, want unknown", tp)
	}
}

func TestSniffFormatResetsPosition(t *testing.T) {
	r := bytes.NewReader(append([]byte("fLaC"), bytes.Repeat([]byte{0xAA}, 64)...))
	if _, err := r.Seek(10, io.SeekStart); err != nil {
		t.Fatal(err)
	}
	tp, ok := sniffFormat(r)
	if !ok || tp != Flac {
		t.Fatalf("type=%v ok=%v, want Flac true", tp, ok)
	}
	if pos, _ := r.Seek(0, io.SeekCurrent); pos != 0 {
		t.Fatalf("reader position=%d, want reset to 0", pos)
	}
}
