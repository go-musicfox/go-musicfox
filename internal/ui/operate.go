package ui

import (
	"log"
	"os"
	"path/filepath"
	"slices"
	"strconv"

	"github.com/anhoder/foxful-cli/model"
	"github.com/buger/jsonparser"
	"github.com/go-musicfox/netease-music/service"
	"github.com/skratchdot/open-golang/open"

	"github.com/go-musicfox/go-musicfox/internal/storage"
	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/internal/types"
	"github.com/go-musicfox/go-musicfox/utils/app"
	"github.com/go-musicfox/go-musicfox/utils/likelist"
	"github.com/go-musicfox/go-musicfox/utils/menux"
	"github.com/go-musicfox/go-musicfox/utils/netease"
	"github.com/go-musicfox/go-musicfox/utils/notify"
	"github.com/go-musicfox/go-musicfox/utils/storagex"
	_struct "github.com/go-musicfox/go-musicfox/utils/struct"
)

// likePlayingSong like/unlike playing song
func likePlayingSong(m *Netease, isLike bool) model.Page {
	loading := model.NewLoading(m.MustMain())
	loading.Start()
	defer loading.Complete()

	if m.player.curSongIndex >= len(m.player.playlist) {
		return nil
	}

	if _struct.CheckUserInfo(m.user) == _struct.NeedLogin {
		page, _ := m.ToLoginPage(func() model.Page {
			likePlayingSong(m, isLike)
			return nil
		})
		return page
	}

	// 防止出现空，兜底
	if m.user.MyLikePlaylistID == 0 {
		userPlaylists := service.UserPlaylistService{
			Uid:    strconv.FormatInt(m.user.UserId, 10),
			Limit:  strconv.Itoa(1),
			Offset: strconv.Itoa(0),
		}
		code, response := userPlaylists.UserPlaylist()
		codeType := _struct.CheckCode(code)
		if codeType == _struct.NeedLogin {
			page, _ := m.ToLoginPage(func() model.Page {
				likePlayingSong(m, isLike)
				return nil
			})
			return page
		} else if codeType != _struct.Success {
			return nil
		}
		var err error
		m.user.MyLikePlaylistID, err = jsonparser.GetInt(response, "playlist", "[0]", "id")
		if err != nil {
			log.Printf("获取歌单ID失败: %+v\n", err)
			return nil
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
		notify.Notify(notify.NotifyContent{
			Title:   msg,
			Text:    m.player.playlist[m.player.curSongIndex].Name,
			Url:     types.AppGithubUrl,
			GroupId: types.GroupID,
		})
		return nil
	}

	go func() {
		likelist.RefreshLikeList(m.user.UserId)
		m.Rerender(false)
	}()

	if isLike {
		notify.Notify(notify.NotifyContent{
			Title:   "已添加到我喜欢的歌曲",
			Text:    m.player.playlist[m.player.curSongIndex].Name,
			Url:     netease.WebUrlOfPlaylist(m.user.MyLikePlaylistID),
			GroupId: types.GroupID,
		})
	} else {
		notify.Notify(notify.NotifyContent{
			Title:   "已从我喜欢的歌曲移除",
			Text:    m.player.playlist[m.player.curSongIndex].Name,
			Url:     netease.WebUrlOfPlaylist(m.user.MyLikePlaylistID),
			GroupId: types.GroupID,
		})
	}
	return nil
}

// logout 登出
func logout() {
	table := storage.NewTable()
	_ = table.DeleteByKVModel(storage.User{})
	(&storage.LastfmUser{}).Clear()
	notify.Notify(notify.NotifyContent{
		Title:   "登出成功",
		Text:    "已清理用户信息",
		Url:     types.AppGithubUrl,
		GroupId: types.GroupID,
	})
	_ = os.Remove(filepath.Join(app.DataRootDir(), "cookie"))
}

// likeSelectedSong like/unlike selected song
func likeSelectedSong(m *Netease, isLike bool) model.Page {
	loading := model.NewLoading(m.MustMain())
	loading.Start()
	defer loading.Complete()

	var (
		main = m.MustMain()
		menu = main.CurMenu()
	)
	me, ok := menu.(SongsMenu)
	selectedIndex := menu.RealDataIndex(main.SelectedIndex())
	if !ok || selectedIndex >= len(me.Songs()) {
		return nil
	}
	songs := me.Songs()

	if _struct.CheckUserInfo(m.user) == _struct.NeedLogin {
		page, _ := m.ToLoginPage(func() model.Page {
			likeSelectedSong(m, isLike)
			return nil
		})
		return page
	}

	// 防止出现空，兜底
	if m.user.MyLikePlaylistID == 0 {
		userPlaylists := service.UserPlaylistService{
			Uid:    strconv.FormatInt(m.user.UserId, 10),
			Limit:  strconv.Itoa(1),
			Offset: strconv.Itoa(0),
		}
		code, response := userPlaylists.UserPlaylist()
		codeType := _struct.CheckCode(code)
		if codeType == _struct.NeedLogin {
			page, _ := m.ToLoginPage(func() model.Page {
				likeSelectedSong(m, isLike)
				return nil
			})
			return page
		} else if codeType != _struct.Success {
			return nil
		}
		var err error
		m.user.MyLikePlaylistID, err = jsonparser.GetInt(response, "playlist", "[0]", "id")
		if err != nil {
			log.Printf("获取歌单ID失败: %+v\n", err)
			return nil
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
		notify.Notify(notify.NotifyContent{
			Title:   msg,
			Text:    songs[selectedIndex].Name,
			Url:     types.AppGithubUrl,
			GroupId: types.GroupID,
		})
		return nil
	}

	if isLike {
		notify.Notify(notify.NotifyContent{
			Title:   "已添加到我喜欢的歌曲",
			Text:    songs[selectedIndex].Name,
			Url:     netease.WebUrlOfPlaylist(m.user.MyLikePlaylistID),
			GroupId: types.GroupID,
		})
	} else {
		notify.Notify(notify.NotifyContent{
			Title:   "已从我喜欢的歌曲移除",
			Text:    songs[selectedIndex].Name,
			Url:     netease.WebUrlOfPlaylist(m.user.MyLikePlaylistID),
			GroupId: types.GroupID,
		})
	}
	return nil
}

// trashPlayingSong 标记为不喜欢
func trashPlayingSong(m *Netease) model.Page {
	loading := model.NewLoading(m.MustMain())
	loading.Start()
	defer loading.Complete()

	if m.player.curSongIndex >= len(m.player.playlist) {
		return nil
	}

	if _struct.CheckUserInfo(m.user) == _struct.NeedLogin {
		page, _ := m.ToLoginPage(func() model.Page {
			trashPlayingSong(m)
			return nil
		})
		return page
	}

	trashService := service.FmTrashService{
		SongID: strconv.FormatInt(m.player.playlist[m.player.curSongIndex].Id, 10),
	}
	trashService.FmTrash()

	notify.Notify(notify.NotifyContent{
		Title:   "已标记为不喜欢",
		Text:    m.player.playlist[m.player.curSongIndex].Name,
		Url:     types.AppGithubUrl,
		GroupId: types.GroupID,
	})
	return nil
}

// trashSelectedSong 标记为不喜欢
func trashSelectedSong(m *Netease) model.Page {
	loading := model.NewLoading(m.MustMain())
	loading.Start()
	defer loading.Complete()

	var (
		main = m.MustMain()
		menu = main.CurMenu()
	)
	me, ok := menu.(SongsMenu)
	selectedIndex := menu.RealDataIndex(main.SelectedIndex())
	if !ok || selectedIndex >= len(me.Songs()) {
		return nil
	}
	songs := me.Songs()

	if _struct.CheckUserInfo(m.user) == _struct.NeedLogin {
		page, _ := m.ToLoginPage(func() model.Page {
			trashSelectedSong(m)
			return nil
		})
		return page
	}

	trashService := service.FmTrashService{
		SongID: strconv.FormatInt(songs[selectedIndex].Id, 10),
	}
	trashService.FmTrash()

	notify.Notify(notify.NotifyContent{
		Title:   "已标记为不喜欢",
		Text:    songs[selectedIndex].Name,
		Url:     types.AppGithubUrl,
		GroupId: types.GroupID,
	})
	return nil
}

// 下载当前音乐
func downloadSelectedSong(m *Netease) {
	loading := model.NewLoading(m.MustMain())
	loading.Start()
	defer loading.Complete()

	var (
		main = m.MustMain()
		menu = main.CurMenu()
	)
	me, ok := menu.(SongsMenu)
	selectedIndex := menu.RealDataIndex(main.SelectedIndex())
	if !ok || selectedIndex >= len(me.Songs()) {
		return
	}
	songs := me.Songs()
	go storagex.DownloadMusic(songs[selectedIndex])
}

func downloadPlayingSong(m *Netease) {
	loading := model.NewLoading(m.MustMain())
	loading.Start()
	defer loading.Complete()

	if m.player.curSongIndex >= len(m.player.playlist) {
		return
	}

	go storagex.DownloadMusic(m.player.playlist[m.player.curSongIndex])
}

func albumOfPlayingSong(m *Netease) {
	loading := model.NewLoading(m.MustMain())
	loading.Start()
	defer loading.Complete()

	var (
		main = m.MustMain()
		menu = main.CurMenu()
	)
	if m.player.curSongIndex >= len(m.player.playlist) {
		return
	}

	curSong := m.player.playlist[m.player.curSongIndex]
	// 避免重复进入
	if detail, ok := menu.(*AlbumDetailMenu); ok && detail.albumId == curSong.Album.Id {
		return
	}

	main.EnterMenu(NewAlbumDetailMenu(newBaseMenu(m), curSong.Album.Id), &model.MenuItem{Title: curSong.Album.Name, Subtitle: "「" + curSong.Name + "」所属专辑"})
}

func albumOfSelectedSong(m *Netease) {
	loading := model.NewLoading(m.MustMain())
	loading.Start()
	defer loading.Complete()

	var (
		main = m.MustMain()
		menu = main.CurMenu()
	)
	me, ok := menu.(SongsMenu)
	selectedIndex := menu.RealDataIndex(main.SelectedIndex())
	if !ok || selectedIndex >= len(me.Songs()) {
		return
	}
	songs := me.Songs()

	// 避免重复进入
	if detail, ok := menu.(*AlbumDetailMenu); ok && detail.albumId == songs[selectedIndex].Album.Id {
		return
	}

	main.EnterMenu(NewAlbumDetailMenu(newBaseMenu(m), songs[selectedIndex].Album.Id), &model.MenuItem{Title: songs[selectedIndex].Album.Name, Subtitle: "「" + songs[selectedIndex].Name + "」所属专辑"})
}

func artistOfPlayingSong(m *Netease) {
	loading := model.NewLoading(m.MustMain())
	loading.Start()
	defer loading.Complete()

	var (
		main = m.MustMain()
		menu = main.CurMenu()
	)
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
		if detail, ok := menu.(*ArtistDetailMenu); ok && detail.artistId == curSong.Artists[0].Id {
			return
		}
		main.EnterMenu(NewArtistDetailMenu(newBaseMenu(m), curSong.Artists[0].Id, curSong.Artists[0].Name), &model.MenuItem{Title: curSong.Artists[0].Name, Subtitle: "「" + curSong.Name + "」所属歌手"})
		return
	}
	// 避免重复进入
	if artists, ok := menu.(*ArtistsOfSongMenu); ok && artists.song.Id == curSong.Id {
		return
	}
	main.EnterMenu(NewArtistsOfSongMenu(newBaseMenu(m), curSong), &model.MenuItem{Title: "「" + curSong.Name + "」所属歌手"})
}

func artistOfSelectedSong(m *Netease) {
	loading := model.NewLoading(m.MustMain())
	loading.Start()
	defer loading.Complete()

	var (
		main = m.MustMain()
		menu = main.CurMenu()
	)
	me, ok := menu.(SongsMenu)
	selectedIndex := menu.RealDataIndex(main.SelectedIndex())
	if !ok || selectedIndex >= len(me.Songs()) {
		return
	}
	songs := me.Songs()
	song := songs[selectedIndex]
	artistCount := len(song.Artists)
	if artistCount <= 0 {
		return
	}
	if artistCount == 1 {
		// 避免重复进入
		if detail, ok := menu.(*ArtistDetailMenu); ok && detail.artistId == song.Artists[0].Id {
			return
		}
		main.EnterMenu(NewArtistDetailMenu(newBaseMenu(m), song.Artists[0].Id, song.Artists[0].Name), &model.MenuItem{Title: song.Artists[0].Name, Subtitle: "「" + song.Name + "」所属歌手"})
		return
	}
	// 避免重复进入
	if artists, ok := menu.(*ArtistsOfSongMenu); ok && artists.song.Id == song.Id {
		return
	}
	main.EnterMenu(NewArtistsOfSongMenu(newBaseMenu(m), song), &model.MenuItem{Title: "「" + song.Name + "」所属歌手"})
}

func openPlayingSongInWeb(m *Netease) {
	loading := model.NewLoading(m.MustMain())
	loading.Start()
	defer loading.Complete()

	if m.player.curSongIndex >= len(m.player.playlist) {
		return
	}
	curSong := m.player.playlist[m.player.curSongIndex]

	_ = open.Start(netease.WebUrlOfSong(curSong.Id))
}

func openSelectedItemInWeb(m *Netease) {
	loading := model.NewLoading(m.MustMain())
	loading.Start()
	defer loading.Complete()

	var (
		main = m.MustMain()
		menu = main.CurMenu()
	)
	selectedIndex := menu.RealDataIndex(main.SelectedIndex())

	// 打开歌曲
	if songMenu, ok := menu.(SongsMenu); ok && selectedIndex < len(songMenu.Songs()) {
		_ = open.Start(netease.WebUrlOfSong(songMenu.Songs()[selectedIndex].Id))
		return
	}

	// 打开歌单
	if playlistMenu, ok := menu.(PlaylistsMenu); ok && selectedIndex < len(playlistMenu.Playlists()) {
		_ = open.Start(netease.WebUrlOfPlaylist(playlistMenu.Playlists()[selectedIndex].Id))
		return
	}

	// 打开专辑
	if albumMenu, ok := menu.(AlbumsMenu); ok && selectedIndex < len(albumMenu.Albums()) {
		_ = open.Start(netease.WebUrlOfAlbum(albumMenu.Albums()[selectedIndex].Id))
		return
	}

	// 打开歌手
	if artistMenu, ok := menu.(ArtistsMenu); ok && selectedIndex < len(artistMenu.Artists()) {
		_ = open.Start(netease.WebUrlOfArtist(artistMenu.Artists()[selectedIndex].Id))
		return
	}
}

// collectSelectedPlaylist 收藏选中歌单
func collectSelectedPlaylist(m *Netease, isCollect bool) model.Page {
	loading := model.NewLoading(m.MustMain())
	loading.Start()
	defer loading.Complete()

	if _struct.CheckUserInfo(m.user) == _struct.NeedLogin {
		page, _ := m.ToLoginPage(func() model.Page {
			collectSelectedPlaylist(m, isCollect)
			return nil
		})
		return page
	}

	var (
		main = m.MustMain()
		menu = main.CurMenu()
	)
	me, ok := menu.(PlaylistsMenu)
	if !ok || main.SelectedIndex() >= len(me.Playlists()) {
		return nil
	}
	playlists := me.Playlists()

	t := "1"
	if !isCollect {
		t = "0"
	}
	s := service.PlaylistSubscribeService{ID: strconv.FormatInt(playlists[main.SelectedIndex()].Id, 10), T: t}
	if code, resp := s.PlaylistSubscribe(); code != 200 {
		var msg string
		if msg, _ = jsonparser.GetString(resp, "message"); msg == "" {
			msg, _ = jsonparser.GetString(resp, "data", "message")
		}
		if msg == "" {
			msg = "收藏歌单或移除歌单失败"
		}
		notify.Notify(notify.NotifyContent{
			Title:   msg,
			Text:    playlists[main.SelectedIndex()].Name,
			Url:     types.AppGithubUrl,
			GroupId: types.GroupID,
		})
		return nil
	}

	if isCollect {
		notify.Notify(notify.NotifyContent{
			Title:   "已收藏歌单",
			Text:    playlists[main.SelectedIndex()].Name,
			Url:     types.AppGithubUrl,
			GroupId: types.GroupID,
		})
	} else {
		notify.Notify(notify.NotifyContent{
			Title:   "已移除收藏歌单",
			Text:    playlists[main.SelectedIndex()].Name,
			Url:     types.AppGithubUrl,
			GroupId: types.GroupID,
		})
	}
	return nil
}

// appendSongsToCurPlaylist 添加歌曲到播放列表
func appendSongsToCurPlaylist(m *Netease, addToNext bool) {
	loading := model.NewLoading(m.MustMain())
	loading.Start()
	defer loading.Complete()

	var (
		main          = m.MustMain()
		menu          = main.CurMenu()
		selectedIndex = menu.RealDataIndex(main.SelectedIndex())
		subMenu       = menu.SubMenu(m.App, selectedIndex)
		appendSongs   []structs.Song
		notifyURL     string
	)

	sm, isSongsMenu := menu.(SongsMenu)
	subSm, subIsSongsMenu := subMenu.(SongsMenu)
	switch {
	case isSongsMenu: // 当前菜单是 SongMenu
		if selectedIndex >= len(sm.Songs()) {
			return
		}
		song := sm.Songs()[selectedIndex]
		appendSongs = append(appendSongs, song)
		notifyURL = netease.WebUrlOfSong(song.Id)
	case subIsSongsMenu: // 选中项菜单是 SongsMenu
		// 触发 BeforeEnterMenuHook 获取歌曲
		if ok, _ := subSm.BeforeEnterMenuHook()(m.Main()); !ok {
			return
		}
		appendSongs = subSm.Songs()
	default:
		return
	}

	var notifyTitle string
	if addToNext && len(m.player.playlist) > 0 {
		// 添加为下一曲
		targetIndex := m.player.curSongIndex + 1
		m.player.playlist = slices.Concat(m.player.playlist[:targetIndex], appendSongs, m.player.playlist[targetIndex:])
		notifyTitle = "已添加到下一曲"
	} else {
		// 添加到播放列表末尾
		m.player.playlist = append(m.player.playlist, appendSongs...)
		notifyTitle = "已添加到播放列表末尾"
	}

	// 替换播放中数据，避免数据错乱
	m.player.playingMenu = nil
	m.player.playingMenuKey += "modified"

	if curPlaylist, ok := menu.(*CurPlaylist); ok {
		curPlaylist.songs = m.player.playlist
		curPlaylist.menus = menux.GetViewFromSongs(m.player.playlist)
		main.RefreshMenuList()
	}

	notify.Notify(notify.NotifyContent{
		Title:   notifyTitle,
		Text:    menu.MenuViews()[selectedIndex].Title,
		Url:     notifyURL,
		GroupId: types.GroupID,
	})
}

// openAddSongToUserPlaylistMenu 打开添加歌曲到用户歌单菜单
func openAddSongToUserPlaylistMenu(m *Netease, isSelected, isAdd bool) model.Page {
	loading := model.NewLoading(m.MustMain())
	loading.Start()
	defer loading.Complete()

	if _struct.CheckUserInfo(m.user) == _struct.NeedLogin {
		page, _ := m.ToLoginPage(func() model.Page {
			openAddSongToUserPlaylistMenu(m, isSelected, isAdd)
			return nil
		})
		return page
	}

	var (
		main = m.MustMain()
		menu = main.CurMenu()
	)
	switch me := menu.(type) {
	case SongsMenu:
		if menu.RealDataIndex(main.SelectedIndex()) >= len(me.Songs()) {
			return nil
		}
	default:
		if isSelected {
			return nil
		}
	}
	// 避免重复进入
	if _, ok := menu.(*AddToUserPlaylistMenu); ok {
		return nil
	}
	var song structs.Song
	var subtitle string
	if isSelected {
		song = menu.(SongsMenu).Songs()[menu.RealDataIndex(main.SelectedIndex())]
	} else {
		song = m.player.curSong
	}
	if isAdd {
		subtitle = "将「" + song.Name + "」加入歌单"
	} else {
		subtitle = "将「" + song.Name + "」从歌单中删除"
	}
	main.EnterMenu(NewAddToUserPlaylistMenu(newBaseMenu(m), m.user.UserId, song, isAdd), &model.MenuItem{Title: "我的歌单", Subtitle: subtitle})
	return nil
}

// addSongToUserPlaylist 添加歌曲到用户歌单
func addSongToUserPlaylist(m *Netease, isAdd bool) model.Page {
	loading := model.NewLoading(m.MustMain())
	loading.Start()
	defer loading.Complete()

	if _struct.CheckUserInfo(m.user) == _struct.NeedLogin {
		page, _ := m.ToLoginPage(func() model.Page {
			addSongToUserPlaylist(m, isAdd)
			return nil
		})
		return page
	}

	var (
		main = m.MustMain()
		menu = main.CurMenu()
	)
	me := menu.(*AddToUserPlaylistMenu)
	playlist := me.playlists[menu.RealDataIndex(main.SelectedIndex())]

	var op string
	if isAdd {
		op = "add"
	} else {
		op = "del"
	}
	likeService := service.PlaylistTracksService{
		TrackIds: []string{strconv.FormatInt(me.song.Id, 10)},
		Op:       op,
		Pid:      strconv.FormatInt(playlist.Id, 10),
	}
	if code, resp := likeService.PlaylistTracks(); code != 200 {
		var msg string
		if msg, _ = jsonparser.GetString(resp, "message"); msg == "" {
			msg, _ = jsonparser.GetString(resp, "data", "message")
		}
		if msg == "" && isAdd {
			msg = "加入歌单失败"
		} else if msg == "" && !isAdd {
			msg = "从歌单中删除失败"
		}
		notify.Notify(notify.NotifyContent{
			Title:   msg,
			Text:    me.song.Name,
			Url:     types.AppGithubUrl,
			GroupId: types.GroupID,
		})
		main.BackMenu()
		return nil
	}

	var title string
	if isAdd {
		title = "已添加到歌单「" + playlist.Name + "」"
	} else {
		title = "已从歌单「" + playlist.Name + "」中删除"
	}
	notify.Notify(notify.NotifyContent{
		Title:   title,
		Text:    me.song.Name,
		Url:     netease.WebUrlOfPlaylist(playlist.Id),
		GroupId: types.GroupID,
	})
	main.BackMenu()
	switch mt := menu.(type) {
	case *PlaylistDetailMenu:
		// 刷新菜单
		if !isAdd && mt.playlistId == playlist.Id {
			t := main.MenuTitle()
			main.BackMenu()
			_, page := menu.BeforeEnterMenuHook()(main)
			main.EnterMenu(menu, t)
			return page
		}
	default:
	}
	return nil
}

// 从播放列表删除选中歌曲,仅在当前播放列表界面有效
func delSongFromPlaylist(m *Netease) model.Page {
	loading := model.NewLoading(m.MustMain())
	loading.Start()
	defer loading.Complete()

	var (
		main = m.MustMain()
		menu = main.CurMenu()
	)
	me, ok := menu.(*CurPlaylist)
	selectedIndex := menu.RealDataIndex(main.SelectedIndex())
	if !ok || selectedIndex >= len(me.Songs()) {
		return nil
	}
	// 防止切片越界
	if len(m.player.playlist) == 0 {
		return nil
	}
	// 选中歌曲为当前播放歌曲时处理逻辑
	if m.player.curSongIndex == selectedIndex && m.player.curSong.Id == me.Songs()[selectedIndex].Id {
		// 防止用户快速删除当前播放歌曲导致错位
		if m.player.State() >= types.Playing && m.player.PassedTime().Seconds() < 2 {
			return nil
		}
		// 末尾歌曲删除向前退
		if m.player.curSongIndex+1 >= len(m.player.playlist) {
			m.player.curSongIndex = len(m.player.playlist) - 1
			// 不在只剩一个歌曲的情况下重新播放歌曲
			if len(m.player.playlist) > 1 {
				m.player.PreviousSong(false)
			}
		} else {
			_ = m.player.PlaySong(m.player.playlist[m.player.curSongIndex+1], DurationNext)
		}
	}
	// 以下2行 为防止切片越界
	m.player.playlist = append(m.player.playlist[:selectedIndex], m.player.playlist[selectedIndex+1:]...)
	songs := m.player.playlist
	me.menus = menux.GetViewFromSongs(songs)
	me.songs = songs
	// 更新当前歌曲下标
	if selectedIndex < m.player.curSongIndex {
		m.player.curSongIndex = m.player.curSongIndex - 1
	}
	// 更新游标位置
	if main.SelectedIndex() >= len(me.Songs()) {
		main.SetSelectedIndex(len(me.Songs()) - 1)
	}

	// 替换播放中数据，避免数据错乱
	m.player.playingMenu = nil
	m.player.playingMenuKey += "modified"

	// 如果播放列表中已经没有歌曲，停止播放
	if len(m.player.playlist) == 0 {
		m.player.Stop()
	}

	main.RefreshMenuList()
	return nil
}

func clearSongCache(m *Netease) {
	m.MustMain().EnterMenu(NewClearSongCacheMenu(newBaseMenu(m), m), &model.MenuItem{Title: "清除缓存", Subtitle: "确定清除缓存"})
}
