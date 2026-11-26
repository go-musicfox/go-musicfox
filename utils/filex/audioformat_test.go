package filex

import (
	"bytes"
	"io"
	"testing"
)

func TestSniffAudioFormatMagicNumbers(t *testing.T) {
	for _, test := range []struct {
		name string
		data []byte
		want string
	}{
		{name: "flac", data: []byte("fLaC\x00\x00\x00\x22...STREAMINFO"), want: "flac"},
		{name: "ogg", data: []byte("OggS\x00\x02junk"), want: "ogg"},
		{name: "wav", data: append([]byte("RIFFxxxxWAVE"), bytes.Repeat([]byte{0}, 64)...), want: "wav"},
		{name: "mp3 with id3", data: append([]byte("ID3\x04\x00\x00\x00\x00\x00\x00"), bytes.Repeat([]byte{0}, 64)...), want: "mp3"},
		{name: "mp3 frame sync", data: []byte{0xFF, 0xFB, 0x90, 0x00, 0x01, 0x02}, want: "mp3"},
	} {
		t.Run(test.name, func(t *testing.T) {
			test.data = append(test.data, bytes.Repeat([]byte{0xAA}, 64)...)
			got, ok := SniffAudioFormat(bytes.NewReader(test.data))
			if !ok || got != test.want {
				t.Fatalf("ext=%q ok=%v, want %q true", got, ok, test.want)
			}
		})
	}
}

func TestSniffAudioFormatUnknownKeepsDeclaration(t *testing.T) {
	r := bytes.NewReader([]byte("\x00\x00\x00\x18ftypM4A \x00\x00\x00\x00"))
	if ext, ok := SniffAudioFormat(r); ok {
		t.Fatalf("sniffed ext=%q, want unknown", ext)
	}
	if pos, _ := r.Seek(0, io.SeekCurrent); pos != 0 {
		t.Fatalf("reader position=%d, want reset to 0", pos)
	}
}

func TestSniffAudioFormatEmpty(t *testing.T) {
	if ext, ok := SniffAudioFormat(bytes.NewReader(nil)); ok {
		t.Fatalf("sniffed ext=%q on empty reader, want unknown", ext)
	}
}

func TestSniffAudioFormatResetsPosition(t *testing.T) {
	r := bytes.NewReader(append([]byte("fLaC"), bytes.Repeat([]byte{0xAA}, 64)...))
	if _, err := r.Seek(10, io.SeekStart); err != nil {
		t.Fatal(err)
	}
	ext, ok := SniffAudioFormat(r)
	if !ok || ext != "flac" {
		t.Fatalf("ext=%q ok=%v, want flac true", ext, ok)
	}
	if pos, _ := r.Seek(0, io.SeekCurrent); pos != 0 {
		t.Fatalf("reader position=%d, want reset to 0", pos)
	}
}
