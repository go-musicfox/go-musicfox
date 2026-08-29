package headless

import (
	"os"
	"sync"
	"testing"

	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/core"
	"github.com/go-musicfox/go-musicfox/internal/storage"
	"github.com/go-musicfox/go-musicfox/internal/types"
	apputils "github.com/go-musicfox/go-musicfox/utils/app"
)

// The beep speaker can only be initialized once per process, and the engine's
// ctrl goroutine processes playback signals asynchronously (reading configs and
// storage while doing so). The package therefore runs all tests against ONE
// shared engine with process-lifetime config/storage setup, set up here in
// TestMain so no test ever tears the state out from under the engine's
// background goroutines.
var (
	sharedEngineOnce sync.Once
	sharedEngineVal  *core.Engine

	testRoot string
)

func TestMain(m *testing.M) {
	var err error
	testRoot, err = os.MkdirTemp("", "musicfox-headless-test-*")
	if err != nil {
		panic(err)
	}

	prevRoot := os.Getenv("MUSICFOX_ROOT")
	prevConfig := configs.AppConfig
	prevDB := storage.DBManager

	// The first app path access bootstraps the path manager from MUSICFOX_ROOT,
	// so any db/cache files land in the temp root, never in user dirs.
	_ = os.Setenv("MUSICFOX_ROOT", testRoot)
	// MUSICFOX_ROOT alone does not redirect the path manager: bootstrapOnce is
	// triggered early by the slogx init, so initPaths already cached the XDG
	// (real user dir) paths. SetupPureRoot forces the temp root — otherwise
	// headless.DialAddr() resolves to the real daemon socket and tests such as
	// TestDialSubscribeNoDaemon would observe the user's live daemon instead of
	// an empty test root.
	apputils.SetupPureRoot(testRoot)
	configs.AppConfig = &configs.Config{Player: configs.PlayerConfig{Engine: types.BeepPlayer}}
	storage.DBManager = new(storage.LocalDBManager)

	code := m.Run()

	storage.DBManager = prevDB
	configs.AppConfig = prevConfig
	if prevRoot == "" {
		_ = os.Unsetenv("MUSICFOX_ROOT")
	} else {
		_ = os.Setenv("MUSICFOX_ROOT", prevRoot)
	}
	_ = os.RemoveAll(testRoot)
	os.Exit(code)
}

// newTestEngine returns the shared real core engine (no Startup). Its
// engine.Player() is usable immediately after construction.
func newTestEngine(t *testing.T) *core.Engine {
	t.Helper()
	sharedEngineOnce.Do(func() {
		sharedEngineVal = core.NewEngine(core.EngineOptions{})
	})
	return sharedEngineVal
}
