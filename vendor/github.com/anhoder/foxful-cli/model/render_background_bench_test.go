package model

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
)

// benchFrame builds a representative 200x40 frame: full-width styled rows
// (visualizer output), short styled rows (menu items), plain text (lyrics),
// and empty rows.
func benchFrame() (string, lipgloss.Style) {
	bg := lipgloss.NewStyle().Background(lipgloss.Color("#000000"))
	var b strings.Builder
	// 20 visualizer rows: every cell carries its own SGR.
	visRow := strings.Repeat("\x1b[38;2;120;80;40;48;2;0;0;0m \x1b[m", 200/1)
	// 200 visible cells requires 200 blocks of 1 cell.
	visRow = strings.Repeat("\x1b[38;2;120;80;40;48;2;0;0;0m \x1b[m", 200)
	for i := 0; i < 20; i++ {
		b.WriteString(visRow)
		b.WriteByte('\n')
	}
	// 10 menu rows: styled short text.
	for i := 0; i < 10; i++ {
		b.WriteString("\x1b[38;2;255;255;255m => 12. 歌曲标题很长的中文歌名\x1b[m")
		b.WriteByte('\n')
	}
	// 8 plain lyric rows + 1 empty row.
	for i := 0; i < 8; i++ {
		b.WriteString(strings.Repeat("歌词内容", 25))
		b.WriteByte('\n')
	}
	b.WriteString("status bar \x1b[38;2;1;2;3m 12:34\x1b[m")
	return b.String(), bg
}

func BenchmarkRenderAppBackgroundManual(b *testing.B) {
	content, bg := benchFrame()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = renderAppBackground(content, 200, bg, backgroundRect{})
	}
}

func BenchmarkRenderAppBackgroundLipgloss(b *testing.B) {
	content, bg := benchFrame()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = renderAppBackgroundLipgloss(content, 200, bg, backgroundRect{})
	}
}
