package ui

import (
	"fmt"
	"log/slog"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/anhoder/foxful-cli/model"
	"github.com/anhoder/foxful-cli/style"
	"github.com/mattn/go-runewidth"

	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/keybindings"
	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/internal/types"
	"github.com/go-musicfox/go-musicfox/utils/likelist"
	"github.com/go-musicfox/go-musicfox/utils/mathx"
)

type EventHandler struct {
	netease         *Netease
	keyToOperateMap map[string]keybindings.OperateType // KeyStr -> OperateType
	mouseVolumeStep int
}

func NewEventHandler(netease *Netease) *EventHandler {
	step := configs.AppConfig.Player.MouseVolumeStep
	if step <= 0 {
		step = 1
	} else if step > 20 {
		step = 20
	}
	handler := &EventHandler{
		netease:         netease,
		mouseVolumeStep: step,
		keyToOperateMap: keybindings.BuildKeyToOperateTypeMap(configs.EffectiveKeybindings),
	}
	slog.Info(fmt.Sprintf("事件处理器已初始化，加载了 %d 个有效按键绑定映射。", len(handler.keyToOperateMap)))
	return handler
}

// KeyMsgHandle 处理键盘事件
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
			main.EnterMenu(NewCurPlaylist(newBaseMenu(h.netease), player.Playlist()), &model.MenuItem{Title: model.T(MsgMenuCurrentPlaylist), Subtitle: subTitle})
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
		newTime := player.PassedTime() - time.Second*1
		if newTime < 0 {
			newTime = 0
		}
		player.Seek(newTime)
	case keybindings.OpSeekBackward5s:
		newTime := player.PassedTime() - time.Second*5
		if newTime < 0 {
			newTime = 0
		}
		player.Seek(newTime)
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
		newPage := likeSong(h.netease, true, false)
		return true, newPage, app.Tick(time.Nanosecond)
	case keybindings.OpDislikePlayingSong:
		newPage := likeSong(h.netease, false, false)
		return true, newPage, app.Tick(time.Nanosecond)
	case keybindings.OpLogout:
		logout()
		return true, nil, tea.Quit
	case keybindings.OpVolumeDown: // half-width, full-width and katakana
		player.DownVolume()
	case keybindings.OpVolumeUp:
		player.UpVolume()
	case keybindings.OpDownloadPlayingSong:
		downloadSong(h.netease, false)
	case keybindings.OpDownloadSelectedSong:
		downloadSong(h.netease, true)
	case keybindings.OpTrashPlayingSong:
		// trash playing song
		newPage := trashSong(h.netease, false)
		return true, newPage, app.Tick(time.Nanosecond)
	case keybindings.OpTrashSelectedSong:
		// trash selected song
		newPage := trashSong(h.netease, true)
		return true, newPage, app.Tick(time.Nanosecond)
	case keybindings.OpLikeSelectedSong: // half-width, full-width, Japanese, Chinese and French
		// like selected song
		newPage := likeSong(h.netease, true, true)
		return true, newPage, app.Tick(time.Nanosecond)
	case keybindings.OpDislikeSelectedSong:
		// unlike selected song
		newPage := likeSong(h.netease, false, true)
		return true, newPage, app.Tick(time.Nanosecond)
	case keybindings.OpHelp:
		// 帮助（Markdown 弹窗，支持滚动/缩放/Esc 关闭）
		showHelpPopup(h.netease.App)
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
		findSimilarSongs(h.netease, false)
	case keybindings.OpOpenSimiSongsOfSelectedSong:
		// 与当前选中歌曲相似的歌曲
		findSimilarSongs(h.netease, true)
	case keybindings.OpAlbumOfPlayingSong:
		// 当前歌曲所属专辑
		goToAlbumOfSong(h.netease, false)
	case keybindings.OpAlbumOfSelectedSong:
		// 选中歌曲所属专辑
		goToAlbumOfSong(h.netease, true)
	case keybindings.OpArtistOfPlayingSong:
		// 当前歌曲所属歌手
		goToArtistOfSong(h.netease, false)
	case keybindings.OpArtistOfSelectedSong:
		// 选中歌曲所属歌手
		goToArtistOfSong(h.netease, true)
	case keybindings.OpOpenPlayingSongInWeb:
		// 网页打开当前歌曲
		openInWeb(h.netease, false, main.SelectedIndex())
	case keybindings.OpOpenSelectedItemInWeb:
		// 网页打开选中项
		openInWeb(h.netease, true, main.SelectedIndex())
	case keybindings.OpCollectSelectedPlaylist:
		// 收藏选中歌单
		newPage := collectSelectedPlaylist(h.netease, true)
		return true, newPage, app.Tick(time.Nanosecond)
	case keybindings.OpDiscollectSelectedPlaylist:
		// 取消收藏选中歌单
		newPage := collectSelectedPlaylist(h.netease, false)
		return true, newPage, app.Tick(time.Nanosecond)
	case keybindings.OpSubscribeAlbumOfPlayingSong:
		// 收藏播放中歌曲的专辑
		newPage := subscribeAlbum(h.netease, true, false)
		return true, newPage, app.Tick(time.Nanosecond)
	case keybindings.OpUnsubscribeAlbumOfPlayingSong:
		// 取消收藏播放中歌曲的专辑
		newPage := subscribeAlbum(h.netease, false, false)
		return true, newPage, app.Tick(time.Nanosecond)
	case keybindings.OpSubscribeArtistOfPlayingSong:
		// 收藏播放中歌曲的歌手
		newPage := subscribeArtist(h.netease, true, false)
		return true, newPage, app.Tick(time.Nanosecond)
	case keybindings.OpUnsubscribeArtistOfPlayingSong:
		// 取消收藏播放中歌曲的歌手
		newPage := subscribeArtist(h.netease, false, false)
		return true, newPage, app.Tick(time.Nanosecond)
	case keybindings.OpSubscribeAlbumOfSelectedSong:
		// 收藏选中歌曲的专辑
		newPage := subscribeAlbum(h.netease, true, true)
		return true, newPage, app.Tick(time.Nanosecond)
	case keybindings.OpUnsubscribeAlbumOfSelectedSong:
		// 取消收藏选中歌曲的专辑
		newPage := subscribeAlbum(h.netease, false, true)
		return true, newPage, app.Tick(time.Nanosecond)
	case keybindings.OpSubscribeArtistOfSelectedSong:
		// 收藏选中歌曲的歌手
		newPage := subscribeArtist(h.netease, true, true)
		return true, newPage, app.Tick(time.Nanosecond)
	case keybindings.OpUnsubscribeArtistOfSelectedSong:
		// 取消收藏选中歌曲的歌手
		newPage := subscribeArtist(h.netease, false, true)
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
		downloadSongLrc(h.netease, false)
	case keybindings.OpDownloadSelectedSongLrc:
		downloadSongLrc(h.netease, true)
	case keybindings.OpActionOfSelected:
		action(h.netease, false)
	case keybindings.OpActionOfPlayingSong:
		action(h.netease, true)
	case keybindings.OpSharePlayingItem:
		shareItem(h.netease, false, main.SelectedIndex())
	case keybindings.OpShareSelectItem:
		shareItem(h.netease, true, main.SelectedIndex())
	case keybindings.OpToggleSortOrder:
		if djMenu, ok := menu.(*DjRadioDetailMenu); ok {
			djMenu.ToggleSortOrder()
			loading := model.NewLoading(h.netease.MustMain())
			loading.Start()
			defer loading.Complete()
			reloadSuccess, _ := djMenu.Reload()
			if reloadSuccess {
				main.RefreshMenuList()
				return true, main, app.RerenderCmd(true)
			}
		}

	case keybindings.OpSwitchTheme:
		registry := configs.CurrentThemeRegistry()
		newSS := registry.NextStyleSet(style.HasDarkBackground())
		if newSS != nil {
			app.SetStyleSet(*newSS)
			h.netease.notifyThemeSwitch(app, "切换主题", registry.CurrentName(style.HasDarkBackground()))
			return true, main, app.RerenderCmd(true)
		}
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
	main := h.netease.MustMain()
	playOrToggle(h.netease, main.CurMenu().RealDataIndex(main.SelectedIndex()))
}

func playOrToggle(netease *Netease, selectedIndex int) {
	var (
		songs         []structs.Song
		main          = netease.MustMain()
		menu          = main.CurMenu()
		player        = netease.player
		inPlayingMenu = player.InPlayingMenu()
	)
	if me, ok := menu.(SongsMenu); ok {
		songs = me.Songs()
	}

	if me, ok := menu.(Menu); !ok || !me.IsPlayable() || selectedIndex < 0 || len(songs) == 0 || selectedIndex > len(songs)-1 {
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

// MouseMsgHandle 处理鼠标事件。
// 作为 foxful-cli 的 MouseController 优先运行：仅处理 go-musicfox 特有区域
// （播放栏、进度条、非菜单区滚轮），其余返回 false 委托给 foxful-cli
// 内置处理（hover、右键 ContextMenu、单击/双击、返回按钮、面包屑、tab、
// 菜单区滚轮滚动、侧键菜单导航）。
func (h *EventHandler) MouseMsgHandle(msg tea.MouseMsg, a *model.App) (stopPropagation bool, newPage model.Page, cmd tea.Cmd) {
	var (
		player = h.netease.player
		main   = a.MustMain()
	)

	// 登录页需要自己处理鼠标点击（输入框/按钮），这里不要吞掉事件
	if _, ok := a.CurPage().(*LoginPage); ok {
		return false, nil, nil
	}

	mouse := msg.Mouse()
	switch msg.(type) {
	case tea.MouseClickMsg:
		switch mouse.Button {
		case tea.MouseLeft:
			// 播放栏点击（播放模式/歌曲信息/歌手）— musicfox 特有
			if handled, page, cmd := h.handlePlayerBarClick(msg, a, main); handled {
				return true, page, cmd
			}
			// 进度条 seek — musicfox 特有
			if handled, m, cmd := h.handleProgressBarSeek(msg, a, main); handled {
				return handled, m, cmd
			}
			// 其他区域（菜单项/返回按钮/面包屑/tab）委托 foxful-cli
			return false, nil, nil

		}
		// 侧键（Backward/Forward/Middle）及其他左右键未命中 → 委托 foxful-cli
		return false, nil, nil

	case tea.MouseWheelMsg:
		switch mouse.Button {
		case tea.MouseWheelUp, tea.MouseWheelDown:
			// 菜单区滚轮 → 委托 foxful-cli 滚动菜单；其他区域 → 调音量。
			// 对齐 foxful-cli 内部 0-based mouse.Y 坐标系，勿加偏移。
			if mouse.Y >= main.MenuStartRow() && mouse.Y < main.MenuBottomRow() {
				return false, nil, nil
			}
			if mouse.Button == tea.MouseWheelDown {
				if mouse.Mod == tea.ModCtrl {
					player.SetVolume(max(player.Volume()-h.mouseVolumeStep, 0))
				} else {
					player.DownVolume()
				}
			} else {
				if mouse.Mod == tea.ModCtrl {
					player.SetVolume(min(player.Volume()+h.mouseVolumeStep, 100))
				} else {
					player.UpVolume()
				}
			}
			return true, main, a.Tick(time.Nanosecond)
		case tea.MouseWheelLeft, tea.MouseWheelRight:
			// 水平滚轮不映射切歌，避免触控板手势转换为多条消息时重复切歌。
			// 用户可继续使用键盘快捷键切歌。
			return false, nil, nil
		}
		return false, nil, nil

	case tea.MouseMotionMsg:
		// 播放栏 hover 跟踪 — musicfox 特有
		if handled, cmd := h.handlePlaybarMotion(msg, a, main); handled {
			return true, nil, cmd
		}
		// 其他 motion → 委托 foxful-cli（hover 效果 + 指针形状）
		return false, nil, nil
	}

	// 其他事件（MouseReleaseMsg 等）→ 委托 foxful-cli
	return false, nil, nil
}

// handlePlayerBarClick 处理播放栏点击事件
func (h *EventHandler) handlePlayerBarClick(msg tea.MouseMsg, a *model.App, main *model.Main) (bool, model.Page, tea.Cmd) {
	playModeRow := main.EffectiveWindowHeight(a) - PlayModeRowOffset
	if msg.Mouse().Y != playModeRow {
		return false, nil, nil
	}

	// 播放模式点击
	if handled, page, cmd := h.handlePlayModeClick(msg, a, main); handled {
		return handled, page, cmd
	}

	// 播放栏其他元素点击
	return h.handlePlayerBarElementsClick(msg, a, main)
}

// handlePlayModeClick 播放模式点击
func (h *EventHandler) handlePlayModeClick(msg tea.MouseMsg, a *model.App, main *model.Main) (bool, model.Page, tea.Cmd) {
	player := h.netease.player
	menuStartColumn := main.MenuStartColumn()

	if menuStartColumn > MenuArrowWidth {
		playModeEndX := menuStartColumn + PlayModeClickWidth
		if msg.Mouse().X >= menuStartColumn-MenuArrowWidth && msg.Mouse().X <= playModeEndX {
			player.SwitchMode()
			return true, main, a.Tick(time.Nanosecond)
		}
	}
	return false, nil, nil
}

// handlePlayerBarElementsClick 播放栏其他元素点击
func (h *EventHandler) handlePlayerBarElementsClick(msg tea.MouseMsg, a *model.App, main *model.Main) (bool, model.Page, tea.Cmd) {
	player := h.netease.player
	curSong := player.CurSong()

	mouse := msg.Mouse()

	if curSong.Id == 0 {
		return false, nil, nil
	}

	menuStartColumn := main.MenuStartColumn()
	leftPad := 0
	if !main.CenterEverything() && menuStartColumn-MenuArrowWidth > 0 {
		leftPad = menuStartColumn - MenuArrowWidth
	}

	currentX := leftPad

	// Mode & Volume 宽度
	if menuStartColumn-MenuArrowWidth > 0 {
		modeStr := fmt.Sprintf("[%s] ", player.Mode().Name())
		modeWidth := runewidth.StringWidth(modeStr)
		volStr := fmt.Sprintf("%d%% ", player.Volume())
		volWidth := runewidth.StringWidth(volStr)
		currentX += modeWidth + volWidth
	}

	// 播放状态
	stateText := "_ z Z Z "
	if player.State() == types.Playing {
		stateText = "♫ ♪ ♫ ♪ "
	}
	stateWidth := runewidth.StringWidth(stateText)

	if mouse.X >= currentX && mouse.X < currentX+stateWidth {
		h.playOrToggleHandle()
		return true, main, a.Tick(time.Nanosecond)
	}
	currentX += stateWidth

	// 喜欢状态
	heartWidth := 0
	if curSong.Id > 0 {
		_ = likelist.IsLikeSong(curSong.Id)
		heartWidth = runewidth.StringWidth("♥ ")
	}
	if heartWidth > 0 {
		if mouse.X >= currentX && mouse.X < currentX+heartWidth {
			isLiked := likelist.IsLikeSong(curSong.Id)
			newPage := likeSong(h.netease, !isLiked, false)
			return true, newPage, a.Tick(time.Nanosecond)
		}
		currentX += heartWidth
	}

	// 歌曲名
	songName := curSong.Name
	songShownWidth := runewidth.StringWidth(songName)
	if !main.CenterEverything() {
		prefixLen := SongInfoPrefixBaseWidth
		if main.MenuStartColumn()-MenuArrowWidth > 0 {
			prefixLen += SongInfoPrefixExtraWidth
		}
		maxSongWidth := a.WindowWidth() - main.MenuStartColumn() - prefixLen
		if songShownWidth > maxSongWidth {
			songShownWidth = maxSongWidth
		}
	}

	if mouse.X >= currentX && mouse.X < currentX+songShownWidth {
		action(h.netease, true)
		return true, main, a.Tick(time.Nanosecond)
	}
	currentX += songShownWidth + 1

	// 歌手
	if mouse.X >= currentX {
		goToArtistOfSong(h.netease, false)
		return true, main, a.Tick(time.Nanosecond)
	}

	return false, nil, nil
}

// handlePlaybarMotion tracks the hovered playbar element and switches the mouse pointer.
func (h *EventHandler) handlePlaybarMotion(msg tea.MouseMsg, a *model.App, main *model.Main) (bool, tea.Cmd) {
	player := h.netease.player
	mouse := msg.Mouse()
	playModeRow := main.EffectiveWindowHeight(a) - PlayModeRowOffset
	progressBarRow := main.EffectiveWindowHeight(a) - 1

	newHover := PlaybarElementNone
	needsPointer := false

	progressBarWidth := a.WindowWidth() - ProgressTimeDisplayWidth
	if mouse.Y == progressBarRow && mouse.X < progressBarWidth {
		newHover = PlaybarElementProgressBar
		needsPointer = true
	} else if mouse.Y == playModeRow {
		curSong := player.CurSong()
		if curSong.Id == 0 {
			menuStartColumn := main.MenuStartColumn()
			if menuStartColumn > MenuArrowWidth {
				playModeEndX := menuStartColumn + PlayModeClickWidth
				if mouse.X >= menuStartColumn-MenuArrowWidth && mouse.X <= playModeEndX {
					newHover = PlaybarElementMode
					needsPointer = true
				}
			}
		} else {
			menuStartColumn := main.MenuStartColumn()
			leftPad := 0
			if !main.CenterEverything() && menuStartColumn-MenuArrowWidth > 0 {
				leftPad = menuStartColumn - MenuArrowWidth
			}
			currentX := leftPad

			if menuStartColumn-MenuArrowWidth > 0 {
				modeWidth := runewidth.StringWidth(fmt.Sprintf("[%s] ", player.Mode().Name()))
				volWidth := runewidth.StringWidth(fmt.Sprintf("%d%% ", player.Volume()))
				if mouse.X >= currentX && mouse.X < currentX+modeWidth {
					newHover = PlaybarElementMode
					needsPointer = true
				}
				currentX += modeWidth + volWidth
			}

			stateText := "_ z Z Z "
			if player.State() == types.Playing {
				stateText = "♫ ♪ ♫ ♪ "
			}
			stateWidth := runewidth.StringWidth(stateText)
			if mouse.X >= currentX && mouse.X < currentX+stateWidth {
				newHover = PlaybarElementState
				needsPointer = true
			}
			currentX += stateWidth

			heartWidth := runewidth.StringWidth("♥ ")
			if mouse.X >= currentX && mouse.X < currentX+heartWidth {
				newHover = PlaybarElementHeart
				needsPointer = true
			}
			currentX += heartWidth

			songShownWidth := runewidth.StringWidth(curSong.Name)
			if !main.CenterEverything() {
				prefixLen := SongInfoPrefixBaseWidth
				if main.MenuStartColumn()-MenuArrowWidth > 0 {
					prefixLen += SongInfoPrefixExtraWidth
				}
				maxSongWidth := a.WindowWidth() - main.MenuStartColumn() - prefixLen
				if songShownWidth > maxSongWidth {
					songShownWidth = maxSongWidth
				}
			}
			if mouse.X >= currentX && mouse.X < currentX+songShownWidth {
				newHover = PlaybarElementSongName
				needsPointer = true
			}
			currentX += songShownWidth + 1

			if mouse.X >= currentX {
				newHover = PlaybarElementArtist
				needsPointer = true
			}
		}
	}

	if newHover != h.netease.playbarHoveredElement {
		h.netease.playbarHoveredElement = newHover
		pointer := "default"
		if needsPointer {
			pointer = "pointer"
		}
		return true, tea.Sequence(a.RerenderCmd(true), a.SetMousePointer(pointer))
	}

	// Hover 未变化时，若鼠标仍停留在播放栏可点击元素上，必须消费该事件，
	// 并重新发送 SetMousePointer，否则 foxful-cli 的 mouseMotionHandle 在后续
	// motion 事件中会因 isOverClickableElement 不识别播放栏而重置指针为 "default"。
	if newHover != PlaybarElementNone {
		return true, a.SetMousePointer("pointer")
	}

	return false, nil
}

func (h *EventHandler) handleProgressBarSeek(msg tea.MouseMsg, a *model.App, main *model.Main) (bool, model.Page, tea.Cmd) {
	player := h.netease.player
	x, y := msg.Mouse().X, msg.Mouse().Y
	progressBarWidth := a.WindowWidth() - ProgressTimeDisplayWidth
	if y+1 == main.EffectiveWindowHeight(a) && x+1 <= progressBarWidth {
		allDuration := int(player.CurMusic().Duration.Seconds())
		if allDuration == 0 {
			return true, main, nil
		}
		duration := float64(x) * player.CurMusic().Duration.Seconds() / float64(progressBarWidth)
		player.Seek(time.Second * time.Duration(duration))
		if player.State() != types.Playing {
			player.Resume()
		}
		return true, main, a.Tick(time.Nanosecond)
	}
	return false, main, nil
}
