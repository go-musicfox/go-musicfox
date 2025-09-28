package ui

import (
	"fmt"
	"strings"

	"github.com/anhoder/foxful-cli/model"
	"github.com/anhoder/foxful-cli/util"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mattn/go-runewidth"
	"github.com/muesli/termenv"

	"github.com/go-musicfox/go-musicfox/internal/types"
	"github.com/go-musicfox/go-musicfox/utils/likelist"
)

// SongInfoRenderer is a dedicated UI component for rendering the current song's information.
type SongInfoRenderer struct {
	netease *Netease
	state   playerRendererState
}

// NewSongInfoRenderer creates a new song info renderer component.
func NewSongInfoRenderer(netease *Netease, state playerRendererState) *SongInfoRenderer {
	return &SongInfoRenderer{
		netease: netease,
		state:   state,
	}
}

// Update handles UI messages.
func (r *SongInfoRenderer) Update(msg tea.Msg, a *model.App) {}

// View renders the song information component.
func (r *SongInfoRenderer) View(a *model.App, main *model.Main) (view string, lines int) {
	// Every part of the song view is expressed as a segment: unformatted text followed by a color specification
	// This makes computing the total length of the song view easier
	type Segment struct {
		text  string
		color termenv.Color
	}

	var (
		builder  strings.Builder
		segments []Segment
		song     = r.state.CurSong()
	)

	// Helper for adding a new segment
	addSegment := func(text string, color termenv.Color) {
		segments = append(segments, Segment{text, color})
	}
	// Helper for adding text whose color we don't care about
	addText := func(text string) {
		segments = append(segments, Segment{text, termenv.ANSIBrightBlack})
	}

	prefixLen := 10
	if main.MenuStartColumn()-4 > 0 {
		prefixLen += 12
		if !main.CenterEverything() {
			addSegment(strings.Repeat(" ", main.MenuStartColumn()-4), termenv.ANSIBrightBlack)
		}
		{
			msg := r.state.Mode().Name()
			addSegment(fmt.Sprintf("[%s] ", msg), termenv.ANSIBrightMagenta)
		}
		addSegment(fmt.Sprintf("%d%% ", r.state.Volume()), termenv.ANSIBrightBlue)
	}
	if r.state.State() == types.Playing {
		addSegment("♫ ♪ ♫ ♪ ", termenv.ANSIBrightYellow)
	} else {
		addSegment("_ z Z Z ", termenv.ANSIYellow)
	}

	if song.Id > 0 {
		var color termenv.ANSIColor
		if likelist.IsLikeSong(song.Id) {
			color = termenv.ANSIRed
		} else {
			color = termenv.ANSIWhite
		}
		addSegment("♥ ", color)
	}

	if r.state.CurSongIndex() < len(r.state.Playlist()) {
		// 按剩余长度截断字符串
		songName := song.Name
		if !main.CenterEverything() {
			songName = runewidth.Truncate(songName, r.netease.WindowWidth()-main.MenuStartColumn()-prefixLen, "") // 多减，避免剩余1个中文字符
		}
		addSegment(songName, util.GetPrimaryColor())
		addText(" ")

		var artists strings.Builder
		for i, v := range song.Artists {
			if i != 0 {
				artists.WriteString(",")
			}
			artists.WriteString(v.Name)
		}

		artistString := artists.String()
		if !main.CenterEverything() {
			// 按剩余长度截断字符串
			remainLen := r.netease.WindowWidth() - main.MenuStartColumn() - prefixLen - runewidth.StringWidth(song.Name)
			artistString = runewidth.Truncate(
				runewidth.FillRight(artistString, remainLen),
				remainLen, "")
		}
		addSegment(artistString, termenv.ANSIBrightBlack)
	}

	if main.CenterEverything() {
		totalWidth := 0
		widthLimit := r.netease.WindowWidth() - 4
		for index, segment := range segments {
			segmentWidth := runewidth.StringWidth(segment.text)
			if totalWidth+segmentWidth > widthLimit {
				segmentWidth = max(0, widthLimit-totalWidth)
				segments[index].text = runewidth.Truncate(segment.text, segmentWidth, "")
			}
			totalWidth += segmentWidth
		}
		paddingLeft := (r.netease.WindowWidth() - totalWidth) / 2
		builder.WriteString(strings.Repeat(" ", paddingLeft))
		for _, segment := range segments {
			builder.WriteString(util.SetFgStyle(segment.text, segment.color))
		}
		builder.WriteString(strings.Repeat(" ", r.netease.WindowWidth()-paddingLeft-totalWidth))
	} else {
		// simply concatenate every segment with the specified color
		for _, segment := range segments {
			builder.WriteString(util.SetFgStyle(segment.text, segment.color))
		}
	}

	return builder.String(), 1
}
