package ui

import (
	"errors"
	"github.com/anhoder/netease-music/service"
	"github.com/buger/jsonparser"
	"github.com/mattn/go-runewidth"
	"github.com/muesli/termenv"
	"go-musicfox/constants"
	"go-musicfox/ds"
	"go-musicfox/lyric"
	"go-musicfox/utils"
	"strconv"
	"strings"
)

// PlayDuration 下首歌的方向
type PlayDuration uint8

const (
	DurationNext PlayDuration = iota
	DurationPrev
)

type Player struct {
	model *NeteaseModel

	playlist       []ds.Song
	curSongIndex   int
	playingMenuKey string // 正在播放的菜单Key
	isIntelligence bool   // 智能模式（心动模式）

	lrcTimer      *lyric.LRCTimer
	lyrics        [5]string
	showLyric     bool
	lyricStartRow int
	lyricLines    int

	playErrCount int // 错误计数，当错误连续超过5次，停止播放

	*utils.Player
}

func NewPlayer(model *NeteaseModel) *Player {
	player := &Player{
		model:  model,
		Player: utils.NewPlayer(),
	}

	// done监听
	go func() {
		for {
			<-player.Player.Done()
			player.NextSong()
			model.Rerender()
		}
	}()

	return player
}

func (p *Player) playerView(top *int) string {
	var playerBuilder strings.Builder

	playerBuilder.WriteString(p.LyricView())
	*top = p.model.WindowHeight-4

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
func (p *Player) PlaySong(songId int64, duration PlayDuration) error {
	loading := NewLoading(p.model)
	loading.start()
	defer loading.complete()

	p.LocatePlayingSong()
	urlService := service.SongUrlService{}
	urlService.ID = strconv.FormatInt(songId, 10)
	urlService.Br = constants.PlayerSongBr
	code, response := urlService.SongUrl()
	if code != 200 {
		return errors.New(string(response))
	}

	url, err1 := jsonparser.GetString(response, "data", "[0]", "url")
	musicType, err2 := jsonparser.GetString(response, "data", "[0]", "type")
	if err1 != nil || err2 != nil || (musicType != "mp3" && musicType != "flac") {
		p.State = utils.Stopped
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

	go p.updateLyric(songId)

	switch musicType {
	case "mp3":
		p.Player.Play(utils.Mp3, url)
	case "flac":
		p.Player.Play(utils.Flac, url)
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
	_ = p.PlaySong(song.Id, DurationNext)
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
	_ = p.PlaySong(song.Id, DurationPrev)
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

	p.model.Rerender()
}

func (p *Player) updateLyric(songId int64) {
	p.lyrics = [5]string{}
	if p.lrcTimer != nil {
		p.lrcTimer.Stop()
	}
	lrcFile, _ := lyric.ReadLRC(strings.NewReader("[00:00.00] 暂无歌词～"))
	defer func() {
		p.lrcTimer = lyric.NewLRCTimer(lrcFile)
		p.lrcTimer.AddListener(p.lyricListener)
		p.lrcTimer.SetTimer(p.Timer())
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
