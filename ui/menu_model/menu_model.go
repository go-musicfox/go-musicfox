package ui

// menu
type iMenu interface {
	// is cur menu playable?
	isPlayable() bool

	// should reset when enter cur menu?
	enterResetPlaylist() bool

	// is dj menu?
	isDj() bool

	// get menu unique key
	getMenuKey() string

	getSubMenuViews() []string

	getSubMenu(index int) iMenu

	hookBottomOut() []iMenu
}