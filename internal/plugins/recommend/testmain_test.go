package recommend

import (
	"fmt"
	"os"
	"testing"

	"github.com/go-musicfox/go-musicfox/internal/framework"
)

// TestMain starts the plugin (and the playlist plugin, whose playlist_detail
// provider the cluster's cross-menu jump tests build through) once — P5 moved
// the registration window from package init() to each plugin's Start, so the
// package-global registries only hold the providers of plugins that were
// actually started.
func TestMain(m *testing.M) {
	ctx := &framework.Context{}
	constructors := framework.PluginConstructors()
	for _, id := range []string{"playlist", "recommend"} {
		ctor, ok := constructors[id]
		if !ok {
			fmt.Fprintln(os.Stderr, "plugin", id, "constructor not registered")
			os.Exit(1)
		}
		if err := ctor().Start(ctx); err != nil {
			fmt.Fprintln(os.Stderr, "plugin", id, "Start error:", err)
			os.Exit(1)
		}
	}
	os.Exit(m.Run())
}
