package track

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/utils/netease"
)

func TestSniffFileFormat(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		wantExt string
		wantOK  bool
	}{
		{
			name:    "flac magic",
			data:    append([]byte("fLaC"), make([]byte, 4096)...),
			wantExt: "flac",
			wantOK:  true,
		},
		{
			name:    "ogg magic",
			data:    append([]byte("OggS"), make([]byte, 4096)...),
			wantExt: "ogg",
			wantOK:  true,
		},
		{
			name:    "wav riff wave",
			data:    append([]byte("RIFF\x24\x08\x00\x00WAVEfmt "), make([]byte, 4096)...),
			wantExt: "wav",
			wantOK:  true,
		},
		{
			name:    "id3v2 mp3",
			data:    append([]byte("ID3\x04\x00\x00\x00\x00\x00\x00"), make([]byte, 4096)...),
			wantExt: "mp3",
			wantOK:  true,
		},
		{
			name: "mpeg frame sync mp3",
			data: append([]byte{
				// 合法 MPEG1 Layer3 帧头：0xFF 0xFB 0x90 0x00
				0xFF, 0xFB, 0x90, 0x00,
			}, make([]byte, 4096)...),
			wantExt: "mp3",
			wantOK:  true,
		},
		{
			name:    "aac adts header rejected",
			data:    append([]byte{0xFF, 0xF1, 0x50, 0x80}, make([]byte, 4096)...), // layer 位为保留值 00
			wantExt: "",
			wantOK:  false,
		},
		{
			name:    "m4a ftyp rejected",
			data:    append([]byte{0x00, 0x00, 0x00, 0x20, 'f', 't', 'y', 'p'}, make([]byte, 4096)...),
			wantExt: "",
			wantOK:  false,
		},
		{
			name:    "empty input",
			data:    []byte{},
			wantExt: "",
			wantOK:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ext, ok := sniffFileFormat(bytes.NewReader(tt.data))
			if ext != tt.wantExt || ok != tt.wantOK {
				t.Fatalf("sniffFileFormat() = (%q, %v), want (%q, %v)", ext, ok, tt.wantExt, tt.wantOK)
			}
		})
	}
}

func TestSniffFileFormatFromPath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "song.mp3")
	if err := os.WriteFile(path, append([]byte("fLaC"), make([]byte, 4096)...), 0644); err != nil {
		t.Fatalf("write test file: %v", err)
	}

	ext, ok := sniffFileFormatFromPath(path)
	if !ok || ext != "flac" {
		t.Fatalf("sniffFileFormatFromPath() = (%q, %v), want (%q, %v)", ext, ok, "flac", true)
	}
}

type downloadFetcherStub struct {
	info *netease.PlayableInfo
	data []byte
}

func (f downloadFetcherStub) FetchPlayableInfo(context.Context, int64) (*netease.PlayableInfo, error) {
	return f.info, nil
}

func (f downloadFetcherStub) FetchStream(context.Context, PlayableSource) (io.ReadCloser, error) {
	return io.NopCloser(bytes.NewReader(f.data)), nil
}

func (f downloadFetcherStub) FetchLyric(context.Context, int64) (structs.LRCData, error) {
	return structs.LRCData{}, nil
}

func (f downloadFetcherStub) FetchCloudLyric(context.Context, int64, int64) (structs.LRCData, error) {
	return structs.LRCData{}, nil
}

type noopTagger struct{}

func (noopTagger) SetSongTag(string, structs.Song) error { return nil }

var downloadTestSong = structs.Song{
	Id:   1,
	Name: "崂山道士",
	Artists: []structs.Artist{
		{Name: "马思唯"},
	},
}

func newDownloadTestManager(t *testing.T, fetcher Fetcher) *Manager {
	t.Helper()
	return NewManager(
		WithDownloadDir(t.TempDir()),
		WithFetcher(fetcher),
		WithTagger(noopTagger{}),
	)
}

func TestPersistRemoteSourceRenamesFlacContentToFlac(t *testing.T) {
	// 网易云接口声明 mp3，实际内容为 FLAC（issue #642 场景）
	info := &netease.PlayableInfo{URL: "https://example.com/song.mp3", MusicType: "mp3"}
	manager := newDownloadTestManager(t, downloadFetcherStub{
		info: info,
		data: append([]byte("fLaC"), make([]byte, 4096)...),
	})

	path, err := manager.persistRemoteSource(context.Background(), PlayableSource{
		Song: downloadTestSong,
		Type: SourceRemote,
		Info: info,
	})
	if err != nil {
		t.Fatalf("persist remote source: %v", err)
	}

	if filepath.Ext(path) != ".flac" {
		t.Fatalf("downloaded file extension = %q, want .flac", filepath.Ext(path))
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("downloaded file missing: %v", err)
	}
	if _, err := os.Stat(strings.TrimSuffix(path, ".flac") + ".mp3"); !os.IsNotExist(err) {
		t.Fatalf("stale .mp3 file should not exist, stat err = %v", err)
	}
}

func TestPersistRemoteSourceRenamesMp3ContentFromFlacDecl(t *testing.T) {
	// 反向场景：声明 flac，实际内容为 MP3
	info := &netease.PlayableInfo{URL: "https://example.com/song.flac", MusicType: "flac"}
	manager := newDownloadTestManager(t, downloadFetcherStub{
		info: info,
		data: append([]byte("ID3\x04\x00\x00\x00\x00\x00\x00"), make([]byte, 4096)...),
	})

	path, err := manager.persistRemoteSource(context.Background(), PlayableSource{
		Song: downloadTestSong,
		Type: SourceRemote,
		Info: info,
	})
	if err != nil {
		t.Fatalf("persist remote source: %v", err)
	}

	if filepath.Ext(path) != ".mp3" {
		t.Fatalf("downloaded file extension = %q, want .mp3", filepath.Ext(path))
	}
}

func TestPersistRemoteSourceKeepsDeclaredExtWhenFormatUnknown(t *testing.T) {
	// 无法识别的内容（如 M4A）维持声明后缀，不强行改写
	info := &netease.PlayableInfo{URL: "https://example.com/song.mp3", MusicType: "mp3"}
	manager := newDownloadTestManager(t, downloadFetcherStub{
		info: info,
		data: append([]byte{0x00, 0x00, 0x00, 0x20, 'f', 't', 'y', 'p'}, make([]byte, 4096)...),
	})

	path, err := manager.persistRemoteSource(context.Background(), PlayableSource{
		Song: downloadTestSong,
		Type: SourceRemote,
		Info: info,
	})
	if err != nil {
		t.Fatalf("persist remote source: %v", err)
	}

	if filepath.Ext(path) != ".mp3" {
		t.Fatalf("downloaded file extension = %q, want .mp3 (declared)", filepath.Ext(path))
	}
}

func TestDownloadSongCorrectsStaleDownloadedFile(t *testing.T) {
	// 旧版本下载的 .mp3 命名 FLAC 文件，再次下载时按内容修正为 .flac
	manager := newDownloadTestManager(t, downloadFetcherStub{
		info: &netease.PlayableInfo{URL: "https://example.com/song.mp3", MusicType: "mp3"},
	})
	downloadDir := manager.downloadDir

	staleName, err := manager.nameGen.Song(downloadTestSong, "mp3")
	if err != nil {
		t.Fatalf("generate stale filename: %v", err)
	}
	stalePath := filepath.Join(downloadDir, staleName)
	if err := os.WriteFile(stalePath, append([]byte("fLaC"), make([]byte, 4096)...), 0644); err != nil {
		t.Fatalf("write stale file: %v", err)
	}

	path, err := manager.DownloadSong(context.Background(), downloadTestSong)
	if !errorsIsExist(err) {
		t.Fatalf("DownloadSong error = %v, want os.ErrExist", err)
	}
	if filepath.Ext(path) != ".flac" {
		t.Fatalf("returned path extension = %q, want .flac", filepath.Ext(path))
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("corrected file missing: %v", err)
	}
	if _, err := os.Stat(stalePath); !os.IsNotExist(err) {
		t.Fatalf("stale .mp3 file should have been renamed, stat err = %v", err)
	}
}

func errorsIsExist(err error) bool {
	return err != nil && os.IsExist(err)
}
