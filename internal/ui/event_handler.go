package ui

import (
	"time"

	"github.com/anhoder/foxful-cli/model"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/internal/types"
)

type OperateType string

const (
	OperateTypeEnter                              = "enter"
	OperateTypeCurPlaylist                        = "curPlaylist"
	OperateTypeSpace                              = "space"
	OperateTypeToggle                             = "toggle"
	OperateTypeForwardFiveSec                     = "forwardFiveSec"
	OperateTypeForwardTenSec                      = "forwardTenSec"
	OperateTypeBackwardOneSec                     = "backwardOneSec"
	OperateTypeBackwardFiveSec                    = "backwardFiceSec"
	OperateTypePrevious                           = "previous"
	OperateTypeNext                               = "next"
	OperateTypeSwitchPlayMode                     = "switchPlayMode"
	OperateTypeIntelligence                       = "intelligence"
	OperateTypeLikePlayingSong                    = "likePlayingSong"
	OperateTypeLikeSelectedSong                   = "likeSelectedSong"
	OperateTypeDislikePlayingSong                 = "dislikePlayingSong"
	OperateTypeDislikeSelectedSong                = "dislikeSelectedSong"
	OperateTypeLogout                             = "logout"
	OperateTypeDownVolume                         = "downVolume"
	OperateTypeUpVolume                           = "upVolume"
	OperateTypeDownloadPlayigSong                 = "downloadPlayingSong"
	OperateTypeDownloadSelectedSong               = "downloadSelectedSong"
	OperateTypeTrashPlayingSong                   = "trashPlayingSong"
	OperateTypeTrashSelectedSong                  = "trashSelectedSong"
	OperateTypeHelp                               = "help"
	OperateTypeAddSelectedSongToUserPlaylist      = "addSelectedSongToUserPlaylist"
	OperateTypeRemoveSelectedSongFromUserPlaylist = "removeSelectedSongFromUserPlaylist"
	OperateTypeAddPlayingSongToUserPlaylist       = "addPlayingSongToUserPlaylist"
	OperateTypeRemovePlayingSongFromUserPlaylist  = "removePlayingSongFromUserPlaylist"
	OperateTypeOpenAlbumOfPlayingSong             = "openAlbumOfPlayingSong"
	OperateTypeOpenAlbumOfSelectedSong            = "openAlbumOfSelectedSong"
	OperateTypeOpenArtistOfPlayingSong            = "openArtistOfPlayingSong"
	OperateTypeOpenArtistOfSelectedSong           = "openArtistOfSelectedSong"
	OperateTypeOpenPlayingSongInWeb               = "openPlayingSongInWeb"
	OperateTypeOpenSelectedItemInWeb              = "openSelectedItemInWeb"
	OperateTypeCollectSelectedPlaylist            = "collectSelectedPlaylist"
	OperateTypeDiscollectSelectedPlaylist         = "discollectSelectedPlaylist"
	OperateTypeDelSongFromCurPlaylist             = "delSongFromCurPlaylist"
	OperateTypeAddSongToNext                      = "addSongToNext"
	OperateTypeAppendSongToCurPlaylist            = "appendSongToCurPlaylist"
	OperateTypeClearSongCache                     = "clearSongCache"
	OperateTypeRerender                           = "rerender"
)

type EventHandler struct {
	netease *Netease
}

func NewEventHandler(netease *Netease) *EventHandler {
	return &EventHandler{
		netease: netease,
	}
}

var keyOperateMapping = map[string]OperateType{
	"enter":     OperateTypeEnter,
	"c":         OperateTypeCurPlaylist,
	"C":         OperateTypeCurPlaylist,
	" ":         OperateTypeSpace,
	"　":         OperateTypeSpace,
	"v":         OperateTypeForwardFiveSec,
	"V":         OperateTypeForwardTenSec,
	"x":         OperateTypeBackwardOneSec,
	"X":         OperateTypeBackwardFiveSec,
	"[":         OperateTypePrevious,
	"【":         OperateTypePrevious,
	"]":         OperateTypeNext,
	"】":         OperateTypeNext,
	"p":         OperateTypeSwitchPlayMode,
	"P":         OperateTypeIntelligence,
	",":         OperateTypeLikePlayingSong,
	"，":         OperateTypeLikePlayingSong,
	".":         OperateTypeDislikePlayingSong,
	"。":         OperateTypeDislikePlayingSong,
	"w":         OperateTypeLogout,
	"W":         OperateTypeLogout,
	"=":         OperateTypeUpVolume,
	"＝":         OperateTypeUpVolume,
	"-":         OperateTypeDownVolume,
	"−":         OperateTypeDownVolume,
	"ー":         OperateTypeDownVolume,
	"d":         OperateTypeDownloadPlayigSong,
	"D":         OperateTypeDownloadSelectedSong,
	"t":         OperateTypeTrashPlayingSong,
	"T":         OperateTypeTrashSelectedSong,
	"<":         OperateTypeLikeSelectedSong,
	"〈":         OperateTypeLikeSelectedSong,
	"＜":         OperateTypeLikeSelectedSong,
	"《":         OperateTypeLikeSelectedSong,
	"«":         OperateTypeLikeSelectedSong,
	">":         OperateTypeDislikeSelectedSong,
	"〉":         OperateTypeDislikeSelectedSong,
	"＞":         OperateTypeDislikeSelectedSong,
	"》":         OperateTypeDislikeSelectedSong,
	"»":         OperateTypeDislikeSelectedSong,
	"?":         OperateTypeHelp,
	"？":         OperateTypeHelp,
	"tab":       OperateTypeAddSelectedSongToUserPlaylist,
	"shift+tab": OperateTypeRemoveSelectedSongFromUserPlaylist,
	"`":         OperateTypeAddPlayingSongToUserPlaylist,
	"~":         OperateTypeRemovePlayingSongFromUserPlaylist,
	"～":         OperateTypeRemovePlayingSongFromUserPlaylist,
	"a":         OperateTypeOpenAlbumOfPlayingSong,
	"A":         OperateTypeOpenAlbumOfSelectedSong,
	"s":         OperateTypeOpenArtistOfPlayingSong,
	"S":         OperateTypeOpenArtistOfSelectedSong,
	"o":         OperateTypeOpenPlayingSongInWeb,
	"O":         OperateTypeOpenSelectedItemInWeb,
	";":         OperateTypeCollectSelectedPlaylist,
	":":         OperateTypeCollectSelectedPlaylist,
	"：":         OperateTypeCollectSelectedPlaylist,
	"；":         OperateTypeCollectSelectedPlaylist,
	"'":         OperateTypeDiscollectSelectedPlaylist,
	"\"":        OperateTypeDiscollectSelectedPlaylist,
	"\\":        OperateTypeDelSongFromCurPlaylist,
	"、":         OperateTypeDelSongFromCurPlaylist,
	"e":         OperateTypeAddSongToNext,
	"E":         OperateTypeAppendSongToCurPlaylist,
	"u":         OperateTypeClearSongCache,
	"U":         OperateTypeClearSongCache,
	"r":         OperateTypeRerender,
	"R":         OperateTypeRerender,
}

func (h *EventHandler) KeyMsgHandle(msg tea.KeyMsg, _ *model.App) (bool, model.Page, tea.Cmd) {
	if ot, ok := keyOperateMapping[msg.String()]; ok {
		return h.handle(ot)
	}

	return false, nil, nil
}

func (h *EventHandler) handle(ot OperateType) (bool, model.Page, tea.Cmd) {
	var (
		player = h.netease.player
		app    = h.netease.App
		main   = app.MustMain()
		menu   = main.CurMenu()
	)
	switch ot {
	case OperateTypeEnter:
		return h.enterKeyHandle()
	case OperateTypeCurPlaylist:
		if _, ok := menu.(*CurPlaylist); !ok {
			var subTitle string
			if !player.playlistUpdateAt.IsZero() {
				subTitle = player.playlistUpdateAt.Format("[更新于2006-01-02 15:04:05]")
			}
			main.EnterMenu(NewCurPlaylist(newBaseMenu(h.netease), player.playlist), &model.MenuItem{Title: "当前播放列表", Subtitle: subTitle})
			player.LocatePlayingSong()
		}
	case OperateTypeSpace:
		h.spaceKeyHandle()
	case OperateTypeToggle:
		if h.netease.player.State() == types.Stopped {
			h.spaceKeyHandle()
		} else {
			h.netease.player.Toggle()
		}
	case OperateTypeForwardFiveSec:
		player.Seek(player.PassedTime() + time.Second*5)
	case OperateTypeForwardTenSec:
		player.Seek(player.PassedTime() + time.Second*10)
	case OperateTypeBackwardOneSec:
		player.Seek(player.PassedTime() - time.Second*1)
	case OperateTypeBackwardFiveSec:
		player.Seek(player.PassedTime() - time.Second*5)
	case OperateTypePrevious:
		player.PreviousSong(true)
	case OperateTypeNext:
		player.NextSong(true)
	case OperateTypeSwitchPlayMode:
		player.SetPlayMode(0)
	case OperateTypeIntelligence:
		newPage := player.Intelligence(false)
		return true, newPage, app.Tick(time.Nanosecond)
	case OperateTypeLikePlayingSong:
		newPage := likePlayingSong(h.netease, true)
		return true, newPage, app.Tick(time.Nanosecond)
	case OperateTypeDislikePlayingSong:
		newPage := likePlayingSong(h.netease, false)
		return true, newPage, app.Tick(time.Nanosecond)
	case OperateTypeLogout:
		logout()
		return true, nil, tea.Quit
	case OperateTypeDownVolume: // half-width, full-width and katakana
		player.DownVolume()
	case OperateTypeUpVolume:
		player.UpVolume()
	case OperateTypeDownloadPlayigSong:
		downloadPlayingSong(h.netease)
	case OperateTypeDownloadSelectedSong:
		downloadSelectedSong(h.netease)
	case OperateTypeTrashPlayingSong:
		// trash playing song
		newPage := trashPlayingSong(h.netease)
		return true, newPage, app.Tick(time.Nanosecond)
	case OperateTypeTrashSelectedSong:
		// trash selected song
		newPage := trashSelectedSong(h.netease)
		return true, newPage, app.Tick(time.Nanosecond)
	case OperateTypeLikeSelectedSong: // half-width, full-width, Japanese, Chinese and French
		// like selected song
		newPage := likeSelectedSong(h.netease, true)
		return true, newPage, app.Tick(time.Nanosecond)
	case OperateTypeDislikeSelectedSong:
		// unlike selected song
		newPage := likeSelectedSong(h.netease, false)
		return true, newPage, app.Tick(time.Nanosecond)
	case OperateTypeHelp:
		// 帮助
		main.EnterMenu(NewHelpMenu(newBaseMenu(h.netease)), &model.MenuItem{Title: "帮助"})
	case OperateTypeAddSelectedSongToUserPlaylist:
		newPage := openAddSongToUserPlaylistMenu(h.netease, true, true)
		return true, newPage, app.Tick(time.Nanosecond)
	case OperateTypeRemoveSelectedSongFromUserPlaylist:
		newPage := openAddSongToUserPlaylistMenu(h.netease, true, false)
		return true, newPage, app.Tick(time.Nanosecond)
	case OperateTypeAddPlayingSongToUserPlaylist:
		newPage := openAddSongToUserPlaylistMenu(h.netease, false, true)
		return true, newPage, app.Tick(time.Nanosecond)
	case OperateTypeRemovePlayingSongFromUserPlaylist:
		newPage := openAddSongToUserPlaylistMenu(h.netease, false, false)
		return true, newPage, app.Tick(time.Nanosecond)
	case OperateTypeOpenAlbumOfPlayingSong:
		// 当前歌曲所属专辑
		albumOfPlayingSong(h.netease)
	case OperateTypeOpenAlbumOfSelectedSong:
		// 选中歌曲所属专辑
		albumOfSelectedSong(h.netease)
	case OperateTypeOpenArtistOfPlayingSong:
		// 当前歌曲所属歌手
		artistOfPlayingSong(h.netease)
	case OperateTypeOpenArtistOfSelectedSong:
		// 选中歌曲所属歌手
		artistOfSelectedSong(h.netease)
	case OperateTypeOpenPlayingSongInWeb:
		// 网页打开当前歌曲
		openPlayingSongInWeb(h.netease)
	case OperateTypeOpenSelectedItemInWeb:
		// 网页打开选中项
		openSelectedItemInWeb(h.netease)
	case OperateTypeCollectSelectedPlaylist:
		// 收藏选中歌单
		newPage := collectSelectedPlaylist(h.netease, true)
		return true, newPage, app.Tick(time.Nanosecond)
	case OperateTypeDiscollectSelectedPlaylist:
		// 取消收藏选中歌单
		newPage := collectSelectedPlaylist(h.netease, false)
		return true, newPage, app.Tick(time.Nanosecond)
	case OperateTypeDelSongFromCurPlaylist:
		// 从播放列表删除歌曲,仅在当前播放列表界面有效
		newPage := delSongFromPlaylist(h.netease)
		return true, newPage, app.Tick(time.Nanosecond)
	case OperateTypeAddSongToNext:
		// 追加到下一曲播放
		addSongToPlaylist(h.netease, true)
	case OperateTypeAppendSongToCurPlaylist:
		// 追加到播放列表末尾
		addSongToPlaylist(h.netease, false)
	case OperateTypeClearSongCache:
		// 清除歌曲缓存
		clearSongCache(h.netease)
	case OperateTypeRerender:
		// rerender
		return true, main, app.RerenderCmd(true)
	default:
		return false, nil, nil
	}

	return true, nil, nil
}

func (h *EventHandler) enterKeyHandle() (stopPropagation bool, newPage model.Page, cmd tea.Cmd) {
	loading := model.NewLoading(h.netease.MustMain())
	loading.Start()
	defer loading.Complete()

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
		isSameSong    bool
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
		case types.Paused:
			h.netease.player.Resume()
		case types.Playing:
			h.netease.player.Paused()
		case types.Stopped:
			_ = player.PlaySong(player.playlist[player.curSongIndex], DurationNext)
		}
		return
	}

	if isSameSong = songs[selectedIndex].Id == player.playlist[player.curSongIndex].Id; inPlayingMenu && isSameSong {
		switch player.State() {
		case types.Paused:
			player.Resume()
		case types.Playing:
			player.Paused()
		case types.Stopped:
			_ = player.PlaySong(player.playlist[player.curSongIndex], DurationNext)
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
	if player.mode == types.PmIntelligent {
		player.SetPlayMode(0)
	}
	if !isSameSong {
		_ = player.PlaySong(player.playlist[selectedIndex], DurationNext)
	}
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
			if player.State() != types.Playing {
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
