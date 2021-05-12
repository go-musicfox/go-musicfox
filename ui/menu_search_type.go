package ui

type SearchTypeMenu struct {
    menus    []MenuItem
}

func NewSearchTypeMenu() *SearchTypeMenu {
    typeMenu := new(SearchTypeMenu)
    typeMenu.menus = []MenuItem{
        {Title: "按单曲"},
        {Title: "按专辑"},
        {Title: "按歌手"},
        {Title: "按歌单"},
        {Title: "按用户"},
        {Title: "按歌词"},
        {Title: "按电台"},
    }

    return typeMenu
}

func (m *SearchTypeMenu) MenuData() interface{} {
    return nil
}

func (m *SearchTypeMenu) BeforeBackMenuHook() Hook {
    return nil
}

func (m *SearchTypeMenu) IsPlayable() bool {
    return false
}

func (m *SearchTypeMenu) ResetPlaylistWhenPlay() bool {
    return false
}

func (m *SearchTypeMenu) GetMenuKey() string {
    return "search_type"
}

func (m *SearchTypeMenu) MenuViews() []MenuItem {
    return m.menus
}

func (m *SearchTypeMenu) SubMenu(_ *NeteaseModel, index int) IMenu {

    typeArr := []SearchType{
        StSingleSong,
        StAlbum,
        StSinger,
        StPlaylist,
        StUser,
        StLyric,
        StRadio,
    }

    if index >= len(typeArr) {
        return nil
    }

    return NewSearchResultMenu(typeArr[index])
}

func (m *SearchTypeMenu) BeforePrePageHook() Hook {
    // Nothing to do
    return nil
}

func (m *SearchTypeMenu) BeforeNextPageHook() Hook {
    // Nothing to do
    return nil
}

func (m *SearchTypeMenu) BeforeEnterMenuHook() Hook {
    // Nothing to do
    return nil
}

func (m *SearchTypeMenu) BottomOutHook() Hook {
    // Nothing to do
    return nil
}

func (m *SearchTypeMenu) TopOutHook() Hook {
    // Nothing to do
    return nil
}

