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
	if layout.barLines != 9 || layout.topPadding != 1 || layout.bottomPadding != 0 {
		t.Fatalf("layout = %+v, want 1 top, 9 bars, and 0 bottom", layout)
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
		t.Fatalf("full spectrum row = %q, want full block", lines[2])
	}
}

func TestSpectrumRenderUsesConfiguredCharacters(t *testing.T) {
	previousConfig := configs.AppConfig
	configs.AppConfig = &configs.Config{}
	configs.AppConfig.Main.Visualizer.BarHalfBlock = "~"
	configs.AppConfig.Main.Visualizer.BarFullBlock = "@"
	configs.AppConfig.Main.Visualizer.BarEmptyBlock = "."
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

// --- Style dispatch tests ---

func TestSpectrumRenderStyleDispatch(t *testing.T) {
	setConfig := func(style string) {
		configs.AppConfig = &configs.Config{}
		configs.AppConfig.Main.Visualizer.Style = style
	}
	t.Cleanup(func() { configs.AppConfig = &configs.Config{} })

	renderer := &SpectrumRenderer{}
	frame := player.SpectrumFrame{}
	frame.Levels[0] = 1
	frame.LevelsL[0] = 1

	// Unknown/empty style defaults to bar.
	setConfig("")
	plain := stripAnsiCodes(renderer.render(frame, 4, 1))
	if !strings.Contains(plain, "█") {
		t.Fatalf("default style should render bar: %q", plain)
	}

	// Explicit bar style.
	setConfig("bar")
	plain = stripAnsiCodes(renderer.render(frame, 4, 1))
	if !strings.Contains(plain, "█") {
		t.Fatalf("bar style output: %q", plain)
	}

	// Line style produces braille characters.
	setConfig("line")
	plain = stripAnsiCodes(renderer.render(frame, 4, 2))
	if len(strings.TrimSpace(plain)) == 0 {
		t.Fatalf("line style produced empty output")
	}

	// Dot style produces non-empty output.
	setConfig("dot")
	plain = stripAnsiCodes(renderer.render(frame, 4, 2))
	if len(strings.TrimSpace(plain)) == 0 {
		t.Fatalf("dot style produced empty output")
	}

	// Mirror bar style.
	setConfig("mirror_bar")
	plain = stripAnsiCodes(renderer.render(frame, 8, 2))
	if !strings.Contains(plain, "█") {
		t.Fatalf("mirror_bar style output: %q", plain)
	}
}

func TestSpectrumRenderEmptySignalIsBlank(t *testing.T) {
	configs.AppConfig = &configs.Config{}
	t.Cleanup(func() { configs.AppConfig = &configs.Config{} })

	renderer := &SpectrumRenderer{}
	frame := player.SpectrumFrame{} // all zeros

	// bar/mirror_bar should be blank when no signal.
	for _, style := range []string{"bar", "mirror_bar"} {
		configs.AppConfig.Main.Visualizer.Style = style
		view := renderer.render(frame, 8, 3)
		plain := stripAnsiCodes(view)
		lines := strings.Split(strings.TrimSuffix(plain, "\n"), "\n")
		if len(lines) != 3 {
			t.Fatalf("%s: got %d lines, want 3", style, len(lines))
		}
		for i, line := range lines {
			if strings.TrimSpace(line) != "" {
				t.Fatalf("%s: line %d non-empty for zero signal: %q", style, i, line)
			}
		}
	}

	// line/dot should show baseline even with zero signal.
	for _, style := range []string{"line", "dot"} {
		configs.AppConfig.Main.Visualizer.Style = style
		view := renderer.render(frame, 8, 3)
		plain := stripAnsiCodes(view)
		lines := strings.Split(strings.TrimSuffix(plain, "\n"), "\n")
		if len(lines) != 3 {
			t.Fatalf("%s: got %d lines, want 3", style, len(lines))
		}
		// All rows except bottom should be empty.
		for i := 0; i < 2; i++ {
			if strings.TrimSpace(lines[i]) != "" {
				t.Fatalf("%s: non-bottom row %d non-empty: %q", style, i, lines[i])
			}
		}
	}
}

// --- Mirror bar tests ---

func TestMirrorBarSymmetry(t *testing.T) {
	ramp := make([]color.Color, 32)
	for i := range ramp {
		ramp[i] = color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
	}

	// Full level: should be symmetric after stripping ANSI.
	full := stripAnsiCodes(renderMirrorBarLine(1.0, 8, ramp, '▌', '█', ' '))
	fullRunes := []rune(full)
	if len(fullRunes) != 8 {
		t.Fatalf("full mirror bar length = %d, want 8: %q", len(fullRunes), full)
	}
	for i := 0; i < len(fullRunes)/2; i++ {
		if fullRunes[i] != fullRunes[len(fullRunes)-1-i] {
			t.Fatalf("mirror bar not symmetric at full level: %q", full)
		}
	}

	// Half level.
	half := stripAnsiCodes(renderMirrorBarLine(0.5, 8, ramp, '▌', '█', ' '))
	halfRunes := []rune(half)
	if len(halfRunes) != 8 {
		t.Fatalf("half mirror bar length = %d, want 8: %q", len(halfRunes), half)
	}
	for i := 0; i < len(halfRunes)/2; i++ {
		if halfRunes[i] != halfRunes[len(halfRunes)-1-i] {
			t.Fatalf("mirror bar not symmetric at half level: %q", half)
		}
	}

	// Odd width (7): no center gap, halves are adjacent.
	odd := stripAnsiCodes(renderMirrorBarLine(1.0, 7, ramp, '▌', '█', ' '))
	oddRunes := []rune(odd)
	if len(oddRunes) != 7 {
		t.Fatalf("odd width bar length = %d, want 7: %q", len(oddRunes), odd)
	}
}

// --- Braille helper tests ---

func TestBrailleCell(t *testing.T) {
	// Empty braille.
	if brailleCell(0) != '\u2800' {
		t.Fatalf("empty braille = %U", brailleCell(0))
	}
	// All dots.
	if brailleCell(0xFF) != '\u28FF' {
		t.Fatalf("full braille = %U", brailleCell(0xFF))
	}
	// Single dot: dot 1 = bit 0.
	if brailleCell(0x01) != '\u2801' {
		t.Fatalf("dot-1 braille = %U", brailleCell(0x01))
	}
}

func TestSetBrailleDot(t *testing.T) {
	grid := [][]byte{
		make([]byte, 2),
		make([]byte, 2),
	}

	// Set dot at subCol=0, subRow=0 (left column, top of cell) → dot 1.
	setBrailleDot(grid, 0, 0)
	if grid[0][0] != 0x01 {
		t.Fatalf("grid[0][0] = %#x, want 0x01", grid[0][0])
	}

	// Set dot at subCol=1, subRow=0 (right column, top of cell) → dot 4.
	setBrailleDot(grid, 1, 0)
	if grid[0][0] != 0x09 { // 0x01 | 0x08
		t.Fatalf("grid[0][0] = %#x, want 0x09", grid[0][0])
	}

	// Set dot at subCol=3, subRow=7 → col=1, row=1, x=1, y=3 → dot 8.
	setBrailleDot(grid, 3, 7)
	if grid[1][1] != 0x80 {
		t.Fatalf("grid[1][1] = %#x, want 0x80", grid[1][1])
	}

	// Out of bounds should not panic.
	setBrailleDot(grid, -1, 0)
	setBrailleDot(grid, 0, -1)
	setBrailleDot(grid, 100, 0)
	setBrailleDot(grid, 0, 100)
}

func TestHasSignal(t *testing.T) {
	if hasSignal(player.SpectrumFrame{}) {
		t.Fatal("empty frame should have no signal")
	}
	frame := player.SpectrumFrame{}
	frame.Levels[player.SpectrumBandCount/2] = 0.5
	if !hasSignal(frame) {
		t.Fatal("frame with data should have signal")
	}
}

// --- Render mirror bar (full integration) ---

func TestSpectrumMirrorBarRender(t *testing.T) {
	configs.AppConfig = &configs.Config{}
	configs.AppConfig.Main.Visualizer.Style = "mirror_bar"
	t.Cleanup(func() { configs.AppConfig = &configs.Config{} })

	renderer := &SpectrumRenderer{}
	frame := player.SpectrumFrame{}
	frame.LevelsL[0] = 1
	frame.Levels[0] = 1

	view := renderer.render(frame, 8, 3)
	plain := stripAnsiCodes(view)
	lines := strings.Split(strings.TrimSuffix(plain, "\n"), "\n")
	if len(lines) != 3 {
		t.Fatalf("line count = %d, want 3", len(lines))
	}
	// Bottom row (low freq) should have bars on both sides.
	bottom := lines[2]
	if !strings.Contains(bottom, "█") {
		t.Fatalf("mirror bar bottom row empty: %q", bottom)
	}
}

// --- Render line/dot style produces braille ---

func TestSpectrumLineRenderProducesBraille(t *testing.T) {
	configs.AppConfig = &configs.Config{}
	configs.AppConfig.Main.Visualizer.Style = "line"
	t.Cleanup(func() { configs.AppConfig = &configs.Config{} })

	renderer := &SpectrumRenderer{}
	frame := player.SpectrumFrame{}
	for i := range frame.LevelsL {
		frame.LevelsL[i] = 0.5
	}

	view := renderer.render(frame, 16, 4)
	plain := stripAnsiCodes(view)
	// Should contain braille characters (U+2800+).
	hasBraille := false
	for _, r := range plain {
		if r >= '\u2800' && r <= '\u28FF' {
			hasBraille = true
			break
		}
	}
	if !hasBraille {
		t.Fatalf("line style produced no braille chars: %q", plain)
	}
}

func TestSpectrumDotRenderProducesBraille(t *testing.T) {
	configs.AppConfig = &configs.Config{}
	configs.AppConfig.Main.Visualizer.Style = "dot"
	t.Cleanup(func() { configs.AppConfig = &configs.Config{} })

	renderer := &SpectrumRenderer{}
	frame := player.SpectrumFrame{}
	frame.LevelsL[0] = 1

	view := renderer.render(frame, 16, 4)
	plain := stripAnsiCodes(view)
	hasBraille := false
	for _, r := range plain {
		if r >= '\u2800' && r <= '\u28FF' {
			hasBraille = true
			break
		}
	}
	if !hasBraille {
		t.Fatalf("dot style produced no braille chars: %q", plain)
	}
}

// --- Block character rendering (dotChar configured) ---

func TestSpectrumLineBlockRender(t *testing.T) {
	configs.AppConfig = &configs.Config{}
	configs.AppConfig.Main.Visualizer.Style = "line"
	configs.AppConfig.Main.Visualizer.LineMode = "block"
	configs.AppConfig.Main.Visualizer.LineFullBlock = "█"
	t.Cleanup(func() { configs.AppConfig = &configs.Config{} })

	renderer := &SpectrumRenderer{}
	frame := player.SpectrumFrame{}
	for i := range frame.LevelsL {
		frame.LevelsL[i] = 0.5
	}

	view := renderer.render(frame, 16, 4)
	plain := stripAnsiCodes(view)
	lines := strings.Split(strings.TrimSuffix(plain, "\n"), "\n")
	if len(lines) != 4 {
		t.Fatalf("line block: got %d lines, want 4", len(lines))
	}

	// At level=0.5, height=4: targetRow = 3 - round(1.5) = 1.
	// Line at row 1, fill-below fills rows 2 and 3.
	if strings.TrimSpace(lines[0]) != "" {
		t.Fatalf("line block top row should be empty: %q", lines[0])
	}
	if strings.TrimSpace(lines[3]) == "" {
		t.Fatalf("line block bottom row should be filled: %q", lines[3])
	}

	// Must contain block chars, not braille.
	for _, r := range plain {
		if r >= '\u2800' && r <= '\u28FF' {
			t.Fatalf("line block produced braille despite dotChar set: %q", plain)
		}
	}
	if !strings.Contains(plain, "█") {
		t.Fatalf("line block missing block char: %q", plain)
	}
}

func TestSpectrumDotBlockRender(t *testing.T) {
	configs.AppConfig = &configs.Config{}
	configs.AppConfig.Main.Visualizer.Style = "dot"
	configs.AppConfig.Main.Visualizer.DotMode = "block"
	configs.AppConfig.Main.Visualizer.DotFullBlock = "█"
	t.Cleanup(func() { configs.AppConfig = &configs.Config{} })

	renderer := &SpectrumRenderer{}
	frame := player.SpectrumFrame{}
	frame.LevelsL[0] = 1

	view := renderer.render(frame, 16, 4)
	plain := stripAnsiCodes(view)

	// Must contain block chars, not braille.
	for _, r := range plain {
		if r >= '\u2800' && r <= '\u28FF' {
			t.Fatalf("dot block produced braille despite dotChar set: %q", plain)
		}
	}
	if !strings.Contains(plain, "█") {
		t.Fatalf("dot block missing block char: %q", plain)
	}
}

func TestSpectrumDotCharEmptyUsesBraille(t *testing.T) {
	configs.AppConfig = &configs.Config{}
	configs.AppConfig.Main.Visualizer.Style = "dot"
	configs.AppConfig.Main.Visualizer.DotMode = "" // default braille
	t.Cleanup(func() { configs.AppConfig = &configs.Config{} })

	renderer := &SpectrumRenderer{}
	frame := player.SpectrumFrame{}
	frame.LevelsL[0] = 1

	view := renderer.render(frame, 16, 4)
	plain := stripAnsiCodes(view)
	hasBraille := false
	for _, r := range plain {
		if r >= '\u2800' && r <= '\u28FF' {
			hasBraille = true
			break
		}
	}
	if !hasBraille {
		t.Fatalf("dot style with empty dotChar should use braille: %q", plain)
	}
}

type spectrumTestProvider struct{}

func (spectrumTestProvider) Spectrum() player.SpectrumFrame {
	return player.SpectrumFrame{}
}

func (spectrumTestProvider) RawSamples() player.RawSampleFrame {
	return player.RawSampleFrame{}
}
