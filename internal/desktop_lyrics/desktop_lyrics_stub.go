//go:build !darwin

package desktop_lyrics

import (
	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/player"
)

type stubController struct{}

func (stubController) Show()                                           {}
func (stubController) Hide()                                           {}
func (stubController) IsVisible() bool                                 { return false }
func (stubController) Update(_, _ LyricLine, _ int, _ int64, _ bool)   {}
func (stubController) UpdateSpectrum(_ player.SpectrumFrame)           {}
func (stubController) UpdateRawSamples(_ player.RawSampleFrame)        {}
func (stubController) SetSpectrumAvailable(_ bool)                     {}
func (stubController) Close()                                          {}

// newController returns a stub on non-macOS platforms.
func newController(cfg configs.DesktopLyricsConfig) Controller {
	return stubController{}
}
