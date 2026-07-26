package ui

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
	"github.com/mattn/go-runewidth"

	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/internal/types"
)

func TestSongInfoHoverDoesNotLeakANSISequences(t *testing.T) {
	app, netease := newFormPageTestApp(t)
	song := structs.Song{
		Id:   1,
		Name: "粗糙",
		Artists: []structs.Artist{
			{Name: "测试歌手"},
		},
	}
	state := songInfoTestState{song: song}
	renderer := NewSongInfoRenderer(netease, state)

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
			netease.playbarHoveredElement = tt.hovered
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
