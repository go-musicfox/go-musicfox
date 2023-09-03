package ui

import (
	"github.com/anhoder/foxful-cli/model"
	"github.com/go-musicfox/go-musicfox/internal/structs"
)

// Menu menu interface
type Menu interface {
	model.Menu

	// IsPlayable 当前菜单是否可播放？
	IsPlayable() bool

	// IsLocatable 当前菜单是否支持播放自动定位
	IsLocatable() bool
}

// DjMenu dj menu interface
type DjMenu interface {
	Menu
}

type SongsMenu interface {
	Menu
	Songs() []structs.Song
}

type PlaylistsMenu interface {
	Menu
	Playlists() []structs.Playlist
}

type AlbumsMenu interface {
	Menu
	Albums() []structs.Album
}

type ArtistsMenu interface {
	Menu
	Artists() []structs.Artist
}

type baseMenu struct {
	model.DefaultMenu
	netease *Netease
}

func newBaseMenu(netease *Netease) baseMenu {
	return baseMenu{
		netease: netease,
	}
}

func (e *baseMenu) IsPlayable() bool {
	return false
}

func (e *baseMenu) IsLocatable() bool {
	return true
}
