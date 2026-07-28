package ui

import "testing"

func TestMenuToPageDefersBeforeEnterUntilPageTransition(t *testing.T) {
	app, netease := newFormPageTestApp(t)
	target := NewSearchPage(netease)
	called := false
	menu := NewMenuToPage(newBaseMenu(netease), target, func() { called = true })
	if called {
		t.Fatal("before-enter callback ran while constructing the deferred menu")
	}

	completed, page := menu.BeforeEnterMenuHook()(app.MustMain())
	if completed {
		t.Fatal("menu-to-page hook completed menu entry, want page transition")
	}
	if page != target {
		t.Fatalf("hook returned %T, want target page", page)
	}
	if !called {
		t.Fatal("before-enter callback did not run during page transition")
	}
}
