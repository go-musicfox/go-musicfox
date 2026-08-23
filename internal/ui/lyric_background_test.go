package ui

import (
	"image/color"
	"strings"
	"testing"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/anhoder/foxful-cli/style"
	"github.com/charmbracelet/x/ansi"

	"github.com/go-musicfox/go-musicfox/internal/lyric"
	"github.com/go-musicfox/go-musicfox/internal/structs"
	appui "github.com/go-musicfox/go-musicfox/utils/app"
)

// withAppBackgroundTheme installs a theme whose app background is set, so glyph
// cells that fail to paint a background show up as transparent in the scan.
func withAppBackgroundTheme(t *testing.T, appBg color.Color) {
	t.Helper()
	theme := style.DefaultDarkTheme()
	theme.AppBackground = style.Highlight{Bg: appBg}
	style.SetStyleSet(style.NewStyleSet(theme))
	t.Cleanup(func() { style.SetStyleSet(style.DefaultStyleSet()) })
}

func assertLyricRowsLeaveLeadingPaddingTransparent(t *testing.T, view string, width int) {
	t.Helper()
	for row, line := range strings.Split(view, "\n") {
		plain := ansi.Strip(line)
		if strings.TrimSpace(plain) == "" {
			continue
		}
		leadingPadding := len(plain) - len(strings.TrimLeft(plain, " "))
		screen := formStyledScreen(line)
		for x := range width {
			cell := screen.CellAt(x, 0)
			if cell == nil || cell.Width == 0 {
				continue
			}
			if x < leadingPadding && cell.Style.Bg != nil {
				t.Fatalf("lyric padding cell (%d,%d) has background %v, want transparent", x, row, cell.Style.Bg)
			}
			if x >= leadingPadding && cell.Style.Bg == nil {
				t.Fatalf("lyric content cell (%d,%d) content=%q has transparent background", x, row, cell.Content)
			}
		}
	}
}

// TestLyricTextLeavesCoverPaddingTransparent verifies that the leading lyric
// padding remains transparent for the cover while lyric glyphs and trailing
// fill still carry the app background.
func TestLyricTextLeavesCoverPaddingTransparent(t *testing.T) {
	app, netease := newFormPageTestApp(t)
	appBg := lipgloss.Color("#1E1E2E")
	withAppBackgroundTheme(t, appBg)

	r := &LyricRenderer{svc: newMenuServices(netease), lyricLines: 3, lyricNowScrollBar: appui.NewXScrollBar()}
	r.lyrics = []string{"上一句歌词", "当前高亮歌词", "下一句歌词"}

	var b strings.Builder
	r.buildLyricsTraditional(app.MustMain(), &b)
	view := b.String()

	const marker = "当前高亮歌词"
	plain := ansi.Strip(view)
	if !strings.Contains(plain, marker) {
		t.Fatalf("lyric view missing highlighted line %q: %q", marker, plain)
	}

	assertLyricRowsLeaveLeadingPaddingTransparent(t, view, netease.WindowWidth())
}

// TestLyricTraditionalFillsFullWidth guards the trailing fill of plain (non-ANSI)
// lyric lines: each rendered row must span the full window width. Regression
// guard for the right-edge margin staying unpainted because the trailing-fill
// width was measured from the ANSI-wrapped line (escape bytes inflated the
// count, driving remainingWidth negative). The cell scan misses this because
// the unpainted cells simply do not exist — the row is shorter than the window.
func TestLyricTraditionalFillsFullWidth(t *testing.T) {
	app, netease := newFormPageTestApp(t)
	withAppBackgroundTheme(t, lipgloss.Color("#1E1E2E"))

	r := &LyricRenderer{svc: newMenuServices(netease), lyricLines: 3, lyricNowScrollBar: appui.NewXScrollBar()}
	// Plain strings mirror the real "no lyric" placeholders (作曲：… / 纯音乐，请欣赏).
	r.lyrics = []string{"作曲：和乐府", "纯音乐，请欣赏", ""}

	var b strings.Builder
	r.buildLyricsTraditional(app.MustMain(), &b)

	width := netease.WindowWidth()
	for row, line := range strings.Split(b.String(), "\n") {
		plain := ansi.Strip(line)
		if strings.TrimSpace(plain) == "" && row >= r.lyricLines {
			continue // trailing newline
		}
		if got := lipgloss.Width(plain); got != width {
			t.Fatalf("lyric row %d painted width = %d, want full window width %d; "+
				"right-edge margin unpainted, cover bleeds through", row, got, width)
		}
	}
}

// TestLyricCenteredTextLeavesCoverPaddingTransparent is the centered-layout counterpart.
func TestLyricCenteredTextLeavesCoverPaddingTransparent(t *testing.T) {
	app, netease := newFormPageTestApp(t)
	appBg := lipgloss.Color("#1E1E2E")
	withAppBackgroundTheme(t, appBg)

	main := app.MustMain()

	r := &LyricRenderer{svc: newMenuServices(netease), lyricLines: 3, lyricNowScrollBar: appui.NewXScrollBar()}
	r.lyrics = []string{"上一句", "居中高亮歌词", "下一句"}

	var b strings.Builder
	r.buildLyricsCentered(main, &b)
	view := b.String()

	const marker = "居中高亮歌词"
	if !strings.Contains(ansi.Strip(view), marker) {
		t.Fatalf("centered lyric view missing %q: %q", marker, ansi.Strip(view))
	}

	assertLyricRowsLeaveLeadingPaddingTransparent(t, view, netease.WindowWidth())
}

// TestLyricYRCLineLeavesCoverPaddingTransparent guards the word-by-word path:
// leading padding stays transparent for the cover, while lyrics and translations
// retain their app-background cells.
func TestLyricYRCLineLeavesCoverPaddingTransparent(t *testing.T) {
	app, netease := newFormPageTestApp(t)
	appBg := lipgloss.Color("#1E1E2E")
	withAppBackgroundTheme(t, appBg)

	r := &LyricRenderer{svc: newMenuServices(netease), lyricLines: 3, lyricNowScrollBar: appui.NewXScrollBar()}
	current := lyric.YRCLine{StartTime: 0, EndTime: 3000, Words: []lyric.YRCWord{
		{Word: "当前", StartTime: 0, EndTime: 1000},
		{Word: "歌词", StartTime: 1000, EndTime: 2000},
	}, TranslatedLyric: "translated"}
	sibling := lyric.YRCLine{StartTime: 0, EndTime: 1, Words: []lyric.YRCWord{
		{Word: "相邻歌词"},
	}, TranslatedLyric: "sibling-trans"}
	r.lyrics = []string{
		r.buildYRCLineString(sibling, -1, true),   // previous (inactive)
		r.buildYRCLineString(current, 1500, true), // current (highlighted + translation)
		r.buildYRCLineString(sibling, -1, true),   // next (inactive)
	}

	var b strings.Builder
	r.buildLyricsTraditional(app.MustMain(), &b)
	view := b.String()

	if !strings.Contains(ansi.Strip(view), "translated") {
		t.Fatalf("YRC view missing current-line translation: %q", ansi.Strip(view))
	}

	assertLyricRowsLeaveLeadingPaddingTransparent(t, view, netease.WindowWidth())
}

// TestThemeSwitchInvalidatesRendererCaches verifies that switching the global
// StyleSet (theme switch) invalidates the cached output of the player
// renderers, so the next frame repaints with the new theme's colors instead of
// replaying stale-colored cache entries ("清空原有渲染内容").
func TestThemeSwitchInvalidatesRendererCaches(t *testing.T) {
	app, netease := newFormPageTestApp(t)

	firstBg := lipgloss.Color("#1E1E2E")
	theme := style.DefaultDarkTheme()
	theme.AppBackground = style.Highlight{Bg: firstBg}
	style.SetStyleSet(style.NewStyleSet(theme))
	t.Cleanup(func() { style.SetStyleSet(style.DefaultStyleSet()) })

	song := structs.Song{Name: "缓存歌曲", Duration: time.Minute}

	songRenderer := NewSongInfoRenderer(newMenuServices(netease), songInfoTestState{song: song})
	progRenderer := NewProgressRenderer(newMenuServices(netease), progressTestState{song: song, passed: time.Minute})

	main := app.MustMain()
	songView1, _ := songRenderer.View(app, main)
	progView1, _ := progRenderer.View(app, main)

	// Switch to a theme with a distinct app background.
	secondBg := lipgloss.Color("#3A2E1E")
	theme2 := style.DefaultDarkTheme()
	theme2.AppBackground = style.Highlight{Bg: secondBg}
	style.SetStyleSet(style.NewStyleSet(theme2))

	songView2, _ := songRenderer.View(app, main)
	progView2, _ := progRenderer.View(app, main)

	if songView1 == songView2 {
		t.Fatal("song-info view unchanged after theme switch; stale cache replayed old colors")
	}
	if progView1 == progView2 {
		t.Fatal("progress view unchanged after theme switch; stale cache replayed old colors")
	}

	// The repainted frames must carry the NEW background, not the old one.
	assertRowBackground(t, netease, songView2, secondBg, "song-info")
	assertRowBackground(t, netease, progView2, secondBg, "progress")
}

func assertRowBackground(t *testing.T, netease *Netease, view string, want color.Color, label string) {
	t.Helper()
	line := strings.Split(view, "\n")[0]
	screen := formStyledScreen(line)
	width := netease.WindowWidth()
	for x := 0; x < width; x++ {
		cell := screen.CellAt(x, 0)
		if cell == nil || cell.Width == 0 || cell.Style.Bg == nil {
			continue
		}
		if colorsEqual(cell.Style.Bg, want) {
			return
		}
	}
	t.Fatalf("%s row carries no cell with the new theme background %v", label, want)
}
