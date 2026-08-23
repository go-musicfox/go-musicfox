package ui

// PROTO: this file is the Phase 3.4.1 compile check for
// docs/plugin_development.md. It mirrors the documented example-plugin
// snippets (hello menu + hello page) against the REAL provider registry so the
// doc's API usage is proven to compile and behave as documented. It is a
// verification artifact, NOT a shipped plugin. Key names use the "example_"
// prefix that production code never registers.
//
// If the docs are edited, update this file to match. Do not use these types in
// production code.

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/anhoder/foxful-cli/model"
)

// docExampleHelloMenu mirrors docs/plugin_development.md's "hello menu"
// example: a static-item-list menu registered under key "example_hello".
type docExampleHelloMenu struct {
	baseMenu
	menus []model.MenuItem
}

func (m *docExampleHelloMenu) GetMenuKey() string { return "example_hello" }

func (m *docExampleHelloMenu) MenuViews() []model.MenuItem { return m.menus }

func (m *docExampleHelloMenu) SubMenu(_ *model.App, _ int) model.Menu { return nil }

func (m *docExampleHelloMenu) BeforeEnterMenuHook() model.Hook {
	return func(_ *model.Main) (bool, model.Page) {
		m.menus = []model.MenuItem{
			{Title: "你好，插件世界！", Subtitle: "这是静态菜单项 A"},
			{Title: "插件菜单项 B", Subtitle: "不触发任何跳转"},
		}
		return true, nil
	}
}

// docExampleHelloPageOpts mirrors the doc's page-plugin opts contract.
type docExampleHelloPageOpts struct{}

// docExampleHelloPage mirrors docs/plugin_development.md's minimal page plugin.
type docExampleHelloPage struct{}

func (p *docExampleHelloPage) IgnoreQuitKeyMsg(tea.KeyMsg) bool { return false }

func (p *docExampleHelloPage) Type() model.PageType { return "example_hello_page" }

func (p *docExampleHelloPage) Update(tea.Msg, *model.App) (model.Page, tea.Cmd) { return p, nil }

func (p *docExampleHelloPage) View(*model.App) string { return "你好，插件页面！" }

func (p *docExampleHelloPage) Msg() tea.Msg { return nil }

// TestDocPluginExampleHelloMenu registers and builds the doc's hello menu
// through the real registry API (RegisterMenu -> BuildMenu).
func TestDocPluginExampleHelloMenu(t *testing.T) {
	RegisterMenu("example_hello", func(base baseMenu, _ NoArgMenuOpts) (Menu, error) {
		return &docExampleHelloMenu{baseMenu: base}, nil
	})

	menu, err := BuildMenu("example_hello", testBase, NoArgMenuOpts{})
	if err != nil {
		t.Fatalf("BuildMenu(example_hello) error = %v", err)
	}
	hello, ok := menu.(*docExampleHelloMenu)
	if !ok {
		t.Fatalf("BuildMenu(example_hello) = %T, want *docExampleHelloMenu", menu)
	}
	if hello.GetMenuKey() != "example_hello" {
		t.Fatalf("GetMenuKey() = %q, want %q", hello.GetMenuKey(), "example_hello")
	}

	// BeforeEnterMenuHook populates the static item list (doc behavior).
	if ok, _ := hello.BeforeEnterMenuHook()(nil); !ok {
		t.Fatal("hello menu BeforeEnterMenuHook returned ok=false")
	}
	if len(hello.MenuViews()) != 2 {
		t.Fatalf("MenuViews() = %v, want 2 static items", hello.MenuViews())
	}
}

// TestDocPluginExampleHelloPage registers and builds the doc's hello page
// through the real page registry API (RegisterPage -> BuildPage).
func TestDocPluginExampleHelloPage(t *testing.T) {
	RegisterPage("example_hello_page", func(_ docExampleHelloPageOpts) (model.Page, error) {
		return &docExampleHelloPage{}, nil
	})

	page, err := BuildPage("example_hello_page", docExampleHelloPageOpts{})
	if err != nil {
		t.Fatalf("BuildPage(example_hello_page) error = %v", err)
	}
	if _, ok := page.(*docExampleHelloPage); !ok {
		t.Fatalf("BuildPage(example_hello_page) = %T, want *docExampleHelloPage", page)
	}
}
