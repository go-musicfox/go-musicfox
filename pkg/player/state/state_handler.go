//go:build !darwin
// +build !darwin

package state

type Handler struct {
}

func NewHandler(_ Player) *Handler {
	return &Handler{}
}

func (s *Handler) registerCommands() {
}

func (s *Handler) SetPlaybackState(_ uint8) {
}

func (s *Handler) SetPlayingInfo(_ PlayingInfo) {
}

func (s *Handler) Release() {
}
