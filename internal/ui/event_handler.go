package ui

import (
	"slices"
	"time"

	"github.com/anhoder/foxful-cli/model"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/internal/types"
	"github.com/go-musicfox/go-musicfox/utils/mathx"
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
	OperateTypeAppendSongsToNext                  = "appendSongsToNext"
	OperateTypeAppendSongsAfterCurPlaylist        = "appendSongsAfterCurPlaylist"
	OperateTypeClearSongCache                     = "clearSongCache"
	OperateTypeRerender                           = "rerender"
	OperateTypePageDown                           = "pageDown"
	OperateTypePageUp                             = "pageUp"
	OperateTypeDownloadLrc                        = "downloadLrc"
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
	"e":         OperateTypeAppendSongsToNext,
	"E":         OperateTypeAppendSongsAfterCurPlaylist,
	"u":         OperateTypeClearSongCache,
	"U":         OperateTypeClearSongCache,
	"r":         OperateTypeRerender,
	"R":         OperateTypeRerender,
	"ctrl+d":    OperateTypePageDown,
	"ctrl+u":    OperateTypePageUp,
	"ctrl+l":    OperateTypeDownloadLrc,
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
			main.EnterMenu(NewCurPlaylist(newBaseMenu(h.netease), player.Playlist()), &model.MenuItem{Title: "当前播放列表", Subtitle: subTitle})
			player.LocatePlayingSong()
		}
	case OperateTypeSpace:
		h.spaceKeyHandle()
	case OperateTypeToggle:
		if player.State() == types.Stopped {
			h.spaceKeyHandle()
		} else {
			player.Toggle()
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
		player.SwitchMode()
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
	case OperateTypeAppendSongsToNext:
		// 追加到下一曲播放
		appendSongsToCurPlaylist(h.netease, true)
	case OperateTypeAppendSongsAfterCurPlaylist:
		// 追加到播放列表末尾
		appendSongsToCurPlaylist(h.netease, false)
	case OperateTypeClearSongCache:
		// 清除歌曲缓存
		clearSongCache(h.netease)
	case OperateTypeRerender:
		// rerender
		return true, main, app.RerenderCmd(true)
	case OperateTypePageDown:
		oldPage := main.CurPage()
		main.NextPage()
		if oldPage != main.CurPage() {
			curIndex := mathx.Min(main.SelectedIndex()+main.PageSize(), len(menu.MenuViews())-1)
			main.SetSelectedIndex(curIndex)
		}
	case OperateTypePageUp:
		oldPage := main.CurPage()
		main.PrePage()
		if oldPage != main.CurPage() {
			curIndex := mathx.Max(main.SelectedIndex()-main.PageSize(), 0)
			main.SetSelectedIndex(curIndex)
		}
	case OperateTypeDownloadLrc:
		DownLoadLrc(h.netease)
	default:
		return false, nil, nil
	}

	return true, nil, nil
}

func (h *EventHandler) enterKeyHandle() (stopPropagation bool, newPage model.Page, cmd tea.Cmd) {
	loading := model.NewLoading(h.netease.MustMain())
	loading.Start()
	defer loading.Complete()

	menu := h.netease.MustMain().CurMenu()
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
		main          = h.netease.MustMain()
		menu          = main.CurMenu()
		player        = h.netease.player
		inPlayingMenu = player.InPlayingMenu()
	)
	if me, ok := menu.(SongsMenu); ok {
		songs = me.Songs()
	}

	selectedIndex := menu.RealDataIndex(main.SelectedIndex())
	if me, ok := menu.(Menu); !ok || !me.IsPlayable() || len(songs) == 0 || selectedIndex > len(songs)-1 {
		if player.CurSongIndex() > len(player.Playlist())-1 {
			return
		}
		switch player.State() {
		case types.Paused:
			player.Resume()
		case types.Playing:
			player.Pause()
		case types.Stopped:
			player.StartPlay()
		}
		return
	}

	if inPlayingMenu && songs[selectedIndex].Id == player.CurSong().Id {
		switch player.State() {
		case types.Paused:
			player.Resume()
		case types.Playing:
			player.Pause()
		case types.Stopped:
			player.StartPlay()
		}
		return
	}

	newPlaylist := make([]structs.Song, len(songs))
	copy(newPlaylist, songs)
	player.songManager.init(selectedIndex, newPlaylist)

	player.playingMenuKey = menu.GetMenuKey()
	if me, ok := menu.(Menu); ok {
		player.playingMenu = me
	}

	player.playlistUpdateAt = time.Now()
	if player.intelligent {
		player.SetMode(player.lastMode)
	}
	player.StartPlay()
}

func (h *EventHandler) MouseMsgHandle(msg tea.MouseMsg, a *model.App) (stopPropagation bool, newPage model.Page, cmd tea.Cmd) {
	var (
		player = h.netease.player
		main   = a.MustMain()
	)
	switch {
	case msg.Button == tea.MouseButtonLeft && slices.Contains([]tea.MouseAction{tea.MouseActionPress, tea.MouseActionMotion}, msg.Action):
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
	case msg.Button == tea.MouseButtonWheelDown && msg.Action == tea.MouseActionPress:
		player.DownVolume()
	case msg.Button == tea.MouseButtonWheelUp && msg.Action == tea.MouseActionPress:
		player.UpVolume()
	case msg.Button == tea.MouseButtonWheelLeft && msg.Action == tea.MouseActionPress:
		player.PreviousSong(true)
	case msg.Button == tea.MouseButtonWheelRight && msg.Action == tea.MouseActionPress:
		player.NextSong(true)
	}

	return true, main, a.Tick(time.Nanosecond)
}
