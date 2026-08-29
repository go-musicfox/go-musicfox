package ui

import (
	"strings"
	"testing"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/anhoder/foxful-cli/style"
	"github.com/charmbracelet/x/ansi"
	"github.com/mattn/go-runewidth"

	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/internal/types"
)

func TestSongInfoHoverDoesNotLeakANSISequences(t *testing.T) {
	app, shell := newFormPageTestApp(t)
	song := structs.Song{
		Id:   1,
		Name: "粗糙",
		Artists: []structs.Artist{
			{Name: "测试歌手"},
		},
	}
	state := songInfoTestState{song: song}
	renderer := NewSongInfoRenderer(newMenuServices(shell), state)

	tests := []struct {
		name    string
		hovered PlaybarElement
		want    string
	}{
		{name: "song name", hovered: PlaybarElementSongName, want: song.Name},
		{name: "artist", hovered: PlaybarElementArtist, want: song.Artists[0].Name},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shell.playbarHoveredElement = tt.hovered
			view, _ := renderer.View(app, app.MustMain())
			plain := ansi.Strip(view)
			if !strings.Contains(plain, tt.want) {
				t.Fatalf("rendered song info does not contain %q: %q", tt.want, plain)
			}
			if strings.Contains(plain, "[38;") || strings.Contains(plain, "[m") {
				t.Fatalf("hover leaked ANSI escape text: %q", plain)
			}
			textIndex := strings.Index(plain, tt.want)
			textColumn := runewidth.StringWidth(plain[:textIndex])
			cell := formStyledScreen(view).CellAt(textColumn, 0)
			if cell == nil {
				t.Fatal("hovered text cell is missing")
			}
			if cell.Style.Underline != ansi.UnderlineNone {
				t.Fatalf("hovered text underline = %v, want none", cell.Style.Underline)
			}
		})
	}
}

// TestSongInfoPaintsAppBackgroundOnGlyphs verifies that, when a theme sets an
// app background color, the song info text glyph cells carry that background
// (component→app→transparent chain) instead of leaving transparent cells that
// would reveal content drawn beneath the TUI (e.g. the cover image).
func TestSongInfoPaintsAppBackgroundOnGlyphs(t *testing.T) {
	app, shell := newFormPageTestApp(t)

	appBg := lipgloss.Color("#1E1E2E")
	theme := style.DefaultDarkTheme()
	theme.AppBackground = style.Highlight{Bg: appBg}
	style.SetStyleSet(style.NewStyleSet(theme))
	t.Cleanup(func() { style.SetStyleSet(style.DefaultStyleSet()) })

	song := structs.Song{
		Id:      1,
		Name:    "背景测试",
		Artists: []structs.Artist{{Name: "测试歌手"}},
	}
	renderer := NewSongInfoRenderer(newMenuServices(shell), songInfoTestState{song: song})
	view, _ := renderer.View(app, app.MustMain())

	plain := ansi.Strip(view)
	textIndex := strings.Index(plain, song.Name)
	if textIndex < 0 {
		t.Fatalf("rendered song info does not contain %q: %q", song.Name, plain)
	}
	textColumn := runewidth.StringWidth(plain[:textIndex])
	cell := formStyledScreen(view).CellAt(textColumn, 0)
	if cell == nil {
		t.Fatal("song name glyph cell is missing")
	}
	if cell.Style.Bg == nil {
		t.Fatalf("song name glyph cell has transparent background; " +
			"cover image would bleed through the text")
	}
	if !colorsEqual(cell.Style.Bg, appBg) {
		t.Fatalf("song name glyph background = %v, want app background %v", cell.Style.Bg, appBg)
	}
}

// TestSongInfoFillsEntireRowWithAppBackground verifies that the whole song-info
// content row — including the trailing padding after the artist name — carries
// the app background, so no cell at the right edge stays transparent.
func TestSongInfoFillsEntireRowWithAppBackground(t *testing.T) {
	app, shell := newFormPageTestApp(t)

	appBg := lipgloss.Color("#1E1E2E")
	theme := style.DefaultDarkTheme()
	theme.AppBackground = style.Highlight{Bg: appBg}
	style.SetStyleSet(style.NewStyleSet(theme))
	t.Cleanup(func() { style.SetStyleSet(style.DefaultStyleSet()) })

	song := structs.Song{
		Name:    "测试歌曲",
		Artists: []structs.Artist{{Name: "歌手"}},
	}
	renderer := NewSongInfoRenderer(newMenuServices(shell), songInfoTestState{song: song})
	view, _ := renderer.View(app, app.MustMain())

	// Inspect only the content row (row 0); row 1 is the blank separator.
	line := strings.Split(view, "\n")[0]
	screen := formStyledScreen(line)
	width := shell.WindowWidth()
	for x := 0; x < width; x++ {
		cell := screen.CellAt(x, 0)
		if cell == nil || cell.Width == 0 {
			continue
		}
		if cell.Style.Bg == nil {
			t.Fatalf("song-info cell (%d,0) content=%q has transparent background; "+
				"cover image would bleed through the trailing padding", x, cell.Content)
		}
	}
}

// TestProgressTimePaintsAppBackground verifies that the progress-bar row —
// including the time display and the separator spaces around it — carries the
// app background across its full width, leaving no transparent cell.
func TestProgressTimePaintsAppBackground(t *testing.T) {
	app, shell := newFormPageTestApp(t)

	appBg := lipgloss.Color("#1E1E2E")
	theme := style.DefaultDarkTheme()
	theme.AppBackground = style.Highlight{Bg: appBg}
	style.SetStyleSet(style.NewStyleSet(theme))
	t.Cleanup(func() { style.SetStyleSet(style.DefaultStyleSet()) })

	song := structs.Song{
		Name:     "测试歌曲",
		Duration: 4*time.Minute + 30*time.Second,
	}
	renderer := NewProgressRenderer(newMenuServices(shell), progressTestState{song: song, passed: time.Minute})
	view, _ := renderer.View(app, app.MustMain())

	plain := ansi.Strip(view)
	// The time text must be present, else the test would not exercise it.
	if !strings.Contains(plain, "01:00/04:30") {
		t.Fatalf("progress view missing expected time text: %q", plain)
	}

	screen := formStyledScreen(view)
	width := shell.WindowWidth()
	for x := 0; x < width; x++ {
		cell := screen.CellAt(x, 0)
		if cell == nil || cell.Width == 0 {
			continue
		}
		if cell.Style.Bg == nil {
			t.Fatalf("progress cell (%d,0) content=%q has transparent background; "+
				"cover image would bleed through the progress row", x, cell.Content)
		}
	}
}

type songInfoTestState struct {
	song structs.Song
}

func (s songInfoTestState) CurSong() structs.Song     { return s.song }
func (s songInfoTestState) CurSongIndex() int         { return 0 }
func (s songInfoTestState) PassedTime() time.Duration { return 0 }
func (s songInfoTestState) State() types.State        { return types.Paused }
func (s songInfoTestState) Volume() int               { return 50 }
func (s songInfoTestState) Mode() types.Mode          { return types.PmListLoop }
func (s songInfoTestState) Playlist() []structs.Song  { return []structs.Song{s.song} }

type progressTestState struct {
	song   structs.Song
	passed time.Duration
}

func (s progressTestState) CurSong() structs.Song     { return s.song }
func (s progressTestState) CurSongIndex() int         { return 0 }
func (s progressTestState) PassedTime() time.Duration { return s.passed }
func (s progressTestState) State() types.State        { return types.Playing }
func (s progressTestState) Volume() int               { return 50 }
func (s progressTestState) Mode() types.Mode          { return types.PmListLoop }
func (s progressTestState) Playlist() []structs.Song  { return []structs.Song{s.song} }
