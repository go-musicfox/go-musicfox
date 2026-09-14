package kitty

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/anhoder/foxful-cli/layout"
	"github.com/charmbracelet/x/ansi"
)

// This integration test checks bytes sent to the outer terminal, not tmux's
// pane buffer: tmux 3.7c preserves combining marks in the buffer but drops them
// from incremental right-pane output.
func TestUnicodePlaceholderTmuxOutput(t *testing.T) {
	if runtime.GOOS == "windows" || os.Getenv("MUSICFOX_TMUX_INTEGRATION") != "1" {
		t.Skip("set MUSICFOX_TMUX_INTEGRATION=1; requires tmux and Python 3 on Unix")
	}
	type fixture struct {
		Frames []string `json:"frames"`
		Want   []string `json:"want"`
	}
	var f fixture
	var rows []string
	for row := range 5 {
		cells := UnicodePlaceholderRow(42, row, 10)
		rows = append(rows, strings.Repeat(" ", 15)+cells+" lyric")
		f.Want = append(f.Want, ansi.Strip(cells))
	}
	base := strings.Join(rows, "\n")
	// Use the same compositor as App.compositeNotifications. It preserves
	// graphemes but removes control sequences embedded in the base content.
	toast := layout.NewCompositor(layout.NewLayer(base), layout.NewLayer("Playing next song").X(45)).Render()
	partial := "\x1b[2;1H" + strings.Split(toast, "\n")[1] + " updated"
	f.Frames = []string{base, toast, partial, base}
	// Test-only bypass for checking tmux builds that have fixed the bug.
	if os.Getenv("MUSICFOX_TMUX_TEST_NO_SYNC") != "1" {
		for i, frame := range f.Frames {
			f.Frames[i] = captureTmuxOutput(t, frame)
		}
	}
	data, err := json.Marshal(f)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "rows.json")
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("python3", "testdata/tmux_placeholder_probe.py", path)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("tmux placeholder probe: %v\n%s", err, output)
	} else {
		t.Logf("%s", output)
	}
}
