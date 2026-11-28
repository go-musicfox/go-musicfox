package player

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMpdConfigMusicDirectory(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "mpd.conf")
	config := "# comment\n" +
		"music_directory \"/srv/music\" # inline comment\n" +
		"audio_output {}\n"
	if err := os.WriteFile(configPath, []byte(config), 0o644); err != nil {
		t.Fatal(err)
	}

	if got := mpdConfigMusicDirectory(configPath); got != "/srv/music" {
		t.Fatalf("mpdConfigMusicDirectory() = %q, want %q", got, "/srv/music")
	}
	if got := mpdConfigMusicDirectory(""); got != "" {
		t.Fatalf("mpdConfigMusicDirectory(\"\") = %q, want empty", got)
	}
}

func TestMpdConfigMusicDirectoryBareValue(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "mpd.conf")
	if err := os.WriteFile(configPath, []byte("music_directory /srv/music\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if got := mpdConfigMusicDirectory(configPath); got != "/srv/music" {
		t.Fatalf("mpdConfigMusicDirectory() = %q, want %q", got, "/srv/music")
	}
}

func TestRelativeLocalURI(t *testing.T) {
	root := t.TempDir()
	musicDir := filepath.Join(root, "music")
	inside := filepath.Join(musicDir, "album", "song.mp3")
	outside := filepath.Join(root, "downloads", "song.mp3")
	for _, path := range []string{inside, outside} {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("audio"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	if got, ok := relativeLocalURI(musicDir, inside); !ok || got != "album/song.mp3" {
		t.Fatalf("relativeLocalURI(inside) = %q, %t; want %q, true", got, ok, "album/song.mp3")
	}
	if got, ok := relativeLocalURI(musicDir, outside); ok {
		t.Fatalf("relativeLocalURI(outside) = %q, %t; want not ok", got, ok)
	}
}

func TestStageLocalFile(t *testing.T) {
	root := t.TempDir()
	musicDir := filepath.Join(root, "music")
	if err := os.MkdirAll(musicDir, 0o755); err != nil {
		t.Fatal(err)
	}
	source := filepath.Join(root, "downloads", "song.mp3")
	if err := os.MkdirAll(filepath.Dir(source), 0o755); err != nil {
		t.Fatal(err)
	}
	want := []byte("audio data")
	if err := os.WriteFile(source, want, 0o644); err != nil {
		t.Fatal(err)
	}

	uri, staged, err := stageLocalFile(musicDir, source)
	if err != nil {
		t.Fatalf("stageLocalFile() error = %v", err)
	}
	if uri != "musicfox-staging/song.mp3" {
		t.Fatalf("stageLocalFile() uri = %q, want %q", uri, "musicfox-staging/song.mp3")
	}
	if staged != filepath.Join(musicDir, mpdStagingDir, "song.mp3") {
		t.Fatalf("stageLocalFile() staged = %q", staged)
	}
	got, err := os.ReadFile(staged)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(want) {
		t.Fatalf("staged content = %q, want %q", got, want)
	}
}

func TestStageLocalFileInsideMusicDir(t *testing.T) {
	musicDir := t.TempDir()
	source := filepath.Join(musicDir, "song.mp3")
	if err := os.WriteFile(source, []byte("audio"), 0o644); err != nil {
		t.Fatal(err)
	}

	uri, staged, err := stageLocalFile(musicDir, source)
	if err != nil {
		t.Fatalf("stageLocalFile() error = %v", err)
	}
	if uri != "song.mp3" || staged != "" {
		t.Fatalf("stageLocalFile() = %q, %q; want %q, empty", uri, staged, "song.mp3")
	}
}
