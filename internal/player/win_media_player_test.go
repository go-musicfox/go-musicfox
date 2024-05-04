package player

import (
	"path/filepath"
	"runtime"
	"time"

	"github.com/go-musicfox/go-musicfox/internal/structs"
)

func Example_winMediaPlayer_Play() {
	_, path, _, _ := runtime.Caller(0)
	uri := "file:///" + filepath.Join(filepath.Dir(filepath.Dir(filepath.Dir(path))), "testdata", "a.flac")

	player := NewWinMediaPlayer()
	player.Play(UrlMusic{
		Url:  uri,
		Type: Flac,
		Song: structs.Song{
			Id:       1,
			Name:     "test",
			Duration: time.Hour,
		},
	})

	time.Sleep(time.Hour)

	// Output:
	// test
}
