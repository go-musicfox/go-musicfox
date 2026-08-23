package lastfm

import (
	"testing"

	"github.com/anhoder/foxful-cli/model"

	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/framework"
	lastfmclient "github.com/go-musicfox/go-musicfox/internal/lastfm"
	"github.com/go-musicfox/go-musicfox/internal/storage"
	ui "github.com/go-musicfox/go-musicfox/internal/ui"
)

// withTestLastfmDB makes the lastfm client storage-safe: lastfm.NewClient reads
// the API account through storage.DBManager, which the runtime bootstrap
// normally sets; a zero-value manager avoids a nil receiver (GetByKVModel
// errors are swallowed by InitFromStorage).
func withTestLastfmDB(t *testing.T) {
	t.Helper()
	previousDB := storage.DBManager
	storage.DBManager = &storage.LocalDBManager{}
	t.Cleanup(func() {
		_ = storage.DBManager.Close()
		storage.DBManager = previousDB
	})
	// lastfm.NewClient → NewTracker reads configs.AppConfig.Reporter.Lastfm;
	// the runtime bootstrap loads it, tests provide a dummy.
	previousConfig := configs.AppConfig
	configs.AppConfig = &configs.Config{}
	t.Cleanup(func() { configs.AppConfig = previousConfig })
}

// TestLastfmMenuBuildsViaRegistryFactory proves the plugin menu is
// constructible through the real registry factory (moved from the ui
// "last_fm" provider, Phase 3.9): the factory in registry.go builds a *Lastfm
// from a ui.BaseMenu base.
func TestLastfmMenuBuildsViaRegistryFactory(t *testing.T) {
	menu, err := ui.BuildMenu("last_fm", ui.BaseMenu{}, ui.NoArgMenuOpts{})
	if err != nil {
		t.Fatalf("BuildMenu(last_fm) error = %v", err)
	}
	lf, ok := menu.(*Lastfm)
	if !ok {
		t.Fatalf("BuildMenu(last_fm) = %T, want *Lastfm", menu)
	}
	if key := lf.GetMenuKey(); key != "last_fm" {
		t.Fatalf("GetMenuKey() = %q, want last_fm", key)
	}
}

// TestLastfmProfileConstruction proves the profile menu (not registered — it is
// reached directly from the Lastfm menu's SubMenu) builds from the exported
// base with its original key.
func TestLastfmProfileConstruction(t *testing.T) {
	profile := NewLastfmProfile(ui.BaseMenu{})
	if key := profile.GetMenuKey(); key != "lastfm_profile" {
		t.Fatalf("GetMenuKey() = %q, want lastfm_profile", key)
	}
}

// TestLastfmAuthPageBuildsViaRegistryFactory proves the page provider
// registration (moved from ui): "lastfm_auth" builds a *LastfmAuthPage through
// ui.BuildPage with an accessor carried in the opts.
func TestLastfmAuthPageBuildsViaRegistryFactory(t *testing.T) {
	page, err := ui.BuildPage("lastfm_auth", LastfmAuthPageOpts{Svc: ui.NewMenuServices(nil)})
	if err != nil {
		t.Fatalf("BuildPage(lastfm_auth) error = %v", err)
	}
	auth, ok := page.(*LastfmAuthPage)
	if !ok {
		t.Fatalf("BuildPage(lastfm_auth) = %T, want *LastfmAuthPage", page)
	}
	if auth.Type() == "" {
		t.Fatal("page.Type() is empty")
	}
}

// TestLastfmCustomAPIPageBuildsViaRegistryFactory proves "lastfm_custom_api"
// builds through ui.BuildPage with service resolution via a test context: the
// constructor reloads the stored API account through svc.Lastfm(), which the
// accessor resolves from the framework context.
func TestLastfmCustomAPIPageBuildsViaRegistryFactory(t *testing.T) {
	withTestLastfmDB(t)
	ctx := &framework.Context{}
	ctx.Provide(ui.ServiceLastfm, lastfmclient.NewClient())
	page, err := ui.BuildPage("lastfm_custom_api", LastfmCustomAPIPageOpts{Svc: ui.NewMenuServices(ctx)})
	if err != nil {
		t.Fatalf("BuildPage(lastfm_custom_api) error = %v", err)
	}
	if _, ok := page.(*LastfmCustomAPIPage); !ok {
		t.Fatalf("BuildPage(lastfm_custom_api) = %T, want *LastfmCustomAPIPage", page)
	}
}

// TestServiceResolutionViaTestContext proves the plugin's service access path:
// a ui.MenuServices rooted at a framework context resolves the registered
// Last.fm client, and degrades to nil (with a warning log) when the service is
// missing.
func TestServiceResolutionViaTestContext(t *testing.T) {
	withTestLastfmDB(t)
	ctx := &framework.Context{}
	ctx.Provide(ui.ServiceLastfm, lastfmclient.NewClient())
	svc := ui.NewMenuServices(ctx)
	if got := svc.Lastfm(); got == nil {
		t.Fatal("svc.Lastfm() = nil, want the registered client")
	}
	if got := ui.NewMenuServices(nil).Lastfm(); got != nil {
		t.Fatalf("svc.Lastfm() without a context = %v, want nil", got)
	}
}

// TestLastfmQRAuthPageConstruction proves the QR page (not registered — it is
// reached only from the auth page's authByQRCode) constructs with its original
// wiring.
func TestLastfmQRAuthPageConstruction(t *testing.T) {
	page := NewLastfmQRAuthPage(ui.NewMenuServices(nil), nil, nil)
	if page.Type() != LastfmQRAuthPageType {
		t.Fatalf("Type() = %q, want %q", page.Type(), LastfmQRAuthPageType)
	}
}

// TestLastfmMainMenuItemRegistered proves the plugin declares its own main-menu
// entry (the built-in entry was removed from menu_main.go, Phase 3.9).
func TestLastfmMainMenuItemRegistered(t *testing.T) {
	found := false
	for _, item := range ui.MainMenuPluginItems() {
		if item.Key == "last_fm" && item.Title == "LastFM" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("last_fm main-menu item not registered")
	}
}

var _ model.Page = (*LastfmAuthPage)(nil)      // provider contract sanity check
var _ model.Page = (*LastfmCustomAPIPage)(nil) // provider contract sanity check
var _ model.Page = (*LastfmQRAuthPage)(nil)    // provider contract sanity check
