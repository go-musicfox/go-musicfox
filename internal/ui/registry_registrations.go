package ui

import "github.com/anhoder/foxful-cli/model"

// Production provider registrations (Phase 3.2.1/3.2.2). The 5 sample menus
// move here from the Phase 3.0 prototype (registry_proto_b.go, kept as
// evidence); demo migrations add album_detail / user_playlist / dj_radio_detail
// and the login/search pages. Phase 3.3.1 extends the set to every menu_*.go
// menu (batches 1-4). The Last.fm menu and pages moved into the
// internal/plugins/lastfm plugin with their registrations (Phase 3.9), the
// whole DJ/radio cluster (dj_* + radio_dj_type) moved into
// internal/plugins/dj (Phase 3.9.x), the whole album cluster
// (album_menu / album_* + album_detail) moved into internal/plugins/album
// (Phase 3.9.x), the whole artist cluster (artist_detail / artist_* +
// hot_artists) moved into internal/plugins/artist (Phase 3.9.x), the
// recommend cluster (daily_songs / daily_playlists / personal_fm /
// recent_songs / ranks) moved into internal/plugins/recommend (Phase 3.9.x),
// the playlist & cloud cluster (user_playlist / user_collect /
// high_quality_playlists / could) and the playlist_detail menu moved into
// internal/plugins/playlist (Phase 3.9.x), the search cluster (search_type /
// search_result / the search page) moved into internal/plugins/search
// (Phase 3.9.x), and the song cluster (simi_songs / add_to_user_playlist)
// moved into internal/plugins/song (Phase 3.9.x).
func init() {
	// playlist_detail was one of the Phase 3.0 prototype sample menus; its
	// provider registration moved into the internal/plugins/playlist plugin
	// (Phase 3.9.x). Its key and ui.PlaylistDetailOpts stay — the operate /
	// high-quality-playlists / cloud-disk call sites jump into it unchanged.

	// artist_detail was one of the Phase 3.0 prototype sample menus; its
	// provider registration moved into the internal/plugins/artist plugin with
	// the artist cluster (Phase 3.9.x). Its key and ui.ArtistDetailOpts stay —
	// the search-result / artists-of-song / subscribed-artists / operate call
	// sites jump into it unchanged.

	// search_result moved into the internal/plugins/search plugin with the
	// search cluster (Phase 3.9.x). Its key and ui.SearchResultOpts stay — the
	// search-type / operate call sites jump into it unchanged.

	// add_to_user_playlist moved into the internal/plugins/song plugin with
	// the song cluster (Phase 3.9.x). Its key and ui.AddToUserPlaylistOpts
	// stay — the operate call site jumps into it unchanged.

	// ranks (排行榜) was a no-arg menu registered here; its provider moved into
	// the internal/plugins/recommend plugin with the recommend cluster
	// (Phase 3.9.x) and now also declares the 排行榜 main-menu item.

	// Phase 3.2.2 demo migrations — non-prototype menus proven on the
	// production registry (call sites in menu_search_result.go). album_detail
	// moved into the internal/plugins/album plugin with the album cluster
	// (Phase 3.9.x); its key and ui.AlbumDetailOpts stay, and the search-result
	// / artist-album / operate call sites jump into it unchanged. user_playlist
	// / user_collect / high_quality_playlists / could moved into the
	// internal/plugins/playlist plugin with the playlist & cloud cluster
	// (Phase 3.9.x); ui.UserPlaylistOpts stays (the search-result jump site
	// uses it), and the menu_search_result call site jumps into user_playlist
	// unchanged.

	// simi_songs moved into the internal/plugins/song plugin with the song
	// cluster (Phase 3.9.x). Its key and ui.SimiSongsOpts stay — the
	// search-result / operate call sites jump into it unchanged.

	// --- Phase 3.3.1 batch 4: main menu cluster + misc ---

	// personal_fm / recent_songs / daily_playlists / daily_songs were no-arg
	// menus registered here; their providers moved into the
	// internal/plugins/recommend plugin with the recommend cluster
	// (Phase 3.9.x), which also declares their main-menu items.

	// search_type moved into the internal/plugins/search plugin with the
	// search cluster (Phase 3.9.x), which also declares its 搜索 main-menu item
	// (RegisterMainMenuItemAfter("search_type", "搜索", "album_menu", nil)).

	// Bootstrap menus (main_menu / local_search) are registered for startup
	// assertion completeness; the app-bootstrap call sites (internal/commands)
	// construct them directly from a BaseMenu.
	RegisterMenu("local_search", func(base baseMenu, _ NoArgMenuOpts) (Menu, error) {
		return NewLocalSearchMenu(base), nil
	})

	RegisterMenu("main_menu", func(base baseMenu, _ NoArgMenuOpts) (Menu, error) {
		return NewMainMenu(base), nil
	})

	// --- Phase 3.3.3: last hardcoded menu constructions ---

	RegisterMenu("cur_playlist", func(base baseMenu, opts CurPlaylistOpts) (Menu, error) {
		return NewCurPlaylist(base, opts.Songs), nil
	})

	RegisterMenu("action_menu", func(base baseMenu, opts ActionMenuOpts) (Menu, error) {
		return NewActionMenu(base, opts.From, opts.CurPlaying), nil
	})

	RegisterPage("login", func(opts LoginPageOpts) (model.Page, error) {
		return NewLoginPage(opts.Netease), nil
	})

	// --- Phase 3.3.2 page migrations ---

	// The search page moved into the internal/plugins/search plugin with the
	// search cluster (Phase 3.9.x). The shell keeps the built instance as a
	// singleton because its wordsInput/result/searchType state is read back by
	// SearchResultMenu and operate.go through the shell (3.3.3 addresses those
	// readers).
}
