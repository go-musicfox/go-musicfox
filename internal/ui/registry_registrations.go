package ui

import "github.com/anhoder/foxful-cli/model"

// Production provider registrations (Phase 3.2.1/3.2.2). The 5 sample menus
// move here from the Phase 3.0 prototype (registry_proto_b.go, kept as
// evidence); demo migrations add album_detail / user_playlist / dj_radio_detail
// and the login/search pages. Phase 3.3.1 extends the set to every menu_*.go
// menu (batches 1-4). The Last.fm menu and pages moved into the
// internal/plugins/lastfm plugin with their registrations (Phase 3.9), the
// whole DJ/radio cluster (dj_* + radio_dj_type) moved into
// internal/plugins/dj (Phase 3.9.x), and the whole album cluster
// (album_menu / album_* + album_detail) moved into internal/plugins/album
// (Phase 3.9.x).
func init() {
	RegisterMenu("playlist_detail", func(base baseMenu, opts PlaylistDetailOpts) (Menu, error) {
		return NewPlaylistDetailMenu(base, opts.PlaylistID), nil
	})

	RegisterMenu("artist_detail", func(base baseMenu, opts ArtistDetailOpts) (Menu, error) {
		return NewArtistDetailMenu(base, opts.ArtistID, opts.Name), nil
	})

	RegisterMenu("search_result", func(base baseMenu, opts SearchResultOpts) (Menu, error) {
		return NewSearchResultMenu(base, opts.SearchType), nil
	})

	RegisterMenu("add_to_user_playlist", func(base baseMenu, opts AddToUserPlaylistOpts) (Menu, error) {
		return NewAddToUserPlaylistMenu(base, opts.UserID, opts.Song, opts.IsAdd), nil
	})

	// No-arg menu: the shared NoArgMenuOpts placeholder keeps the generic
	// signature; call sites use the mustBuildNoArg / buildMenuOrToast helpers.
	RegisterMenu("ranks", func(base baseMenu, _ NoArgMenuOpts) (Menu, error) {
		return NewRanksMenu(base), nil
	})

	// Phase 3.2.2 demo migrations — non-prototype menus proven on the
	// production registry (call sites in menu_search_result.go). album_detail
	// moved into the internal/plugins/album plugin with the album cluster
	// (Phase 3.9.x); its key and ui.AlbumDetailOpts stay, and the search-result
	// / artist-album / operate call sites jump into it unchanged.
	RegisterMenu("user_playlist", func(base baseMenu, opts UserPlaylistOpts) (Menu, error) {
		return NewUserPlaylistMenu(base, opts.UserID), nil
	})

	RegisterMenu("high_quality_playlists", func(base baseMenu, _ NoArgMenuOpts) (Menu, error) {
		return NewHighQualityPlaylistsMenu(base), nil
	})

	// --- Phase 3.3.1 batch 3: artist cluster ---

	RegisterMenu("artist_of_song", func(base baseMenu, opts ArtistsOfSongOpts) (Menu, error) {
		return NewArtistsOfSongMenu(base, opts.Song), nil
	})

	RegisterMenu("artist_album", func(base baseMenu, opts ArtistAlbumOpts) (Menu, error) {
		return NewArtistAlbumMenu(base, opts.ArtistID), nil
	})

	RegisterMenu("artist_song", func(base baseMenu, opts ArtistSongOpts) (Menu, error) {
		return NewArtistSongMenu(base, opts.ArtistID), nil
	})

	RegisterMenu("artists_sub_list", func(base baseMenu, _ NoArgMenuOpts) (Menu, error) {
		return NewArtistsSubscribeListMenu(base), nil
	})

	RegisterMenu("hot_artists", func(base baseMenu, _ NoArgMenuOpts) (Menu, error) {
		return NewHotArtistsMenu(base), nil
	})

	RegisterMenu("simi_songs", func(base baseMenu, opts SimiSongsOpts) (Menu, error) {
		return NewSimilarSongsMenu(base, opts.Song), nil
	})

	RegisterMenu("user_collect", func(base baseMenu, _ NoArgMenuOpts) (Menu, error) {
		return NewUserCollectionMenu(base), nil
	})

	// --- Phase 3.3.1 batch 4: main menu cluster + misc ---

	RegisterMenu("personal_fm", func(base baseMenu, _ NoArgMenuOpts) (Menu, error) {
		return NewPersonalFmMenu(base), nil
	})

	RegisterMenu("could", func(base baseMenu, _ NoArgMenuOpts) (Menu, error) {
		return NewCloudMenu(base), nil
	})

	RegisterMenu("recent_songs", func(base baseMenu, _ NoArgMenuOpts) (Menu, error) {
		return NewRecentSongsMenu(base), nil
	})

	RegisterMenu("daily_playlists", func(base baseMenu, _ NoArgMenuOpts) (Menu, error) {
		return NewDailyRecommendPlaylistMenu(base), nil
	})

	RegisterMenu("daily_songs", func(base baseMenu, _ NoArgMenuOpts) (Menu, error) {
		return NewDailyRecommendSongsMenu(base), nil
	})

	RegisterMenu("search_type", func(base baseMenu, _ NoArgMenuOpts) (Menu, error) {
		return NewSearchTypeMenu(base), nil
	})

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

	// Search page: the shell keeps the built instance as a singleton because
	// its wordsInput/result/searchType state is read back by SearchResultMenu
	// and operate.go through the shell (3.3.3 addresses those readers).
	RegisterPage("search", func(opts SearchPageOpts) (model.Page, error) {
		return NewSearchPage(opts.Netease), nil
	})
}
