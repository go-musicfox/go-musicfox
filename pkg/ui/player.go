package ui

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/go-musicfox/go-musicfox/pkg/configs"
	"github.com/go-musicfox/go-musicfox/pkg/constants"
	"github.com/go-musicfox/go-musicfox/pkg/lastfm"
	"github.com/go-musicfox/go-musicfox/pkg/lyric"
	"github.com/go-musicfox/go-musicfox/pkg/player"
	"github.com/go-musicfox/go-musicfox/pkg/state_handler"
	"github.com/go-musicfox/go-musicfox/pkg/storage"
	"github.com/go-musicfox/go-musicfox/pkg/structs"
	"github.com/go-musicfox/go-musicfox/utils"
	"github.com/go-musicfox/go-musicfox/utils/like_list"

	"github.com/buger/jsonparser"
	"github.com/go-musicfox/netease-music/service"
	"github.com/mattn/go-runewidth"
	"github.com/muesli/termenv"
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
	model  *NeteaseModel
	cancel context.CancelFunc

	playlist         []structs.Song // 歌曲列表
	playlistUpdateAt time.Time      // 播放列表更新时间
	curSongIndex     int            // 当前歌曲的下标
	curSong          structs.Song   // 当前歌曲信息（防止播放列表发生变动后，歌曲信息不匹配）
	playingMenuKey   string         // 正在播放的菜单Key
	playingMenu      Menu
	playedTime       time.Duration // 已经播放的时长

	lrcTimer          *lyric.LRCTimer   // 歌词计时器
	lyrics            [5]string         // 歌词信息，保留5行
	showLyric         bool              // 显示歌词
	lyricStartRow     int               // 歌词开始行
	lyricLines        int               // 歌词显示行数，3或5
	lyricNowScrollBar *utils.XScrollBar // 当前歌词滚动

	// 播放进度条
	progressLastWidth float64
	progressRamp      []string

	playErrCount int // 错误计数，当错误连续超过5次，停止播放
	mode         player.Mode
	stateHandler *state_handler.Handler
	ctrl         chan CtrlSignal

	player.Player // 播放器
}

func NewPlayer(model *NeteaseModel) *Player {
	p := &Player{
		model:             model,
		mode:              player.PmListLoop,
		ctrl:              make(chan CtrlSignal),
		lyricNowScrollBar: utils.NewXScrollBar(),
	}
	var ctx context.Context
	ctx, p.cancel = context.WithCancel(context.Background())

	p.Player = player.NewPlayerFromConfig()
	p.stateHandler = state_handler.NewHandler(p)

	// remote control
	go func() {
		defer utils.Recover(false)
		for {
			select {
			case <-ctx.Done():
				return
			case signal := <-p.ctrl:
				p.handleControlSignal(signal)
			}
		}
	}()

	// 状态监听
	go func() {
		defer utils.Recover(false)
		for {
			select {
			case <-ctx.Done():
				return
			case s := <-p.Player.StateChan():
				p.stateHandler.SetPlayingInfo(p.PlayingInfo())
				if s == player.Stopped {
					// 上报lastfm
					lastfm.Report(p.model.lastfm, lastfm.ReportPhaseComplete, p.curSong, p.PassedTime())
					// 自动切歌且播放时间不少于(实际歌曲时间-20)秒时，才上报至网易云
					if p.CurMusic().Duration.Seconds()-p.playedTime.Seconds() < 20 {
						utils.ReportSongEnd(p.curSong.Id, p.PlayingInfo().TrackID, p.PassedTime())
					}
					p.Next()
				} else {
					p.model.Rerender(false)
				}
			}
		}
	}()

	// 时间监听
	go func() {
		defer utils.Recover(false)
		for {
			select {
			case <-ctx.Done():
				return
			case duration := <-p.TimeChan():
				// 200ms 为刷新间隔，刷新间隔修改时此处需要保持同步
				p.playedTime += time.Millisecond * 200
				if duration.Seconds()-p.CurMusic().Duration.Seconds() > 10 {
					// 上报
					lastfm.Report(p.model.lastfm, lastfm.ReportPhaseComplete, p.curSong, p.PassedTime())
					p.NextSong()
				}
				if p.lrcTimer != nil {
					select {
					case p.lrcTimer.Timer() <- duration + time.Millisecond*time.Duration(configs.ConfigRegistry.MainLyricOffset):
					default:
					}
				}

				p.model.Rerender(false)
			}
		}
	}()

	return p
}

// playerView 播放器UI，包含 lyricView, songView, progressView
func (p *Player) playerView(top *int) string {
	var playerBuilder strings.Builder

	playerBuilder.WriteString(p.lyricView())

	playerBuilder.WriteString(p.songView())
	playerBuilder.WriteString("\n\n")

	playerBuilder.WriteString(p.progressView())

	*top = p.model.WindowHeight

	return playerBuilder.String()
}

// lyricView 歌词显示UI
func (p *Player) lyricView() string {
	endRow := p.model.WindowHeight - 4

	if !p.showLyric {
		if endRow-p.model.menuBottomRow > 0 {
			return strings.Repeat("\n", endRow-p.model.menuBottomRow)
		} else {
			return ""
		}
	}

	var lyricBuilder strings.Builder
	if p.lyricStartRow > p.model.menuBottomRow {
		lyricBuilder.WriteString(strings.Repeat("\n", p.lyricStartRow-p.model.menuBottomRow))
	}

	var startCol int
	if p.model.doubleColumn {
		startCol = p.model.menuStartColumn + 3
	} else {
		startCol = p.model.menuStartColumn - 4
	}

	maxLen := p.model.WindowWidth - startCol - 4
	switch p.lyricLines {
	// 3行歌词
	case 3:
		for i := 1; i <= 3; i++ {
			if startCol > 0 {
				lyricBuilder.WriteString(strings.Repeat(" ", startCol))
			}
			if i == 2 {
				lyricLine := p.lyricNowScrollBar.Tick(maxLen, p.lyrics[i])
				lyricBuilder.WriteString(SetFgStyle(lyricLine, termenv.ANSIBrightCyan))
			} else {
				lyricLine := runewidth.Truncate(runewidth.FillRight(p.lyrics[i], maxLen), maxLen, "")
				lyricBuilder.WriteString(SetFgStyle(lyricLine, termenv.ANSIBrightBlack))
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
				lyricBuilder.WriteString(SetFgStyle(lyricLine, termenv.ANSIBrightCyan))
			} else {
				lyricLine := runewidth.Truncate(runewidth.FillRight(p.lyrics[i], maxLen), maxLen, "")
				lyricBuilder.WriteString(SetFgStyle(lyricLine, termenv.ANSIBrightBlack))
			}
			lyricBuilder.WriteString("\n")
		}
	}

	if endRow-p.lyricStartRow-p.lyricLines > 0 {
		lyricBuilder.WriteString(strings.Repeat("\n", endRow-p.lyricStartRow-p.lyricLines))
	}

	return lyricBuilder.String()
}

// songView 歌曲信息UI
func (p *Player) songView() string {
	var builder strings.Builder

	var prefixLen = 10
	if p.model.menuStartColumn-4 > 0 {
		prefixLen += 12
		builder.WriteString(strings.Repeat(" ", p.model.menuStartColumn-4))
		builder.WriteString(SetFgStyle(fmt.Sprintf("[%s] ", player.ModeName(p.mode)), termenv.ANSIBrightMagenta))
		builder.WriteString(SetFgStyle(fmt.Sprintf("%d%% ", p.Volume()), termenv.ANSIBrightBlue))
	}
	if p.State() == player.Playing {
		builder.WriteString(SetFgStyle("♫ ♪ ♫ ♪ ", termenv.ANSIBrightYellow))
	} else {
		builder.WriteString(SetFgStyle("_ z Z Z ", termenv.ANSIYellow))
	}

	if p.curSong.Id > 0 {
		if like_list.IsLikeSong(p.curSong.Id) {
			builder.WriteString(SetFgStyle("♥ ", termenv.ANSIRed))
		} else {
			builder.WriteString(SetFgStyle("♥ ", termenv.ANSIWhite))
		}
	}

	if p.curSongIndex < len(p.playlist) {
		// 按剩余长度截断字符串
		truncateSong := runewidth.Truncate(p.curSong.Name, p.model.WindowWidth-p.model.menuStartColumn-prefixLen, "") // 多减，避免剩余1个中文字符
		builder.WriteString(SetFgStyle(truncateSong, GetPrimaryColor()))
		builder.WriteString(" ")

		var artists strings.Builder
		for i, v := range p.curSong.Artists {
			if i != 0 {
				artists.WriteString(",")
			}

			artists.WriteString(v.Name)
		}

		// 按剩余长度截断字符串
		remainLen := p.model.WindowWidth - p.model.menuStartColumn - prefixLen - runewidth.StringWidth(p.curSong.Name)
		truncateArtists := runewidth.Truncate(
			runewidth.FillRight(artists.String(), remainLen),
			remainLen, "")
		builder.WriteString(SetFgStyle(truncateArtists, termenv.ANSIBrightBlack))
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

	width := float64(p.model.WindowWidth - 14)
	if progressStartColor == "" || progressEndColor == "" || len(p.progressRamp) == 0 {
		progressStartColor, progressEndColor = GetRandomRgbColor(true)
	}
	if width != p.progressLastWidth || len(p.progressRamp) == 0 {
		p.progressRamp = makeRamp(progressStartColor, progressEndColor, width)
		p.progressLastWidth = width
	}

	progressView := Progress(int(width), int(math.Round(width*float64(progress)/100)), p.progressRamp)

	if allDuration/60 >= 100 {
		times := SetFgStyle(fmt.Sprintf("%03d:%02d/%03d:%02d", passedDuration/60, passedDuration%60, allDuration/60, allDuration%60), GetPrimaryColor())
		return progressView + " " + times
	} else {
		times := SetFgStyle(fmt.Sprintf("%02d:%02d/%02d:%02d", passedDuration/60, passedDuration%60, allDuration/60, allDuration%60), GetPrimaryColor())
		return progressView + " " + times + " "
	}

}

// InPlayingMenu 是否处于正在播放的菜单中
func (p *Player) InPlayingMenu() bool {
	return p.model.menu.GetMenuKey() == p.playingMenuKey || p.model.menu.GetMenuKey() == CurPlaylistKey
}

// CompareWithCurPlaylist 与当前播放列表对比，是否一致
func (p *Player) CompareWithCurPlaylist(playlist []structs.Song) bool {

	if len(playlist) != len(p.playlist) {
		return false
	}

	// 如果前20个一致，则认为相同
	for i := 0; i < 20 && i < len(playlist); i++ {
		if playlist[i].Id != p.playlist[i].Id {
			return false
		}
	}

	return true
}

// LocatePlayingSong 定位到正在播放的音乐
func (p *Player) LocatePlayingSong() {
	if !p.model.menu.IsLocatable() {
		return
	}

	menu, ok := p.model.menu.(SongsMenu)
	if !ok {
		return
	}
	if !p.InPlayingMenu() || !p.CompareWithCurPlaylist(menu.Songs()) {
		return
	}

	var pageDelta = p.curSongIndex/p.model.menuPageSize - (p.model.menuCurPage - 1)
	if pageDelta > 0 {
		for i := 0; i < pageDelta; i++ {
			nextPage(p.model)
		}
	} else if pageDelta < 0 {
		for i := 0; i > pageDelta; i-- {
			prePage(p.model)
		}
	}
	p.model.selectedIndex = p.curSongIndex
}

// PlaySong 播放歌曲
func (p *Player) PlaySong(song structs.Song, direction PlayDirection) error {
	loading := NewLoading(p.model)
	loading.start()
	defer loading.complete()

	table := storage.NewTable()
	_ = table.SetByKVModel(storage.PlayerSnapshot{}, storage.PlayerSnapshot{
		CurSongIndex:     p.curSongIndex,
		Playlist:         p.playlist,
		PlaylistUpdateAt: p.playlistUpdateAt,
	})
	p.curSong = song
	p.playedTime = 0

	p.LocatePlayingSong()
	p.Player.Paused()
	url, musicType, err := utils.GetSongUrl(song.Id)
	if url == "" || err != nil {
		p.progressRamp = []string{}
		p.playErrCount++
		if p.playErrCount >= 3 {
			return nil
		}
		switch direction {
		case DurationPrev:
			p.PreviousSong()
		case DurationNext:
			p.NextSong()
		}
		return nil
	}

	if configs.ConfigRegistry.MainShowLyric {
		go p.updateLyric(song.Id)
	}

	p.Player.Play(player.UrlMusic{
		Url:  url,
		Song: song,
		Type: player.SongTypeMapping[musicType],
	})

	// 上报
	lastfm.Report(p.model.lastfm, lastfm.ReportPhaseStart, p.curSong, p.PassedTime())

	go utils.Notify(utils.NotifyContent{
		Title:   "正在播放: " + song.Name,
		Text:    fmt.Sprintf("%s - %s", song.ArtistName(), song.Album.Name),
		Icon:    utils.AddResizeParamForPicUrl(song.PicUrl, 60),
		Url:     utils.WebUrlOfSong(song.Id),
		GroupId: constants.GroupID,
	})

	p.playErrCount = 0

	return nil
}

// NextSong 下一曲
func (p *Player) NextSong() {
	if len(p.playlist) == 0 || p.curSongIndex >= len(p.playlist)-1 {
		if p.mode == player.PmIntelligent {
			p.Intelligence(true)
		}

		if p.InPlayingMenu() {
			if p.model.doubleColumn && p.curSongIndex%2 == 0 {
				moveRight(p.model)
			} else {
				moveDown(p.model)
			}
		} else if p.playingMenu != nil {
			if bottomHook := p.playingMenu.BottomOutHook(); bottomHook != nil {
				bottomHook(p.model)
			}
		}
	}

	switch p.mode {
	case player.PmListLoop, player.PmIntelligent:
		p.curSongIndex++
		if p.curSongIndex > len(p.playlist)-1 {
			p.curSongIndex = 0
		}
	case player.PmSingleLoop:
	case player.PmRandom:
		if len(p.playlist)-1 < 0 {
			return
		}
		if len(p.playlist)-1 == 0 {
			p.curSongIndex = 0
		} else {
			p.curSongIndex = rand.Intn(len(p.playlist) - 1)
		}
	case player.PmOrder:
		if p.curSongIndex >= len(p.playlist)-1 {
			return
		}
		p.curSongIndex++
	}

	if p.curSongIndex > len(p.playlist)-1 {
		return
	}
	song := p.playlist[p.curSongIndex]
	_ = p.PlaySong(song, DurationNext)
}

// PreviousSong 上一曲
func (p *Player) PreviousSong() {
	if len(p.playlist) == 0 || p.curSongIndex >= len(p.playlist)-1 {
		if p.mode == player.PmIntelligent {
			p.Intelligence(true)
		}

		if p.InPlayingMenu() {
			if p.model.doubleColumn && p.curSongIndex%2 == 0 {
				moveUp(p.model)
			} else {
				moveLeft(p.model)
			}
		} else if p.playingMenu != nil {
			if topHook := p.playingMenu.TopOutHook(); topHook != nil {
				topHook(p.model)
			}
		}
	}

	switch p.mode {
	case player.PmListLoop, player.PmIntelligent:
		p.curSongIndex--
		if p.curSongIndex < 0 {
			p.curSongIndex = len(p.playlist) - 1
		}
	case player.PmSingleLoop:
	case player.PmRandom:
		if len(p.playlist)-1 < 0 {
			return
		}
		if len(p.playlist) == 0 {
			p.curSongIndex = 0
		} else {
			p.curSongIndex = rand.Intn(len(p.playlist) - 1)
		}
	case player.PmOrder:
		if p.curSongIndex <= 0 {
			return
		}
		p.curSongIndex--
	}

	if p.curSongIndex < 0 {
		return
	}
	song := p.playlist[p.curSongIndex]
	_ = p.PlaySong(song, DurationPrev)
}

// SetPlayMode 播放模式切换
func (p *Player) SetPlayMode(playMode player.Mode) {
	if playMode > 0 {
		p.mode = playMode
	} else {
		switch p.mode {
		case player.PmListLoop, player.PmOrder, player.PmSingleLoop:
			p.mode++
		case player.PmRandom:
			p.mode = player.PmListLoop
		default:
			p.mode = player.PmListLoop
		}
	}

	table := storage.NewTable()
	_ = table.SetByKVModel(storage.PlayMode{}, p.mode)
}

// Close 关闭
func (p *Player) Close() {
	p.cancel()
	if p.stateHandler != nil {
		p.stateHandler.Release()
	}
	p.Player.Close()
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

// updateLyric 更新歌词UI
func (p *Player) updateLyric(songId int64) {
	p.lyrics = [5]string{}
	if p.lrcTimer != nil {
		p.lrcTimer.Stop()
	}
	lrcFile, _ := lyric.ReadLRC(strings.NewReader("[00:00.00] 暂无歌词~"))
	tranLRCFile, _ := lyric.ReadTranslateLRC(strings.NewReader("[00:00.00]"))
	defer func() {
		p.lrcTimer = lyric.NewLRCTimer(lrcFile, tranLRCFile)
		p.lrcTimer.AddListener(p.lyricListener)
		p.lrcTimer.Start()
	}()

	lrcService := service.LyricService{
		ID: strconv.FormatInt(songId, 10),
	}
	code, response := lrcService.Lyric()
	if code != 200 {
		return
	}

	if lrc, err := jsonparser.GetString(response, "lrc", "lyric"); err == nil && lrc != "" {
		if file, err := lyric.ReadLRC(strings.NewReader(lrc)); err == nil {
			lrcFile = file
		}
	}
	if configs.ConfigRegistry.MainShowLyricTrans {
		if lrc, err := jsonparser.GetString(response, "tlyric", "lyric"); err == nil && lrc != "" {
			if file, err := lyric.ReadTranslateLRC(strings.NewReader(lrc)); err == nil {
				tranLRCFile = file
			}
		}
	}
}

// Intelligence 智能/心动模式
func (p *Player) Intelligence(appendMode bool) {
	playlist, ok := p.model.menu.(*PlaylistDetailMenu)
	if !ok {
		return
	}

	selectedIndex := p.model.menu.RealDataIndex(p.model.selectedIndex)
	if selectedIndex >= len(playlist.songs) {
		return
	}

	if utils.CheckUserInfo(p.model.user) == utils.NeedLogin {
		NeedLoginHandle(p.model, nil)
		return
	}

	intelligenceService := service.PlaymodeIntelligenceListService{
		SongId:       strconv.FormatInt(playlist.songs[selectedIndex].Id, 10),
		PlaylistId:   strconv.FormatInt(playlist.playlistId, 10),
		StartMusicId: strconv.FormatInt(playlist.songs[selectedIndex].Id, 10),
	}
	code, response := intelligenceService.PlaymodeIntelligenceList()
	codeType := utils.CheckCode(code)
	if codeType == utils.NeedLogin {
		NeedLoginHandle(p.model, func(m *NeteaseModel, newMenu Menu, newTitle *MenuItem) {
			p.Intelligence(appendMode)
		})
		return
	} else if codeType != utils.Success {
		return
	}
	songs := utils.GetIntelligenceSongs(response)

	if appendMode {
		p.playlist = append(p.playlist, songs...)
		p.playlistUpdateAt = time.Now()
		p.curSongIndex++
	} else {
		p.playlist = append([]structs.Song{playlist.songs[selectedIndex]}, songs...)
		p.playlistUpdateAt = time.Now()
		p.curSongIndex = 0
	}
	p.mode = player.PmIntelligent
	p.playingMenuKey = "Intelligent"

	_ = p.PlaySong(p.playlist[p.curSongIndex], DurationNext)
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
		p.Player.Paused()
	case CtrlResume:
		p.Player.Resume()
	case CtrlStop:
		p.Player.Stop()
	case CtrlToggle:
		p.Player.Toggle()
	case CtrlPrevious:
		p.PreviousSong()
	case CtrlNext:
		p.NextSong()
	case CtrlSeek:
		p.Player.Seek(signal.Duration)
		if p.lrcTimer != nil {
			p.lrcTimer.Rewind()
		}
		p.stateHandler.SetPlayingInfo(p.PlayingInfo())
	case CtrlRerender:
		p.model.Rerender(false)
	}
}
