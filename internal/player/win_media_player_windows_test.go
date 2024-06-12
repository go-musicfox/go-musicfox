package player

import (
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/go-musicfox/go-musicfox/internal/structs"
)

func TestWinMediaPlayer_Play(t *testing.T) {
	if runtime.GOOS != "windows" {
		return
	}
	_, path, _, _ := runtime.Caller(0)
	uri := "file:///" + filepath.Join(filepath.Dir(filepath.Dir(filepath.Dir(path))), "testdata", "a.mp3")

	player := NewWinMediaPlayer()
	player.Play(URLMusic{
		URL:  uri,
		Type: Flac,
		Song: structs.Song{
			Id:       1,
			Name:     "test",
			Duration: time.Hour,
		},
	})
	<-time.After(time.Second * 2)
	if player.PassedTime() < time.Second*2 {
		t.Fatal("win media player not work")
	}
}
