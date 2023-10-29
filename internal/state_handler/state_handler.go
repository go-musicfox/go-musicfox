//go:build !darwin && !linux

package state_handler

import "time"

type Handler struct {
}

func NewHandler(Controller, PlayingInfo) *Handler {
	return &Handler{}
}

func (s *Handler) SetPosition(time.Duration) {
}

func (s *Handler) SetPlayingInfo(PlayingInfo) {
}

func (s *Handler) Release() {
}
