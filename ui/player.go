package ui

import (
    "errors"
    "fmt"
    "github.com/anhoder/netease-music/service"
    "github.com/buger/jsonparser"
    "github.com/mattn/go-runewidth"
    "github.com/muesli/termenv"
    "go-musicfox/constants"
    "go-musicfox/ds"
    "go-musicfox/lyric"
    "go-musicfox/utils"
    "math"
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
    PmOrder       PlayMode = "顺序" // 顺序播放
    PmListLoop             = "列表" // 列表循环
    PmSingleLoop           = "单曲" // 单曲循环
    PmRandom               = "随机" // 随机播放
    PmIntelligent          = "智能" // 智能模式
)

type Player struct {
    model *NeteaseModel

    playlist       []ds.Song // 歌曲列表
    curSongIndex   int       // 当前歌曲的下标
    playingMenuKey string    // 正在播放的菜单Key
    isIntelligence bool      // 智能模式（心动模式）

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

    *utils.Player
}

func NewPlayer(model *NeteaseModel) *Player {
    player := &Player{
        model:  model,
        mode: PmOrder,              // 默认顺序，TODO
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

func (p *Player) playerView(top *int) string {
    var playerBuilder strings.Builder

    playerBuilder.WriteString(p.LyricView())

    playerBuilder.WriteString("\n")
    playerBuilder.WriteString(p.SongView())
    playerBuilder.WriteString("\n\n")

    playerBuilder.WriteString(p.ProgressView())

    *top = p.model.WindowHeight

    return playerBuilder.String()
}

func (p *Player) LyricView() string {
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
            lyricBuilder.WriteString(strings.Repeat(" ", p.model.menuStartColumn+4))
            if i == 2 {
                lyricBuilder.WriteString(SetFgStyle(runewidth.FillRight(p.lyrics[i], p.model.WindowWidth-p.model.menuStartColumn-4), termenv.ANSICyan))
            } else {
                lyricBuilder.WriteString(SetFgStyle(runewidth.FillRight(p.lyrics[i], p.model.WindowWidth-p.model.menuStartColumn-4), termenv.ANSIBrightBlack))
            }

            lyricBuilder.WriteString("\n")
        }
    // 5行歌词
    case 5:
        for i := 0; i < 5; i++ {
            lyricBuilder.WriteString(strings.Repeat(" ", p.model.menuStartColumn+4))
            if i == 2 {
                lyricBuilder.WriteString(SetFgStyle(runewidth.FillRight(p.lyrics[i], p.model.WindowWidth-p.model.menuStartColumn-4), termenv.ANSICyan))
            } else {
                lyricBuilder.WriteString(SetFgStyle(runewidth.FillRight(p.lyrics[i], p.model.WindowWidth-p.model.menuStartColumn-4), termenv.ANSIBrightBlack))
            }
            lyricBuilder.WriteString("\n")
        }
    }

    if endRow-p.lyricStartRow-p.lyricLines > 0 {
        lyricBuilder.WriteString(strings.Repeat("\n", endRow-p.lyricStartRow-p.lyricLines))
    }

    return lyricBuilder.String()
}

func (p *Player) SongView() string {
    var builder strings.Builder

    builder.WriteString(strings.Repeat(" ", p.model.menuStartColumn-6))
    builder.WriteString(SetFgStyle(fmt.Sprintf("[%s] ", p.mode), termenv.ANSIMagenta))
    if p.State == utils.Playing {
        builder.WriteString(SetFgStyle("♫  ♪ ♫  ♪  ", GetPrimaryColor()))
    } else {
        builder.WriteString(SetFgStyle("_ _ z Z Z  ", GetPrimaryColor()))
    }

    if p.curSongIndex < len(p.playlist) {
        builder.WriteString(SetFgStyle(p.playlist[p.curSongIndex].Name, GetPrimaryColor()))
        builder.WriteString(" ")


        var artists strings.Builder
        for i, v := range p.playlist[p.curSongIndex].Artists {
            if i != 0 {
                artists.WriteString(",")
            }

            artists.WriteString(v.Name)
        }

        builder.WriteString(SetFgStyle(artists.String(), termenv.ANSIBrightBlack))
    }

    return builder.String()
}

func (p *Player) ProgressView() string {
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

    return nil
}

// NextSong 下一曲
func (p *Player) NextSong() {
    if len(p.playlist) == 0 || p.curSongIndex >= len(p.playlist)-1 {
        //if p.isIntelligence {
        //    p.intelligence(true)
        //}
        if p.model.doubleColumn && p.curSongIndex%2 == 0 {
            moveRight(p.model)
        } else {
            moveDown(p.model)
        }
    }
    //switch (_playMode) {
    //case PlaylistMode.LIST_LOOP:
    //    _curSongIndex = _curSongIndex >= songs.length - 1 ? 0 : _curSongIndex + 1;
    //    break;
    //case PlaylistMode.SINGLE_LOOP:
    //    break;
    //case PlaylistMode.SHUFFLE:
    //    _curSongIndex = Random().nextInt(songs.length - 1);
    //    break;
    //case PlaylistMode.ORDER:
    //    if (_curSongIndex >= songs.length - 1) return;
    //    _curSongIndex++;
    //    break;
    //}
    p.curSongIndex++
    if p.curSongIndex > len(p.playlist)-1 {
        p.curSongIndex = 0
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
        //if p.isIntelligence {
        //    p.intelligence(true)
        //}
        if p.model.doubleColumn && p.curSongIndex%2 == 0 {
            moveUp(p.model)
        } else {
            moveLeft(p.model)
        }
    }
    //switch (_playMode) {
    //case PlaylistMode.LIST_LOOP:
    //    _curSongIndex = _curSongIndex >= songs.length - 1 ? 0 : _curSongIndex + 1;
    //    break;
    //case PlaylistMode.SINGLE_LOOP:
    //    break;
    //case PlaylistMode.SHUFFLE:
    //    _curSongIndex = Random().nextInt(songs.length - 1);
    //    break;
    //case PlaylistMode.ORDER:
    //    if (_curSongIndex >= songs.length - 1) return;
    //    _curSongIndex++;
    //    break;
    //}
    p.curSongIndex--
    if p.curSongIndex < 0 {
        p.curSongIndex = len(p.playlist) - 1
    }

    if p.curSongIndex < 0 {
        return
    }
    song := p.playlist[p.curSongIndex]
    _ = p.PlaySong(song, DurationPrev)
}

// Close 关闭
func (p *Player) Close() {
    p.Player.Close()
}

func (p *Player) lyricListener(_ int64, content string, last bool, index int) {
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
