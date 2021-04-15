package ui

import (
    "errors"
    "github.com/anhoder/netease-music/service"
    "github.com/buger/jsonparser"
    "go-musicfox/constants"
    "go-musicfox/ds"
    "go-musicfox/utils"
    "strconv"
    "strings"
)

type Player struct {
    model         *NeteaseModel

    playlist       []ds.Song
    curSongIndex   int
    playingMenuKey string // 正在播放的菜单Key
    isIntelligence bool   // 智能模式（心动模式）

    lyric         map[float64]string
    showLyric     bool
    lyricQueue    *[]string
    lyricStartRow int
    lyricLines    int

    *utils.Player
}

func NewPlayer(model *NeteaseModel) *Player {
    player := &Player{
        model: model,
        Player: utils.NewPlayer(),
    }

    // done监听
    go func() {
        for {
            <- player.Player.Done()
            player.NextSong()
            model.Rerender()
        }
    }()

    return player
}

func (p *Player) playerView() string {
    var playerBuilder strings.Builder

    playerBuilder.WriteString(p.LyricView())

    return playerBuilder.String()

}

func (p *Player) LyricView() string {
    startRow := p.model.menuBottomRow + 1
    endRow := p.model.WindowHeight - 4

    if len(p.lyric) == 0 || !p.showLyric {
        return strings.Repeat("\n", endRow - startRow)
    }

    var lyricBuilder strings.Builder
    lyricBuilder.WriteString(strings.Repeat("\n", p.lyricStartRow - p.model.menuBottomRow))

    return lyricBuilder.String()
}

func (p *Player) SingView() string {
    return ""
}

func (p *Player) ProgressView() string {
    return ""
}

// InPlayingMenu 是否处于正在播放的菜单中
func (p *Player) InPlayingMenu() bool {
    return p.model.menu.GetMenuKey() == p.playingMenuKey
}

// CompareWithCurPlaylist 与当前播放列表对比，是否一致
func (p *Player) CompareWithCurPlaylist(playlist interface{}) bool {
    list, ok := playlist.([]ds.Song)
    if !ok {
       return false
    }

    if len(list) != len(p.playlist) {
        return false
    }

    // 如果前10个一致，则认为相同
    for i := 0; i < 10 && i < len(list); i++ {
        if list[i].Id != p.playlist[i].Id {
            return false
        }
    }

    return true
}

// LocatePlayingSong 定位到正在播放的音乐
func (p *Player) LocatePlayingSong() {
    if !p.InPlayingMenu() || !p.CompareWithCurPlaylist(p.model.menuData) {
        return
    }

    var pageDelta = p.curSongIndex / p.model.menuPageSize - (p.model.menuCurPage - 1)
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
func (p *Player) PlaySong(songId int64) error {
    p.LocatePlayingSong()
    urlService := service.SongUrlService{}
    urlService.ID = strconv.FormatInt(songId, 10)
    urlService.Br = constants.PlayerSongBr
    code, response := urlService.SongUrl()
    if code != 200 {
        return errors.New(response)
    }

    url, err := jsonparser.GetString([]byte(response), "data", "[0]", "url")
    if err != nil {
       p.NextSong()
       return nil
    }

    p.Player.Play(utils.Mp3, url)

    return nil
}

// NextSong 下一曲
func (p *Player) NextSong() {
    if len(p.playlist) == 0 || p.curSongIndex >= len(p.playlist) - 1 {
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
    if p.curSongIndex > len(p.playlist) - 1 {
        p.curSongIndex = 0
    }

    if p.curSongIndex > len(p.playlist) - 1 {
        return
    }
    song := p.playlist[p.curSongIndex]
    _ = p.PlaySong(song.Id)
}

// Close 关闭
func (p *Player) Close() {
    p.Player.Close()
}
