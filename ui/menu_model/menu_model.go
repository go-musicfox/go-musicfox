package ui

// menu interface
type iMenu interface {
    // is cur menu playable?
    isPlayable() bool

    // should reset when enter cur menu?
    enterResetPlaylist() bool

    // get menu unique key.
    getMenuKey() string

    // get sub menu list.
    getSubMenuViews() []string

    // get sub menu by index.
    getSubMenu(index int) iMenu

    // get extra view. The extra view will render when getSubMenuViews return empty.
    getExtraView() string

    // hook on bottom out
    onBottomOut() []iMenu
}

// dj menu interface
type iDjMenu interface {
    iMenu
}