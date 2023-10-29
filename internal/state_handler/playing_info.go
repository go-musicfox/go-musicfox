package state_handler

import (
	"time"

	"github.com/go-musicfox/go-musicfox/internal/types"
)

type PlayingInfo struct {
	TotalDuration  time.Duration
	PassedDuration time.Duration
	State          types.State
	Volume         int
	TrackID        int64
	PicUrl         string
	Name           string
	Artist         string
	Album          string
	AlbumArtist    string
	AsText         string
}
