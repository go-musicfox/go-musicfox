package ui

import (
	"testing"

	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/player"
)

func benchSpectrumRender(b *testing.B, style string) {
	previousConfig := configs.AppConfig
	configs.AppConfig = &configs.Config{}
	configs.AppConfig.Main.Visualizer.Enable = true
	configs.AppConfig.Main.Visualizer.Style = style
	b.Cleanup(func() { configs.AppConfig = previousConfig })

	r := &SpectrumRenderer{}
	frame := player.SpectrumFrame{}
	for i := range frame.LevelsL {
		frame.LevelsL[i] = float64(i) / player.SpectrumBandCount
		frame.LevelsR[i] = 1 - float64(i)/player.SpectrumBandCount
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.render(frame, 200, 20)
	}
}

func BenchmarkSpectrumRenderLine200x20(b *testing.B) { benchSpectrumRender(b, "line") }
func BenchmarkSpectrumRenderBar200x20(b *testing.B)  { benchSpectrumRender(b, "bar") }
