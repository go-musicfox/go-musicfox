package lastfm

import (
	"fmt"
	"os"
	"testing"

	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/framework"
	lastfmclient "github.com/go-musicfox/go-musicfox/internal/lastfm"
	"github.com/go-musicfox/go-musicfox/internal/storage"
	ui "github.com/go-musicfox/go-musicfox/internal/ui"
)

// TestMain starts the plugin once (P5: the registration window moved from
// package init() to the plugin Start) so the package-global registries hold
// its menu/page providers and the LastFM main-menu entry before the tests run.
//
// The plugin's Deps resolves ServiceLastfm, so the context must carry the
// *lastfm.Client. Constructing the client reads the storage manager and
// configs.AppConfig (lastfm.NewClient → InitFromStorage / getAPIKey); tests
// provide a minimal bootstrap (the same shape as withTestLastfmDB below, kept
// for the per-test client flows).
func TestMain(m *testing.M) {
	previousDB := storage.DBManager
	storage.DBManager = &storage.LocalDBManager{}
	previousConfig := configs.AppConfig
	configs.AppConfig = &configs.Config{}

	code := 1
	ctx := &framework.Context{}
	ctx.Provide(ui.ServiceLastfm, lastfmclient.NewClient())
	if err := (&Plugin{}).Start(ctx); err != nil {
		fmt.Fprintln(os.Stderr, "lastfm plugin Start error:", err)
	} else {
		code = m.Run()
	}

	_ = storage.DBManager.Close()
	storage.DBManager = previousDB
	configs.AppConfig = previousConfig
	os.Exit(code)
}
