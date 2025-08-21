package ui

import (
	"fmt"
	"log/slog"
	"slices"
	"time"

	"github.com/anhoder/foxful-cli/model"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/keybindings"
	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/internal/types"
	"github.com/go-musicfox/go-musicfox/utils/mathx"
)

type EventHandler struct {
	netease         *Netease
	keyToOperateMap map[string]keybindings.OperateType // KeyStr -> OperateType
}

func NewEventHandler(netease *Netease) *EventHandler {
	handler := &EventHandler{
		netease:         netease,
		keyToOperateMap: keybindings.BuildKeyToOperateTypeMap(configs.ConfigRegistry.Keybindings),
	}
	slog.Info(fmt.Sprintf("事件处理器已初始化，加载了 %d 个有效按键绑定映射。", len(handler.keyToOperateMap)))
	return handler
}

func (h *EventHandler) KeyMsgHandle(msg tea.KeyMsg, _ *model.App) (bool, model.Page, tea.Cmd) {
	if op, ok := h.keyToOperateMap[msg.String()]; ok {
		stopPropagation, newPage, cmd := h.handle(op)
		return stopPropagation, newPage, cmd
	}
	return false, nil, nil // false: 允许 foxful-cli 处理
}

func (h *EventHandler) handle(op keybindings.OperateType) (bool, model.Page, tea.Cmd) {
	var (
		player = h.netease.player
		app    = h.netease.App
		main   = app.MustMain()
		menu   = main.CurMenu()
	)

	switch op {
	case keybindings.OpEnter:
		return h.enterKeyHandle()
	case keybindings.OpCurPlaylist:
		if _, ok := menu.(*CurPlaylist); !ok {
			var subTitle string
			if !player.playlistUpdateAt.IsZero() {
				subTitle = player.playlistUpdateAt.Format("[更新于2006-01-02 15:04:05]")
			}
			main.EnterMenu(NewCurPlaylist(newBaseMenu(h.netease), player.Playlist()), &model.MenuItem{Title: "当前播放列表", Subtitle: subTitle})
			player.LocatePlayingSong()
		}
	case keybindings.OpPlayOrToggle:
		h.playOrToggleHandle()
	case keybindings.OpToggle:
		if player.State() == types.Stopped {
			h.playOrToggleHandle()
		} else {
			player.Toggle()
		}
	case keybindings.OpSeekForward5s:
		player.Seek(player.PassedTime() + time.Second*5)
	case keybindings.OpSeekForward10s:
		player.Seek(player.PassedTime() + time.Second*10)
	case keybindings.OpSeekBackward1s:
		player.Seek(player.PassedTime() - time.Second*1)
	case keybindings.OpSeekBackward5s:
		player.Seek(player.PassedTime() - time.Second*5)
	case keybindings.OpPrevious:
		player.PreviousSong(true)
	case keybindings.OpNext:
		player.NextSong(true)
	case keybindings.OpSwitchPlayMode:
		player.SwitchMode()
	case keybindings.OpIntelligence:
		newPage := player.Intelligence(false)
		return true, newPage, app.Tick(time.Nanosecond)
	case keybindings.OpLikePlayingSong:
		newPage := likePlayingSong(h.netease, true)
		return true, newPage, app.Tick(time.Nanosecond)
	case keybindings.OpDislikePlayingSong:
		newPage := likePlayingSong(h.netease, false)
		return true, newPage, app.Tick(time.Nanosecond)
	case keybindings.OpLogout:
		logout()
		return true, nil, tea.Quit
	case keybindings.OpVolumeDown: // half-width, full-width and katakana
		player.DownVolume()
	case keybindings.OpVolumeUp:
		player.UpVolume()
	case keybindings.OpDownloadPlayingSong:
		downloadPlayingSong(h.netease)
	case keybindings.OpDownloadSelectedSong:
		downloadSelectedSong(h.netease)
	case keybindings.OpTrashPlayingSong:
		// trash playing song
		newPage := trashPlayingSong(h.netease)
		return true, newPage, app.Tick(time.Nanosecond)
	case keybindings.OpTrashSelectedSong:
		// trash selected song
		newPage := trashSelectedSong(h.netease)
		return true, newPage, app.Tick(time.Nanosecond)
	case keybindings.OpLikeSelectedSong: // half-width, full-width, Japanese, Chinese and French
		// like selected song
		newPage := likeSelectedSong(h.netease, true)
		return true, newPage, app.Tick(time.Nanosecond)
	case keybindings.OpDislikeSelectedSong:
		// unlike selected song
		newPage := likeSelectedSong(h.netease, false)
		return true, newPage, app.Tick(time.Nanosecond)
	case keybindings.OpHelp:
		// 帮助
		main.EnterMenu(NewHelpMenu(newBaseMenu(h.netease)), &model.MenuItem{Title: "帮助"})
	case keybindings.OpAddSelectedToUserPlaylist:
		newPage := openAddSongToUserPlaylistMenu(h.netease, true, true)
		return true, newPage, app.Tick(time.Nanosecond)
	case keybindings.OpRemoveSelectedFromUserPlaylist:
		newPage := openAddSongToUserPlaylistMenu(h.netease, true, false)
		return true, newPage, app.Tick(time.Nanosecond)
	case keybindings.OpAddPlayingToUserPlaylist:
		newPage := openAddSongToUserPlaylistMenu(h.netease, false, true)
		return true, newPage, app.Tick(time.Nanosecond)
	case keybindings.OpRemovePlayingFromUserPlaylist:
		newPage := openAddSongToUserPlaylistMenu(h.netease, false, false)
		return true, newPage, app.Tick(time.Nanosecond)
	case keybindings.OpOpenSimiSongsOfPlayingSong:
		// 与当前歌曲相似的歌曲
		simiSongsOfPlayingSong(h.netease)
	case keybindings.OpOpenSimiSongsOfSelectedSong:
		// 与当前选中歌曲相似的歌曲
		simiSongsOfSelectedSong(h.netease)
	case keybindings.OpAlbumOfPlayingSong:
		// 当前歌曲所属专辑
		albumOfPlayingSong(h.netease)
	case keybindings.OpAlbumOfSelectedSong:
		// 选中歌曲所属专辑
		albumOfSelectedSong(h.netease)
	case keybindings.OpArtistOfPlayingSong:
		// 当前歌曲所属歌手
		artistOfPlayingSong(h.netease)
	case keybindings.OpArtistOfSelectedSong:
		// 选中歌曲所属歌手
		artistOfSelectedSong(h.netease)
	case keybindings.OpOpenPlayingSongInWeb:
		// 网页打开当前歌曲
		openPlayingSongInWeb(h.netease)
	case keybindings.OpOpenSelectedItemInWeb:
		// 网页打开选中项
		openSelectedItemInWeb(h.netease)
	case keybindings.OpCollectSelectedPlaylist:
		// 收藏选中歌单
		newPage := collectSelectedPlaylist(h.netease, true)
		return true, newPage, app.Tick(time.Nanosecond)
	case keybindings.OpDiscollectSelectedPlaylist:
		// 取消收藏选中歌单
		newPage := collectSelectedPlaylist(h.netease, false)
		return true, newPage, app.Tick(time.Nanosecond)
	case keybindings.OpDeleteSongFromPlaylist:
		// 从播放列表删除歌曲,仅在当前播放列表界面有效
		newPage := delSongFromPlaylist(h.netease)
		return true, newPage, app.Tick(time.Nanosecond)
	case keybindings.OpAppendSongsToNext:
		// 追加到下一曲播放
		appendSongsToCurPlaylist(h.netease, true)
	case keybindings.OpAppendSongsToEnd:
		// 追加到播放列表末尾
		appendSongsToCurPlaylist(h.netease, false)
	case keybindings.OpClearSongCache:
		// 清除歌曲缓存
		clearSongCache(h.netease)
	case keybindings.OpRerenderUI:
		// rerender
		return true, main, app.RerenderCmd(true)
	case keybindings.OpPageDown:
		oldPage := main.CurPage()
		main.NextPage()
		if oldPage != main.CurPage() {
			curIndex := mathx.Min(main.SelectedIndex()+main.PageSize(), len(menu.MenuViews())-1)
			main.SetSelectedIndex(curIndex)
		}
	case keybindings.OpPageUp:
		oldPage := main.CurPage()
		main.PrePage()
		if oldPage != main.CurPage() {
			curIndex := mathx.Max(main.SelectedIndex()-main.PageSize(), 0)
			main.SetSelectedIndex(curIndex)
		}
	case keybindings.OpDownloadPlayingSongLrc:
		downloadPlayingSongLrc(h.netease)
	case keybindings.OpActionOfSelected:
		action(h.netease, false)
	case keybindings.OpActionOfPlayingSong:
		action(h.netease, true)
	case keybindings.OpSharePlayingItem:
		shareSelectItem(h.netease)
	case keybindings.OpShareSelectItem:
		shareSelectItem(h.netease)
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

func (h *EventHandler) playOrToggleHandle() {
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
	_ = player.playlistManager.Initialize(selectedIndex, newPlaylist)

	player.playingMenuKey = menu.GetMenuKey()
	if me, ok := menu.(Menu); ok {
		player.playingMenu = me
	}

	// 如果当前不是心动模式菜单
	mode := player.Mode()
	if !inPlayingMenu && mode == types.PmIntelligent && player.playingMenuKey != "Intelligent" {
		player.SetMode(types.PmListLoop)
	}

	player.playlistUpdateAt = time.Now()
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
