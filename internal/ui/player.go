package ui

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/anhoder/foxful-cli/model"
	"github.com/anhoder/foxful-cli/util"
	"github.com/buger/jsonparser"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-musicfox/netease-music/service"
	"github.com/mattn/go-runewidth"
	"github.com/muesli/termenv"

	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/lastfm"
	"github.com/go-musicfox/go-musicfox/internal/lyric"
	"github.com/go-musicfox/go-musicfox/internal/player"
	control "github.com/go-musicfox/go-musicfox/internal/remote_control"
	"github.com/go-musicfox/go-musicfox/internal/storage"
	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/internal/types"
	"github.com/go-musicfox/go-musicfox/utils/app"
	"github.com/go-musicfox/go-musicfox/utils/errorx"
	"github.com/go-musicfox/go-musicfox/utils/likelist"
	"github.com/go-musicfox/go-musicfox/utils/netease"
	"github.com/go-musicfox/go-musicfox/utils/notify"
	"github.com/go-musicfox/go-musicfox/utils/storagex"
	_struct "github.com/go-musicfox/go-musicfox/utils/struct"
)

// PlayDirection 下首歌的方向
type PlayDirection uint8

const (
	DurationNext PlayDirection = iota
	DurationPrev
)

type CtrlType string

type CtrlSignal struct {
	Type     CtrlType
	Duration time.Duration
}

const (
	CtrlResume   CtrlType = "Resume"
	CtrlPaused   CtrlType = "Paused"
	CtrlStop     CtrlType = "Stop"
	CtrlToggle   CtrlType = "Toggle"
	CtrlPrevious CtrlType = "Previous"
	CtrlNext     CtrlType = "Next"
	CtrlSeek     CtrlType = "Seek"
	CtrlRerender CtrlType = "Rerender"
)

// Player 网易云音乐播放器
type Player struct {
	netease *Netease
	cancel  context.CancelFunc

	playlistUpdateAt time.Time   // 播放列表更新时间
	songManager      songManager // 歌曲管理器
	// TODO: 将心动模式单独抽象出来，减少耦合
	intelligent    bool // 是否处于心动模式
	lastMode       types.Mode
	playingMenuKey string // 正在播放的菜单Key
	playingMenu    Menu

	lrcFile           *lyric.LRCFile
	transLrcFile      *lyric.TranslateLRCFile
	lrcTimer          *lyric.LRCTimer // 歌词计时器
	lyrics            [5]string       // 歌词信息，保留5行
	showLyric         bool            // 显示歌词
	lyricStartRow     int             // 歌词开始行
	lyricLines        int             // 歌词显示行数，3或5
	lyricNowScrollBar *app.XScrollBar // 当前歌词滚动

	// 播放进度条
	progressLastWidth float64
	progressRamp      []string

	playErrCount int // 错误计数，当错误连续超过5次，停止播放
	stateHandler *control.RemoteControl
	ctrl         chan CtrlSignal

	renderTicker *tickerByPlayer // renderTicker 用于渲染

	player.Player // 播放器
}

func NewPlayer(n *Netease) *Player {
	p := &Player{
		netease:           n,
		songManager:       &listLoopSongManager{},
		ctrl:              make(chan CtrlSignal),
		lyricNowScrollBar: app.NewXScrollBar(),
	}
	var ctx context.Context
	ctx, p.cancel = context.WithCancel(context.Background())

	p.Player = player.NewPlayerFromConfig()
	p.stateHandler = control.NewRemoteControl(p, p.PlayingInfo())

	p.renderTicker = newTickerByPlayer(p)

	// remote control
	errorx.WaitGoStart(func() {
		for {
			select {
			case <-ctx.Done():
				return
			case signal := <-p.ctrl:
				p.handleControlSignal(signal)
			}
		}
	})

	// 状态监听
	errorx.WaitGoStart(func() {
		for {
			select {
			case <-ctx.Done():
				return
			case s := <-p.StateChan():
				p.stateHandler.SetPlayingInfo(p.PlayingInfo())
				if s != types.Stopped {
					p.netease.Rerender(false)
					break
				}
				p.NextSong(false)
			}
		}
	})

	// 时间监听
	errorx.WaitGoStart(func() {
		for {
			select {
			case <-ctx.Done():
				return
			case duration := <-p.TimeChan():
				p.stateHandler.SetPosition(p.PassedTime())
				if duration.Seconds()-p.CurMusic().Duration.Seconds() > 10 {
					p.NextSong(false)
				}
				if p.lrcTimer != nil {
					select {
					case p.lrcTimer.Timer() <- duration + time.Millisecond*time.Duration(configs.ConfigRegistry.Main.LyricOffset):
					default:
					}
				}
				if p.renderTicker != nil {
					select {
					case p.renderTicker.c <- time.Now():
					default:
					}
				}

				p.netease.Rerender(false)
			}
		}
	})

	return p
}

func (p *Player) Update(_ tea.Msg, _ *model.App) {
	main := p.netease.MustMain()
	// 播放器歌词
	spaceHeight := p.netease.WindowHeight() - 5 - main.MenuBottomRow()
	if spaceHeight < 3 || !configs.ConfigRegistry.Main.ShowLyric {
		// 不显示歌词
		p.showLyric = false
	} else {
		p.showLyric = true
		if spaceHeight >= 5 {
			// 5行歌词
			p.lyricStartRow = (p.netease.WindowHeight()-3+main.MenuBottomRow())/2 - 3
			p.lyricLines = 5
		} else {
			// 3行歌词
			p.lyricStartRow = (p.netease.WindowHeight()-3+main.MenuBottomRow())/2 - 2
			p.lyricLines = 3
		}
	}
}

func (p *Player) View(a *model.App, main *model.Main) (view string, lines int) {
	var playerBuilder strings.Builder
	playerBuilder.WriteString(p.lyricView())
	playerBuilder.WriteString(p.songView())
	playerBuilder.WriteString("\n\n")
	playerBuilder.WriteString(p.progressView())
	return playerBuilder.String(), a.WindowHeight() - main.MenuBottomRow()
}

// lyricView 歌词显示UI
func (p *Player) lyricView() string {
	var (
		endRow = p.netease.WindowHeight() - 4
		main   = p.netease.MustMain()
	)

	if !p.showLyric {
		if endRow-main.MenuBottomRow() > 0 {
			return strings.Repeat("\n", endRow-main.MenuBottomRow())
		} else {
			return ""
		}
	}

	var lyricBuilder strings.Builder
	if p.lyricStartRow > main.MenuBottomRow() {
		lyricBuilder.WriteString(strings.Repeat("\n", p.lyricStartRow-main.MenuBottomRow()))
	}

	if main.CenterEverything() {
		p.buildLyricsCentered(main, &lyricBuilder)
	} else {
		p.buildLyricsTraditional(main, &lyricBuilder)
	}

	if endRow-p.lyricStartRow-p.lyricLines > 0 {
		lyricBuilder.WriteString(strings.Repeat("\n", endRow-p.lyricStartRow-p.lyricLines))
	}

	return lyricBuilder.String()
}

func (p *Player) buildLyricsCentered(main *model.Main, lyricBuilder *strings.Builder) {
	windowWidth := p.netease.WindowWidth()
	highlightLine := 2
	startLine := highlightLine - (p.lyricLines-1)/2
	endLine := highlightLine + (p.lyricLines-1)/2
	extraPadding := 8 + max(0, (windowWidth-40)/5)
	lyricsMaxLength := windowWidth - extraPadding
	for i := startLine; i <= endLine; i++ {
		line := p.lyrics[i]
		if i == highlightLine {
			line = p.lyricNowScrollBar.Tick(lyricsMaxLength, line)
			line = strings.Trim(line, " ")
		}
		line = runewidth.Truncate(line, lyricsMaxLength, "")
		lineLength := runewidth.StringWidth(line)
		paddingLeft := (windowWidth - lineLength) / 2
		lyricBuilder.WriteString(strings.Repeat(" ", paddingLeft))
		if i == highlightLine {
			line = util.SetFgStyle(line, termenv.ANSIBrightCyan)
		} else {
			line = util.SetFgStyle(line, termenv.ANSIBrightBlack)
		}
		lyricBuilder.WriteString(line)
		lyricBuilder.WriteString(strings.Repeat(" ", windowWidth-paddingLeft-lineLength))
		lyricBuilder.WriteString("\n")
	}
}

func (p *Player) buildLyricsTraditional(main *model.Main, lyricBuilder *strings.Builder) {
	var startCol int
	if main.IsDualColumn() {
		startCol = main.MenuStartColumn() + 3
	} else {
		startCol = main.MenuStartColumn() - 4
	}

	maxLen := p.netease.WindowWidth() - startCol - 4
	switch p.lyricLines {
	// 3行歌词
	case 3:
		for i := 1; i <= 3; i++ {
			if startCol > 0 {
				lyricBuilder.WriteString(strings.Repeat(" ", startCol))
			}
			if i == 2 {
				lyricLine := p.lyricNowScrollBar.Tick(maxLen, p.lyrics[i])
				lyricBuilder.WriteString(util.SetFgStyle(lyricLine, termenv.ANSIBrightCyan))
			} else {
				lyricLine := runewidth.Truncate(runewidth.FillRight(p.lyrics[i], maxLen), maxLen, "")
				lyricBuilder.WriteString(util.SetFgStyle(lyricLine, termenv.ANSIBrightBlack))
			}
			lyricBuilder.WriteString("\n")
		}
	// 5行歌词
	case 5:
		for i := 0; i < 5; i++ {
			if startCol > 0 {
				lyricBuilder.WriteString(strings.Repeat(" ", startCol))
			}
			if i == 2 {
				lyricLine := p.lyricNowScrollBar.Tick(maxLen, p.lyrics[i])
				lyricBuilder.WriteString(util.SetFgStyle(lyricLine, termenv.ANSIBrightCyan))
			} else {
				lyricLine := runewidth.Truncate(runewidth.FillRight(p.lyrics[i], maxLen), maxLen, "")
				lyricBuilder.WriteString(util.SetFgStyle(lyricLine, termenv.ANSIBrightBlack))
			}
			lyricBuilder.WriteString("\n")
		}
	}
}

// songView 歌曲信息UI
func (p *Player) songView() string {

	// Every part of the song view is expressed as a segment: unformatted text followed by a color specification
	// This makes computing the total length of the song view easier
	type Segment struct {
		text  string
		color termenv.Color
	}

	var (
		builder  strings.Builder
		main     = p.netease.MustMain()
		segments []Segment
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
			var msg string
			if p.intelligent {
				msg = "心动"
			} else {
				msg = p.songManager.modeName()
			}
			addSegment(fmt.Sprintf("[%s] ", msg), termenv.ANSIBrightMagenta)
		}
		addSegment(fmt.Sprintf("%d%% ", p.Volume()), termenv.ANSIBrightBlue)
	}
	if p.State() == types.Playing {
		addSegment("♫ ♪ ♫ ♪ ", termenv.ANSIBrightYellow)
	} else {
		addSegment("_ z Z Z ", termenv.ANSIYellow)
	}

	if p.CurSong().Id > 0 {
		var color termenv.ANSIColor
		if likelist.IsLikeSong(p.CurSong().Id) {
			color = termenv.ANSIRed
		} else {
			color = termenv.ANSIWhite
		}
		addSegment("♥ ", color)
	}

	if p.CurSongIndex() < len(p.Playlist()) {
		// 按剩余长度截断字符串
		songName := p.CurSong().Name
		if !main.CenterEverything() {
			songName = runewidth.Truncate(songName, p.netease.WindowWidth()-main.MenuStartColumn()-prefixLen, "") // 多减，避免剩余1个中文字符
		}
		addSegment(songName, util.GetPrimaryColor())
		addText(" ")

		var artists strings.Builder
		for i, v := range p.CurSong().Artists {
			if i != 0 {
				artists.WriteString(",")
			}
			artists.WriteString(v.Name)
		}

		artistString := artists.String()
		if !main.CenterEverything() {
			// 按剩余长度截断字符串
			remainLen := p.netease.WindowWidth() - main.MenuStartColumn() - prefixLen - runewidth.StringWidth(p.CurSong().Name)
			artistString = runewidth.Truncate(
				runewidth.FillRight(artistString, remainLen),
				remainLen, "")
		}
		addSegment(artistString, termenv.ANSIBrightBlack)
	}

	if main.CenterEverything() {
		totalWidth := 0
		widthLimit := p.netease.WindowWidth() - 4
		for index, segment := range segments {
			segmentWidth := runewidth.StringWidth(segment.text)
			if totalWidth+segmentWidth > widthLimit {
				segmentWidth = max(0, widthLimit-totalWidth)
				segments[index].text = runewidth.Truncate(segment.text, segmentWidth, "")
			}
			totalWidth += segmentWidth
		}
		paddingLeft := (p.netease.WindowWidth() - totalWidth) / 2
		builder.WriteString(strings.Repeat(" ", paddingLeft))
		for _, segment := range segments {
			builder.WriteString(util.SetFgStyle(segment.text, segment.color))
		}
		builder.WriteString(strings.Repeat(" ", p.netease.WindowWidth()-paddingLeft-totalWidth))
	} else {
		// simply concatenate every segment with the specified color
		for _, segment := range segments {
			builder.WriteString(util.SetFgStyle(segment.text, segment.color))
		}
	}

	return builder.String()
}

// progressView 进度条UI
func (p *Player) progressView() string {
	allDuration := int(p.CurMusic().Duration.Seconds())
	if allDuration == 0 {
		return ""
	}
	passedDuration := int(p.PassedTime().Seconds())
	progress := passedDuration * 100 / allDuration

	width := float64(p.netease.WindowWidth() - 14)
	start, end := model.GetProgressColor()
	if width != p.progressLastWidth || len(p.progressRamp) == 0 {
		p.progressRamp = util.MakeRamp(start, end, width)
		p.progressLastWidth = width
	}

	progressView := model.Progress(&p.netease.Options().ProgressOptions, int(width), int(math.Round(width*float64(progress)/100)), p.progressRamp)

	if allDuration/60 >= 100 {
		times := util.SetFgStyle(fmt.Sprintf("%03d:%02d/%03d:%02d", passedDuration/60, passedDuration%60, allDuration/60, allDuration%60), util.GetPrimaryColor())
		return progressView + " " + times
	} else {
		times := util.SetFgStyle(fmt.Sprintf("%02d:%02d/%02d:%02d", passedDuration/60, passedDuration%60, allDuration/60, allDuration%60), util.GetPrimaryColor())
		return progressView + " " + times + " "
	}
}

// InPlayingMenu 是否处于正在播放的菜单中
func (p *Player) InPlayingMenu() bool {
	key := p.netease.MustMain().CurMenu().GetMenuKey()
	return key == p.playingMenuKey || key == CurPlaylistKey
}

// CompareWithCurPlaylist 与当前播放列表对比，是否一致
func (p *Player) CompareWithCurPlaylist(playlist []structs.Song) bool {
	if len(playlist) != len(p.Playlist()) {
		return false
	}

	// 如果前20个一致，则认为相同
	for i := 0; i < 20 && i < len(playlist); i++ {
		if playlist[i].Id != p.Playlist()[i].Id {
			return false
		}
	}

	return true
}

// LocatePlayingSong 定位到正在播放的音乐
func (p *Player) LocatePlayingSong() {
	var (
		main        = p.netease.MustMain()
		curMenu, ok = main.CurMenu().(Menu)
	)
	if !ok {
		return
	}

	if !curMenu.IsLocatable() {
		return
	}

	menu, ok := curMenu.(SongsMenu)
	if !ok {
		return
	}
	if !p.InPlayingMenu() || !p.CompareWithCurPlaylist(menu.Songs()) {
		return
	}

	pageDelta := p.CurSongIndex()/main.PageSize() - (main.CurPage() - 1)
	if pageDelta > 0 {
		for i := 0; i < pageDelta; i++ {
			p.netease.MustMain().NextPage()
		}
	} else if pageDelta < 0 {
		for i := 0; i > pageDelta; i-- {
			p.netease.MustMain().PrePage()
		}
	}
	main.SetSelectedIndex(p.CurSongIndex())
}

// PlaySong 播放歌曲
func (p *Player) PlaySong(song structs.Song, direction PlayDirection) {
	if song.Id != p.CurSong().Id {
		p.reportEnd() // 上一首播放结束
	}

	loading := model.NewLoading(p.netease.MustMain())
	loading.Start()
	defer loading.Complete()

	table := storage.NewTable()
	_ = table.SetByKVModel(storage.PlayerSnapshot{}, storage.PlayerSnapshot{
		CurSongIndex:     p.CurSongIndex(),
		Playlist:         p.Playlist(),
		PlaylistUpdateAt: p.playlistUpdateAt,
	})

	p.LocatePlayingSong()
	p.Pause()
	url, musicType, err := storagex.PlayableURLSong(song)
	if url == "" || err != nil {
		p.progressRamp = []string{}
		p.playErrCount++
		if p.playErrCount >= configs.ConfigRegistry.Player.MaxPlayErrCount {
			return
		}
		switch direction {
		case DurationPrev:
			p.PreviousSong(false)
		case DurationNext:
			p.NextSong(false)
		}
		return
	}

	go p.updateLyric(song.Id)

	p.Play(player.URLMusic{
		URL:  url,
		Song: song,
		Type: player.SongTypeMapping[musicType],
	})
	slog.Info("Start play song", slog.String("url", url), slog.String("type", musicType), slog.Any("song", song))

	// 上报开始播放
	p.reportStart()

	go notify.Notify(notify.NotifyContent{
		Title:   "正在播放: " + song.Name,
		Text:    fmt.Sprintf("%s - %s", song.ArtistName(), song.Album.Name),
		Icon:    app.AddResizeParamForPicUrl(song.PicUrl, 60),
		Url:     netease.WebUrlOfSong(song.Id),
		GroupId: types.GroupID,
	})

	p.playErrCount = 0
}

func (p *Player) StartPlay() {
	if len(p.Playlist()) <= p.CurSongIndex() {
		return
	}
	p.PlaySong(p.CurSong(), DurationNext)
}

func (p *Player) Mode() types.Mode {
	return p.songManager.mode()
}

func (p *Player) Playlist() []structs.Song {
	return p.songManager.getPlaylist()
}

func (p *Player) InitSongManager(index int, playlist []structs.Song) {
	p.songManager.init(index, playlist)
}

func (p *Player) CurSongIndex() int {
	return p.songManager.getCurSongIndex()
}

func (p *Player) CurSong() structs.Song {
	if len(p.Playlist()) <= p.CurSongIndex() {
		return structs.Song{}
	}
	return p.Playlist()[p.CurSongIndex()]
}

// NextSong 下一曲
func (p *Player) NextSong(manual bool) {
	index := p.CurSongIndex()
	playlistLen := len(p.Playlist())
	if playlistLen == 0 || index >= playlistLen-1 {
		if p.Mode() == types.PmIntelligent {
			p.Intelligence(true)
		}

		main := p.netease.MustMain()
		if p.InPlayingMenu() {
			if main.IsDualColumn() && index%2 == 0 {
				p.netease.MustMain().MoveRight()
			} else {
				p.netease.MustMain().MoveDown()
			}
		} else if p.playingMenu != nil {
			if bottomHook := p.playingMenu.BottomOutHook(); bottomHook != nil {
				bottomHook(main)
			}
		}
	}

	p.songManager.nextSong(manual).ifSome(func(song structs.Song) {
		p.PlaySong(song, DurationNext)
	})
}

// PreviousSong 上一曲
func (p *Player) PreviousSong(manual bool) {
	index := p.CurSongIndex()
	playlistLen := len(p.Playlist())
	if playlistLen == 0 || index >= playlistLen-1 {
		main := p.netease.MustMain()
		if p.InPlayingMenu() {
			if main.IsDualColumn() && index%2 == 0 {
				p.netease.MustMain().MoveUp()
			} else {
				p.netease.MustMain().MoveLeft()
			}
		} else if p.playingMenu != nil {
			if topHook := p.playingMenu.TopOutHook(); topHook != nil {
				topHook(main)
			}
		}
	}

	p.songManager.prevSong(manual).ifSome(func(song structs.Song) {
		p.PlaySong(song, DurationNext)
	})
}

func (p *Player) Seek(duration time.Duration) {
	p.Player.Seek(duration)
	if p.lrcTimer != nil {
		p.lrcTimer.Rewind()
	}
	p.stateHandler.SetPlayingInfo(p.PlayingInfo())
}

// SetMode 设置播放模式
func (p *Player) SetMode(playMode types.Mode) {
	if p.lastMode != p.netease.player.Mode() {
		p.lastMode = p.netease.player.Mode()
	}
	switch playMode {
	case types.PmIntelligent:
		p.intelligent = true
		ordered := p.songManager.ordered()
		p.songManager = &ordered
	case types.PmOrdered:
		p.intelligent = false
		ordered := p.songManager.ordered()
		p.songManager = &ordered
	case types.PmSingleLoop:
		p.intelligent = false
		single := p.songManager.singleLoop()
		p.songManager = &single
	case types.PmListRandom:
		p.intelligent = false
		random := p.songManager.listRandom()
		p.songManager = &random
	case types.PmInfRandom:
		p.intelligent = false
		random := p.songManager.infRandom()
		p.songManager = &random
	case types.PmListLoop:
		fallthrough
	default:
		p.intelligent = false
		list := p.songManager.listLoop()
		p.songManager = &list
	}

	table := storage.NewTable()
	_ = table.SetByKVModel(storage.PlayMode{}, playMode)
}

// SwitchMode 顺序切换播放模式
func (p *Player) SwitchMode() {
	mode := p.Mode()
	mode = (mode + 1) % types.NormalPmLength
	p.SetMode(mode)

	table := storage.NewTable()
	_ = table.SetByKVModel(storage.PlayMode{}, p.Mode())
}

// Close 关闭
func (p *Player) Close() error {
	// 退出前上报
	p.reportEnd()

	p.cancel()
	if p.stateHandler != nil {
		p.stateHandler.Release()
	}
	p.Player.Close()
	return nil
}

// lyricListener 歌词变更监听
func (p *Player) lyricListener(_ int64, content, transContent string, _ bool, index int) {
	curIndex := len(p.lyrics) / 2

	// before
	for i := 0; i < curIndex; i++ {
		if f, tf := p.lrcTimer.GetLRCFragment(index - curIndex + i); f != nil {
			p.lyrics[i] = f.Content
			if tf != nil && tf.Content != "" {
				p.lyrics[i] += " [" + tf.Content + "]"
			}
		} else {
			p.lyrics[i] = ""
		}
	}

	// cur
	p.lyrics[curIndex] = content
	if transContent != "" {
		p.lyrics[curIndex] += " [" + transContent + "]"
	}

	// after
	for i := 1; i < len(p.lyrics)-curIndex; i++ {
		if f, tf := p.lrcTimer.GetLRCFragment(index + i); f != nil {
			p.lyrics[curIndex+i] = f.Content
			if tf != nil && tf.Content != "" {
				p.lyrics[curIndex+i] += " [" + tf.Content + "]"
			}
		} else {
			p.lyrics[curIndex+i] = ""
		}
	}
}

// getLyric 获取歌曲歌词
func (p *Player) getLyric(songId int64) {
	p.lrcFile, _ = lyric.ReadLRC(strings.NewReader("[00:00.00] 暂无歌词~"))
	p.transLrcFile, _ = lyric.ReadTranslateLRC(strings.NewReader("[00:00.00]"))
	lrcService := service.LyricService{
		ID: strconv.FormatInt(songId, 10),
	}
	code, response := lrcService.Lyric()
	if code != 200 {
		return
	}

	if lrc, err := jsonparser.GetString(response, "lrc", "lyric"); err == nil && lrc != "" {
		if file, err := lyric.ReadLRC(strings.NewReader(lrc)); err == nil {
			p.lrcFile = file
		}
	}
	if configs.ConfigRegistry.Main.ShowLyricTrans {
		if lrc, err := jsonparser.GetString(response, "tlyric", "lyric"); err == nil && lrc != "" {
			if file, err := lyric.ReadTranslateLRC(strings.NewReader(lrc)); err == nil {
				p.transLrcFile = file
			}
		}
	}
	if p.stateHandler != nil {
		p.stateHandler.SetPlayingInfo(p.PlayingInfo())
	}
}

// updateLyric 更新歌词UI
func (p *Player) updateLyric(songId int64) {
	p.getLyric(songId)
	if !configs.ConfigRegistry.Main.ShowLyric {
		return
	}
	p.lyrics = [5]string{}
	if p.lrcTimer != nil {
		p.lrcTimer.Stop()
	}
	p.lrcTimer = lyric.NewLRCTimer(p.lrcFile, p.transLrcFile)
	p.lrcTimer.AddListener(p.lyricListener)
	p.lrcTimer.Start()
}

// Intelligence 智能/心动模式
func (p *Player) Intelligence(appendMode bool) model.Page {
	var (
		main    = p.netease.MustMain()
		curMenu = main.CurMenu()
	)
	playlist, ok := curMenu.(*PlaylistDetailMenu)
	if !ok {
		return nil
	}

	selectedIndex := curMenu.RealDataIndex(main.SelectedIndex())
	if selectedIndex >= len(playlist.songs) {
		return nil
	}

	if _struct.CheckUserInfo(p.netease.user) == _struct.NeedLogin {
		page, _ := p.netease.ToLoginPage(nil)
		return page
	}

	intelligenceService := service.PlaymodeIntelligenceListService{
		SongId:       strconv.FormatInt(playlist.songs[selectedIndex].Id, 10),
		PlaylistId:   strconv.FormatInt(playlist.playlistId, 10),
		StartMusicId: strconv.FormatInt(playlist.songs[selectedIndex].Id, 10),
	}
	code, response := intelligenceService.PlaymodeIntelligenceList()
	codeType := _struct.CheckCode(code)
	if codeType == _struct.NeedLogin {
		page, _ := p.netease.ToLoginPage(func() model.Page {
			p.Intelligence(appendMode)
			return nil
		})
		return page
	} else if codeType != _struct.Success {
		return nil
	}
	songs := _struct.GetIntelligenceSongs(response)

	var song structs.Song
	if appendMode {
		p.songManager.init(p.CurSongIndex(), append(p.Playlist(), songs...))
		p.playlistUpdateAt = time.Now()
		song = p.songManager.nextSong(true).unwrap()
	} else {
		p.SetMode(types.PmIntelligent)
		p.playingMenuKey = "Intelligent"
		p.songManager.init(0, append([]structs.Song{playlist.songs[selectedIndex]}, songs...))
		p.playlistUpdateAt = time.Now()
		song = p.Playlist()[0]
	}

	p.PlaySong(song, DurationNext)
	return nil
}

func (p *Player) UpVolume() {
	p.Player.UpVolume()

	if v, ok := p.Player.(storage.VolumeStorable); ok {
		table := storage.NewTable()
		_ = table.SetByKVModel(storage.Volume{}, v.Volume())
	}

	p.stateHandler.SetPlayingInfo(p.PlayingInfo())
}

func (p *Player) DownVolume() {
	p.Player.DownVolume()

	if v, ok := p.Player.(storage.VolumeStorable); ok {
		table := storage.NewTable()
		_ = table.SetByKVModel(storage.Volume{}, v.Volume())
	}

	p.stateHandler.SetPlayingInfo(p.PlayingInfo())
}

func (p *Player) SetVolume(volume int) {
	p.Player.SetVolume(volume)

	p.stateHandler.SetPlayingInfo(p.PlayingInfo())
}

func (p *Player) handleControlSignal(signal CtrlSignal) {
	switch signal.Type {
	case CtrlPaused:
		p.Pause()
	case CtrlResume:
		p.Resume()
	case CtrlStop:
		p.Stop()
	case CtrlToggle:
		p.Toggle()
	case CtrlPrevious:
		p.PreviousSong(true)
	case CtrlNext:
		p.NextSong(true)
	case CtrlSeek:
		p.Seek(signal.Duration)
	case CtrlRerender:
		p.netease.Rerender(false)
	}
}

func (p *Player) PlayingInfo() control.PlayingInfo {
	song := p.CurSong()
	return control.PlayingInfo{
		TotalDuration:  song.Duration,
		PassedDuration: p.PassedTime(),
		State:          p.State(),
		Volume:         p.Volume(),
		TrackID:        song.Id,
		PicUrl:         song.PicUrl,
		Name:           song.Name,
		Album:          song.Album.Name,
		Artist:         song.ArtistName(),
		AlbumArtist:    song.Album.ArtistName(),
		LRCText:        p.lrcFile.AsText(p.transLrcFile),
	}
}

func (p *Player) RenderTicker() model.Ticker {
	return p.renderTicker
}

func (p *Player) buildNeteaseReportService() *service.ReportService {
	svc := &service.ReportService{
		ID:      p.CurSong().Id,
		Alg:     p.CurSong().Alg,
		Type:    "song",
		Time:    int64(p.PassedTime().Seconds()),
		EndType: "playend",
	}

	switch m := p.playingMenu.(type) {
	case *PlaylistDetailMenu:
		svc.SourceType = "list"
		svc.SourceId = strconv.FormatInt(m.PlaylistId(), 10)
	case *AlbumDetailMenu:
		svc.SourceType = "album"
		svc.SourceId = strconv.FormatInt(m.AlbumId(), 10)
	case *DailyRecommendSongsMenu:
		svc.SourceType = "dailySongRecommend"
	case *PersonalFmMenu:
		svc.SourceType = "userfm"
	case *DjRecommendMenu:
		svc.Type = "dj"
	case nil:
		if p.CurSong().Album.Id != 0 {
			svc.SourceType = "album"
			svc.SourceId = strconv.FormatInt(p.CurSong().Album.Id, 10)
		}
	}

	if p.CurSong().Duration.Seconds()-p.PassedTime().Seconds() > 10 {
		svc.EndType = "ui"
	}

	return svc
}

func (p *Player) reportStart() {
	p.buildNeteaseReportService().Playstart()
	lastfm.Report(p.netease.lastfm, lastfm.ReportPhaseStart, p.CurSong(), p.PassedTime())
}

func (p *Player) reportEnd() {
	// 播放时间不少于20秒时，才上报
	if p.PassedTime().Seconds() < 20 {
		return
	}

	svc := p.buildNeteaseReportService()
	switch {
	case math.Abs(p.CurSong().Duration.Seconds()-p.PassedTime().Seconds()) <= 10:
		// 播放结束
		svc.EndType = "playend"
	case time.Since(p.playlistUpdateAt).Seconds() <= 5:
		// 更换了歌单
		svc.EndType = "interrupt"
	default:
		svc.EndType = "ui"
	}
	svc.Playend()

	// 播放过一半, 上报lastfm
	if p.PassedTime().Seconds() >= p.CurSong().Duration.Seconds()/2 {
		lastfm.Report(p.netease.lastfm, lastfm.ReportPhaseComplete, p.CurSong(), p.PassedTime())
	}
}
