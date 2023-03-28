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
	"github.com/go-musicfox/go-musicfox/pkg/lyric"
	"github.com/go-musicfox/go-musicfox/pkg/player"
	"github.com/go-musicfox/go-musicfox/pkg/state_handler"
	"github.com/go-musicfox/go-musicfox/pkg/storage"
	"github.com/go-musicfox/go-musicfox/pkg/structs"
	"github.com/go-musicfox/go-musicfox/utils"

	"github.com/anhoder/netease-music/service"
	"github.com/buger/jsonparser"
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

type ReportPhase uint8

const (
	ReportPhaseStart ReportPhase = iota
	ReportPhaseComplete
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

	lrcTimer      *lyric.LRCTimer // 歌词计时器
	lyrics        [5]string       // 歌词信息，保留5行
	showLyric     bool            // 显示歌词
	lyricStartRow int             // 歌词开始行
	lyricLines    int             // 歌词显示行数，3或5
	lyricNow      string          // 当前播放的歌词 = lyrics[2]

	// 播放进度条
	progressLastWidth float64
	progressRamp      []string

	playErrCount int // 错误计数，当错误连续超过5次，停止播放
	mode         player.Mode
	stateHandler *state_handler.Handler
	ctrl         chan CtrlSignal

	player.Player // 播放器
}

// 调用自增长 闭包
func increment(max int) func(change bool) int {
	i := 0
	return func(change bool) int {
		i++
		if i > max || change {
			i = 1
		}
		return i
	}
}

// 更新当前正在播放的歌词，实现水平滚动
func (p *Player) updateCurrentLyric() {
	ticker := time.NewTicker(time.Millisecond * 100)
	f := increment(1000)
	var lrcTmp string
	for {
		select {
		case <-ticker.C:
			i := f(false)
			if lrcTmp != p.lyrics[2] {
				i = f(true)
				lrcTmp = p.lyrics[2]
			}
			var tmp string
			length := runewidth.StringWidth(p.lyrics[2])
			width := p.model.WindowWidth - p.model.menuStartColumn - 4
			// 歌词首末补偿，歌词开头结尾等待15*100ms
			// 100ms由上述ticker间隔决定
			a := i%(length+15) - 15
			if length < width || a < 1 {
				tmp = runewidth.TruncateLeft(lrcTmp, 0, "")
			} else if a+width <= length {
				tmp = runewidth.TruncateLeft(lrcTmp, a, "")
			} else {
				tmp = runewidth.TruncateLeft(lrcTmp, length-width, "")
			}
			p.lyricNow = runewidth.Truncate(runewidth.FillRight(tmp, width), width, "")
		}
	}
}

func NewPlayer(model *NeteaseModel) *Player {
	p := &Player{
		model: model,
		mode:  player.PmListLoop,
		ctrl:  make(chan CtrlSignal),
	}
	var ctx context.Context
	ctx, p.cancel = context.WithCancel(context.Background())

	p.Player = player.NewPlayerFromConfig()
	p.stateHandler = state_handler.NewHandler(p)
	// 更新当前歌词
	go p.updateCurrentLyric()

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
					p.report(ReportPhaseComplete)
					p.Next()
				} else {
					p.Rerender()
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
				if duration.Seconds()-p.CurMusic().Duration.Seconds() > 10 {
					// 上报
					p.report(ReportPhaseComplete)
					p.NextSong()
				}
				if p.lrcTimer != nil {
					select {
					case p.lrcTimer.Timer() <- duration + time.Millisecond*time.Duration(configs.ConfigRegistry.MainLyricOffset):
					default:
					}
				}
				p.model.Rerender()
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

	switch p.lyricLines {
	// 3行歌词
	case 3:
		for i := 1; i <= 3; i++ {
			if p.model.menuStartColumn+3 > 0 {
				lyricBuilder.WriteString(strings.Repeat(" ", p.model.menuStartColumn+3))
			}
			lyricLine := runewidth.Truncate(runewidth.FillRight(p.lyrics[i], p.model.WindowWidth-p.model.menuStartColumn-4), p.model.WindowWidth-p.model.menuStartColumn-4, "")
			if i == 2 {
				lyricBuilder.WriteString(SetFgStyle(p.lyricNow, termenv.ANSIBrightCyan))
			} else {
				lyricBuilder.WriteString(SetFgStyle(lyricLine, termenv.ANSIBrightBlack))
			}

			lyricBuilder.WriteString("\n")
		}
	// 5行歌词
	case 5:
		for i := 0; i < 5; i++ {
			if p.model.menuStartColumn+3 > 0 {
				lyricBuilder.WriteString(strings.Repeat(" ", p.model.menuStartColumn+3))
			}
			lyricLine := runewidth.Truncate(
				runewidth.FillRight(p.lyrics[i], p.model.WindowWidth-p.model.menuStartColumn-4),
				p.model.WindowWidth-p.model.menuStartColumn-4,
				"")
			if i == 2 {
				lyricBuilder.WriteString(SetFgStyle(p.lyricNow, termenv.ANSIBrightCyan))
			} else {
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

	if p.model.menuStartColumn-4 > 0 {
		builder.WriteString(strings.Repeat(" ", p.model.menuStartColumn-4))
		builder.WriteString(SetFgStyle(fmt.Sprintf("[%s] ", player.ModeName(p.mode)), termenv.ANSIBrightMagenta))
	}
	builder.WriteString(SetFgStyle(fmt.Sprintf("%d%% ", p.Volume()), termenv.ANSIBrightBlue))
	if p.State() == player.Playing {
		builder.WriteString(SetFgStyle("♫ ♪ ♫ ♪ ", termenv.ANSIBrightYellow))
	} else {
		builder.WriteString(SetFgStyle("_ z Z Z ", termenv.ANSIYellow))
	}

	if p.curSongIndex < len(p.playlist) {
		prefixLen := 22
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

	fullSize := int(math.Round(width * float64(progress) / 100))
	var fullCells string
	for i := 0; i < fullSize && i < len(p.progressRamp); i++ {
		fullCells += termenv.String(string(configs.ConfigRegistry.ProgressFullChar)).Foreground(termProfile.Color(p.progressRamp[i])).String()
	}

	emptySize := 0
	if int(width)-fullSize > 0 {
		emptySize = int(width) - fullSize
	}
	emptyCells := SetFgStyle(strings.Repeat(string(configs.ConfigRegistry.ProgressEmptyChar), emptySize), termenv.ANSIBrightBlack)

	if allDuration/60 >= 100 {
		times := SetFgStyle(fmt.Sprintf("%03d:%02d/%03d:%02d", passedDuration/60, passedDuration%60, allDuration/60, allDuration%60), GetPrimaryColor())
		return fmt.Sprintf("%s%s %s", fullCells, emptyCells, times)
	} else {
		times := SetFgStyle(fmt.Sprintf("%02d:%02d/%02d:%02d", passedDuration/60, passedDuration%60, allDuration/60, allDuration%60), GetPrimaryColor())
		return fmt.Sprintf("%s%s  %s ", fullCells, emptyCells, times)
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
	p.report(ReportPhaseStart)

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
		if len(p.playlist) == 0 {
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

	p.model.Rerender()
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
	} else {
		p.lyrics[curIndex] = ""
	}

	// after
	for i := 0; i < len(p.lyrics)-curIndex; i++ {
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

func (p *Player) report(phase ReportPhase) {
	switch phase {
	case ReportPhaseStart:
		go func(song structs.Song) {
			_ = p.model.lastfm.UpdateNowPlaying(map[string]interface{}{
				"artist":   song.ArtistName(),
				"track":    song.Name,
				"album":    song.Album.Name,
				"duration": song.Duration,
			})
		}(p.curSong)
	case ReportPhaseComplete:
		duration := p.curSong.Duration.Seconds()
		passedTime := p.PassedTime().Seconds()
		if duration <= passedTime || passedTime >= duration/2 {
			go func(song structs.Song, passed time.Duration) {
				_ = p.model.lastfm.Scrobble(map[string]interface{}{
					"artist":    song.ArtistName(),
					"track":     song.Name,
					"album":     song.Album.Name,
					"timestamp": time.Now().Unix(),
					"duration":  song.Duration.Seconds(),
				})
			}(p.curSong, p.PassedTime())
		}
	}
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
		p.model.Rerender()
	}
}
