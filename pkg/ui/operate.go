package ui

import (
	"log"
	"math"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/go-musicfox/go-musicfox/utils/like_list"
	"github.com/skratchdot/open-golang/open"

	"github.com/go-musicfox/go-musicfox/pkg/constants"
	"github.com/go-musicfox/go-musicfox/pkg/player"
	"github.com/go-musicfox/go-musicfox/pkg/storage"
	"github.com/go-musicfox/go-musicfox/pkg/structs"
	"github.com/go-musicfox/go-musicfox/utils"

	"github.com/buger/jsonparser"
	"github.com/go-musicfox/netease-music/service"
)

type menuStackItem struct {
	menuList      []MenuItem
	selectedIndex int
	menuCurPage   int
	menuTitle     *MenuItem
	menu          Menu
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

// 上移到顶部
func moveTop(m *NeteaseModel) {
	if m.doubleColumn {
		m.selectedIndex = m.selectedIndex % 2
	} else {
		m.selectedIndex = 0
	}
	m.menuCurPage = 1
}

// 下移到底部
func moveBottom(m *NeteaseModel) {
	if m.doubleColumn && len(m.menuList)%2 == 0 {
		m.selectedIndex = len(m.menuList) + (m.selectedIndex%2 - 2)
	} else if m.doubleColumn && m.selectedIndex%2 != 0 {
		m.selectedIndex = len(m.menuList) - 2
	} else {
		m.selectedIndex = len(m.menuList) - 1
	}
	m.menuCurPage = int(math.Ceil(float64(len(m.menuList)) / float64(m.menuPageSize)))
	if m.doubleColumn && m.selectedIndex%2 != 0 && len(m.menuList)%m.menuPageSize == 1 {
		m.menuCurPage -= 1
	}
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
func enterMenu(m *NeteaseModel, newMenu Menu, newTitle *MenuItem) {
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
	if menu, ok := m.menu.(SongsMenu); ok {
		songs = menu.Songs()
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

		var newPlaylists = make([]structs.Song, len(songs))
		copy(newPlaylists, songs)
		m.player.playlist = newPlaylists

		m.player.playlistUpdateAt = time.Now()
		if m.player.mode == player.PmIntelligent {
			m.player.SetPlayMode(0)
		}
		_ = m.player.PlaySong(m.player.playlist[selectedIndex], DurationNext)
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
		NeedLoginHandle(m, func(m *NeteaseModel, newMenu Menu, newTitle *MenuItem) {
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
			NeedLoginHandle(m, func(m *NeteaseModel, newMenu Menu, newTitle *MenuItem) {
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
	if code, resp := likeService.PlaylistTracks(); code != 200 {
		var msg string
		if msg, _ = jsonparser.GetString(resp, "message"); msg == "" {
			msg, _ = jsonparser.GetString(resp, "data", "message")
		}
		if msg == "" {
			msg = "加入或移出歌单失败"
		}
		utils.Notify(utils.NotifyContent{
			Title:   msg,
			Text:    m.player.playlist[m.player.curSongIndex].Name,
			Url:     constants.AppGithubUrl,
			GroupId: constants.GroupID,
		})
		return
	}

	go func() {
		like_list.RefreshLikeList(m.user.UserId)
		m.Rerender(false)
	}()

	if isLike {
		utils.Notify(utils.NotifyContent{
			Title:   "已添加到我喜欢的歌曲",
			Text:    m.player.playlist[m.player.curSongIndex].Name,
			Url:     utils.WebUrlOfPlaylist(m.user.MyLikePlaylistID),
			GroupId: constants.GroupID,
		})
	} else {
		utils.Notify(utils.NotifyContent{
			Title:   "已从我喜欢的歌曲移除",
			Text:    m.player.playlist[m.player.curSongIndex].Name,
			Url:     utils.WebUrlOfPlaylist(m.user.MyLikePlaylistID),
			GroupId: constants.GroupID,
		})
	}
}

// logout 登出
func logout() {
	table := storage.NewTable()
	_ = table.DeleteByKVModel(storage.User{})
	(&storage.LastfmUser{}).Clear()
	utils.Notify(utils.NotifyContent{
		Title:   "登出成功",
		Text:    "已清理用户信息",
		Url:     constants.AppGithubUrl,
		GroupId: constants.GroupID,
	})
	_ = os.Remove(path.Join(utils.GetLocalDataDir(), "cookie"))
}

// likeSelectedSong like/unlike selected song
func likeSelectedSong(m *NeteaseModel, isLike bool) {
	loading := NewLoading(m)
	loading.start()
	defer loading.complete()

	menu, ok := m.menu.(SongsMenu)
	selectedIndex := m.menu.RealDataIndex(m.selectedIndex)
	if !ok || selectedIndex >= len(menu.Songs()) {
		return
	}
	songs := menu.Songs()

	if utils.CheckUserInfo(m.user) == utils.NeedLogin {
		NeedLoginHandle(m, func(m *NeteaseModel, newMenu Menu, newTitle *MenuItem) {
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
			NeedLoginHandle(m, func(m *NeteaseModel, newMenu Menu, newTitle *MenuItem) {
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
	if code, resp := likeService.PlaylistTracks(); code != 200 {
		var msg string
		if msg, _ = jsonparser.GetString(resp, "message"); msg == "" {
			msg, _ = jsonparser.GetString(resp, "data", "message")
		}
		if msg == "" {
			msg = "加入或移出歌单失败"
		}
		utils.Notify(utils.NotifyContent{
			Title:   msg,
			Text:    songs[selectedIndex].Name,
			Url:     constants.AppGithubUrl,
			GroupId: constants.GroupID,
		})
		return
	}

	if isLike {
		utils.Notify(utils.NotifyContent{
			Title:   "已添加到我喜欢的歌曲",
			Text:    songs[selectedIndex].Name,
			Url:     utils.WebUrlOfPlaylist(m.user.MyLikePlaylistID),
			GroupId: constants.GroupID,
		})
	} else {
		utils.Notify(utils.NotifyContent{
			Title:   "已从我喜欢的歌曲移除",
			Text:    songs[selectedIndex].Name,
			Url:     utils.WebUrlOfPlaylist(m.user.MyLikePlaylistID),
			GroupId: constants.GroupID,
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
		NeedLoginHandle(m, func(m *NeteaseModel, newMenu Menu, newTitle *MenuItem) {
			trashPlayingSong(m)
		})
		return
	}

	trashService := service.FmTrashService{
		SongID: strconv.FormatInt(m.player.playlist[m.player.curSongIndex].Id, 10),
	}
	trashService.FmTrash()

	utils.Notify(utils.NotifyContent{
		Title:   "已标记为不喜欢",
		Text:    m.player.playlist[m.player.curSongIndex].Name,
		Url:     constants.AppGithubUrl,
		GroupId: constants.GroupID,
	})
}

// trashSelectedSong 标记为不喜欢
func trashSelectedSong(m *NeteaseModel) {
	loading := NewLoading(m)
	loading.start()
	defer loading.complete()

	menu, ok := m.menu.(SongsMenu)
	selectedIndex := m.menu.RealDataIndex(m.selectedIndex)
	if !ok || selectedIndex >= len(menu.Songs()) {
		return
	}
	songs := menu.Songs()

	if utils.CheckUserInfo(m.user) == utils.NeedLogin {
		NeedLoginHandle(m, func(m *NeteaseModel, newMenu Menu, newTitle *MenuItem) {
			trashSelectedSong(m)
		})
		return
	}

	trashService := service.FmTrashService{
		SongID: strconv.FormatInt(songs[selectedIndex].Id, 10),
	}
	trashService.FmTrash()

	utils.Notify(utils.NotifyContent{
		Title:   "已标记为不喜欢",
		Text:    songs[selectedIndex].Name,
		Url:     constants.AppGithubUrl,
		GroupId: constants.GroupID,
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

	menu, ok := m.menu.(SongsMenu)
	selectedIndex := m.menu.RealDataIndex(m.selectedIndex)
	if !ok || selectedIndex >= len(menu.Songs()) {
		return
	}
	songs := menu.Songs()
	go utils.DownloadMusic(songs[selectedIndex])
}

func downloadPlayingSong(m *NeteaseModel) {
	loading := NewLoading(m)
	loading.start()
	defer loading.complete()

	if m.player.curSongIndex >= len(m.player.playlist) {
		return
	}

	go utils.DownloadMusic(m.player.playlist[m.player.curSongIndex])
}

func albumOfPlayingSong(m *NeteaseModel) {
	loading := NewLoading(m)
	loading.start()
	defer loading.complete()

	if m.player.curSongIndex >= len(m.player.playlist) {
		return
	}

	curSong := m.player.playlist[m.player.curSongIndex]
	// 避免重复进入
	if detail, ok := m.menu.(*AlbumDetailMenu); ok && detail.albumId == curSong.Album.Id {
		return
	}

	enterMenu(m, NewAlbumDetailMenu(curSong.Album.Id), &MenuItem{Title: curSong.Album.Name, Subtitle: "「" + curSong.Name + "」所属专辑"})
}

func albumOfSelectedSong(m *NeteaseModel) {
	loading := NewLoading(m)
	loading.start()
	defer loading.complete()

	menu, ok := m.menu.(SongsMenu)
	selectedIndex := m.menu.RealDataIndex(m.selectedIndex)
	if !ok || selectedIndex >= len(menu.Songs()) {
		return
	}
	songs := menu.Songs()

	// 避免重复进入
	if detail, ok := m.menu.(*AlbumDetailMenu); ok && detail.albumId == songs[selectedIndex].Album.Id {
		return
	}

	enterMenu(m, NewAlbumDetailMenu(songs[selectedIndex].Album.Id), &MenuItem{Title: songs[selectedIndex].Album.Name, Subtitle: "「" + songs[selectedIndex].Name + "」所属专辑"})
}

func artistOfPlayingSong(m *NeteaseModel) {
	loading := NewLoading(m)
	loading.start()
	defer loading.complete()

	if m.player.curSongIndex >= len(m.player.playlist) {
		return
	}
	curSong := m.player.playlist[m.player.curSongIndex]
	artistCount := len(curSong.Artists)
	if artistCount <= 0 {
		return
	}
	if artistCount == 1 {
		// 避免重复进入
		if detail, ok := m.menu.(*ArtistDetailMenu); ok && detail.artistId == curSong.Artists[0].Id {
			return
		}
		enterMenu(m, NewArtistDetailMenu(curSong.Artists[0].Id, curSong.Artists[0].Name), &MenuItem{Title: curSong.Artists[0].Name, Subtitle: "「" + curSong.Name + "」所属歌手"})
		return
	}
	// 避免重复进入
	if artists, ok := m.menu.(*ArtistsOfSongMenu); ok && artists.song.Id == curSong.Id {
		return
	}
	enterMenu(m, NewArtistsOfSongMenu(curSong), &MenuItem{Title: "「" + curSong.Name + "」所属歌手"})
}

func artistOfSelectedSong(m *NeteaseModel) {
	loading := NewLoading(m)
	loading.start()
	defer loading.complete()

	menu, ok := m.menu.(SongsMenu)
	selectedIndex := m.menu.RealDataIndex(m.selectedIndex)
	if !ok || selectedIndex >= len(menu.Songs()) {
		return
	}
	songs := menu.Songs()
	song := songs[selectedIndex]
	artistCount := len(song.Artists)
	if artistCount <= 0 {
		return
	}
	if artistCount == 1 {
		// 避免重复进入
		if detail, ok := m.menu.(*ArtistDetailMenu); ok && detail.artistId == song.Artists[0].Id {
			return
		}
		enterMenu(m, NewArtistDetailMenu(song.Artists[0].Id, song.Artists[0].Name), &MenuItem{Title: song.Artists[0].Name, Subtitle: "「" + song.Name + "」所属歌手"})
		return
	}
	// 避免重复进入
	if artists, ok := m.menu.(*ArtistsOfSongMenu); ok && artists.song.Id == song.Id {
		return
	}
	enterMenu(m, NewArtistsOfSongMenu(song), &MenuItem{Title: "「" + song.Name + "」所属歌手"})
}

func openPlayingSongInWeb(m *NeteaseModel) {
	loading := NewLoading(m)
	loading.start()
	defer loading.complete()

	if m.player.curSongIndex >= len(m.player.playlist) {
		return
	}
	curSong := m.player.playlist[m.player.curSongIndex]

	_ = open.Start(utils.WebUrlOfSong(curSong.Id))
}

func openSelectedItemInWeb(m *NeteaseModel) {
	loading := NewLoading(m)
	loading.start()
	defer loading.complete()

	selectedIndex := m.menu.RealDataIndex(m.selectedIndex)

	// 打开歌曲
	if songMenu, ok := m.menu.(SongsMenu); ok && selectedIndex < len(songMenu.Songs()) {
		_ = open.Start(utils.WebUrlOfSong(songMenu.Songs()[selectedIndex].Id))
		return
	}

	// 打开歌单
	if playlistMenu, ok := m.menu.(PlaylistsMenu); ok && selectedIndex < len(playlistMenu.Playlists()) {
		_ = open.Start(utils.WebUrlOfPlaylist(playlistMenu.Playlists()[selectedIndex].Id))
		return
	}

	// 打开专辑
	if albumMenu, ok := m.menu.(AlbumsMenu); ok && selectedIndex < len(albumMenu.Albums()) {
		_ = open.Start(utils.WebUrlOfAlbum(albumMenu.Albums()[selectedIndex].Id))
		return
	}

	// 打开歌手
	if artistMenu, ok := m.menu.(ArtistsMenu); ok && selectedIndex < len(artistMenu.Artists()) {
		_ = open.Start(utils.WebUrlOfArtist(artistMenu.Artists()[selectedIndex].Id))
		return
	}
}

// collectSelectedPlaylist 收藏选中歌单
func collectSelectedPlaylist(m *NeteaseModel, isCollect bool) {
	loading := NewLoading(m)
	loading.start()
	defer loading.complete()

	menu, ok := m.menu.(PlaylistsMenu)
	if !ok || m.selectedIndex >= len(menu.Playlists()) {
		return
	}
	playlists := menu.Playlists()

	var t = "1"
	if !isCollect {
		t = "0"
	}
	s := service.PlaylistSubscribeService{ID: strconv.FormatInt(playlists[m.selectedIndex].Id, 10), T: t}
	if code, resp := s.PlaylistSubscribe(); code != 200 {
		var msg string
		if msg, _ = jsonparser.GetString(resp, "message"); msg == "" {
			msg, _ = jsonparser.GetString(resp, "data", "message")
		}
		if msg == "" {
			msg = "收藏歌单或移除歌单失败"
		}
		utils.Notify(utils.NotifyContent{
			Title:   msg,
			Text:    playlists[m.selectedIndex].Name,
			Url:     constants.AppGithubUrl,
			GroupId: constants.GroupID,
		})
		return
	}

	if isCollect {
		utils.Notify(utils.NotifyContent{
			Title:   "已收藏歌单",
			Text:    playlists[m.selectedIndex].Name,
			Url:     constants.AppGithubUrl,
			GroupId: constants.GroupID,
		})
	} else {
		utils.Notify(utils.NotifyContent{
			Title:   "已移除收藏歌单",
			Text:    playlists[m.selectedIndex].Name,
			Url:     constants.AppGithubUrl,
			GroupId: constants.GroupID,
		})
	}
}

// addSongToPlaylist 添加歌曲到播放列表
func addSongToPlaylist(m *NeteaseModel, addToNext bool) {
	loading := NewLoading(m)
	loading.start()
	defer loading.complete()

	menu, ok := m.menu.(SongsMenu)
	selectedIndex := m.menu.RealDataIndex(m.selectedIndex)
	if !ok || selectedIndex >= len(menu.Songs()) {
		return
	}
	songs := menu.Songs()

	var notifyTitle string
	if addToNext {
		// 添加为下一曲
		targetIndex := m.player.curSongIndex + 1
		m.player.playlist = append(m.player.playlist, structs.Song{})
		copy(m.player.playlist[targetIndex+1:], m.player.playlist[targetIndex:])
		m.player.playlist[targetIndex] = songs[selectedIndex]
		notifyTitle = "已添加为下一曲播放"
	} else {
		// 添加到播放列表末尾
		m.player.playlist = append(m.player.playlist, songs[selectedIndex])
		notifyTitle = "已添加到播放列表末尾"
	}
	// 替换播放中数据，避免数据错乱
	m.player.playingMenu = nil
	m.player.playingMenuKey += "modified"

	if menu, ok := m.menu.(*CurPlaylist); ok {
		menu.songs = m.player.playlist
		menu.menus = GetViewFromSongs(m.player.playlist)
		m.refreshMenuList()
	}

	utils.Notify(utils.NotifyContent{
		Title:   notifyTitle,
		Text:    songs[selectedIndex].Name,
		Url:     utils.WebUrlOfSong(songs[selectedIndex].Id),
		GroupId: constants.GroupID,
	})
}

// 从播放列表删除选中歌曲,仅在当前播放列表界面有效
func delSongFromPlaylist(m *NeteaseModel) {
	loading := NewLoading(m)
	loading.start()
	defer loading.complete()

	menu, ok := m.menu.(*CurPlaylist)
	selectedIndex := m.menu.RealDataIndex(m.selectedIndex)
	if !ok || selectedIndex >= len(menu.Songs()) {
		return
	}
	// 选中歌曲为当前播放歌曲时处理逻辑
	if m.player.curSongIndex == selectedIndex && m.player.curSong.Id == menu.Songs()[selectedIndex].Id {
		// 防止用户快速删除当前播放歌曲导致错位
		if m.player.State() >= player.Playing && m.player.playedTime.Seconds() < 2 {
			return
		}
		// 末尾歌曲删除向前退
		if m.player.curSongIndex+1 >= len(m.player.playlist) {
			m.player.curSongIndex = len(m.player.playlist) - 1
			m.player.Previous()
		} else {
			m.player.PlaySong(m.player.playlist[m.player.curSongIndex+1], DurationNext)

		}
	}
	// 以下2行 为防止切片越界
	m.player.playlist = append(m.player.playlist[:selectedIndex], m.player.playlist[selectedIndex+1:]...)
	songs := m.player.playlist
	menu.menus = GetViewFromSongs(songs)
	menu.songs = songs
	// 更新当前歌曲下标
	if selectedIndex < m.player.curSongIndex {
		m.player.curSongIndex = m.player.curSongIndex - 1
	}
	// 更新游标位置
	if m.selectedIndex >= len(menu.Songs()) {
		m.selectedIndex = len(menu.Songs()) - 1
	}

	// 替换播放中数据，避免数据错乱
	m.player.playingMenu = nil
	m.player.playingMenuKey += "modified"

	m.refreshMenuList()
}

func clearSongCache(m *NeteaseModel) {
	loading := NewLoading(m)
	loading.start()
	defer loading.complete()
	err := utils.ClearMusicCache()
	if err != nil {
		utils.Notify(utils.NotifyContent{
			Title:   "清除缓存失败",
			Text:    err.Error(),
			GroupId: constants.GroupID,
		})
	} else {
		utils.Notify(utils.NotifyContent{
			Title:   "清除缓存成功",
			Text:    "缓存已清除",
			GroupId: constants.GroupID,
		})
	}
}
