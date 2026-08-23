package ui

import (
	"fmt"
	"image/color"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/anhoder/foxful-cli/model"
	"github.com/anhoder/foxful-cli/style"
	"github.com/anhoder/foxful-cli/util"
	"github.com/mattn/go-runewidth"

	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/internal/types"
	"github.com/go-musicfox/go-musicfox/utils/likelist"
)

// SongInfoRenderer is a dedicated UI component for rendering the current song's information.
type SongInfoRenderer struct {
	svc   *menuServices
	state playerRendererState

	cachedView     string
	cachedLines    int
	cachedSongID   int64
	cachedState    types.State
	cachedVolume   int
	cachedMode     types.Mode
	cachedLike     bool
	cachedWidth    int
	cachedCentered bool
	cachedHover    PlaybarElement
	cachedStyleGen uint64
}

// NewSongInfoRenderer creates a new song info renderer component.
func NewSongInfoRenderer(svc *menuServices, state playerRendererState) *SongInfoRenderer {
	return &SongInfoRenderer{
		svc:   svc,
		state: state,
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
		width          = r.svc.App().WindowWidth()
		centered       = main.CenterEverything()
		hoveredElement = r.svc.PlaybarHoveredElement()
	)

	var isLike bool
	if song.Id > 0 {
		isLike = likelist.IsLikeSong(song.Id)
	}

	// Output caching: skip full rebuild when nothing changed. styleGen guards
	// against replaying stale-colored output after a theme switch.
	styleGen := style.StyleGeneration()
	if song.Id == r.cachedSongID && state == r.cachedState && volume == r.cachedVolume &&
		mode == r.cachedMode && isLike == r.cachedLike && width == r.cachedWidth &&
		centered == r.cachedCentered && hoveredElement == r.cachedHover &&
		styleGen == r.cachedStyleGen {
		return r.cachedView, r.cachedLines
	}

	var (
		builder           strings.Builder
		segments          []Segment
		appColors         = configs.GetCurrentAppColors()
		artistColor       = configs.SafeGetForeground(appColors.PlaybarArtist, configs.PlaybarArtistColor)
		modeColor         = configs.SafeGetForeground(appColors.PlaybarMode, configs.PlaybarModeColor)
		volumeColor       = configs.SafeGetForeground(appColors.PlaybarVolume, configs.PlaybarVolumeColor)
		playingColor      = configs.SafeGetForeground(appColors.PlaybarPlaying, configs.PlaybarPlayingColor)
		pausedColor       = configs.SafeGetForeground(appColors.PlaybarPaused, configs.PlaybarPausedColor)
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
	// Resolve the background once: the app background paints text glyph cells so
	// they never reveal content drawn beneath the TUI. nil leaves it transparent.
	segmentBg := configs.ResolveBackground(nil)
	renderSegment := func(segment Segment) string {
		s := lipgloss.NewStyle().
			Foreground(segment.color).
			Underline(segment.underline).
			Bold(segment.bold)
		if segmentBg != nil {
			s = s.Background(segmentBg)
		}
		return s.Render(segment.text)
	}

	if main.MenuStartColumn()-MenuArrowWidth > 0 {
		if !main.CenterEverything() {
			// Store plain spaces (not pre-rendered); renderSegment paints them with
			// segmentBg. Pre-rendering here would make runewidth.StringWidth count the
			// ANSI escape bytes, inflating totalWidth and shrinking the trailing fill.
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
		availableWidth := r.svc.App().WindowWidth()

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
		widthLimit := r.svc.App().WindowWidth() - SongInfoHorizontalPadding
		for index, segment := range segments {
			segmentWidth := runewidth.StringWidth(segment.text)
			if totalWidth+segmentWidth > widthLimit {
				segmentWidth = max(0, widthLimit-totalWidth)
				segments[index].text = runewidth.Truncate(segment.text, segmentWidth, "")
			}
			totalWidth += segmentWidth
		}
		paddingLeft := (r.svc.App().WindowWidth() - totalWidth) / 2
		builder.WriteString(style.CurrentStyleSet().AppBackground.Render(strings.Repeat(" ", paddingLeft)))
		for _, segment := range segments {
			builder.WriteString(renderSegment(segment))
		}
		builder.WriteString(style.CurrentStyleSet().AppBackground.Render(strings.Repeat(" ", r.svc.App().WindowWidth()-paddingLeft-totalWidth)))
	} else {
		// Left-aligned: concatenate segments, then fill trailing space to window width.
		totalWidth := 0
		for _, segment := range segments {
			builder.WriteString(renderSegment(segment))
			totalWidth += runewidth.StringWidth(segment.text)
		}
		remainingWidth := r.svc.App().WindowWidth() - totalWidth
		if remainingWidth > 0 {
			builder.WriteString(style.CurrentStyleSet().AppBackground.Render(strings.Repeat(" ", remainingWidth)))
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
	r.cachedSongID = song.Id
	r.cachedState = state
	r.cachedVolume = volume
	r.cachedMode = mode
	r.cachedLike = isLike
	r.cachedWidth = width
	r.cachedCentered = centered
	r.cachedHover = hoveredElement
	r.cachedStyleGen = styleGen

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
