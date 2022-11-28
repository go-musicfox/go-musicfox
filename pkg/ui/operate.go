package ui

import (
	"log"
	"math"
	"os"
	"strconv"
	"time"

	"github.com/anhoder/netease-music/service"
	"github.com/buger/jsonparser"
	"go-musicfox/pkg/constants"
	"go-musicfox/pkg/player"
	"go-musicfox/pkg/storage"
	"go-musicfox/pkg/structs"
	"go-musicfox/utils"
)

type menuStackItem struct {
	menuList      []MenuItem
	selectedIndex int
	menuCurPage   int
	menuTitle     *MenuItem
	menu          IMenu
}

// 上移
func moveUp(m *NeteaseModel) {
	topHook := m.menu.TopOutHook()
	if m.doubleColumn {
		if m.selectedIndex-2 < 0 && topHook != nil {
			loading := NewLoading(m)
			loading.start()
			if res := topHook(m); !res {
				loading.complete()
				return
			}
			// 更新菜单UI
			m.menuList = m.menu.MenuViews()
			loading.complete()
		}
		if m.selectedIndex-2 < 0 {
			return
		}
		m.selectedIndex -= 2
	} else {
		if m.selectedIndex-1 < 0 && topHook != nil {
			loading := NewLoading(m)
			loading.start()
			if res := topHook(m); !res {
				loading.complete()
				return
			}
			m.menuList = m.menu.MenuViews()
			loading.complete()
		}
		if m.selectedIndex-1 < 0 {
			return
		}
		m.selectedIndex--
	}
	if m.selectedIndex < (m.menuCurPage-1)*m.menuPageSize {
		prePage(m)
	}
}

// 下移
func moveDown(m *NeteaseModel) {
	bottomHook := m.menu.BottomOutHook()
	if m.doubleColumn {
		if m.selectedIndex+2 > len(m.menuList)-1 && bottomHook != nil {
			loading := NewLoading(m)
			loading.start()
			if res := bottomHook(m); !res {
				loading.complete()
				return
			}
			m.menuList = m.menu.MenuViews()
			loading.complete()
		}
		if m.selectedIndex+2 > len(m.menuList)-1 {
			return
		}
		m.selectedIndex += 2
	} else {
		if m.selectedIndex+1 > len(m.menuList)-1 && bottomHook != nil {
			loading := NewLoading(m)
			loading.start()
			if res := bottomHook(m); !res {
				loading.complete()
				return
			}
			m.menuList = m.menu.MenuViews()
			loading.complete()
		}
		if m.selectedIndex+1 > len(m.menuList)-1 {
			return
		}
		m.selectedIndex++
	}
	if m.selectedIndex >= m.menuCurPage*m.menuPageSize {
		nextPage(m)
	}
}

// 左移
func moveLeft(m *NeteaseModel) {
	if !m.doubleColumn || m.selectedIndex%2 == 0 || m.selectedIndex-1 < 0 {
		return
	}
	m.selectedIndex--
}

// 右移
func moveRight(m *NeteaseModel) {
	if !m.doubleColumn || m.selectedIndex%2 != 0 {
		return
	}
	if bottomHook := m.menu.BottomOutHook(); m.selectedIndex >= len(m.menuList)-1 && bottomHook != nil {
		loading := NewLoading(m)
		loading.start()
		if res := bottomHook(m); !res {
			loading.complete()
			return
		}
		m.menuList = m.menu.MenuViews()
		loading.complete()
	}
	if m.selectedIndex >= len(m.menuList)-1 {
		return
	}
	m.selectedIndex++
}

// 切换到上一页
func prePage(m *NeteaseModel) {
	m.isListeningKey = false
	defer func() {
		m.isListeningKey = true
	}()

	if prePageHook := m.menu.BeforePrePageHook(); prePageHook != nil {
		loading := NewLoading(m)
		loading.start()
		if res := prePageHook(m); !res {
			loading.complete()
			return
		}
		loading.complete()
	}

	if m.menuCurPage <= 1 {
		return
	}
	m.menuCurPage--
}

// 切换到下一页
func nextPage(m *NeteaseModel) {
	m.isListeningKey = false
	defer func() {
		m.isListeningKey = true
	}()

	if nextPageHook := m.menu.BeforeNextPageHook(); nextPageHook != nil {
		loading := NewLoading(m)
		loading.start()
		if res := nextPageHook(m); !res {
			loading.complete()
			return
		}
		loading.complete()
	}
	if m.menuCurPage >= int(math.Ceil(float64(len(m.menuList))/float64(m.menuPageSize))) {
		return
	}

	m.menuCurPage++
}

// 进入菜单
func enterMenu(m *NeteaseModel, newMenu IMenu, newTitle *MenuItem) {
	m.isListeningKey = false
	defer func() {
		m.isListeningKey = true
	}()

	if (newMenu == nil || newTitle == nil) && m.selectedIndex >= len(m.menuList) {
		return
	}

	if newMenu == nil {
		newMenu = m.menu.SubMenu(m, m.selectedIndex)
	}
	if newTitle == nil {
		newTitle = &m.menuList[m.selectedIndex]
	}

	stackItem := &menuStackItem{
		menuList:      m.menuList,
		selectedIndex: m.selectedIndex,
		menuCurPage:   m.menuCurPage,
		menuTitle:     m.menuTitle,
		menu:          m.menu,
	}
	m.menuStack.Push(stackItem)

	if newMenu == nil {
		m.menuStack.Pop()
		return
	}

	if enterMenuHook := newMenu.BeforeEnterMenuHook(); enterMenuHook != nil {
		loading := NewLoading(m)
		loading.start()
		if res := enterMenuHook(m); !res {
			loading.complete()
			m.menuStack.Pop() // 压入的重新弹出
			return
		}

		// 如果位于正在播放的菜单中，更新播放列表
		if newMenu.GetMenuKey() == m.player.playingMenuKey {
			if songs, ok := newMenu.MenuData().([]structs.Song); ok {
				m.player.playlist = songs
				m.player.playlistUpdateAt = time.Now()
			}
		}

		loading.complete()
	}

	if newMenu != nil {
		newMenu.FormatMenuItem(newTitle)
	}

	menuList := newMenu.MenuViews()

	m.menu = newMenu
	m.menuList = menuList
	m.menuTitle = newTitle
	m.selectedIndex = 0
	m.menuCurPage = 1
}

// 菜单返回
func backMenu(m *NeteaseModel) {
	m.isListeningKey = false
	defer func() {
		m.isListeningKey = true
	}()

	if m.menuStack.Len() <= 0 {
		return
	}

	stackItem := m.menuStack.Pop()
	if backMenuHook := m.menu.BeforeBackMenuHook(); backMenuHook != nil {
		loading := NewLoading(m)
		loading.start()
		if res := backMenuHook(m); !res {
			loading.complete()
			m.menuStack.Push(stackItem) // 弹出的重新压入
			return
		}
		loading.complete()
	}
	m.menu.FormatMenuItem(m.menuTitle) // 重新格式化

	stackMenu, ok := stackItem.(*menuStackItem)
	if !ok {
		return
	}

	m.menuList = stackMenu.menuList
	m.menu = stackMenu.menu
	m.menuTitle = stackMenu.menuTitle
	m.menu.FormatMenuItem(m.menuTitle)
	m.selectedIndex = stackMenu.selectedIndex
	m.menuCurPage = stackMenu.menuCurPage
}

// 空格监听
func spaceKeyHandle(m *NeteaseModel) {
	var (
		songs         []structs.Song
		inPlayingMenu = m.player.InPlayingMenu()
	)
	if inPlayingMenu && !m.menu.ResetPlaylistWhenPlay() {
		songs = m.player.playlist
	} else {
		if data, ok := m.menu.MenuData().([]structs.Song); ok {
			songs = data
		}
	}

	selectedIndex := m.menu.RealDataIndex(m.selectedIndex)
	if !m.menu.IsPlayable() || len(songs) == 0 || selectedIndex > len(songs)-1 {
		if m.player.curSongIndex > len(m.player.playlist)-1 {
			return
		}

		switch m.player.State() {
		case player.Paused:
			m.player.Resume()
		case player.Playing:
			m.player.Paused()
		case player.Stopped:
			_ = m.player.PlaySong(m.player.playlist[m.player.curSongIndex], DurationNext)
		}

		return
	}

	if inPlayingMenu && songs[selectedIndex].Id == m.player.playlist[m.player.curSongIndex].Id {
		switch m.player.State() {
		case player.Paused:
			m.player.Resume()
		case player.Playing:
			m.player.Paused()
		}
	} else {
		m.player.curSongIndex = selectedIndex
		m.player.playingMenuKey = m.menu.GetMenuKey()
		m.player.playingMenu = m.menu
		m.player.playlist = songs
		m.player.playlistUpdateAt = time.Now()
		if m.player.mode == player.PmIntelligent {
			m.player.SetPlayMode(0)
		}
		_ = m.player.PlaySong(songs[selectedIndex], DurationNext)
	}
}

// likePlayingSong like/unlike playing song
func likePlayingSong(m *NeteaseModel, isLike bool) {
	loading := NewLoading(m)
	loading.start()
	defer loading.complete()

	if m.player.curSongIndex >= len(m.player.playlist) {
		return
	}

	if utils.CheckUserInfo(m.user) == utils.NeedLogin {
		NeedLoginHandle(m, func(m *NeteaseModel, newMenu IMenu, newTitle *MenuItem) {
			likePlayingSong(m, isLike)
		})
		return
	}

	// 防止出现空，兜底
	if m.user.MyLikePlaylistID == 0 {
		userPlaylists := service.UserPlaylistService{
			Uid:    strconv.FormatInt(m.user.UserId, 10),
			Limit:  strconv.Itoa(1),
			Offset: strconv.Itoa(0),
		}
		code, response := userPlaylists.UserPlaylist()
		codeType := utils.CheckCode(code)
		if codeType == utils.NeedLogin {
			NeedLoginHandle(m, func(m *NeteaseModel, newMenu IMenu, newTitle *MenuItem) {
				likePlayingSong(m, isLike)
			})
			return
		} else if codeType != utils.Success {
			return
		}
		var err error
		m.user.MyLikePlaylistID, err = jsonparser.GetInt(response, "playlist", "[0]", "id")
		if err != nil {
			log.Printf("获取歌单ID失败: %+v\n", err)
			return
		}

		// 写入本地数据库
		table := storage.NewTable()
		_ = table.SetByKVModel(storage.User{}, m.user)
	}

	op := "add"
	if !isLike {
		op = "del"
	}
	likeService := service.PlaylistTracksService{
		TrackIds: []string{strconv.FormatInt(m.player.playlist[m.player.curSongIndex].Id, 10)},
		Op:       op,
		Pid:      strconv.FormatInt(m.user.MyLikePlaylistID, 10),
	}
	likeService.PlaylistTracks()

	if isLike {
		utils.Notify(utils.NotifyContent{
			Title: "已添加到我喜欢的歌曲",
			Text:  m.player.playlist[m.player.curSongIndex].Name,
			Url:   constants.AppGithubUrl,
		})
	} else {
		utils.Notify(utils.NotifyContent{
			Title: "已从我喜欢的歌曲移除",
			Text:  m.player.playlist[m.player.curSongIndex].Name,
			Url:   constants.AppGithubUrl,
		})
	}
}

// logout 登出
func logout() {
	table := storage.NewTable()
	_ = table.DeleteByKVModel(storage.User{})
	(&storage.LastfmUser{}).Clear()
	utils.Notify(utils.NotifyContent{
		Title: "登出成功",
		Text:  "已清理用户信息",
		Url:   constants.AppGithubUrl,
	})
	_ = os.Remove(utils.GetLocalDataDir() + "/cookie")
}

// likeSelectedSong like/unlike selected song
func likeSelectedSong(m *NeteaseModel, isLike bool) {
	loading := NewLoading(m)
	loading.start()
	defer loading.complete()

	songs, ok := m.menu.MenuData().([]structs.Song)
	selectedIndex := m.menu.RealDataIndex(m.selectedIndex)
	if !ok || selectedIndex >= len(songs) {
		return
	}

	if utils.CheckUserInfo(m.user) == utils.NeedLogin {
		NeedLoginHandle(m, func(m *NeteaseModel, newMenu IMenu, newTitle *MenuItem) {
			likeSelectedSong(m, isLike)
		})
		return
	}

	// 防止出现空，兜底
	if m.user.MyLikePlaylistID == 0 {
		userPlaylists := service.UserPlaylistService{
			Uid:    strconv.FormatInt(m.user.UserId, 10),
			Limit:  strconv.Itoa(1),
			Offset: strconv.Itoa(0),
		}
		code, response := userPlaylists.UserPlaylist()
		codeType := utils.CheckCode(code)
		if codeType == utils.NeedLogin {
			NeedLoginHandle(m, func(m *NeteaseModel, newMenu IMenu, newTitle *MenuItem) {
				likeSelectedSong(m, isLike)
			})
			return
		} else if codeType != utils.Success {
			return
		}
		var err error
		m.user.MyLikePlaylistID, err = jsonparser.GetInt(response, "playlist", "[0]", "id")
		if err != nil {
			log.Printf("获取歌单ID失败: %+v\n", err)
			return
		}

		// 写入本地数据库
		table := storage.NewTable()
		_ = table.SetByKVModel(storage.User{}, m.user)
	}

	op := "add"
	if !isLike {
		op = "del"
	}
	likeService := service.PlaylistTracksService{
		TrackIds: []string{strconv.FormatInt(songs[selectedIndex].Id, 10)},
		Op:       op,
		Pid:      strconv.FormatInt(m.user.MyLikePlaylistID, 10),
	}
	likeService.PlaylistTracks()

	if isLike {
		utils.Notify(utils.NotifyContent{
			Title: "已添加到我喜欢的歌曲",
			Text:  songs[selectedIndex].Name,
			Url:   constants.AppGithubUrl,
		})
	} else {
		utils.Notify(utils.NotifyContent{
			Title: "已从我喜欢的歌曲移除",
			Text:  songs[selectedIndex].Name,
			Url:   constants.AppGithubUrl,
		})
	}
}

// trashPlayingSong 标记为不喜欢
func trashPlayingSong(m *NeteaseModel) {
	loading := NewLoading(m)
	loading.start()
	defer loading.complete()

	if m.player.curSongIndex >= len(m.player.playlist) {
		return
	}

	if utils.CheckUserInfo(m.user) == utils.NeedLogin {
		NeedLoginHandle(m, func(m *NeteaseModel, newMenu IMenu, newTitle *MenuItem) {
			trashPlayingSong(m)
		})
		return
	}

	trashService := service.FmTrashService{
		SongID: strconv.FormatInt(m.player.playlist[m.player.curSongIndex].Id, 10),
	}
	trashService.FmTrash()

	utils.Notify(utils.NotifyContent{
		Title: "已标记为不喜欢",
		Text:  m.player.playlist[m.player.curSongIndex].Name,
		Url:   constants.AppGithubUrl,
	})
}

// trashSelectedSong 标记为不喜欢
func trashSelectedSong(m *NeteaseModel) {
	loading := NewLoading(m)
	loading.start()
	defer loading.complete()

	songs, ok := m.menu.MenuData().([]structs.Song)
	selectedIndex := m.menu.RealDataIndex(m.selectedIndex)
	if !ok || selectedIndex >= len(songs) {
		return
	}

	if utils.CheckUserInfo(m.user) == utils.NeedLogin {
		NeedLoginHandle(m, func(m *NeteaseModel, newMenu IMenu, newTitle *MenuItem) {
			trashSelectedSong(m)
		})
		return
	}

	trashService := service.FmTrashService{
		SongID: strconv.FormatInt(songs[selectedIndex].Id, 10),
	}
	trashService.FmTrash()

	utils.Notify(utils.NotifyContent{
		Title: "已标记为不喜欢",
		Text:  songs[selectedIndex].Name,
		Url:   constants.AppGithubUrl,
	})
}

// 搜索当前菜单
func searchMenuHandle(m *NeteaseModel) {
	m.inSearching = false
	enterMenu(m, NewSearchMenu(m.menu, m.searchInput.Value()), &MenuItem{Title: "搜索结果", Subtitle: m.searchInput.Value()})
	m.searchInput.Blur()
	m.searchInput.Reset()
}

// 下载当前音乐
func downloadSelectedSong(m *NeteaseModel) {
	loading := NewLoading(m)
	loading.start()
	defer loading.complete()

	songs, ok := m.menu.MenuData().([]structs.Song)
	selectedIndex := m.menu.RealDataIndex(m.selectedIndex)
	if !ok || selectedIndex >= len(songs) {
		return
	}

	utils.DownloadMusic(songs[selectedIndex])
}

func downloadPlayingSong(m *NeteaseModel) {
	loading := NewLoading(m)
	loading.start()
	defer loading.complete()

	if m.player.curSongIndex >= len(m.player.playlist) {
		return
	}

	utils.DownloadMusic(m.player.playlist[m.player.curSongIndex])
}
