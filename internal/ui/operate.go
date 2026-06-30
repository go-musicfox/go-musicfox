package ui

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http/cookiejar"
	"os"
	"path/filepath"
	"slices"
	"strconv"

	"github.com/anhoder/foxful-cli/model"
	"github.com/buger/jsonparser"
	"github.com/go-musicfox/go-musicfox/internal/composer"
	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/storage"
	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/internal/types"
	"github.com/go-musicfox/go-musicfox/utils/app"
	"github.com/go-musicfox/go-musicfox/utils/clipboard"
	"github.com/go-musicfox/go-musicfox/utils/errorx"
	"github.com/go-musicfox/go-musicfox/utils/likelist"
	"github.com/go-musicfox/go-musicfox/utils/menux"
	"github.com/go-musicfox/go-musicfox/utils/netease"
	"github.com/go-musicfox/go-musicfox/utils/notify"
	"github.com/go-musicfox/go-musicfox/utils/slogx"
	_struct "github.com/go-musicfox/go-musicfox/utils/struct"
	"github.com/go-musicfox/netease-music/service"
	"github.com/go-musicfox/netease-music/util"
	"github.com/skratchdot/open-golang/open"
)

// getTargetSong 根据上下文获取目标歌曲。
// isSelected: true 获取当前选择歌曲，false 获取当前播放歌曲
// 返回获取到的歌曲和一个布尔值，表示是否成功获取。
func getTargetSong(n *Netease, isSelected bool) (structs.Song, bool) {
	if !isSelected {
		if n.player.CurSongIndex() >= len(n.player.Playlist()) {
			return structs.Song{}, false
		}
		song := n.player.CurSong()
		return song, song.Id != 0
	}

	main := n.MustMain()
	menu := main.CurMenu()
	songsMenu, ok := menu.(SongsMenu)
	if !ok {
		return structs.Song{}, false
	}

	selectedIndex := menu.RealDataIndex(main.SelectedIndex())
	songs := songsMenu.Songs()
	if selectedIndex >= len(songs) {
		return structs.Song{}, false
	}

	return songs[selectedIndex], true
}

// logout 登出
func logout() {

	logout_service := service.LogoutService{}
	code, _, err := logout_service.Logout()
	if err != nil {
		slog.Error("调用退出登录api错误", slog.String("error", err.Error()))
	} else if code != 200 {
		slog.Warn("退出登录接口返回状态非200", "code", code)
	}

	table := storage.NewTable()
	if err := table.DeleteByKVModel(storage.User{}); err != nil {
		slog.Error("清理用户数据库失败", slogx.Error(err))
	}

	if emptyJar, err := cookiejar.New(nil); err == nil {
		util.SetGlobalCookieJar(emptyJar)
	}

	cookieFile := filepath.Join(app.DataDir(), "cookie")
	err = os.Remove(cookieFile)
	if err != nil {
		if !os.IsNotExist(err) {
			slog.Error("删除Cookie文件失败", slog.String("path", cookieFile), slogx.Error(err))
		}
	} else {
		slog.Info("Cookie文件已成功删除")
	}

	notify.Notify(notify.NotifyContent{
		Title:   "登出成功",
		Text:    "已清理用户信息",
		Url:     types.AppGithubUrl,
		GroupId: types.GroupID,
	})

}

// getSelectedPlaylist 获取选中的歌单
func getSelectedPlaylist(n *Netease) (structs.Playlist, bool) {
	main := n.MustMain()
	menu := main.CurMenu()
	playlistMenu, ok := menu.(PlaylistsMenu)
	if !ok {
		return structs.Playlist{}, false
	}
	selectedIndex := main.SelectedIndex()
	if selectedIndex >= len(playlistMenu.Playlists()) {
		return structs.Playlist{}, false
	}
	return playlistMenu.Playlists()[selectedIndex], true
}

// likeSong 喜欢或取消喜欢一首歌。
// isLike: true 为喜欢, false 为取消喜欢。
// isSelected: true 操作选中的歌曲, false 操作正在播放的歌曲。
func likeSong(n *Netease, isLike bool, isSelected bool) model.Page {
	coreLogic := func(n *Netease) model.Page {
		song, ok := getTargetSong(n, isSelected)
		if !ok {
			return nil
		}

		// TODO: 提取为函数
		if n.user.MyLikePlaylistID == 0 {
			userPlaylists := service.UserPlaylistService{
				Uid:    strconv.FormatInt(n.user.UserId, 10),
				Limit:  "1",
				Offset: "0",
			}
			code, response := userPlaylists.UserPlaylist()
			if _struct.CheckCode(code) != _struct.Success {
				return nil
			}
			var err error
			n.user.MyLikePlaylistID, err = jsonparser.GetInt(response, "playlist", "[0]", "id")
			if err != nil {
				slog.Error("获取歌单ID失败", "error", err)
				return nil
			}

			// 写入本地数据库
			table := storage.NewTable()
			_ = table.SetByKVModel(storage.User{}, n.user)
		}

		op := "add"
		if !isLike {
			op = "del"
		}
		likeService := service.PlaylistTracksService{
			TrackIds: []string{strconv.FormatInt(song.Id, 10)},
			Op:       op,
			Pid:      strconv.FormatInt(n.user.MyLikePlaylistID, 10),
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
				Text:    n.player.CurSong().Name,
				Url:     types.AppGithubUrl,
				GroupId: types.GroupID,
			})
			return nil
		}

		go func() {
			likelist.RefreshLikeList(n.user.UserId)
			n.Rerender(false)
		}()

		title := "已添加到我喜欢的歌曲"
		if !isLike {
			title = "已从我喜欢的歌曲移除"
		}
		notify.Notify(notify.NotifyContent{
			Title:   title,
			Text:    song.Name,
			Url:     netease.WebUrlOfPlaylist(n.user.MyLikePlaylistID),
			GroupId: types.GroupID,
		})
		return nil
	}

	return NewOperation(n, coreLogic).ShowLoading().NeedsAuth().Execute()
}

// trashSong 将歌曲标记为不喜欢
func trashSong(n *Netease, isSelected bool) model.Page {
	coreLogic := func(n *Netease) model.Page {
		song, ok := getTargetSong(n, isSelected)
		if !ok {
			return nil
		}
		trashService := service.FmTrashService{
			SongID: strconv.FormatInt(song.Id, 10),
		}
		trashService.FmTrash()

		notify.Notify(notify.NotifyContent{
			Title:   "已标记为不喜欢",
			Text:    song.Name,
			Url:     types.AppGithubUrl,
			GroupId: types.GroupID,
		})
		return nil
	}

	return NewOperation(n, coreLogic).ShowLoading().NeedsAuth().Execute()
}

// downloadSong 下载歌曲
func downloadSong(n *Netease, isSelected bool) {
	op := NewOperation(n, func(n *Netease) model.Page {
		song, ok := getTargetSong(n, isSelected)
		if !ok {
			slog.Warn("未获取到下载项")
			return nil
		}
		errorx.Go(func() { handleSongDownload(n, song) }, true)
		return nil
	})
	op.ShowLoading().Execute()
}

// handleSongDownload 歌曲下载处理器
func handleSongDownload(n *Netease, song structs.Song) {
	if song.Id == 0 {
		slog.Error("指定音乐不存在，跳过下载")
		return
	}
	slog.Info("开始下载歌曲", "song", song.Name, "id", song.Id)
	notify.Notify(notify.NotifyContent{
		Title:   "👇🏻 正在下载，请稍候...",
		Text:    song.Name,
		GroupId: types.GroupID,
	})
	path, err := n.trackManager.DownloadSong(context.Background(), song)
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

	withLyric := configs.AppConfig.Storage.DownloadSongWithLyric
	if withLyric && (err == nil || errors.Is(err, os.ErrExist)) {
		slog.Info("歌曲已下载或已存在，开始下载歌词", "song", song.Name, "id", song.Id)
		errorx.Go(func() { handleLyricDownload(n, song) }, true)
	}
}

// downloadSongLrc 下载歌词
func downloadSongLrc(n *Netease, isSelected bool) {
	op := NewOperation(n, func(n *Netease) model.Page {
		song, ok := getTargetSong(n, isSelected)
		if !ok {
			slog.Warn("未获取到下载项")
			return nil
		}
		errorx.Go(func() { handleLyricDownload(n, song) }, true)
		return nil
	})
	op.ShowLoading().Execute()
}

// handleLyricDownload 歌词下载处理器
func handleLyricDownload(n *Netease, song structs.Song) {
	if song.Id == 0 {
		slog.Error("指定音乐不存在，跳过下载")
		return
	}
	slog.Info("开始下载歌词", "song", song.Name, "id", song.Id)

	path, err := n.trackManager.DownloadLyric(context.Background(), song)

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

// findSimilarSongs 查找相似歌曲
func findSimilarSongs(n *Netease, isSelected bool) {
	op := NewOperation(n, func(n *Netease) model.Page {
		song, ok := getTargetSong(n, isSelected)
		if !ok {
			return nil
		}
		main := n.MustMain()
		menu := main.CurMenu()
		if detail, ok := menu.(*SimilarSongsMenu); ok && detail.relateSongId == song.Id {
			return nil // 避免重复进入
		}
		newTitle := &model.MenuItem{Title: "相似歌曲", Subtitle: "与「" + song.Name + "」相似的歌曲"}
		main.EnterMenu(NewSimilarSongsMenu(newBaseMenu(n), song), newTitle)
		return nil
	})
	op.ShowLoading().Execute()
}

// goToAlbumOfSong 查看歌曲所属专辑
func goToAlbumOfSong(n *Netease, isSelected bool) {
	op := NewOperation(n, func(n *Netease) model.Page {
		song, ok := getTargetSong(n, isSelected)
		if !ok {
			return nil
		}
		main := n.MustMain()
		menu := main.CurMenu()
		if detail, ok := menu.(*AlbumDetailMenu); ok && detail.albumId == song.Album.Id {
			return nil // 避免重复进入
		}
		newTitle := &model.MenuItem{Title: song.Album.Name, Subtitle: "「" + song.Name + "」所属专辑"}
		main.EnterMenu(NewAlbumDetailMenu(newBaseMenu(n), song.Album.Id), newTitle)
		return nil
	})
	op.ShowLoading().Execute()
}

// goToArtistOfSong 查看歌曲所属歌手
func goToArtistOfSong(n *Netease, isSelected bool) {
	op := NewOperation(n, func(n *Netease) model.Page {
		song, ok := getTargetSong(n, isSelected)
		if !ok {
			return nil
		}
		main := n.MustMain()
		menu := main.CurMenu()
		artistCount := len(song.Artists)
		if artistCount <= 0 {
			return nil
		}
		if artistCount == 1 {
			if detail, ok := menu.(*ArtistDetailMenu); ok && detail.artistId == song.Artists[0].Id {
				return nil
			}
			newMenu := NewArtistDetailMenu(newBaseMenu(n), song.Artists[0].Id, song.Artists[0].Name)
			newTitle := &model.MenuItem{Title: song.Artists[0].Name, Subtitle: "「" + song.Name + "」所属歌手"}
			main.EnterMenu(newMenu, newTitle)
			return nil
		}
		if artists, ok := menu.(*ArtistsOfSongMenu); ok && artists.song.Id == song.Id {
			return nil
		}
		newTitle := &model.MenuItem{Title: "「" + song.Name + "」所属歌手"}
		main.EnterMenu(NewArtistsOfSongMenu(newBaseMenu(n), song), newTitle)
		return nil
	})
	op.ShowLoading().Execute()
}

// openInWeb 在浏览器中打开
func openInWeb(n *Netease, isSelected bool) {
	op := NewOperation(n, func(n *Netease) model.Page {
		if !isSelected {
			if song, ok := getTargetSong(n, false); ok {
				_ = open.Start(netease.WebUrlOfSong(song.Id))
			}
			return nil
		}

		main := n.MustMain()
		menu := main.CurMenu()
		selectedIndex := menu.RealDataIndex(main.SelectedIndex())

		// FIXME: 暂时这样简化，相应函数应进一步通用化命名
		if sharer, ok := menu.(composer.Sharer); ok {
			item := sharer.ItemToShare(selectedIndex)
			var url string
			switch v := item.(type) {
			case structs.Song:
				url = netease.WebUrlOfSong(v.Id)
			case structs.Playlist:
				url = netease.WebUrlOfPlaylist(v.Id)
			case structs.Album:
				url = netease.WebUrlOfAlbum(v.Id)
			case structs.Artist:
				url = netease.WebUrlOfArtist(v.Id)
			}
			if url != "" {
				slog.Debug("URL has been generated and will open soon", "url", url)
				if err := open.Start(url); err != nil {
					slog.Error("Failed to open URL", "url", url, "err", err)
				}

			}
		}
		return nil
	})
	op.ShowLoading().Execute()
}

// collectSelectedPlaylist 收藏或取消收藏选中歌单
func collectSelectedPlaylist(n *Netease, isCollect bool) model.Page {
	coreLogic := func(n *Netease) model.Page {
		playlist, ok := getSelectedPlaylist(n)
		if !ok {
			return nil
		}
		t := "1"
		if !isCollect {
			t = "0" // API中 '1' 为收藏, '0' 为取消收藏
		}
		s := service.PlaylistSubscribeService{ID: strconv.FormatInt(playlist.Id, 10), T: t}
		if code, resp := s.PlaylistSubscribe(); code != 200 {
			var msg string
			if msg, _ = jsonparser.GetString(resp, "message"); msg == "" {
				msg, _ = jsonparser.GetString(resp, "data", "message")
			}
			if msg == "" {
				msg = "操作失败"
			}
			notify.Notify(notify.NotifyContent{
				Title:   msg,
				Text:    playlist.Name,
				Url:     types.AppGithubUrl,
				GroupId: types.GroupID,
			})
			return nil
		}
		title := "已收藏歌单"
		if !isCollect {
			title = "已取消收藏歌单"
		}
		notify.Notify(notify.NotifyContent{
			Title:   title,
			Text:    playlist.Name,
			Url:     types.AppGithubUrl,
			GroupId: types.GroupID,
			Icon:    "",
		})
		return nil
	}
	return NewOperation(n, coreLogic).ShowLoading().NeedsAuth().Execute()
}

// subscribeAlbum 收藏或取消收藏歌曲的专辑
// isSub: true 为收藏, false 为取消收藏。
// isSelected: true 操作选中的歌曲, false 操作正在播放的歌曲。
func subscribeAlbum(n *Netease, isSub bool, isSelected bool) model.Page {
	coreLogic := func(n *Netease) model.Page {
		song, ok := getTargetSong(n, isSelected)
		if !ok {
			return nil
		}
		if song.Album.Id == 0 {
			notify.Notify(notify.NotifyContent{
				Title:   "操作失败",
				Text:    "歌曲没有专辑信息",
				Url:     types.AppGithubUrl,
				GroupId: types.GroupID,
			})
			return nil
		}
		t := "1"
		if !isSub {
			t = "2" // API 中 '1' 为收藏, '2'（或任何非 1）为取消收藏
		}
		s := service.AlbumSubService{ID: strconv.FormatInt(song.Album.Id, 10), T: t}
		if code, resp := s.AlbumSub(); code != 200 {
			var msg string
			if msg, _ = jsonparser.GetString(resp, "message"); msg == "" {
				msg, _ = jsonparser.GetString(resp, "data", "message")
			}
			if msg == "" {
				msg = "操作失败"
			}
			notify.Notify(notify.NotifyContent{
				Title:   msg,
				Text:    song.Album.Name,
				Url:     types.AppGithubUrl,
				GroupId: types.GroupID,
			})
			return nil
		}
		title := "已收藏专辑"
		if !isSub {
			title = "已取消收藏专辑"
		}
		notify.Notify(notify.NotifyContent{
			Title:   title,
			Text:    song.Album.Name,
			Url:     types.AppGithubUrl,
			GroupId: types.GroupID,
			Icon:    "",
		})
		return nil
	}
	return NewOperation(n, coreLogic).ShowLoading().NeedsAuth().Execute()
}

// subscribeArtist 收藏或取消收藏歌曲的歌手
// isSub: true 为收藏, false 为取消收藏。
// isSelected: true 操作选中的歌曲, false 操作正在播放的歌曲。
func subscribeArtist(n *Netease, isSub bool, isSelected bool) model.Page {
	coreLogic := func(n *Netease) model.Page {
		song, ok := getTargetSong(n, isSelected)
		if !ok {
			return nil
		}
		if len(song.Artists) == 0 {
			notify.Notify(notify.NotifyContent{
				Title:   "操作失败",
				Text:    "歌曲没有歌手信息",
				Url:     types.AppGithubUrl,
				GroupId: types.GroupID,
			})
			return nil
		}

		artist := song.Artists[0]

		t := "1"
		if !isSub {
			t = "2" // API 中 '1' 为收藏, '2'（或任何非 1）为取消收藏
		}
		s := service.ArtistSubService{T: t, Id: strconv.FormatInt(artist.Id, 10)}
		if code, resp := s.ArtistSub(); code != 200 {
			var msg string
			if msg, _ = jsonparser.GetString(resp, "message"); msg == "" {
				msg, _ = jsonparser.GetString(resp, "data", "message")
			}
			if msg == "" {
				msg = "操作失败"
			}
			notify.Notify(notify.NotifyContent{
				Title:   msg,
				Text:    artist.Name,
				Url:     types.AppGithubUrl,
				GroupId: types.GroupID,
			})
			return nil
		}
		title := "已收藏歌手"
		if !isSub {
			title = "已取消收藏歌手"
		}
		notify.Notify(notify.NotifyContent{
			Title:   title,
			Text:    artist.Name,
			Url:     types.AppGithubUrl,
			GroupId: types.GroupID,
			Icon:    "",
		})
		return nil
	}
	return NewOperation(n, coreLogic).ShowLoading().NeedsAuth().Execute()
}

// appendSongsToCurPlaylist 添加歌曲到播放列表
func appendSongsToCurPlaylist(n *Netease, addToNext bool) {
	op := NewOperation(n, func(n *Netease) model.Page {
		main := n.MustMain()
		menu := main.CurMenu()
		selectedIndex := menu.RealDataIndex(main.SelectedIndex())
		subMenu := menu.SubMenu(n.App, selectedIndex)
		var appendSongs []structs.Song
		var notifyURL string

		sm, isSongsMenu := menu.(SongsMenu)
		subSm, subIsSongsMenu := subMenu.(SongsMenu)
		switch {
		case isSongsMenu: // 当前菜单是 SongMenu
			if selectedIndex >= len(sm.Songs()) {
				return nil
			}
			song := sm.Songs()[selectedIndex]
			appendSongs = append(appendSongs, song)
			notifyURL = netease.WebUrlOfSong(song.Id)
		case subIsSongsMenu: // 选中项菜单是 SongsMenu
			// 触发 BeforeEnterMenuHook 获取歌曲
			if ok, _ := subSm.BeforeEnterMenuHook()(n.Main()); !ok {
				return nil
			}
			appendSongs = subSm.Songs()
		default:
			return nil
		}
		if len(appendSongs) == 0 {
			return nil
		}

		var notifyTitle string
		if addToNext && len(n.player.Playlist()) > 0 {
			// 添加为下一曲
			targetIndex := n.player.CurSongIndex() + 1
			_ = n.player.playlistManager.Initialize(
				n.player.CurSongIndex(),
				slices.Concat(
					n.player.Playlist()[:targetIndex],
					appendSongs,
					n.player.Playlist()[targetIndex:],
				),
			)
			notifyTitle = "已添加到下一曲"
		} else {
			// 添加到播放列表末尾
			_ = n.player.playlistManager.Initialize(
				n.player.CurSongIndex(),
				append(n.player.Playlist(), appendSongs...),
			)
			notifyTitle = "已添加到播放列表末尾"
		}

		// 替换播放中数据，避免数据错乱
		n.player.playingMenu = nil
		n.player.playingMenuKey += "modified"
		if curPlaylist, ok := menu.(*CurPlaylist); ok {
			curPlaylist.songs = n.player.Playlist()
			curPlaylist.menus = menux.GetViewFromSongs(n.player.Playlist())
			main.RefreshMenuList()
		}
		notify.Notify(notify.NotifyContent{
			Title:   notifyTitle,
			Text:    menu.MenuViews()[main.SelectedIndex()].Title,
			Url:     notifyURL,
			GroupId: types.GroupID,
		})
		return nil
	})
	op.ShowLoading().Execute()
}

// openAddSongToUserPlaylistMenu 打开“添加歌曲到歌单”菜单
func openAddSongToUserPlaylistMenu(n *Netease, isSelected, isAdd bool) model.Page {
	coreLogic := func(n *Netease) model.Page {
		song, ok := getTargetSong(n, isSelected)
		if !ok {
			return nil
		}

		main := n.MustMain()
		if _, ok := main.CurMenu().(*AddToUserPlaylistMenu); ok {
			return nil // 避免重复进入
		}

		subtitle := "将「" + song.Name + "」加入歌单"
		if !isAdd {
			subtitle = "将「" + song.Name + "」从歌单中删除"
		}
		newMenu := NewAddToUserPlaylistMenu(newBaseMenu(n), n.user.UserId, song, isAdd)
		newTitle := &model.MenuItem{Title: "我的歌单", Subtitle: subtitle}
		main.EnterMenu(newMenu, newTitle)
		return nil
	}

	return NewOperation(n, coreLogic).ShowLoading().NeedsAuth().Execute()
}

// addSongToUserPlaylist 添加或从歌单删除歌曲（仅在 AddToUserPlaylistMenu 中调用）
func addSongToUserPlaylist(n *Netease, isAdd bool) model.Page {
	coreLogic := func(n *Netease) model.Page {
		main := n.MustMain()
		menu, ok := main.CurMenu().(*AddToUserPlaylistMenu)
		if !ok {
			return nil
		}
		playlist := menu.playlists[menu.RealDataIndex(main.SelectedIndex())]
		op := "add"
		if !isAdd {
			op = "del"
		}
		likeService := service.PlaylistTracksService{
			TrackIds: []string{strconv.FormatInt(menu.song.Id, 10)},
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
				Text:    menu.song.Name,
				Url:     types.AppGithubUrl,
				GroupId: types.GroupID,
			})
			main.BackMenu()
			return nil
		}
		title := "已添加到歌单「" + playlist.Name + "」"
		if !isAdd {
			title = "已从歌单「" + playlist.Name + "」中删除"
		}
		notify.Notify(notify.NotifyContent{
			Title:   title,
			Text:    menu.song.Name,
			Url:     netease.WebUrlOfPlaylist(playlist.Id),
			GroupId: types.GroupID,
		})
		main.BackMenu()

		// 刷新原歌单详情页
		if !isAdd {
			originalMenu, ok := main.CurMenu().(*PlaylistDetailMenu)
			if ok && originalMenu.playlistId == playlist.Id {
				t := main.MenuTitle()
				main.BackMenu()
				_, page := originalMenu.BeforeEnterMenuHook()(main)
				main.EnterMenu(originalMenu, t)
				return page
			}
		}
		return nil
	}
	return NewOperation(n, coreLogic).ShowLoading().NeedsAuth().Execute()
}

// delSongFromPlaylist 从播放列表删除选中歌曲,仅在当前播放列表界面有效
func delSongFromPlaylist(n *Netease) model.Page {
	op := NewOperation(n, func(n *Netease) model.Page {
		main := n.MustMain()
		menu, ok := main.CurMenu().(*CurPlaylist)
		if !ok {
			return nil
		}
		selectedIndex := menu.RealDataIndex(main.SelectedIndex())
		if selectedIndex >= len(menu.Songs()) || len(n.player.Playlist()) == 0 {
			return nil
		}

		if removedSong, err := n.player.playlistManager.RemoveSong(selectedIndex); err == nil {
			n.player.PlaySong(removedSong, DurationNext)
		}
		songs := n.player.Playlist()
		menu.menus = menux.GetViewFromSongs(songs)
		menu.songs = songs

		// 更新游标位置
		if main.SelectedIndex() >= len(songs) {
			main.SetSelectedIndex(len(songs) - 1)
		}

		// 替换播放中数据，避免数据错乱
		n.player.playingMenu = nil
		n.player.playingMenuKey += "modified"

		// 如果播放列表中已经没有歌曲，停止播放
		if len(n.player.Playlist()) == 0 {
			n.player.Stop()
		}

		main.RefreshMenuList()
		return nil
	})
	return op.ShowLoading().Execute()
}

// clearSongCache 清除歌曲缓存
func clearSongCache(n *Netease) {
	action := func() {
		op := NewOperation(n, func(n *Netease) model.Page {
			err := n.trackManager.ClearCache()
			if err != nil {
				slog.Error("清除缓存失败", "error", err)
				notify.Notify(notify.NotifyContent{
					Title:   "清除缓存失败",
					Text:    err.Error(),
					GroupId: types.GroupID,
				})
			} else {
				slog.Info("清除缓存成功")
				notify.Notify(notify.NotifyContent{
					Title:   "清除缓存成功",
					Text:    "缓存已清除",
					GroupId: types.GroupID,
				})
			}
			return nil
		})
		op.ShowLoading().Execute()
	}
	menu := NewConfirmMenu(newBaseMenu(n), []ConfirmItem{
		{title: model.MenuItem{Title: "确定"}, action: action, backLevel: 1},
	})
	n.MustMain().EnterMenu(menu, &model.MenuItem{Title: "清除缓存", Subtitle: "确定清除缓存"})
}

// action 打开操作菜单
func action(n *Netease, curPlaying bool) {
	main := n.MustMain()
	menu := main.CurMenu()

	menuKey := menu.GetMenuKey()
	if menuKey == actionMenuKey {
		slog.Debug("已位于操作菜单，取消")
		return
	}

	var newTitle *model.MenuItem

	// 显式设置 titel 避免被修改后影响原 item
	if curPlaying {
		song, ok := getTargetSong(n, false)
		if ok {
			newTitle = &model.MenuItem{Title: "操作当前播放", Subtitle: song.Name}
		} else {
			newTitle = &model.MenuItem{Title: "操作当前播放", Subtitle: "当前无播放"}
		}
	} else {
		menuItems := menu.MenuViews()
		selectedIndex := menu.RealDataIndex(main.SelectedIndex())
		if selectedIndex >= 0 && selectedIndex < len(menuItems) {
			item := menuItems[selectedIndex]
			newTitle = &model.MenuItem{Title: "操作：" + item.Title, Subtitle: item.Subtitle}
		} else {
			newTitle = &model.MenuItem{Title: "操作", Subtitle: "未知操作对象"}
		}
	}

	newMenu := NewActionMenu(newBaseMenu(n), menuKey, curPlaying)
	main.EnterMenu(newMenu, newTitle)
}

// shareItem 分享项目到剪贴板
func shareItem(n *Netease, isSelected bool) {
	var itemToShare any
	if !isSelected {
		if song, ok := getTargetSong(n, false); ok {
			itemToShare = song
		}
	} else {
		main := n.MustMain()
		menu := main.CurMenu()
		selectedIndex := menu.RealDataIndex(main.SelectedIndex())
		if sharer, ok := menu.(composer.Sharer); ok {
			itemToShare = sharer.ItemToShare(selectedIndex)
		} else { // 兼容旧菜单类型
			switch me := menu.(type) {
			case SongsMenu:
				if selectedIndex < len(me.Songs()) {
					itemToShare = me.Songs()[selectedIndex]
				}
			case AlbumsMenu:
				if selectedIndex < len(me.Albums()) {
					itemToShare = me.Albums()[selectedIndex]
				}
			case ArtistsMenu:
				if selectedIndex < len(me.Artists()) {
					itemToShare = me.Artists()[selectedIndex]
				}
			case PlaylistsMenu:
				if selectedIndex < len(me.Playlists()) {
					itemToShare = me.Playlists()[selectedIndex]
				}
			default:
				slog.Error(fmt.Sprintf("不支持的分享页面: %T", me))
			}
		}
	}

	if itemToShare == nil {
		return
	}

	str, err := n.shareSvc.Share(itemToShare)
	if err != nil {
		slog.Error("分享失败", "error", err)
		notify.Notify(notify.NotifyContent{
			Title:   "分享失败",
			Text:    err.Error(),
			Url:     types.AppGithubUrl,
			GroupId: types.GroupID,
			Icon:    "",
		})
		return
	}
	if err = clipboard.Write(str); err != nil {
		slog.Error("写入剪贴板失败", "error", err)
		notify.Notify(notify.NotifyContent{
			Title:   "分享失败",
			Text:    err.Error(),
			Url:     types.AppGithubUrl,
			GroupId: types.GroupID,
			Icon:    "",
		})
	}
}

// searchSong 搜索歌名
func searchSong(n *Netease, isSelected bool) {
	op := NewOperation(n, func(n *Netease) model.Page {
		song, ok := getTargetSong(n, isSelected)
		if !ok {
			return nil
		}

		main := n.MustMain()
		searchService := service.SearchService{
			S:     song.Name,
			Type:  strconv.Itoa(int(StSingleSong)),
			Limit: strconv.Itoa(types.SearchPageSize),
		}
		code, response := searchService.Search()
		codeType := _struct.CheckCode(code)
		if codeType != _struct.Success {
			return nil
		}

		n.search.wordsInput.SetValue(song.Name)
		n.search.searchType = StSingleSong
		n.search.result = _struct.GetSongsOfSearchResult(response)

		return main.EnterMenu(NewSearchResultMenu(newBaseMenu(n), StSingleSong), &model.MenuItem{Title: "搜索结果"})
	})
	op.ShowLoading().Execute()
}
