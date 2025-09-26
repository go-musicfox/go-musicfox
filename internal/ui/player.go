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
	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-musicfox/netease-music/service"
	"github.com/mattn/go-runewidth"
	"github.com/muesli/termenv"

	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/lyric"
	"github.com/go-musicfox/go-musicfox/internal/player"
	"github.com/go-musicfox/go-musicfox/internal/playlist"
	control "github.com/go-musicfox/go-musicfox/internal/remote_control"
	"github.com/go-musicfox/go-musicfox/internal/reporter"
	"github.com/go-musicfox/go-musicfox/internal/storage"
	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/internal/types"
	"github.com/go-musicfox/go-musicfox/utils/app"
	"github.com/go-musicfox/go-musicfox/utils/errorx"
	"github.com/go-musicfox/go-musicfox/utils/likelist"
	"github.com/go-musicfox/go-musicfox/utils/netease"
	"github.com/go-musicfox/go-musicfox/utils/notify"
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

	playlistUpdateAt time.Time                // 播放列表更新时间
	playlistManager  playlist.PlaylistManager // 播放列表管理器
	lastMode         types.Mode
	playingMenuKey   string // 正在播放的菜单Key
	playingMenu      Menu

	lyricService *lyric.Service

	// 播放进度条
	progressLastWidth float64
	progressRamp      []string

	playErrCount int // 错误计数，当错误连续超过5次，停止播放
	stateHandler *control.RemoteControl
	ctrl         chan CtrlSignal

	renderTicker *tickerByPlayer // renderTicker 用于渲染

	player.Player // 播放器
	reporter      reporter.Service
}

func NewPlayer(n *Netease, lyricService *lyric.Service) *Player {
	reporterOptions := []reporter.Option{}
	if configs.AppConfig.Reporter.Lastfm.Enable {
		skipDjRadio := configs.AppConfig.Reporter.Lastfm.SkipDjRadio
		reporterOptions = append(reporterOptions, reporter.WithLastFM(n.lastfm.Tracker, skipDjRadio))
	}
	if configs.AppConfig.Reporter.Netease.Enable {
		reporterOptions = append(reporterOptions, reporter.WithNetease())
	}

	p := &Player{
		netease:         n,
		lyricService:    lyricService,
		playlistManager: playlist.NewPlaylistManager(),
		ctrl:            make(chan CtrlSignal),
		reporter:        reporter.NewService(reporterOptions...),
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

				p.lyricService.UpdatePosition(duration)

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

func (p *Player) Update(_ tea.Msg, _ *model.App) {}

func (p *Player) View(a *model.App, main *model.Main) (view string, lines int) {
	var playerBuilder strings.Builder
	playerBuilder.WriteString(p.songView())
	playerBuilder.WriteString("\n\n")
	playerBuilder.WriteString(p.progressView())
	return playerBuilder.String(), a.WindowHeight() - main.MenuBottomRow()
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
			msg := p.playlistManager.GetPlayModeName()
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
	p.reporter.ReportEnd(p.PlayedTime())

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
	url, musicType, err := p.getPlayInfo(song)

	var skip bool
	logger := slog.With(slog.String("url", url), slog.String("type", musicType), slog.Any("song", song))
	if configs.AppConfig.UNM.SkipInvalidTracks {
		skip, _ = netease.HasBannedPathSuffix(url)
	}

	if url == "" || err != nil || skip {
		p.progressRamp = []string{}
		p.playErrCount++
		if skip {
			logger.Info("已拦截无效播放")
		} else {
			logger.Error("Play song error", slog.Any("err", err))
		}
		if p.playErrCount >= configs.AppConfig.Player.MaxPlayErrCount {
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

	go p.lyricService.SetSong(context.Background(), song)

	p.Play(player.URLMusic{
		URL:  url,
		Song: song,
		Type: player.SongTypeMapping[musicType],
	})
	slog.Info("Start play song", slog.String("url", url), slog.String("type", musicType), slog.Any("song", song))

	// 上报开始播放
	p.reporter.ReportStart(song)

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
	return p.playlistManager.GetPlayMode()
}

func (p *Player) Playlist() []structs.Song {
	return p.playlistManager.GetPlaylist()
}

func (p *Player) InitSongManager(index int, playlist []structs.Song) {
	_ = p.playlistManager.Initialize(index, playlist)
}

func (p *Player) CurSongIndex() int {
	return p.playlistManager.GetCurrentIndex()
}

func (p *Player) CurSong() structs.Song {
	index := p.CurSongIndex()
	if index < 0 || len(p.Playlist()) <= index {
		return structs.Song{}
	}
	return p.Playlist()[index]
}

// NextSong 下一曲
func (p *Player) NextSong(manual bool) {
	index := p.CurSongIndex()
	playlistLen := len(p.Playlist())

	// 尝试获取下一首歌曲
	song, err := p.playlistManager.NextSong(manual)
	if err != nil {
		// 如果是心动模式且到达列表末尾，获取更多推荐
		if p.Mode() == types.PmIntelligent {
			p.Intelligence(true)
			return
		}

		// 其他模式的处理逻辑
		if playlistLen == 0 || index >= playlistLen-1 {
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
		return
	}

	p.PlaySong(song, DurationNext)
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

	if song, err := p.playlistManager.PreviousSong(manual); err == nil {
		p.PlaySong(song, DurationNext)
	}
}

func (p *Player) Seek(duration time.Duration) {
	p.Player.Seek(duration)
	p.stateHandler.SetPlayingInfo(p.PlayingInfo())
}

// SetMode 设置播放模式
func (p *Player) SetMode(playMode types.Mode) {
	if p.lastMode != p.netease.player.Mode() {
		p.lastMode = p.netease.player.Mode()
	}

	// 直接使用PlaylistManager设置播放模式
	_ = p.playlistManager.SetPlayMode(playMode)

	table := storage.NewTable()
	_ = table.SetByKVModel(storage.PlayMode{}, playMode)
}

// SwitchMode 顺序切换播放模式
func (p *Player) SwitchMode() {
	mode := p.Mode()
	supportedModes := p.playlistManager.SupportedPlayModes()
	index := 0
	for i, m := range supportedModes {
		if mode != m {
			continue
		}
		index = i + 1
		break
	}

	for {
		if index >= len(supportedModes) {
			index = 0
		}
		if supportedModes[index] == types.PmIntelligent {
			index++
			continue
		}
		break
	}

	p.SetMode(supportedModes[index])
}

// Close 关闭
func (p *Player) Close() error {
	// 退出前上报
	p.reporter.ReportEnd(p.PlayedTime())

	p.cancel()
	if p.stateHandler != nil {
		p.stateHandler.Release()
	}
	p.Player.Close()
	return nil
}

func (p *Player) getPlayInfo(song structs.Song) (string, string, error) {
	source, err := p.netease.trackManager.ResolvePlayableSource(context.Background(), song)
	if err != nil || source.Info == nil {
		return "", "", err
	}
	url := source.Info.URL
	musicType := source.Info.MusicType
	return url, musicType, err
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

	// 获取智能推荐歌曲
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
		_ = p.playlistManager.Initialize(p.CurSongIndex(), append(p.Playlist(), songs...))
		p.playlistUpdateAt = time.Now()
		var err error
		song, err = p.playlistManager.NextSong(true)
		if err != nil {
			return nil
		}
	} else {
		p.SetMode(types.PmIntelligent)
		p.playingMenuKey = "Intelligent"
		_ = p.playlistManager.Initialize(0, append([]structs.Song{playlist.songs[selectedIndex]}, songs...))
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
		LRCText:        p.lyricService.State().FormatAsLRC(),
	}
}

func (p *Player) RenderTicker() model.Ticker {
	return p.renderTicker
}
