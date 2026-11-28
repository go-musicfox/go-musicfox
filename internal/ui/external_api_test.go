package ui

import (
	"testing"
	"time"

	"github.com/go-musicfox/go-musicfox/internal/externalapi"
	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/internal/types"
)

func TestExternalAPITrackMapping(t *testing.T) {
	player := &Player{}
	track := player.externalAPITrack(structs.Song{
		Id:       123,
		Name:     "Example",
		Duration: 3 * time.Minute,
		Artists:  []structs.Artist{{Id: 1, Name: "Artist"}},
		Album:    structs.Album{Id: 2, Name: "Album", PicUrl: "https://example.test/cover.jpg"},
	})
	if track == nil {
		t.Fatal("track is nil")
	}
	if track.ID != "123" || track.Source != "netease" || track.Title != "Example" || track.Duration != 180000 {
		t.Fatalf("unexpected track: %+v", track)
	}
	if len(track.Artists) != 1 || track.Artists[0].Name != "Artist" || track.Album == nil || track.Album.Name != "Album" {
		t.Fatalf("unexpected metadata: %+v", track)
	}
}

func TestExternalAPIStateMapping(t *testing.T) {
	cases := map[types.State]externalapi.PlayerState{
		types.Playing:     externalapi.StatePlaying,
		types.Paused:      externalapi.StatePaused,
		types.Stopped:     externalapi.StateStopped,
		types.Interrupted: externalapi.StateStopped,
		types.Unknown:     externalapi.StateIdle,
	}
	for input, expected := range cases {
		if got := externalAPIState(input); got != expected {
			t.Fatalf("state %v => %q, want %q", input, got, expected)
		}
	}
}
