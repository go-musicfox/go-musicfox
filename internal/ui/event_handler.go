package ui

import (
	"fmt"
	"log/slog"
	"runtime"
	"slices"
	"time"

	"github.com/anhoder/foxful-cli/model"
	tea "github.com/charmbracelet/bubbletea"
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

	// 双击检测相关字段
	lastClickTime time.Time
	lastClickX    int
	lastClickY    int
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
		openInWeb(h.netease, false)
	case keybindings.OpOpenSelectedItemInWeb:
		// 网页打开选中项
		openInWeb(h.netease, true)
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
		downloadSongLrc(h.netease, false)
	case keybindings.OpDownloadSelectedSongLrc:
		downloadSongLrc(h.netease, true)
	case keybindings.OpActionOfSelected:
		action(h.netease, false)
	case keybindings.OpActionOfPlayingSong:
		action(h.netease, true)
	case keybindings.OpSharePlayingItem:
		shareItem(h.netease, false)
	case keybindings.OpShareSelectItem:
		shareItem(h.netease, true)
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

// MouseMsgHandle 处理鼠标事件
func (h *EventHandler) MouseMsgHandle(msg tea.MouseMsg, a *model.App) (stopPropagation bool, newPage model.Page, cmd tea.Cmd) {
	var (
		player = h.netease.player
		main   = a.MustMain()
	)

	// 登录页需要自己处理鼠标点击（输入框/按钮），这里不要吞掉事件
	if _, ok := a.CurPage().(*LoginPage); ok {
		return false, nil, nil
	}

	switch msg.Button {
	case tea.MouseButtonLeft:
		// Handle play mode click
		if msg.Action == tea.MouseActionPress {
			// 计算播放模式显示位置
			// 播放模式在歌曲信息行，位于窗口底部往上第3行（进度条是最后一行，往上数第3行）
			if handled, page, cmd := h.handlePlayerBarClick(msg, a, main); handled {
				return true, page, cmd
			}
		}

		// Handle single and double-click detection for menu items
		if msg.Action == tea.MouseActionPress {
			now := time.Now()
			// 根据操作系统设置双击间隔阈值
			var doubleClickInterval time.Duration
			switch runtime.GOOS {
			case "darwin":
				doubleClickInterval = 400 * time.Millisecond
			case "windows":
				doubleClickInterval = 500 * time.Millisecond
			default:
				doubleClickInterval = 300 * time.Millisecond
			}

			// 计算坐标差的绝对值
			deltaX := msg.X - h.lastClickX
			if deltaX < 0 {
				deltaX = -deltaX
			}
			deltaY := msg.Y - h.lastClickY
			if deltaY < 0 {
				deltaY = -deltaY
			}

			// 检测是否为双击（时间间隔小于阈值且位置相近）
			if now.Sub(h.lastClickTime) <= doubleClickInterval &&
				deltaX <= 2 && deltaY <= 2 {
				// 双击事件处理
				handled, page := h.handleDoubleClick(msg, a, main)
				if handled {
					// 重置双击检测状态
					h.lastClickTime = time.Time{}
					if page != nil {
						return true, page, a.Tick(time.Nanosecond)
					}
					return true, main, a.Tick(time.Nanosecond)
				}
			} else {
				// 单击事件处理：改变焦点
				h.handleSingleClick(msg, a, main)
			}

			// 更新最后点击信息
			h.lastClickTime = now
			h.lastClickX = msg.X
			h.lastClickY = msg.Y
		}

		// Handle progress bar seeking
		if slices.Contains([]tea.MouseAction{tea.MouseActionPress, tea.MouseActionMotion}, msg.Action) {
			x, y := msg.X, msg.Y
			progressBarWidth := a.WindowWidth() - 14
			if y+1 == a.WindowHeight() && x+1 <= progressBarWidth {
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
		}
	case tea.MouseButtonWheelDown:
		if msg.Ctrl {
			currentVolume := player.Volume()
			newVolume := max(currentVolume-h.mouseVolumeStep, 0)
			player.SetVolume(newVolume)
		} else {
			player.DownVolume()
		}
	case tea.MouseButtonWheelUp:
		if msg.Ctrl {
			currentVolume := player.Volume()
			newVolume := min(currentVolume+h.mouseVolumeStep, 100)
			player.SetVolume(newVolume)
		} else {
			player.UpVolume()
		}
	case tea.MouseButtonWheelLeft:
		player.PreviousSong(true)
	case tea.MouseButtonWheelRight:
		player.NextSong(true)
	case tea.MouseButtonForward:
		// 鼠标前进键：上一页
		if msg.Action == tea.MouseActionPress {
			oldPage := main.CurPage()
			main.PrePage()
			if oldPage != main.CurPage() {
				curIndex := mathx.Max(main.SelectedIndex()-main.PageSize(), 0)
				main.SetSelectedIndex(curIndex)
			}
			return true, main, a.Tick(time.Nanosecond)
		}
	case tea.MouseButtonBackward:
		// 鼠标后退键：下一页
		if msg.Action == tea.MouseActionPress {
			oldPage := main.CurPage()
			main.NextPage()
			if oldPage != main.CurPage() {
				curIndex := mathx.Min(main.SelectedIndex()+main.PageSize(), len(main.CurMenu().MenuViews())-1)
				main.SetSelectedIndex(curIndex)
			}
			return true, main, a.Tick(time.Nanosecond)
		}
	case tea.MouseButtonRight:
		// 鼠标右键：根据点击位置判断是菜单还是歌曲信息区域
		if msg.Action == tea.MouseActionPress {
			// 歌曲信息行：等同于对“当前播放”按 m
			playModeRow := a.WindowHeight() - 3
			if msg.Y == playModeRow {
				action(h.netease, true)
				return true, main, a.Tick(time.Nanosecond)
			}

			// 菜单区域：先切换焦点；若当前是“可播放歌曲菜单”，则打开选中项操作菜单
			menuStartRow := main.MenuStartRow()
			menuBottomRow := main.MenuBottomRow()
			y := msg.Y + 1 // 1-based
			isInMenuArea := y >= menuStartRow && y < menuBottomRow
			if isInMenuArea {
				h.handleSingleClick(msg, a, main)

				menu := main.CurMenu()
				if _, ok := menu.(SongsMenu); ok {
					if curMenu, ok := menu.(Menu); ok && curMenu.IsPlayable() {
						action(h.netease, false)
						return true, main, a.Tick(time.Nanosecond)
					}
				}

				return true, main, a.Tick(time.Nanosecond)
			}

			// 非菜单区域：默认触发“当前播放”的操作菜单
			action(h.netease, true)
			return true, main, a.Tick(time.Nanosecond)
		}
	case tea.MouseButtonMiddle:
		// 鼠标中键：返回上一级（等同于ESC键）
		if msg.Action == tea.MouseActionPress {
			main.BackMenu()
			return true, main, a.Tick(time.Nanosecond)
		}
	}

	return true, main, a.Tick(time.Nanosecond)
}

// handleSingleClick 处理鼠标单击事件，单击菜单项时改变焦点
func (h *EventHandler) handleSingleClick(msg tea.MouseMsg, a *model.App, main *model.Main) bool {
	index, ok := h.getClickedIndexFromPosition(msg, a, main)
	if !ok {
		return false
	}
	main.SetSelectedIndex(index)
	return true
}

// handleDoubleClick 处理鼠标双击事件，双击菜单项时进入该菜单
func (h *EventHandler) handleDoubleClick(msg tea.MouseMsg, a *model.App, main *model.Main) (bool, model.Page) {
	clickedIndex, ok := h.getClickedIndexFromPosition(msg, a, main)
	if !ok {
		return false, nil
	}

	// 设置选中索引
	main.SetSelectedIndex(clickedIndex)
	menu := main.CurMenu()
	menuViews := menu.MenuViews()

	// 如果是歌曲菜单，双击播放歌曲
	if songsMenu, ok := menu.(SongsMenu); ok {
		if curMenu, ok := menu.(Menu); ok && curMenu.IsPlayable() {
			songs := songsMenu.Songs()
			selectedIndex := menu.RealDataIndex(clickedIndex)
			if selectedIndex >= 0 && selectedIndex < len(songs) {
				// 调用播放逻辑
				h.playOrToggleHandle()
				return true, nil
			}
		}
	}

	// 非歌曲菜单，进入子菜单
	loading := model.NewLoading(main)
	loading.Start()
	defer loading.Complete()

	submenu := menu.SubMenu(a, menu.RealDataIndex(clickedIndex))
	if submenu != nil {
		menuTitle := &menuViews[clickedIndex]
		newPage := main.EnterMenu(submenu, menuTitle)
		return true, newPage
	}

	return false, nil
}

// getClickedIndexFromPosition 根据鼠标位置计算点击的菜单项索引
func (h *EventHandler) getClickedIndexFromPosition(msg tea.MouseMsg, a *model.App, main *model.Main) (int, bool) {
	menu := main.CurMenu()
	menuViews := menu.MenuViews()

	// 获取菜单显示区域的行范围
	menuStartRow := main.MenuStartRow()
	menuBottomRow := main.MenuBottomRow()
	menuStartColumn := main.MenuStartColumn()

	// 检查点击是否在菜单区域内
	// msg.Y 是 0-based，需要转换为 1-based 与 menuStartRow 比较
	y := msg.Y + 1
	if y < menuStartRow || y >= menuBottomRow {
		return 0, false
	}

	// X坐标：检查是否在菜单的有效区域内
	if msg.X < menuStartColumn-4 {
		return 0, false
	}

	// 计算点击对应的菜单项索引
	actualMenuStartRow := menuStartRow
	relativeRow := y - actualMenuStartRow

	if relativeRow < 0 {
		return 0, false
	}

	pageStartIndex := (main.CurPage() - 1) * main.PageSize()
	var clickedIndex int

	// 双列模式
	if main.IsDualColumn() {
		windowWidth := a.WindowWidth()
		var leftColumnWidth int
		if windowWidth <= 88 {
			leftColumnWidth = (windowWidth - menuStartColumn - 4) / 2
		} else {
			leftColumnWidth = 44
		}

		if msg.X < menuStartColumn+leftColumnWidth {
			clickedIndex = pageStartIndex + relativeRow*2
		} else {
			clickedIndex = pageStartIndex + relativeRow*2 + 1
		}
	} else {
		clickedIndex = pageStartIndex + relativeRow
	}

	if clickedIndex < 0 || clickedIndex >= len(menuViews) {
		return 0, false
	}

	return clickedIndex, true
}

// handlePlayerBarClick 处理播放栏点击事件
func (h *EventHandler) handlePlayerBarClick(msg tea.MouseMsg, a *model.App, main *model.Main) (bool, model.Page, tea.Cmd) {
	playModeRow := a.WindowHeight() - 3
	if msg.Y != playModeRow {
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

	if menuStartColumn > 4 {
		playModeEndX := menuStartColumn + 5
		if msg.X >= menuStartColumn-4 && msg.X <= playModeEndX {
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

	if curSong.Id == 0 {
		return false, nil, nil
	}

	menuStartColumn := main.MenuStartColumn()
	leftPad := 0
	if !main.CenterEverything() && menuStartColumn-4 > 0 {
		leftPad = menuStartColumn - 4
	}

	currentX := leftPad

	// Mode & Volume 宽度
	if menuStartColumn-4 > 0 {
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

	if msg.X >= currentX && msg.X < currentX+stateWidth {
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
		if msg.X >= currentX && msg.X < currentX+heartWidth {
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
		prefixLen := 10
		if main.MenuStartColumn()-4 > 0 {
			prefixLen += 12
		}
		maxSongWidth := a.WindowWidth() - main.MenuStartColumn() - prefixLen
		if songShownWidth > maxSongWidth {
			songShownWidth = maxSongWidth
		}
	}

	if msg.X >= currentX && msg.X < currentX+songShownWidth {
		action(h.netease, true)
		return true, main, a.Tick(time.Nanosecond)
	}
	currentX += songShownWidth + 1

	// 歌手
	if msg.X >= currentX {
		goToArtistOfSong(h.netease, false)
		return true, main, a.Tick(time.Nanosecond)
	}

	return false, nil, nil
}
