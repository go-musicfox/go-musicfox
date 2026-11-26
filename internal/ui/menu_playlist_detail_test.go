package ui

import (
	"testing"

	"github.com/anhoder/foxful-cli/model"

	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/utils/notify"
	_struct "github.com/go-musicfox/go-musicfox/utils/struct"
)

func TestPlaylistDetailMenuBeforeEnterNotifiesOnFailure(t *testing.T) {
	prevLocale := model.DefaultCatalog().Locale()
	t.Cleanup(func() { model.SetLocale(prevLocale) })
	SetupI18n("zh")

	app, neteaseApp := newFormPageTestApp(t)

	prevFetch := fetchSongsOfPlaylist
	t.Cleanup(func() { fetchSongsOfPlaylist = prevFetch })
	fetchSongsOfPlaylist = func(playlistId int64, getAll bool) (_struct.ResCode, []structs.Song) {
		return _struct.NetworkError, nil
	}

	var gotContent notify.NotifyContent
	var gotLevel notify.ToastLevel
	notified := false
	notify.SetToastHook(func(content notify.NotifyContent, level notify.ToastLevel) {
		notified = true
		gotContent = content
		gotLevel = level
	})
	t.Cleanup(func() { notify.SetToastHook(nil) })

	menu := NewPlaylistDetailMenu(newBaseMenu(neteaseApp), 42)
	ok, page := menu.BeforeEnterMenuHook()(app.MustMain())
	if ok {
		t.Fatal("expected BeforeEnterMenuHook to return false on playlist fetch failure")
	}
	if page != nil {
		t.Fatalf("expected nil page on non-auth failure, got %T", page)
	}
	if !notified {
		t.Fatal("expected error toast notification on playlist fetch failure")
	}
	if gotContent.Title != "加载歌单失败" {
		t.Fatalf("expected notification title %q, got %q", "加载歌单失败", gotContent.Title)
	}
	if gotLevel != notify.ToastError {
		t.Fatalf("expected ToastError level, got %v", gotLevel)
	}
	if len(menu.Songs()) != 0 || len(menu.MenuViews()) != 0 {
		t.Fatalf("expected menu to remain empty on failure, got %d songs / %d views", len(menu.Songs()), len(menu.MenuViews()))
	}
}

func TestPlaylistDetailMenuBeforeEnterSucceedsWithSongs(t *testing.T) {
	prevLocale := model.DefaultCatalog().Locale()
	t.Cleanup(func() { model.SetLocale(prevLocale) })
	SetupI18n("zh")

	app, neteaseApp := newFormPageTestApp(t)

	prevFetch := fetchSongsOfPlaylist
	t.Cleanup(func() { fetchSongsOfPlaylist = prevFetch })
	fetchSongsOfPlaylist = func(playlistId int64, getAll bool) (_struct.ResCode, []structs.Song) {
		return _struct.Success, []structs.Song{
			{Id: 1, Name: "Song 1", Artists: []structs.Artist{{Name: "Artist 1"}}},
		}
	}

	notified := false
	notify.SetToastHook(func(content notify.NotifyContent, level notify.ToastLevel) {
		notified = true
	})
	t.Cleanup(func() { notify.SetToastHook(nil) })

	menu := NewPlaylistDetailMenu(newBaseMenu(neteaseApp), 42)
	ok, page := menu.BeforeEnterMenuHook()(app.MustMain())
	if !ok || page != nil {
		t.Fatalf("expected (true, nil), got (%v, %v)", ok, page)
	}
	if notified {
		t.Fatal("did not expect notification on successful playlist load")
	}
	if len(menu.Songs()) != 1 || len(menu.MenuViews()) != 1 {
		t.Fatalf("expected 1 song and 1 menu item, got %d / %d", len(menu.Songs()), len(menu.MenuViews()))
	}
}
