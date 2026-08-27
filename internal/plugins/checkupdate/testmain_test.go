package checkupdate

import (
	"fmt"
	"os"
	"testing"

	"github.com/go-musicfox/go-musicfox/internal/framework"
)

// TestMain starts the plugin once (P5: the registration window moved from
// package init() to the plugin Start) so the package-global registries hold
// its contributions (menu provider / main-menu entry / startup hook) before
// the tests run.
func TestMain(m *testing.M) {
	if err := (&Plugin{}).Start(&framework.Context{}); err != nil {
		fmt.Fprintln(os.Stderr, "checkupdate plugin Start error:", err)
		os.Exit(1)
	}
	os.Exit(m.Run())
}
