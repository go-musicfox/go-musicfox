package ui

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strconv"

	"github.com/anhoder/foxful-cli/model"
	"github.com/buger/jsonparser"
	"github.com/go-musicfox/netease-music/service"
	"github.com/skratchdot/open-golang/open"

	"github.com/go-musicfox/go-musicfox/internal/composer"
	"github.com/go-musicfox/go-musicfox/internal/storage"
	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/internal/types"
	"github.com/go-musicfox/go-musicfox/utils/app"
	"github.com/go-musicfox/go-musicfox/utils/clipboard"
	"github.com/go-musicfox/go-musicfox/utils/likelist"
	"github.com/go-musicfox/go-musicfox/utils/menux"
	"github.com/go-musicfox/go-musicfox/utils/netease"
	"github.com/go-musicfox/go-musicfox/utils/notify"
	_struct "github.com/go-musicfox/go-musicfox/utils/struct"
)

// likePlayingSong like/unlike playing song
func likePlayingSong(m *Netease, isLike bool) model.Page {
	loading := model.NewLoading(m.MustMain())
	loading.Start()
	defer loading.Complete()

	if m.player.CurSongIndex() >= len(m.player.Playlist()) {
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
			slog.Error("获取歌单ID失败", "error", err)
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
		TrackIds: []string{strconv.FormatInt(m.player.CurSong().Id, 10)},
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
			Text:    m.player.CurSong().Name,
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
			Text:    m.player.CurSong().Name,
			Url:     netease.WebUrlOfPlaylist(m.user.MyLikePlaylistID),
			GroupId: types.GroupID,
		})
	} else {
		notify.Notify(notify.NotifyContent{
			Title:   "已从我喜欢的歌曲移除",
			Text:    m.player.CurSong().Name,
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
	notify.Notify(notify.NotifyContent{
		Title:   "登出成功",
		Text:    "已清理用户信息",
		Url:     types.AppGithubUrl,
		GroupId: types.GroupID,
	})
	_ = os.Remove(filepath.Join(app.DataDir(), "cookie"))

	//调用退出登录api
	logout_service := service.LogoutService{}
	code, _, err := logout_service.Logout()
	if err != nil {
		slog.Error("调用退出登录api错误", slog.String("error", err.Error()))
	}
	if code != 200 {
		slog.Error("退出登录状态异常")
	}
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
			slog.Error("获取歌单ID失败", slog.String("error", err.Error()))
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

	if m.player.CurSongIndex() >= len(m.player.Playlist()) {
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
		SongID: strconv.FormatInt(m.player.CurSong().Id, 10),
	}
	trashService.FmTrash()

	notify.Notify(notify.NotifyContent{
		Title:   "已标记为不喜欢",
		Text:    m.player.CurSong().Name,
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

func handleSongDownload(m *Netease, song structs.Song) {
	if song.Id == 0 {
		slog.Error("指定音乐不存在，跳过下载")
	}
	slog.Info("开始下载歌曲", "song", song.Name, "id", song.Id)
	notify.Notify(notify.NotifyContent{
		Title:   "👇🏻 正在下载，请稍候...",
		Text:    song.Name,
		GroupId: types.GroupID,
	})

	path, err := m.trackManager.DownloadSong(context.Background(), song)

	switch {
	case err == nil:
		slog.Info("歌曲下载成功", "song", song.Name, "id", song.Id, "path", path)
		notify.Notify(notify.NotifyContent{
			Title:   "✅ 下载完成",
			Text:    filepath.Base(path),
			GroupId: types.GroupID,
		})
	case errors.Is(err, os.ErrExist):
		slog.Info("歌曲文件已存在，跳过下载", "song", song.Name, "id", song.Id, "path", path)
		notify.Notify(notify.NotifyContent{
			Title:   "🙅🏻 文件已存在",
			Text:    filepath.Base(path),
			GroupId: types.GroupID,
		})
	default:
		slog.Error("歌曲下载失败", "song", song.Name, "id", song.Id, "error", err)
		notify.Notify(notify.NotifyContent{
			Title:   "❌ 下载失败",
			Text:    err.Error(),
			GroupId: types.GroupID,
		})
	}
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

	go handleSongDownload(m, songs[selectedIndex])
}

func downloadPlayingSong(m *Netease) {
	loading := model.NewLoading(m.MustMain())
	loading.Start()
	defer loading.Complete()

	if m.player.CurSongIndex() >= len(m.player.Playlist()) {
		return
	}

	go handleSongDownload(m, m.player.CurSong())
}

func simiSongsOfPlayingSong(m *Netease) {
	loading := model.NewLoading(m.MustMain())
	loading.Start()
	defer loading.Complete()

	var (
		main = m.MustMain()
		menu = main.CurMenu()
	)
	if m.player.CurSongIndex() >= len(m.player.Playlist()) {
		return
	}

	curSong := m.player.CurSong()
	// 避免重复进入
	if detail, ok := menu.(*SimilarSongsMenu); ok && detail.relateSongId == curSong.Id {
		return
	}

	main.EnterMenu(NewSimilarSongsMenu(newBaseMenu(m), curSong), &model.MenuItem{Title: "相似歌曲", Subtitle: "与「" + curSong.Name + "」相似的歌曲"})
}

func simiSongsOfSelectedSong(m *Netease) {
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
	if detail, ok := menu.(*SimilarSongsMenu); ok && detail.relateSongId == songs[selectedIndex].Id {
		return
	}

	main.EnterMenu(NewSimilarSongsMenu(newBaseMenu(m), songs[selectedIndex]), &model.MenuItem{Title: "相似歌曲", Subtitle: "与「" + songs[selectedIndex].Name + "」相似的歌曲"})
}

func albumOfPlayingSong(m *Netease) {
	loading := model.NewLoading(m.MustMain())
	loading.Start()
	defer loading.Complete()

	var (
		main = m.MustMain()
		menu = main.CurMenu()
	)
	if m.player.CurSongIndex() >= len(m.player.Playlist()) {
		return
	}

	curSong := m.player.CurSong()
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
	if m.player.CurSongIndex() >= len(m.player.Playlist()) {
		return
	}
	curSong := m.player.CurSong()
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

	if m.player.CurSongIndex() >= len(m.player.Playlist()) {
		return
	}
	curSong := m.player.CurSong()

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
	if addToNext && len(m.player.Playlist()) > 0 {
		// 添加为下一曲
		targetIndex := m.player.CurSongIndex() + 1
		_ = m.player.playlistManager.Initialize(m.player.CurSongIndex(), slices.Concat(m.player.Playlist()[:targetIndex], appendSongs, m.player.Playlist()[targetIndex:]))
		notifyTitle = "已添加到下一曲"
	} else {
		// 添加到播放列表末尾
		_ = m.player.playlistManager.Initialize(m.player.CurSongIndex(), append(m.player.Playlist(), appendSongs...))
		notifyTitle = "已添加到播放列表末尾"
	}

	// 替换播放中数据，避免数据错乱
	m.player.playingMenu = nil
	m.player.playingMenuKey += "modified"

	if curPlaylist, ok := menu.(*CurPlaylist); ok {
		curPlaylist.songs = m.player.Playlist()
		curPlaylist.menus = menux.GetViewFromSongs(m.player.Playlist())
		main.RefreshMenuList()
	}

	notify.Notify(notify.NotifyContent{
		Title:   notifyTitle,
		Text:    menu.MenuViews()[main.SelectedIndex()].Title,
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
		song = m.player.CurSong()
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
	if len(m.player.Playlist()) == 0 {
		return nil
	}

	if removedSong, err := m.player.playlistManager.RemoveSong(selectedIndex); err == nil {
		m.player.PlaySong(removedSong, DurationNext)
	}
	songs := m.player.Playlist()
	me.menus = menux.GetViewFromSongs(songs)
	me.songs = songs

	// 更新游标位置
	if main.SelectedIndex() >= len(songs) {
		main.SetSelectedIndex(len(songs) - 1)
	}

	// 替换播放中数据，避免数据错乱
	m.player.playingMenu = nil
	m.player.playingMenuKey += "modified"

	// 如果播放列表中已经没有歌曲，停止播放
	if len(m.player.Playlist()) == 0 {
		m.player.Stop()
	}

	main.RefreshMenuList()
	return nil
}

func clearSongCache(m *Netease) {
	action := func() {
		loading := model.NewLoading(m.MustMain())
		loading.Start()
		defer loading.Complete()
		err := m.trackManager.ClearCache()
		if err != nil {
			slog.Error("清除缓存失败", "error", err)
			notify.Notify(notify.NotifyContent{
				Title:   "清除缓存失败",
				Text:    err.Error(),
				GroupId: types.GroupID,
			})
		} else {
			slog.Error("清除缓存成功")
			notify.Notify(notify.NotifyContent{
				Title:   "清除缓存成功",
				Text:    "缓存已清除",
				GroupId: types.GroupID,
			})
		}
	}
	menu := NewConfirmMenu(newBaseMenu(m), []ConfirmItem{
		{title: model.MenuItem{Title: "确定"}, action: action, backLevel: 1},
	})
	m.MustMain().EnterMenu(menu, &model.MenuItem{Title: "清除缓存", Subtitle: "确定清除缓存"})
}

func handleLyricDownload(m *Netease, song structs.Song) {
	if song.Id == 0 {
		slog.Error("指定音乐不存在，跳过下载")
	}
	slog.Info("开始下载歌词", "song", song.Name, "id", song.Id)

	path, err := m.trackManager.DownloadLyric(context.Background(), song)

	switch {
	case err == nil:
		slog.Info("歌词下载成功", "song", song.Name, "id", song.Id, "path", path)
		notify.Notify(notify.NotifyContent{
			Title:   "下载歌词成功",
			Text:    filepath.Base(path),
			Url:     filepath.Base(path),
			GroupId: types.GroupID,
		})
	case errors.Is(err, os.ErrExist):
		slog.Info("歌词文件已存在，跳过下载", "song", song.Name, "id", song.Id, "path", path)
		notify.Notify(notify.NotifyContent{
			Title:   "歌词文件已存在",
			Text:    filepath.Base(path),
			GroupId: types.GroupID,
		})
	default:
		slog.Error("歌词下载失败", "song", song.Name, "id", song.Id, "error", err)
		notify.Notify(notify.NotifyContent{
			Title:   "下载歌词失败",
			Text:    err.Error(),
			Url:     types.AppGithubUrl,
			GroupId: types.GroupID,
		})
	}
}

func downloadPlayingSongLrc(m *Netease) {
	loading := model.NewLoading(m.MustMain())
	loading.Start()
	defer loading.Complete()

	if m.player.CurSongIndex() >= len(m.player.Playlist()) {
		return
	}

	go handleLyricDownload(m, m.player.CurSong())
}

func action(m *Netease, curPlaying bool) {
	var (
		main     = m.MustMain()
		menu     = main.CurMenu()
		newTitle *model.MenuItem
	)

	switch menu.(type) {
	case SongsMenu:
	case PlaylistsMenu:
	default:
		newTitle = &model.MenuItem{Title: "操作当前播放"}
	}

	menuKey := m.MustMain().CurMenu().GetMenuKey()
	if menuKey != actionMenuKey {
		newMenu := NewActionMenu(newBaseMenu(m), menuKey, curPlaying)
		main.EnterMenu(
			newMenu,
			newTitle)
	}
}

func shareToClipBoard(m *Netease, data any) {
	var handleError = func(err error) {
		slog.Error("分享失败", "error", err)
		notify.Notify(notify.NotifyContent{
			Title:   "分享失败",
			Text:    err.Error(),
			Url:     types.AppGithubUrl,
			GroupId: types.GroupID,
		})
	}

	str, err := m.shareSvc.Share(data)
	if err != nil {
		handleError(err)
	}
	if err = clipboard.Write(str); err != nil {
		handleError(err)
	}
}

func shareSelectItem(m *Netease) {
	var (
		main          = m.MustMain()
		menu          = main.CurMenu()
		selectedIndex = menu.RealDataIndex(main.SelectedIndex())
		item          any
	)

	if sharer, ok := menu.(composer.Sharer); ok {
		if item = sharer.ItemToShare(selectedIndex); item != nil {
			shareToClipBoard(m, item)
		}
		return
	}

	// For backward compatibility with legacy menu types.
	switch me := menu.(type) {
	case SongsMenu:
		if selectedIndex >= len(me.Songs()) {
			return
		}
		item = me.Songs()[selectedIndex]
	case AlbumsMenu:
		if selectedIndex >= len(me.Albums()) {
			return
		}
		item = me.Albums()[selectedIndex]
	case ArtistsMenu:
		if selectedIndex >= len(me.Artists()) {
			return
		}
		item = me.Artists()[selectedIndex]
	case PlaylistsMenu:
		if selectedIndex >= len(me.Playlists()) {
			return
		}
		item = me.Playlists()[selectedIndex]
	default:
		slog.Error(fmt.Sprintf("不支持的页面: %T", me))
	}

	if item != nil {
		shareToClipBoard(m, item)
	}
}

func sharePlayingItem(m *Netease) {
	if m.player.CurSongIndex() >= len(m.player.Playlist()) {
		return
	}
	shareToClipBoard(m, m.player.CurSong())
}
