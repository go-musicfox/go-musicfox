// Package song implements the song cluster (相似歌曲 / 添加到歌单) as the
// ninth real plugin. The two menus moved from internal/ui verbatim with their
// provider keys unchanged — simi_songs (相似歌曲) and add_to_user_playlist
// (添加到歌单/从歌单删除) — neither declares a main-menu item (both are jump
// targets, not top-level entries). The shared opts stay in ui —
// ui.SimiSongsOpts / ui.AddToUserPlaylistOpts / ui.CurUser are used by
// operate.go (右键操作表) and this cluster references them. Cross-menu jumps
// stay key-based (search_result / operate -> "simi_songs" and
// "add_to_user_playlist", now provided by this plugin).
package song

import (
	ui "github.com/go-musicfox/go-musicfox/internal/ui"
)

// init is the compile-time registration entry (linked via the internal/plugins
// aggregator blank import, which cmd/musicfox.go pulls in). Every key is
// identical to the one the menu registered under in internal/ui before the
// extraction: simi_songs / add_to_user_playlist. Both are parameterized
// (SimiSongsOpts carries the source song; AddToUserPlaylistOpts carries the
// target user / song / add-or-del flag) and are pure jump targets, not
// main-menu entries.
func init() {
	ui.RegisterMenu("simi_songs", func(base ui.BaseMenu, opts ui.SimiSongsOpts) (ui.Menu, error) {
		return NewSimilarSongsMenu(base, opts.Song), nil
	})
	ui.RegisterMenu("add_to_user_playlist", func(base ui.BaseMenu, opts ui.AddToUserPlaylistOpts) (ui.Menu, error) {
		return NewAddToUserPlaylistMenu(base, opts.UserID, opts.Song, opts.IsAdd), nil
	})
}
