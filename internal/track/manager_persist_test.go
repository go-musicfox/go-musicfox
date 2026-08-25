package track

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-musicfox/go-musicfox/internal/structs"
)

func TestPersistStreamCorrectsSuffixByContent(t *testing.T) {
	dir := t.TempDir()
	m := NewManager(WithDownloadDir(dir))

	song := structs.Song{Id: 123, Name: "CloudSong", Artists: []structs.Artist{{Name: "Singer"}}}
	fileName, err := m.nameGen.Song(song, "mp3")
	if err != nil {
		t.Fatalf("generate name: %v", err)
	}

	job := persistJob{
		stream:        io.NopCloser(bytes.NewReader(append([]byte("fLaC\x00\x00\x00\x22"), bytes.Repeat([]byte{0xAA}, 128)...))),
		finalFilePath: filepath.Join(dir, fileName),
		source:        PlayableSource{Song: song},
		isFromCache:   true,
	}

	path, err := m.persistStream(job)
	if err != nil {
		t.Fatalf("persist stream: %v", err)
	}
	if filepath.Ext(path) != ".flac" {
		t.Fatalf("suffix = %q, want .flac (path=%s)", filepath.Ext(path), path)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("corrected file missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, fileName)); !os.IsNotExist(err) {
		t.Fatalf("declared .mp3 file should be absent, got err=%v", err)
	}
}

func TestPersistStreamKeepsDeclaredSuffixWhenSniffUnknown(t *testing.T) {
	dir := t.TempDir()
	m := NewManager(WithDownloadDir(dir))

	song := structs.Song{Id: 456, Name: "CloudSong", Artists: []structs.Artist{{Name: "Singer"}}}
	fileName, err := m.nameGen.Song(song, "mp3")
	if err != nil {
		t.Fatalf("generate name: %v", err)
	}

	job := persistJob{
		stream:        io.NopCloser(bytes.NewReader([]byte("\x00\x00\x00\x18ftypM4A \x00\x00\x00\x00"))),
		finalFilePath: filepath.Join(dir, fileName),
		source:        PlayableSource{Song: song},
		isFromCache:   true,
	}

	path, err := m.persistStream(job)
	if err != nil {
		t.Fatalf("persist stream: %v", err)
	}
	if filepath.Ext(path) != ".mp3" {
		t.Fatalf("suffix = %q, want .mp3 (path=%s)", filepath.Ext(path), path)
	}
}
