// Package artist implements the artist cluster (热门歌手 / 歌手详情) as the
// fifth real plugin. It is the third bulk cluster extraction after the DJ/radio
// and album clusters: all six artist menus moved from internal/ui verbatim with
// their provider keys unchanged — hot_artists (the 热门歌手 entry menu, which
// now declares the plugin main-menu item), artist_detail (the shared detail
// menu), the two detail sub-menus (artist_song / artist_album), the
// artists-of-song menu (artist_of_song) and the subscribed-artists list
// (artists_sub_list). The cluster navigates internally via ui.BuildMenu /
// ui.BuildMenuOrToast with the same keys as before, and the ui side
// (search_result / operate goToArtistOfSong / user_collection) keeps jumping
// into "artist_detail" / "artist_of_song" / "artists_sub_list" unchanged.
package artist

import (
	"github.com/anhoder/foxful-cli/model"

	ui "github.com/go-musicfox/go-musicfox/internal/ui"
)

// enterMenuCallback mirrors ui.EnterMenuCallback (unexported there): the login
// callback re-enters the requesting menu once login succeeds.
func enterMenuCallback(main *model.Main) ui.LoginCallback {
	return func() model.Page {
		return main.EnterMenu(nil, nil)
	}
}

// ArtistAlbumOpts is the parameter contract of the "artist_album" menu provider
// (hot albums of an artist). Moved from ui together with the cluster — its only
// consumers are inside this package (artist_detail's SubMenu).
type ArtistAlbumOpts struct {
	ArtistID int64
}

// ArtistSongOpts is the parameter contract of the "artist_song" menu provider
// (hot songs of an artist). Moved from ui together with the cluster — its only
// consumers are inside this package (artist_detail's SubMenu).
type ArtistSongOpts struct {
	ArtistID int64
}

// init is the compile-time registration entry (linked via the internal/plugins
// aggregator blank import, which cmd/musicfox.go pulls in). Every key is
// identical to the one the menu registered under in internal/ui before the
// extraction: hot_artists / artist_detail / artist_song / artist_album /
// artist_of_song / artists_sub_list. Note "artist_detail" and "artist_of_song"
// keep their shared opts types in ui (ui.ArtistDetailOpts — ui's search-result
// menu, artists-of-song and subscribed-artists lists and operate.go
// goToArtistOfSong also jump into it; ui.ArtistsOfSongOpts — operate.go's
// multi-artist jump carries it). The hot_artists entry menu declares the
// main-menu item 热门歌手: the built-in entry was removed from menu_main.go
// (plugin items are appended after all built-ins).
func init() {
	ui.RegisterMenu("hot_artists", func(base ui.BaseMenu, _ ui.NoArgMenuOpts) (ui.Menu, error) {
		return NewHotArtistsMenu(base), nil
	})
	ui.RegisterMenu("artist_detail", func(base ui.BaseMenu, opts ui.ArtistDetailOpts) (ui.Menu, error) {
		return NewArtistDetailMenu(base, opts.ArtistID, opts.Name), nil
	})
	ui.RegisterMenu("artist_song", func(base ui.BaseMenu, opts ArtistSongOpts) (ui.Menu, error) {
		return NewArtistSongMenu(base, opts.ArtistID), nil
	})
	ui.RegisterMenu("artist_album", func(base ui.BaseMenu, opts ArtistAlbumOpts) (ui.Menu, error) {
		return NewArtistAlbumMenu(base, opts.ArtistID), nil
	})
	ui.RegisterMenu("artist_of_song", func(base ui.BaseMenu, opts ui.ArtistsOfSongOpts) (ui.Menu, error) {
		return NewArtistsOfSongMenu(base, opts.Song), nil
	})
	ui.RegisterMenu("artists_sub_list", func(base ui.BaseMenu, _ ui.NoArgMenuOpts) (ui.Menu, error) {
		return NewArtistsSubscribeListMenu(base), nil
	})
	// 声明主菜单入口：NewMainMenu 在全部内置项之后追加「热门歌手」（原先为
	// 内置索引 8，现为插件主菜单项，排在全部内置项之后）。
	ui.RegisterMainMenuItem("hot_artists", "热门歌手")
}
