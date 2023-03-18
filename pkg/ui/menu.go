package ui

import "github.com/go-musicfox/go-musicfox/pkg/structs"

type Hook func(m *NeteaseModel) bool

// Menu menu interface
type Menu interface {
	// IsPlayable 当前菜单是否可播放？
	IsPlayable() bool

	// IsSearchable 当前菜单是否可搜索
	IsSearchable() bool

	// IsLocatable 当前菜单是否支持播放自动定位
	IsLocatable() bool

	// RealDataIndex 索引转换
	RealDataIndex(index int) int

	// ResetPlaylistWhenPlay 空格播放时，是否重置播放列表？
	ResetPlaylistWhenPlay() bool

	// GetMenuKey 菜单唯一Key
	GetMenuKey() string

	// MenuViews 获取子菜单View
	MenuViews() []MenuItem

	// FormatMenuItem 进入前的格式化
	FormatMenuItem(item *MenuItem)

	// SubMenu 根据下标获取菜单Model
	SubMenu(model *NeteaseModel, index int) Menu

	// BeforePrePageHook 切换上一页前的Hook
	BeforePrePageHook() Hook

	// BeforeNextPageHook 切换下一页前的Hook
	BeforeNextPageHook() Hook

	// BeforeEnterMenuHook 进入菜单项前的Hook
	BeforeEnterMenuHook() Hook

	// BeforeBackMenuHook 菜单返回前的Hook
	BeforeBackMenuHook() Hook

	// BottomOutHook 触底的Hook
	BottomOutHook() Hook

	// TopOutHook 触顶Hook
	TopOutHook() Hook
}

// DjMenu dj menu interface
type DjMenu interface {
	Menu
}

type SongsMenu interface {
	Songs() []structs.Song
}

type PlaylistsMenu interface {
	Playlists() []structs.Playlist
}

type AlbumsMenu interface {
	Albums() []structs.Album
}

type ArtistsMenu interface {
	Artists() []structs.Artist
}

type DefaultMenu struct {
}

func (e *DefaultMenu) IsPlayable() bool {
	return false
}

func (e *DefaultMenu) IsSearchable() bool {
	return false
}

func (e *DefaultMenu) IsLocatable() bool {
	return true
}

func (e *DefaultMenu) RealDataIndex(index int) int {
	return index
}

func (e *DefaultMenu) ResetPlaylistWhenPlay() bool {
	return false
}

func (e *DefaultMenu) GetMenuKey() string {
	panic("implement me")
}

func (e *DefaultMenu) MenuViews() []MenuItem {
	return nil
}

func (e *DefaultMenu) FormatMenuItem(_ *MenuItem) {
}

func (e *DefaultMenu) SubMenu(_ *NeteaseModel, _ int) Menu {
	return nil
}

func (e *DefaultMenu) BeforePrePageHook() Hook {
	return nil
}

func (e *DefaultMenu) BeforeNextPageHook() Hook {
	return nil
}

func (e *DefaultMenu) BeforeEnterMenuHook() Hook {
	return nil
}

func (e *DefaultMenu) BeforeBackMenuHook() Hook {
	return nil
}

func (e *DefaultMenu) BottomOutHook() Hook {
	return nil
}

func (e *DefaultMenu) TopOutHook() Hook {
	return nil
}
