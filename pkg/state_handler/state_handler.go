//go:build !darwin && !linux

package state_handler

type Handler struct {
}

func NewHandler(_ Controller) *Handler {
	return &Handler{}
}

func (s *Handler) SetPlayingInfo(_ PlayingInfo) {
}

func (s *Handler) Release() {
}
