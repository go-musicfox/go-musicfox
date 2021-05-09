package ui

import (
    "errors"
    "fmt"
    "github.com/anhoder/netease-music/service"
    "github.com/buger/jsonparser"
    "github.com/mattn/go-runewidth"
    "github.com/muesli/termenv"
    "go-musicfox/constants"
    "go-musicfox/db"
    "go-musicfox/ds"
    "go-musicfox/lyric"
    "go-musicfox/utils"
    "math"
    "math/rand"
    "strconv"
    "strings"
)

// PlayDirection 下首歌的方向
type PlayDirection uint8

const (
    DurationNext PlayDirection = iota
    DurationPrev
)

// PlayMode 播放模式
type PlayMode string

const (
    PmListLoop    PlayMode = "列表" // 列表循环
    PmOrder       PlayMode = "顺序" // 顺序播放
    PmSingleLoop  PlayMode = "单曲" // 单曲循环
    PmRandom      PlayMode = "随机" // 随机播放
    PmIntelligent PlayMode = "心动" // 智能模式
)

// Player 网易云音乐播放器
type Player struct {
    model *NeteaseModel

    playlist       []ds.Song // 歌曲列表
    curSongIndex   int       // 当前歌曲的下标
    playingMenuKey string    // 正在播放的菜单Key

    lrcTimer      *lyric.LRCTimer // 歌词计时器
    lyrics        [5]string       // 歌词信息，保留5行
    showLyric     bool            // 显示歌词
    lyricStartRow int             // 歌词开始行
    lyricLines    int             // 歌词显示行数，3或5

    // 播放进度条
    progressLastWidth float64
    progressRamp      []string

    playErrCount int // 错误计数，当错误连续超过5次，停止播放
    mode         PlayMode

    *utils.Player // 播放器
}

func NewPlayer(model *NeteaseModel) *Player {
    player := &Player{
        model:  model,
        mode:   PmListLoop,
        Player: utils.NewPlayer(),
    }

    // done监听
    go func() {
        for {
            select {
            case <-player.Player.Done():
                player.NextSong()
                model.Rerender()
            case duration := <-player.TimeChan():
                if duration.Seconds()-player.CurMusic.Duration.Seconds() > 10 {
                    player.NextSong()
                    break
                }
                if player.lrcTimer != nil {
                    select {
                    case player.lrcTimer.Timer() <- duration:
                    default:
                    }
                }

                player.model.Rerender()
            }
        }
    }()

    return player
}

// playerView 播放器UI，包含 lyricView, songView, progressView
func (p *Player) playerView(top *int) string {
    var playerBuilder strings.Builder

    playerBuilder.WriteString(p.lyricView())

    playerBuilder.WriteString(p.songView())
    playerBuilder.WriteString("\n\n\n")

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
                lyricBuilder.WriteString(SetFgStyle(lyricLine, termenv.ANSIBrightCyan))
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
                lyricBuilder.WriteString(SetFgStyle(lyricLine, termenv.ANSIBrightCyan))
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
        builder.WriteString(SetFgStyle(fmt.Sprintf("[%s] ", p.mode), termenv.ANSIBrightMagenta))
    }
    if p.State == utils.Playing {
        builder.WriteString(SetFgStyle("♫  ♪ ♫  ♪  ", termenv.ANSIBrightYellow))
    } else {
        builder.WriteString(SetFgStyle("_ _ z Z Z  ", termenv.ANSIBrightRed))
    }

    if p.curSongIndex < len(p.playlist) {
        // 按剩余长度截断字符串
        truncateSong := runewidth.Truncate(p.playlist[p.curSongIndex].Name, p.model.WindowWidth-p.model.menuStartColumn-16, "") // 多减，避免剩余1个中文字符
        builder.WriteString(SetFgStyle(truncateSong, GetPrimaryColor()))
        builder.WriteString(" ")

        var artists strings.Builder
        for i, v := range p.playlist[p.curSongIndex].Artists {
            if i != 0 {
                artists.WriteString(",")
            }

            artists.WriteString(v.Name)
        }

        // 按剩余长度截断字符串
        remainLen := p.model.WindowWidth - p.model.menuStartColumn - 16 - runewidth.StringWidth(p.playlist[p.curSongIndex].Name)
        truncateArtists := runewidth.Truncate(
            runewidth.FillRight(artists.String(), remainLen),
            remainLen, "")
        builder.WriteString(SetFgStyle(truncateArtists, termenv.ANSIBrightBlack))
    }

    return builder.String()
}

// progressView 进度条UI
func (p *Player) progressView() string {
    if p.Timer == nil {
        return ""
    }

    width := float64(p.model.WindowWidth - 14)

    if progressStartColor == "" || progressEndColor == "" || len(p.progressRamp) == 0 {
        progressStartColor, progressEndColor = GetRandomRgbColor(true)
    }
    if width != p.progressLastWidth || len(p.progressRamp) == 0 {
        p.progressRamp = makeRamp(progressStartColor, progressEndColor, width)
        p.progressLastWidth = width
    }

    fullSize := int(math.Round(width * float64(p.Progress) / 100))
    var fullCells string
    for i := 0; i < fullSize && i < len(p.progressRamp); i++ {
        fullCells += termenv.String(string(constants.ProgressFullChar)).Foreground(termProfile.Color(p.progressRamp[i])).String()
    }

    emptySize := 0
    if int(width)-fullSize > 0 {
        emptySize = int(width) - fullSize
    }
    emptyCells := strings.Repeat(string(constants.ProgressEmptyChar), emptySize)

    passedDuration := int(p.Timer.Passed().Seconds())
    allDuration := int(p.CurMusic.Duration.Seconds())

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
    return p.model.menu.GetMenuKey() == p.playingMenuKey
}

// CompareWithCurPlaylist 与当前播放列表对比，是否一致
func (p *Player) CompareWithCurPlaylist(playlist []ds.Song) bool {

    if len(playlist) != len(p.playlist) {
        return false
    }

    // 如果前10个一致，则认为相同
    for i := 0; i < 10 && i < len(playlist); i++ {
        if playlist[i].Id != p.playlist[i].Id {
            return false
        }
    }

    return true
}

// LocatePlayingSong 定位到正在播放的音乐
func (p *Player) LocatePlayingSong() {
    if !p.InPlayingMenu() || !p.CompareWithCurPlaylist(p.playlist) {
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
func (p *Player) PlaySong(song ds.Song, duration PlayDirection) error {
    loading := NewLoading(p.model)
    loading.start()
    defer loading.complete()

    p.LocatePlayingSong()
    urlService := service.SongUrlService{}
    urlService.ID = strconv.FormatInt(song.Id, 10)
    urlService.Br = constants.PlayerSongBr
    code, response := urlService.SongUrl()
    if code != 200 {
        return errors.New(string(response))
    }

    url, err1 := jsonparser.GetString(response, "data", "[0]", "url")
    musicType, err2 := jsonparser.GetString(response, "data", "[0]", "type")
    if err1 != nil || err2 != nil || (musicType != "mp3" && musicType != "flac") {
        p.State = utils.Stopped
        p.progressRamp = []string{}
        p.playErrCount++
        if p.playErrCount >= 3 {
            return nil
        }
        switch duration {
        case DurationPrev:
            p.PreSong()
        case DurationNext:
            p.NextSong()
        }
        return nil
    }

    go p.updateLyric(song.Id)

    switch musicType {
    case "mp3":
        p.Player.Play(utils.Mp3, url, song.Duration)
    case "flac":
        p.Player.Play(utils.Flac, url, song.Duration)
    }

    p.playErrCount = 0

    table := db.NewTable()
    _ = table.SetByKVModel(db.PlayerSnapshot{}, db.PlayerSnapshot{
        p.curSongIndex,
        p.playlist,
        p.playingMenuKey,
    })

    return nil
}

// NextSong 下一曲
func (p *Player) NextSong() {
    if len(p.playlist) == 0 || p.curSongIndex >= len(p.playlist)-1 {
        if p.mode == PmIntelligent {
           p.Intelligence(true)
        }
        if p.model.doubleColumn && p.curSongIndex%2 == 0 {
            moveRight(p.model)
        } else {
            moveDown(p.model)
        }
    }

    switch p.mode {
    case PmListLoop, PmIntelligent:
        p.curSongIndex++
        if p.curSongIndex > len(p.playlist)-1 {
            p.curSongIndex = 0
        }
    case PmSingleLoop:
    case PmRandom:
        p.curSongIndex = rand.Intn(len(p.playlist)-1)
    case PmOrder:
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

// PreSong 上一曲
func (p *Player) PreSong() {
    if len(p.playlist) == 0 || p.curSongIndex >= len(p.playlist)-1 {
        if p.mode == PmIntelligent {
           p.Intelligence(true)
        }
        if p.model.doubleColumn && p.curSongIndex%2 == 0 {
            moveUp(p.model)
        } else {
            moveLeft(p.model)
        }
    }

    switch p.mode {
    case PmListLoop, PmIntelligent:
        p.curSongIndex--
        if p.curSongIndex < 0 {
            p.curSongIndex = len(p.playlist) - 1
        }
    case PmSingleLoop:
    case PmRandom:
        p.curSongIndex = rand.Intn(len(p.playlist)-1)
    case PmOrder:
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
func (p *Player) SetPlayMode(playMode PlayMode) {
    if playMode == "" {
        switch p.mode {
        case PmListLoop:
            p.mode = PmOrder
        case PmOrder:
            p.mode = PmSingleLoop
        case PmSingleLoop:
            p.mode = PmRandom
        case PmRandom:
            p.mode = PmListLoop
        default:
            p.mode = PmListLoop
        }
    } else {
        p.mode = playMode
    }

    table := db.NewTable()
    _ = table.SetByKVModel(db.PlayMode{}, p.mode)

    p.model.Rerender()
}

// Close 关闭
func (p *Player) Close() {
    p.Player.Close()
}

// lyricListener 歌词变更监听
func (p *Player) lyricListener(_ int64, content string, _ bool, index int) {
    curIndex := len(p.lyrics) / 2

    // before
    for i := 0; i < curIndex; i++ {
        if f := p.lrcTimer.GetLRCFragment(index - curIndex + i); f != nil {
            p.lyrics[i] = f.Content
        } else {
            p.lyrics[i] = ""
        }
    }

    // cur
    p.lyrics[curIndex] = content

    // after
    for i := 0; i < len(p.lyrics)-curIndex; i++ {
        if f := p.lrcTimer.GetLRCFragment(index + i); f != nil {
            p.lyrics[curIndex+i] = f.Content
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
    defer func() {
        p.lrcTimer = lyric.NewLRCTimer(lrcFile)
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

    if lrc, err := jsonparser.GetString(response, "lrc", "lyric"); err == nil {
        if file, err := lyric.ReadLRC(strings.NewReader(lrc)); err == nil {
            lrcFile = file
        }
    }
}

// Intelligence 智能/心动模式
func (p *Player) Intelligence(appendMode bool)  {
    playlist, ok := p.model.menu.(*PlaylistDetailMenu)
    if !ok {
        return
    }

    if p.model.selectedIndex >= len(playlist.songs) {
        return
    }

    if utils.CheckUserInfo(p.model.user) == utils.NeedLogin {
        NeedLoginHandle(p.model, nil)
        return
    }

    intelligenceService := service.PlaymodeIntelligenceListService{
        SongId:       strconv.FormatInt(playlist.songs[p.model.selectedIndex].Id, 10),
        PlaylistId:   strconv.FormatInt(playlist.PlaylistId, 10),
        StartMusicId: strconv.FormatInt(playlist.songs[p.model.selectedIndex].Id, 10),
    }
    code, response := intelligenceService.PlaymodeIntelligenceList()
    codeType := utils.CheckCode(code)
    if codeType == utils.NeedLogin {
        NeedLoginHandle(p.model, func(m *NeteaseModel) {
            p.Intelligence(appendMode)
        })
        return
    } else if codeType != utils.Success {
        return
    }
    songs := utils.GetIntelligenceSongs(response)

    if appendMode {
        p.playlist = append(p.playlist, songs...)
        p.curSongIndex++
    } else {
        p.playlist = append([]ds.Song{playlist.songs[p.model.selectedIndex]}, songs...)
        p.curSongIndex = 0
    }
    p.mode = PmIntelligent
    p.playingMenuKey = "Intelligent"

    _ = p.PlaySong(p.playlist[p.curSongIndex], DurationNext)
}