package model

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/anhoder/foxful-cli/style"
	"github.com/charmbracelet/x/ansi"
)

// FilePicker is a composable file and directory browser widget. It renders
// a list of files and directories, supports keyboard and mouse navigation,
// and allows entering directories or selecting files.
type FilePicker struct {
	currentDir   string
	entries      []fileEntry
	selected     int
	scrollOffset int
	width        int
	height       int
	focused      bool
	showHidden   bool
	readError    error
}

// fileEntry represents a single file or directory in the picker.
type fileEntry struct {
	name  string
	isDir bool
}

// NewFilePicker creates a new FilePicker starting in the specified directory.
// If startDir is empty, it defaults to the current working directory.
func NewFilePicker(startDir string) *FilePicker {
	if startDir == "" {
		var err error
		startDir, err = os.Getwd()
		if err != nil {
			startDir = "."
		}
	}

	// Clean and make absolute
	startDir, _ = filepath.Abs(startDir)

	fp := &FilePicker{
		currentDir: startDir,
		showHidden: false,
		focused:    true,
	}
	fp.readDirectory()
	return fp
}

// Focus sets the picker to focused state.
func (fp *FilePicker) Focus() {
	fp.focused = true
}

// Blur sets the picker to unfocused state.
func (fp *FilePicker) Blur() {
	fp.focused = false
}

// Focused returns whether the picker is currently focused.
func (fp *FilePicker) Focused() bool {
	return fp.focused
}

// SetSize updates the dimensions available for rendering the picker.
func (fp *FilePicker) SetSize(width, height int) {
	fp.width = width
	fp.height = height
	fp.clampSelection()
}

// SetShowHidden controls whether hidden files (names starting with ".") are shown.
func (fp *FilePicker) SetShowHidden(show bool) {
	if fp.showHidden != show {
		fp.showHidden = show
		fp.readDirectory()
	}
}

// CurrentDir returns the absolute path of the current directory.
func (fp *FilePicker) CurrentDir() string {
	return fp.currentDir
}

// SelectedPath returns the absolute path of the currently selected entry.
// Returns an empty string if no entry is selected.
func (fp *FilePicker) SelectedPath() string {
	if fp.selected < 0 || fp.selected >= len(fp.entries) {
		return ""
	}
	return filepath.Join(fp.currentDir, fp.entries[fp.selected].name)
}

// Selected returns the absolute path and directory status of the selected entry.
// Returns empty string and false if no entry is selected.
func (fp *FilePicker) Selected() (path string, isDir bool) {
	if fp.selected < 0 || fp.selected >= len(fp.entries) {
		return "", false
	}
	entry := fp.entries[fp.selected]
	return filepath.Join(fp.currentDir, entry.name), entry.isDir
}

// readDirectory reads the current directory and populates the entries list.
// Sorts directories before files, each alphabetically.
func (fp *FilePicker) readDirectory() {
	fp.readError = nil
	entries, err := os.ReadDir(fp.currentDir)
	if err != nil {
		fp.readError = err
		fp.entries = nil
		fp.selected = 0
		fp.scrollOffset = 0
		return
	}

	// Filter and convert to fileEntry
	fp.entries = make([]fileEntry, 0, len(entries))
	for _, e := range entries {
		name := e.Name()
		// Skip hidden files unless showHidden is true
		if !fp.showHidden && strings.HasPrefix(name, ".") {
			continue
		}
		fp.entries = append(fp.entries, fileEntry{
			name:  name,
			isDir: e.IsDir(),
		})
	}

	// Sort: directories first, then files, each alphabetically
	sort.Slice(fp.entries, func(i, j int) bool {
		ei, ej := fp.entries[i], fp.entries[j]
		if ei.isDir != ej.isDir {
			return ei.isDir // dirs before files
		}
		return strings.ToLower(ei.name) < strings.ToLower(ej.name)
	})

	// Clamp selection to valid range
	if fp.selected >= len(fp.entries) {
		fp.selected = max(0, len(fp.entries)-1)
	}
	fp.clampSelection()
}

// navigateToParent moves to the parent directory, clamping at the filesystem root.
func (fp *FilePicker) navigateToParent() {
	parent := filepath.Dir(fp.currentDir)
	// Don't navigate above root
	if parent == fp.currentDir {
		return
	}
	fp.currentDir = parent
	fp.selected = 0
	fp.scrollOffset = 0
	fp.readDirectory()
}

// enterDirectory enters the currently selected directory.
func (fp *FilePicker) enterDirectory() {
	if fp.selected < 0 || fp.selected >= len(fp.entries) {
		return
	}
	entry := fp.entries[fp.selected]
	if !entry.isDir {
		return
	}
	fp.currentDir = filepath.Join(fp.currentDir, entry.name)
	fp.selected = 0
	fp.scrollOffset = 0
	fp.readDirectory()
}

// Update handles input messages for the file picker.
func (fp *FilePicker) Update(msg tea.Msg) tea.Cmd {
	if !fp.focused {
		return nil
	}

	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return nil
	}

	switch keyMsg.String() {
	case "up", "k":
		fp.selected = max(0, fp.selected-1)
		fp.ensureVisible()

	case "down", "j":
		fp.selected = min(fp.selected+1, max(0, len(fp.entries)-1))
		fp.ensureVisible()

	case "pgup":
		jump := max(fp.height/2, 1)
		fp.selected = max(0, fp.selected-jump)
		fp.ensureVisible()

	case "pgdown":
		jump := max(fp.height/2, 1)
		fp.selected = min(fp.selected+jump, max(0, len(fp.entries)-1))
		fp.ensureVisible()

	case "home", "g":
		fp.selected = 0
		fp.ensureVisible()

	case "end", "G":
		fp.selected = max(0, len(fp.entries)-1)
		fp.ensureVisible()

	case "enter", "right", "l":
		if fp.selected >= 0 && fp.selected < len(fp.entries) {
			if fp.entries[fp.selected].isDir {
				fp.enterDirectory()
			}
		}

	case "left", "h", "backspace":
		fp.navigateToParent()
	}

	return nil
}

// visibleRows returns the number of entry rows available below the header.
func (fp *FilePicker) visibleRows() int {
	return max(0, fp.height-1) // reserve 1 line for the header
}

// ensureVisible adjusts scrollOffset to keep the selected entry visible.
func (fp *FilePicker) ensureVisible() {
	rows := fp.visibleRows()
	if rows <= 0 {
		return
	}
	if fp.selected < fp.scrollOffset {
		fp.scrollOffset = fp.selected
	}
	if fp.selected >= fp.scrollOffset+rows {
		fp.scrollOffset = fp.selected - rows + 1
	}
	fp.clampSelection()
}

// clampSelection ensures selection and scroll are within valid bounds.
func (fp *FilePicker) clampSelection() {
	if len(fp.entries) == 0 {
		fp.selected = 0
		fp.scrollOffset = 0
		return
	}
	fp.selected = clampInt(fp.selected, 0, len(fp.entries)-1)
	maxScroll := max(0, len(fp.entries)-fp.visibleRows())
	fp.scrollOffset = clampInt(fp.scrollOffset, 0, maxScroll)
}

// View renders the file picker.
func (fp *FilePicker) View() string {
	styles := style.CurrentStyleSet()

	if fp.width <= 0 || fp.height <= 0 {
		return ""
	}

	// Render header: current directory path
	header := fp.renderHeader(styles)

	// Render error if directory read failed
	if fp.readError != nil {
		errorMsg := styles.Error.Render(Tf(MsgReadError, fp.readError.Error()))
		return lipgloss.JoinVertical(lipgloss.Left, header, errorMsg)
	}

	// Render empty state
	if len(fp.entries) == 0 {
		emptyMsg := styles.Muted.Render(T(MsgEmptyDirectory))
		return lipgloss.JoinVertical(lipgloss.Left, header, emptyMsg)
	}

	// Render visible entries
	visibleHeight := fp.visibleRows()
	if visibleHeight <= 0 {
		return header
	}

	endIdx := min(fp.scrollOffset+visibleHeight, len(fp.entries))
	visibleEntries := fp.entries[fp.scrollOffset:endIdx]

	lines := make([]string, 0, len(visibleEntries))
	for i, entry := range visibleEntries {
		globalIdx := fp.scrollOffset + i
		lines = append(lines, fp.renderEntry(entry, globalIdx == fp.selected, styles))
	}

	body := strings.Join(lines, "\n")
	return lipgloss.JoinVertical(lipgloss.Left, header, body)
}

// renderHeader renders the current directory path as a truncated header.
func (fp *FilePicker) renderHeader(styles style.StyleSet) string {
	displayPath := fp.currentDir
	if fp.width > 0 {
		displayPath = ansi.Truncate(displayPath, fp.width, "…")
	}
	return styles.Subtitle.Render(displayPath)
}

// renderEntry renders a single file or directory entry.
func (fp *FilePicker) renderEntry(entry fileEntry, isSelected bool, styles style.StyleSet) string {
	// Build display name with directory indicator
	displayName := entry.name
	if entry.isDir {
		displayName += "/"
	}

	// Truncate to fit width
	if fp.width > 0 {
		displayName = ansi.Truncate(displayName, fp.width, "…")
	}

	// Style based on type and selection
	var styled string
	if isSelected {
		styled = styles.SelectedItem.Render(displayName)
	} else {
		if entry.isDir {
			styled = styles.Info.Render(displayName)
		} else {
			styled = styles.MenuItem.Render(displayName)
		}
	}

	// Pad to full width if needed
	if fp.width > 0 {
		currentWidth := lipgloss.Width(styled)
		if currentWidth < fp.width {
			if isSelected {
				styled = styles.SelectedItem.Width(fp.width).Render(displayName)
			} else {
				padding := strings.Repeat(" ", fp.width-currentWidth)
				styled += padding
			}
		}
	}

	return styled
}
