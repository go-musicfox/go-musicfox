package ui

import (
	"image/color"
	"strings"
	"testing"

	"github.com/go-musicfox/go-musicfox/internal/configs"

	"github.com/go-musicfox/go-musicfox/internal/player"
)

func TestSpectrumLineCountRespondsToAvailableSpace(t *testing.T) {
	previousConfig := configs.AppConfig
	configs.AppConfig = &configs.Config{}
	configs.AppConfig.Main.Visualizer.Enable = true
	defer func() { configs.AppConfig = previousConfig }()

	renderer := &SpectrumRenderer{provider: spectrumTestProvider{}}

	if got := renderer.LineCount(40, 20); got != 10 {
		t.Fatalf("line count = %d, want 10", got)
	}
	// Lyrics have priority; spectrum yields when space is tight.
	if got := renderer.LineCount(30, 20); got != 0 {
		t.Fatalf("line count = %d, want 0", got)
	}
	if got := renderer.LineCount(28, 20); got != 0 {
		t.Fatalf("line count = %d, want 0", got)
	}
}

func TestSpectrumLayoutFillsAvailableSpaceAndSeparatesSongInfo(t *testing.T) {
	previousConfig := configs.AppConfig
	configs.AppConfig = &configs.Config{}
	configs.AppConfig.Main.Visualizer.Enable = true
	t.Cleanup(func() { configs.AppConfig = previousConfig })

	renderer := &SpectrumRenderer{provider: spectrumTestProvider{}}
	layout := renderer.layout(40, 20)
	if layout.barLines != 8 || layout.topPadding != 1 || layout.bottomPadding != 1 {
		t.Fatalf("layout = %+v, want 1 top, 8 bars, and 1 bottom", layout)
	}
	if layout.lines() != 10 {
		t.Fatalf("layout lines = %d, want 10", layout.lines())
	}
}

func TestSpectrumLayoutRespectsConfiguredMaxHeight(t *testing.T) {
	previousConfig := configs.AppConfig
	configs.AppConfig = &configs.Config{}
	configs.AppConfig.Main.Visualizer.Enable = true
	configs.AppConfig.Main.Visualizer.MaxHeight = 3
	t.Cleanup(func() { configs.AppConfig = previousConfig })

	renderer := &SpectrumRenderer{provider: spectrumTestProvider{}}
	if got := renderer.layout(40, 20).barLines; got != 3 {
		t.Fatalf("bar lines = %d, want 3", got)
	}
}

func TestSpectrumRenderWithLayoutAddsPaddingAndSongInfoSeparator(t *testing.T) {
	previousConfig := configs.AppConfig
	configs.AppConfig = &configs.Config{}
	t.Cleanup(func() { configs.AppConfig = previousConfig })

	renderer := &SpectrumRenderer{}
	frame := player.SpectrumFrame{}
	frame.Levels[0] = 1
	plain := stripAnsiCodes(renderer.renderWithLayout(frame, 4, spectrumLayout{
		topPadding:    1,
		barLines:      1,
		bottomPadding: 1,
	}))
	lines := strings.Split(plain, "\n")
	if lines[0] != "" || lines[2] != "" {
		t.Fatalf("layout output = %q, want blank lines above spectrum and before song info", plain)
	}
}

func TestSpectrumRenderUsesRequestedHeight(t *testing.T) {
	previousConfig := configs.AppConfig
	configs.AppConfig = &configs.Config{}
	t.Cleanup(func() { configs.AppConfig = previousConfig })
	renderer := &SpectrumRenderer{}
	frame := player.SpectrumFrame{}
	frame.Levels[0] = 1

	view := renderer.render(frame, 8, 3)
	plain := stripAnsiCodes(view)
	lines := strings.Split(strings.TrimSuffix(plain, "\n"), "\n")
	if len(lines) != 3 {
		t.Fatalf("line count = %d, want 3", len(lines))
	}
	for _, line := range lines {
		if len([]rune(line)) != 8 {
			t.Fatalf("line width = %d, want 8", len([]rune(line)))
		}
	}
	if lines[0] != strings.Repeat(" ", 8) || lines[1] != strings.Repeat(" ", 8) {
		t.Fatalf("empty spectrum rows = %q, %q, want blank cells", lines[0], lines[1])
	}
	if lines[2] != strings.Repeat("█", 8) {
		t.Fatalf("full spectrum row = %q, want Bubbles full block", lines[2])
	}
}

func TestSpectrumRenderUsesConfiguredCharacters(t *testing.T) {
	previousConfig := configs.AppConfig
	configs.AppConfig = &configs.Config{}
	configs.AppConfig.Main.Visualizer.FullCharHalfBlock = "~"
	configs.AppConfig.Main.Visualizer.FullCharFullBlock = "@"
	configs.AppConfig.Main.Visualizer.EmptyCharBlock = "."
	t.Cleanup(func() { configs.AppConfig = previousConfig })

	renderer := &SpectrumRenderer{}
	frame := player.SpectrumFrame{}
	frame.Levels[0] = 1.0 / 8
	partial := strings.TrimSuffix(stripAnsiCodes(renderer.render(frame, 4, 1)), "\n")
	if partial != "~..." {
		t.Fatalf("partial spectrum row = %q, want %q", partial, "~...")
	}
	frame.Levels[0] = 1
	full := strings.TrimSuffix(stripAnsiCodes(renderer.render(frame, 4, 1)), "\n")
	if full != "@@@@" {
		t.Fatalf("full spectrum row = %q, want %q", full, "@@@@")
	}
}

func TestSpectrumFillUnitsPreserveHalfCharacterResolution(t *testing.T) {
	if got := spectrumFillUnits(0.125, 80); got != 20 {
		t.Fatalf("fill units = %d, want 20", got)
	}
}

func TestSpectrumRenderUsesHalfBlockForPartialCell(t *testing.T) {
	previousConfig := configs.AppConfig
	configs.AppConfig = &configs.Config{}
	t.Cleanup(func() { configs.AppConfig = previousConfig })

	renderer := &SpectrumRenderer{}
	frame := player.SpectrumFrame{}
	frame.Levels[0] = 1.0 / 16
	plain := strings.TrimSuffix(stripAnsiCodes(renderer.render(frame, 8, 1)), "\n")
	want := "▌" + strings.Repeat(" ", 7)
	if plain != want {
		t.Fatalf("partial spectrum row = %q, want %q", plain, want)
	}
}

func TestSpectrumHalfBlockStyleUsesForegroundAndBackground(t *testing.T) {
	foreground := color.RGBA{R: 0x12, G: 0x34, B: 0x56, A: 0xff}
	background := color.RGBA{R: 0x65, G: 0x43, B: 0x21, A: 0xff}
	style := spectrumHalfBlockStyle(foreground, background)

	if gotR, gotG, gotB, gotA := style.GetForeground().RGBA(); gotR != 0x1212 || gotG != 0x3434 || gotB != 0x5656 || gotA != 0xffff {
		t.Fatalf("half-block foreground = %#x %#x %#x %#x", gotR, gotG, gotB, gotA)
	}
	if gotR, gotG, gotB, gotA := style.GetBackground().RGBA(); gotR != 0x6565 || gotG != 0x4343 || gotB != 0x2121 || gotA != 0xffff {
		t.Fatalf("half-block background = %#x %#x %#x %#x", gotR, gotG, gotB, gotA)
	}
}

func TestSpectrumEmptyCellsHaveNoStyle(t *testing.T) {
	ramp := make([]color.Color, 16)
	for index := range ramp {
		ramp[index] = color.RGBA{R: 0x12, G: 0x34, B: 0x56, A: 0xff}
	}
	bar := renderSpectrumBar(1.0/16, 8, ramp, '▌', '█', ' ')
	if !strings.HasSuffix(bar, strings.Repeat(" ", 7)) {
		t.Fatalf("empty spectrum cells are styled: %q", bar)
	}
}

func TestSpectrumRowLevelMapsLowFrequenciesToBottom(t *testing.T) {
	frame := player.SpectrumFrame{}
	frame.Levels[0] = 1
	frame.Levels[player.SpectrumBandCount-1] = 0.5

	if got := spectrumRowLevel(frame, 0, 2); got != 0.5 {
		t.Fatalf("top row level = %f, want 0.5", got)
	}
	if got := spectrumRowLevel(frame, 1, 2); got != 1 {
		t.Fatalf("bottom row level = %f, want 1", got)
	}
}

type spectrumTestProvider struct{}

func (spectrumTestProvider) Spectrum() player.SpectrumFrame {
	return player.SpectrumFrame{}
}
