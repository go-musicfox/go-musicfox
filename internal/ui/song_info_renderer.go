package ui

import (
	"fmt"
	"image/color"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/anhoder/foxful-cli/model"
	"github.com/anhoder/foxful-cli/util"
	"github.com/mattn/go-runewidth"

	"github.com/go-musicfox/go-musicfox/internal/types"
	"github.com/go-musicfox/go-musicfox/utils/likelist"
)

// SongInfoRenderer is a dedicated UI component for rendering the current song's information.
type SongInfoRenderer struct {
	netease *Netease
	state   playerRendererState

	cachedView     string
	cachedLines    int
	cachedSongId   int64
	cachedState    types.State
	cachedVolume   int
	cachedMode     types.Mode
	cachedLike     bool
	cachedWidth    int
	cachedCentered bool
	cachedHover    PlaybarElement
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
		text      string
		color     color.Color
		underline bool
		bold      bool
	}

	var (
		song           = r.state.CurSong()
		state          = r.state.State()
		volume         = r.state.Volume()
		mode           = r.state.Mode()
		width          = r.netease.WindowWidth()
		centered       = main.CenterEverything()
		hoveredElement = r.netease.playbarHoveredElement
	)

	var isLike bool
	if song.Id > 0 {
		isLike = likelist.IsLikeSong(song.Id)
	}

	// Output caching: skip full rebuild when nothing changed
	if song.Id == r.cachedSongId && state == r.cachedState && volume == r.cachedVolume &&
		mode == r.cachedMode && isLike == r.cachedLike && width == r.cachedWidth &&
		centered == r.cachedCentered && hoveredElement == r.cachedHover {
		return r.cachedView, r.cachedLines
	}

	var (
		builder  strings.Builder
		segments []Segment
	)

	// Helper for adding a new segment
	addSegment := func(text string, color color.Color, underline, bold bool) {
		segments = append(segments, Segment{text, color, underline, bold})
	}
	// Helper for adding text whose color we don't care about
	addText := func(text string) {
		segments = append(segments, Segment{text, lipgloss.BrightBlack, false, false})
	}
	renderSegment := func(segment Segment) string {
		return lipgloss.NewStyle().
			Foreground(segment.color).
			Underline(segment.underline).
			Bold(segment.bold).
			Render(segment.text)
	}

	prefixLen := SongInfoPrefixBaseWidth
	if main.MenuStartColumn()-MenuArrowWidth > 0 {
		prefixLen += SongInfoPrefixExtraWidth
		if !main.CenterEverything() {
			addSegment(strings.Repeat(" ", main.MenuStartColumn()-MenuArrowWidth), lipgloss.BrightBlack, false, false)
		}
		{
			msg := r.state.Mode().Name()
			modeColor := color.Color(lipgloss.BrightMagenta)
			modeBold := hoveredElement == PlaybarElementMode
			if modeBold {
				modeColor = util.GetPrimaryColor()
			}
			addSegment(fmt.Sprintf("[%s] ", msg), modeColor, false, modeBold)
		}
		addSegment(fmt.Sprintf("%d%% ", r.state.Volume()), lipgloss.BrightBlue, false, false)
	}
	if r.state.State() == types.Playing {
		stateColor := color.Color(lipgloss.BrightYellow)
		stateBold := hoveredElement == PlaybarElementState
		if stateBold {
			stateColor = util.GetPrimaryColor()
		}
		addSegment("♫ ♪ ♫ ♪ ", stateColor, false, stateBold)
	} else {
		stateColor := color.Color(lipgloss.Yellow)
		stateBold := hoveredElement == PlaybarElementState
		if stateBold {
			stateColor = util.GetPrimaryColor()
		}
		addSegment("_ z Z Z ", stateColor, false, stateBold)
	}

	if song.Id > 0 {
		var icolor color.Color
		if isLike {
			icolor = lipgloss.Red
		} else {
			icolor = lipgloss.White
		}
		heartBold := hoveredElement == PlaybarElementHeart
		if heartBold && !isLike {
			icolor = util.GetPrimaryColor()
		}
		addSegment("♥ ", icolor, false, heartBold)
	}

	if r.state.CurSongIndex() < len(r.state.Playlist()) {
		// 按剩余长度截断字符串
		songName := song.Name
		if !main.CenterEverything() {
			songName = runewidth.Truncate(songName, r.netease.WindowWidth()-main.MenuStartColumn()-prefixLen, "") // 多减，避免剩余1个中文字符
		}
		addSegment(songName, util.GetPrimaryColor(), false, false)
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
		artistColor := color.Color(lipgloss.BrightBlack)
		artistHovered := hoveredElement == PlaybarElementArtist
		if artistHovered {
			artistColor = util.GetPrimaryColor()
		}
		addSegment(artistString, artistColor, false, false)
	}

	if main.CenterEverything() {
		totalWidth := 0
		widthLimit := r.netease.WindowWidth() - SongInfoHorizontalPadding
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
			builder.WriteString(renderSegment(segment))
		}
		builder.WriteString(strings.Repeat(" ", r.netease.WindowWidth()-paddingLeft-totalWidth))
	} else {
		// simply concatenate every segment with the specified color
		for _, segment := range segments {
			builder.WriteString(renderSegment(segment))
		}
	}

	// Return the actual number of lines: 1 content line + 1 blank line separator.
	// In the new JoinVertical model, the returned `lines` value is discarded (no longer
	// used to advance a `top` pointer). Only the actual string content matters for layout.
	// Output exactly 2 visual rows: content + single trailing newline = 1 blank separator.
	lines = SongInfoLines

	// Store output cache
	r.cachedView = builder.String() + "\n"
	r.cachedLines = lines
	r.cachedSongId = song.Id
	r.cachedState = state
	r.cachedVolume = volume
	r.cachedMode = mode
	r.cachedLike = isLike
	r.cachedWidth = width
	r.cachedCentered = centered
	r.cachedHover = hoveredElement

	return r.cachedView, r.cachedLines
}
