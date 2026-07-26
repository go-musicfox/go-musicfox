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

	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/structs"
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
		appColors = configs.GetCurrentAppColors()
		artistColor = configs.SafeGetForeground(appColors.PlaybarArtist, configs.PlaybarArtistColor)
		modeColor   = configs.SafeGetForeground(appColors.PlaybarMode, configs.PlaybarModeColor)
		volumeColor = configs.SafeGetForeground(appColors.PlaybarVolume, configs.PlaybarVolumeColor)
		playingColor = configs.SafeGetForeground(appColors.PlaybarPlaying, configs.PlaybarPlayingColor)
		pausedColor  = configs.SafeGetForeground(appColors.PlaybarPaused, configs.PlaybarPausedColor)
		heartLikedColor   = configs.SafeGetForeground(appColors.PlaybarHeartLiked, configs.PlaybarHeartLikedColor)
		heartUnlikedColor = configs.SafeGetForeground(appColors.PlaybarHeartUnliked, configs.PlaybarHeartUnlikedColor)
	)

	// Helper for adding a new segment
	addSegment := func(text string, color color.Color, underline, bold bool) {
		segments = append(segments, Segment{text, color, underline, bold})
	}
	// Helper for adding text whose color we don't care about
	addText := func(text string) {
		segments = append(segments, Segment{text, artistColor, false, false})
	}
	renderSegment := func(segment Segment) string {
		return lipgloss.NewStyle().
			Foreground(segment.color).
			Underline(segment.underline).
			Bold(segment.bold).
			Render(segment.text)
	}

	if main.MenuStartColumn()-MenuArrowWidth > 0 {
		if !main.CenterEverything() {
			addSegment(strings.Repeat(" ", main.MenuStartColumn()-MenuArrowWidth), artistColor, false, false)
		}
		{
			msg := r.state.Mode().Name()
			c := modeColor
			modeBold := hoveredElement == PlaybarElementMode
			if modeBold {
				c = util.GetPrimaryColor()
			}
			addSegment(fmt.Sprintf("[%s] ", msg), c, false, modeBold)
		}
		addSegment(fmt.Sprintf("%d%% ", r.state.Volume()), volumeColor, false, false)
	}
	if r.state.State() == types.Playing {
		c := playingColor
		stateBold := hoveredElement == PlaybarElementState
		if stateBold {
			c = util.GetPrimaryColor()
		}
		addSegment("♫ ♪ ♫ ♪ ", c, false, stateBold)
	} else {
		c := pausedColor
		stateBold := hoveredElement == PlaybarElementState
		if stateBold {
			c = util.GetPrimaryColor()
		}
		addSegment("_ z Z Z ", c, false, stateBold)
	}

	if song.Id > 0 {
		var icolor color.Color
		if isLike {
			icolor = heartLikedColor
		} else {
			icolor = heartUnlikedColor
		}
		heartBold := hoveredElement == PlaybarElementHeart
		if heartBold && !isLike {
			icolor = util.GetPrimaryColor()
		}
		addSegment("♥ ", icolor, false, heartBold)
	}

	if r.state.CurSongIndex() < len(r.state.Playlist()) {
		// 计算已占用的前缀宽度（实际 segment 宽度，而非估算的 prefixLen）
		prefixUsedWidth := 0
		for _, seg := range segments {
			prefixUsedWidth += runewidth.StringWidth(seg.text)
		}
		availableWidth := r.netease.WindowWidth()

		songName := song.Name
		artistString := artistsString(song)

		// 优先保留歌曲名，歌手名用剩余空间
		songMaxWidth := max(0, availableWidth-prefixUsedWidth-1) // -1 for space between
		songName = runewidth.Truncate(songName, songMaxWidth, "")
		addSegment(songName, util.GetPrimaryColor(), false, false)
		addText(" ")

		// 剩余空间全部给歌手名
		artistMaxWidth := max(0, availableWidth-prefixUsedWidth-runewidth.StringWidth(songName)-1)
		artistString = runewidth.Truncate(artistString, artistMaxWidth, "")
		c := artistColor
		artistHovered := hoveredElement == PlaybarElementArtist
		if artistHovered {
			c = util.GetPrimaryColor()
		}
		addSegment(artistString, c, false, false)
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

// artistsString joins the artists' names into a comma-separated string.
func artistsString(song structs.Song) string {
	if len(song.Artists) == 0 {
		return ""
	}
	var b strings.Builder
	for i, v := range song.Artists {
		if i != 0 {
			b.WriteString(",")
		}
		b.WriteString(v.Name)
	}
	return b.String()
}
