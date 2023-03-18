package state_handler

import (
	"time"

	"github.com/go-musicfox/go-musicfox/pkg/player"
)

type PlayingInfo struct {
	TotalDuration  time.Duration
	PassedDuration time.Duration
	State          player.State
	Volume         int
	TrackID        int64
	PicUrl         string
	Name           string
	Artist         string
	Album          string
	AlbumArtist    string
}
