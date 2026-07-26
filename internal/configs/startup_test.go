package configs

import (
	"testing"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/anhoder/foxful-cli/model"
)

func TestStartupConfigToModelIncludesAnimationOptions(t *testing.T) {
	config := StartupConfig{
		Enable:            true,
		ProgressOutBounce: true,
		LoadingSeconds:    3,
		Welcome:           "musicfox",
		Animation:         string(model.StartupAnimationRainbowWave),
		ReducedMotion:     true,
	}

	options := config.ToModel()
	if got, want := options.Animation, model.StartupAnimationRainbowWave; got != want {
		t.Errorf("animation = %q, want %q", got, want)
	}
	if !options.ReducedMotion {
		t.Error("ReducedMotion = false, want true")
	}
	if got, want := options.LoadingDuration, 3*time.Second; got != want {
		t.Errorf("loading duration = %s, want %s", got, want)
	}
}

func TestThemeConfigModelThemesKeepPrimaryColor(t *testing.T) {
	primary := lipgloss.Color("#EA403F")
	dark, light := (ThemeConfig{}).modelThemes(primary)

	if got, want := dark.Primary, primary; got != want {
		t.Errorf("dark primary = %v, want %v", got, want)
	}
	if got, want := light.Primary, primary; got != want {
		t.Errorf("light primary = %v, want %v", got, want)
	}
}
