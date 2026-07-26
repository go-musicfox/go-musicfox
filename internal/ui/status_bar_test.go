package ui

import (
	"testing"

	"github.com/go-musicfox/netease-music/service"

	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/playlist"
	"github.com/go-musicfox/go-musicfox/internal/structs"
)

func TestFormatQueueAndQualityIncludesMusicfox(t *testing.T) {
	previousConfig := configs.AppConfig
	configs.AppConfig = &configs.Config{}
	configs.AppConfig.Player.SongLevel = service.Lossless
	t.Cleanup(func() { configs.AppConfig = previousConfig })

	player := &Player{playlistManager: playlist.NewPlaylistManager()}
	player.InitSongManager(0, []structs.Song{{Id: 1}, {Id: 2}})

	const want = "musicfox · [1/2] · 无损"
	if got := formatQueueAndQuality(player); got != want {
		t.Fatalf("formatQueueAndQuality() = %q, want %q", got, want)
	}
}
