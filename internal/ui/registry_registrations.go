package ui

import "github.com/anhoder/foxful-cli/model"

// Production provider registrations (Phase 3.2.1/3.2.2). The 5 sample menus
// move here from the Phase 3.0 prototype (registry_proto_b.go, kept as
// evidence); demo migrations add album_detail / user_playlist / dj_radio_detail
// and the lastfm_auth page.
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

	RegisterPage("login", func(opts LoginPageOpts) (model.Page, error) {
		return NewLoginPage(opts.Netease), nil
	})

	RegisterPage("lastfm_auth", func(opts LastfmAuthPageOpts) (model.Page, error) {
		return NewLastfmAuthPage(opts.Netease), nil
	})
}
