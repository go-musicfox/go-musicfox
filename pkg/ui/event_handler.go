package ui

import (
	"time"

	"github.com/anhoder/foxful-cli/model"
	tea "github.com/charmbracelet/bubbletea"
	playerpkg "github.com/go-musicfox/go-musicfox/pkg/player"
	"github.com/go-musicfox/go-musicfox/pkg/structs"
)

type EventHandler struct {
	netease *Netease
}

func NewEventHandler(netease *Netease) *EventHandler {
	return &EventHandler{
		netease: netease,
	}
}

func (h *EventHandler) KeyMsgHandle(msg tea.KeyMsg, a *model.App) (bool, model.Page, tea.Cmd) {
	var (
		key    = msg.String()
		player = h.netease.player
		main   = a.MustMain()
		menu   = main.CurMenu()
	)
	switch key {
	case "enter":
		return h.enterKeyHandle()
	case "c", "C":
		if _, ok := menu.(*CurPlaylist); !ok {
			var subTitle string
			if !player.playlistUpdateAt.IsZero() {
				subTitle = player.playlistUpdateAt.Format("[更新于2006-01-02 15:04:05]")
			}
			main.EnterMenu(NewCurPlaylist(newBaseMenu(h.netease), player.playlist), &model.MenuItem{Title: "当前播放列表", Subtitle: subTitle})
			player.LocatePlayingSong()
		}
	case " ", "　":
		h.spaceKeyHandle()
	case "v":
		player.Seek(player.PassedTime() + time.Second*5)
	case "V":
		player.Seek(player.PassedTime() + time.Second*10)
	case "x":
		player.Seek(player.PassedTime() - time.Second*1)
	case "X":
		player.Seek(player.PassedTime() - time.Second*5)
	case "[", "【":
		player.PreviousSong(true)
	case "]", "】":
		player.NextSong(true)
	case "p":
		player.SetPlayMode(0)
	case "P":
		newPage := player.Intelligence(false)
		return true, newPage, a.Tick(time.Nanosecond)
	case ",", "，":
		newPage := likePlayingSong(h.netease, true)
		return true, newPage, a.Tick(time.Nanosecond)
	case ".", "。":
		newPage := likePlayingSong(h.netease, false)
		return true, newPage, a.Tick(time.Nanosecond)
	case "w", "W":
		logout()
		return true, main, tea.Quit
	case "-", "−", "ー": // half-width, full-width and katakana
		player.DownVolume()
	case "=", "＝":
		player.UpVolume()
	case "d":
		downloadPlayingSong(h.netease)
	case "D":
		downloadSelectedSong(h.netease)
	case "t":
		// trash playing song
		newPage := trashPlayingSong(h.netease)
		return true, newPage, a.Tick(time.Nanosecond)
	case "T":
		// trash selected song
		newPage := trashSelectedSong(h.netease)
		return true, newPage, a.Tick(time.Nanosecond)
	case "<", "〈", "＜", "《", "«": // half-width, full-width, Japanese, Chinese and French
		// like selected song
		newPage := likeSelectedSong(h.netease, true)
		return true, newPage, a.Tick(time.Nanosecond)
	case ">", "〉", "＞", "》", "»":
		// unlike selected song
		newPage := likeSelectedSong(h.netease, false)
		return true, newPage, a.Tick(time.Nanosecond)
	case "?", "？":
		// 帮助
		main.EnterMenu(NewHelpMenu(newBaseMenu(h.netease)), &model.MenuItem{Title: "帮助"})
	case "tab":
		newPage := openAddSongToUserPlaylistMenu(h.netease, true, true)
		return true, newPage, a.Tick(time.Nanosecond)
	case "shift+tab":
		newPage := openAddSongToUserPlaylistMenu(h.netease, true, false)
		return true, newPage, a.Tick(time.Nanosecond)
	case "`":
		newPage := openAddSongToUserPlaylistMenu(h.netease, false, true)
		return true, newPage, a.Tick(time.Nanosecond)
	case "~", "～":
		newPage := openAddSongToUserPlaylistMenu(h.netease, false, false)
		return true, newPage, a.Tick(time.Nanosecond)
	case "a":
		// 当前歌曲所属专辑
		albumOfPlayingSong(h.netease)
	case "A":
		// 选中歌曲所属专辑
		albumOfSelectedSong(h.netease)
	case "s":
		// 当前歌曲所属歌手
		artistOfPlayingSong(h.netease)
	case "S":
		// 选中歌曲所属歌手
		artistOfSelectedSong(h.netease)
	case "o":
		// 网页打开当前歌曲
		openPlayingSongInWeb(h.netease)
	case "O":
		// 网页打开选中项
		openSelectedItemInWeb(h.netease)
	case ";", ":", "：", "；":
		// 收藏选中歌单
		newPage := collectSelectedPlaylist(h.netease, true)
		return true, newPage, a.Tick(time.Nanosecond)
	case "'", "\"":
		// 取消收藏选中歌单
		newPage := collectSelectedPlaylist(h.netease, false)
		return true, newPage, a.Tick(time.Nanosecond)
	case "\\", "、":
		// 从播放列表删除歌曲,仅在当前播放列表界面有效
		newPage := delSongFromPlaylist(h.netease)
		return true, newPage, a.Tick(time.Nanosecond)
	case "e":
		// 追加到下一曲播放
		addSongToPlaylist(h.netease, true)
	case "E":
		// 追加到播放列表末尾
		addSongToPlaylist(h.netease, false)
	case "h", "U":
		// 清除歌曲缓存
		clearSongCache(h.netease)
	case "r", "R":
		// rerender
		return true, main, a.RerenderCmd(true)
	default:
		return false, nil, nil
	}

	return true, nil, nil
}

func (h *EventHandler) enterKeyHandle() (stopPropagation bool, newPage model.Page, cmd tea.Cmd) {
	loading := NewLoading(h.netease)
	loading.start()
	defer loading.complete()

	var menu = h.netease.MustMain().CurMenu()
	if _, ok := menu.(*AddToUserPlaylistMenu); ok {
		addSongToUserPlaylist(h.netease, menu.(*AddToUserPlaylistMenu).action)
		return true, h.netease.MustMain(), h.netease.Tick(time.Nanosecond)
	}
	return false, nil, nil
}

// 空格监听
func (h *EventHandler) spaceKeyHandle() {
	var (
		songs         []structs.Song
		inPlayingMenu = h.netease.player.InPlayingMenu()
		main          = h.netease.MustMain()
		menu          = main.CurMenu()
		player        = h.netease.player
	)
	if me, ok := menu.(SongsMenu); ok {
		songs = me.Songs()
	}

	selectedIndex := menu.RealDataIndex(main.SelectedIndex())
	if me, ok := menu.(Menu); !ok || !me.IsPlayable() || len(songs) == 0 || selectedIndex > len(songs)-1 {
		if player.curSongIndex > len(player.playlist)-1 {
			return
		}
		switch player.State() {
		case playerpkg.Paused:
			h.netease.player.Resume()
		case playerpkg.Playing:
			h.netease.player.Paused()
		case playerpkg.Stopped:
			_ = player.PlaySong(player.playlist[player.curSongIndex], DurationNext)
		}
		return
	}

	if inPlayingMenu && songs[selectedIndex].Id == player.playlist[player.curSongIndex].Id {
		switch player.State() {
		case playerpkg.Paused:
			player.Resume()
		case playerpkg.Playing:
			player.Paused()
		}
		return
	}

	player.curSongIndex = selectedIndex
	player.playingMenuKey = menu.GetMenuKey()
	if me, ok := menu.(Menu); ok {
		player.playingMenu = me
	}

	var newPlaylists = make([]structs.Song, len(songs))
	copy(newPlaylists, songs)
	player.playlist = newPlaylists

	player.playlistUpdateAt = time.Now()
	if player.mode == playerpkg.PmIntelligent {
		player.SetPlayMode(0)
	}
	_ = player.PlaySong(player.playlist[selectedIndex], DurationNext)
}

func (h *EventHandler) MouseMsgHandle(msg tea.MouseMsg, a *model.App) (stopPropagation bool, newPage model.Page, cmd tea.Cmd) {
	var (
		player = h.netease.player
		main   = a.MustMain()
	)
	switch msg.Type {
	case tea.MouseLeft:
		x, y := msg.X, msg.Y
		w := len(player.progressRamp)
		if y+1 == a.WindowHeight() && x+1 <= len(player.progressRamp) {
			allDuration := int(player.CurMusic().Duration.Seconds())
			if allDuration == 0 {
				return true, main, nil
			}
			duration := float64(x) * player.CurMusic().Duration.Seconds() / float64(w)
			player.Seek(time.Second * time.Duration(duration))
			if player.State() != playerpkg.Playing {
				player.Resume()
			}
		}
	case tea.MouseWheelDown:
		player.DownVolume()
	case tea.MouseWheelUp:
		player.UpVolume()
	}

	return true, main, a.Tick(time.Nanosecond)
}
