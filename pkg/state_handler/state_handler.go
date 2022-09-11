//go:build !darwin
// +build !darwin

package state_handler

import "go-musicfox/pkg/player"

type Handler struct {
}

func NewHandler(_ player.Player) *Handler {
	return &Handler{}
}

func (s *Handler) registerCommands() {
}

func (s *Handler) SetPlaybackState(_ player.State) {
}

func (s *Handler) SetPlayingInfo(_ PlayingInfo) {
}

func (s *Handler) Release() {
}
