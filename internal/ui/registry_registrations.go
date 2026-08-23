package ui

import "github.com/anhoder/foxful-cli/model"

// Production provider registrations (Phase 3.2.1/3.2.2). The 5 sample menus
// move here from the Phase 3.0 prototype (registry_proto_b.go, kept as
// evidence); demo migrations add album_detail / user_playlist / dj_radio_detail
// and the lastfm_auth page. Phase 3.3.1 extends the set to every menu_*.go
// menu (batches 1-4).
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
	// production registry (call sites in menu_search_result.go).
	RegisterMenu("album_detail", func(base baseMenu, opts AlbumDetailOpts) (Menu, error) {
		return NewAlbumDetailMenu(base, opts.AlbumID), nil
	})

	RegisterMenu("user_playlist", func(base baseMenu, opts UserPlaylistOpts) (Menu, error) {
		return NewUserPlaylistMenu(base, opts.UserID), nil
	})

	RegisterMenu("dj_radio_detail", func(base baseMenu, opts DjRadioDetailOpts) (Menu, error) {
		return NewDjRadioDetailMenu(base, opts.DjRadioID), nil
	})

	// --- Phase 3.3.1 batch 1: DJ / radio cluster ---

	RegisterMenu("dj_category_detail", func(base baseMenu, opts DjCategoryDetailOpts) (Menu, error) {
		return NewDjCategoryDetailMenu(base, opts.CategoryID), nil
	})

	RegisterMenu("dj_category", func(base baseMenu, _ NoArgMenuOpts) (Menu, error) {
		return NewDjCategoryMenu(base), nil
	})

	RegisterMenu("dj_program_rank", func(base baseMenu, _ NoArgMenuOpts) (Menu, error) {
		return NewDjProgramRankMenu(base), nil
	})

	RegisterMenu("dj_program_hour_rank", func(base baseMenu, _ NoArgMenuOpts) (Menu, error) {
		return NewDjProgramHoursRankMenu(base), nil
	})

	RegisterMenu("dj_hot", func(base baseMenu, opts DjHotOpts) (Menu, error) {
		return NewDjHotMenu(base, opts.HotType), nil
	})

	RegisterMenu("dj_sub", func(base baseMenu, _ NoArgMenuOpts) (Menu, error) {
		return NewDjSubListMenu(base), nil
	})

	RegisterMenu("dj_recommend", func(base baseMenu, _ NoArgMenuOpts) (Menu, error) {
		return NewDjRecommendMenu(base), nil
	})

	RegisterMenu("dj_today_recommend", func(base baseMenu, _ NoArgMenuOpts) (Menu, error) {
		return NewDjTodayRecommendMenu(base), nil
	})

	RegisterMenu("radio_dj_type", func(base baseMenu, _ NoArgMenuOpts) (Menu, error) {
		return NewRadioDjTypeMenu(base), nil
	})

	// --- Phase 3.3.1 batch 2: album cluster ---

	RegisterMenu("album_new_area", func(base baseMenu, _ NoArgMenuOpts) (Menu, error) {
		return NewAlbumNewAreaMenu(base), nil
	})

	RegisterMenu("album_top_area", func(base baseMenu, _ NoArgMenuOpts) (Menu, error) {
		return NewAlbumTopAreaMenu(base), nil
	})

	RegisterMenu("album_new_hot", func(base baseMenu, _ NoArgMenuOpts) (Menu, error) {
		return NewAlbumNewestMenu(base), nil
	})

	RegisterMenu("album_menu", func(base baseMenu, _ NoArgMenuOpts) (Menu, error) {
		return NewAlbumListMenu(base), nil
	})

	RegisterMenu("album_top", func(base baseMenu, opts AlbumTopOpts) (Menu, error) {
		return NewAlbumTopMenu(base, opts.Area), nil
	})

	RegisterMenu("album_new", func(base baseMenu, opts AlbumNewOpts) (Menu, error) {
		return NewAlbumNewMenu(base, opts.Area), nil
	})

	RegisterMenu("album_sub_list", func(base baseMenu, _ NoArgMenuOpts) (Menu, error) {
		return NewAlbumSubscribeListMenu(base), nil
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

	RegisterMenu("last_fm", func(base baseMenu, _ NoArgMenuOpts) (Menu, error) {
		return NewLastfm(base), nil
	})

	RegisterPage("login", func(opts LoginPageOpts) (model.Page, error) {
		return NewLoginPage(opts.Netease), nil
	})

	RegisterPage("lastfm_auth", func(opts LastfmAuthPageOpts) (model.Page, error) {
		return NewLastfmAuthPage(opts.svc), nil
	})

	// --- Phase 3.3.2 page migrations ---

	// Search page: the shell keeps the built instance as a singleton because
	// its wordsInput/result/searchType state is read back by SearchResultMenu
	// and operate.go through the shell (3.3.3 addresses those readers).
	RegisterPage("search", func(opts SearchPageOpts) (model.Page, error) {
		return NewSearchPage(opts.Netease), nil
	})

	RegisterPage("lastfm_custom_api", func(opts LastfmCustomAPIPageOpts) (model.Page, error) {
		return NewLastfmCustomAPIPage(opts.svc), nil
	})
}
