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
	"github.com/go-musicfox/go-musicfox/internal/framework"
	ui "github.com/go-musicfox/go-musicfox/internal/ui"
)

// Plugin is the song business plugin (P5 cordis shape): its Start registers
// the two song menu providers — the registration window moves from package
// init() to the frontend scope Start.
type Plugin struct {
	framework.NoopPlugin
}

// Start registers the plugin's contributions inside a ui.WithPlugin scope so
// the attribution stamp records them under "song". Every key is identical to
// the one the menu registered under in internal/ui before the extraction:
// simi_songs / add_to_user_playlist. Both are parameterized (SimiSongsOpts
// carries the source song; AddToUserPlaylistOpts carries the target user /
// song / add-or-del flag) and are pure jump targets, not main-menu entries.
func (p *Plugin) Start(_ *framework.Context) error {
	ui.WithPlugin("song", "单曲", func() {
		ui.RegisterMenu("simi_songs", func(base ui.BaseMenu, opts ui.SimiSongsOpts) (ui.Menu, error) {
			return NewSimilarSongsMenu(base, opts.Song), nil
		})
		ui.RegisterMenu("add_to_user_playlist", func(base ui.BaseMenu, opts ui.AddToUserPlaylistOpts) (ui.Menu, error) {
			return NewAddToUserPlaylistMenu(base, opts.UserID, opts.Song, opts.IsAdd), nil
		})
	})
	return nil
}

// init is the compile-time registration entry (linked via the internal/plugins
// aggregator blank import, which cmd/musicfox.go pulls in) and only declares
// the plugin constructor — actual registrations happen in Start (frontend
// scope).
func init() {
	framework.RegisterPlugin("song", func() framework.Plugin { return &Plugin{} })
}
