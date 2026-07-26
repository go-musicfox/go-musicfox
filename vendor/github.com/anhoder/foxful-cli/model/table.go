package model

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/anhoder/foxful-cli/style"
	"github.com/charmbracelet/x/ansi"
)

// Column defines a table column with title and width.
// Width 0 means auto/flex sizing based on available space.
type Column struct {
	Title string
	Width int
}

// Table is a composable table widget with header, scrollable rows, and keyboard navigation.
// Supports column-based layout with truncation, row selection, and vertical scrolling.
type Table struct {
	columns []Column
	rows    [][]string

	selectedRow  int
	scrollOffset int

	focused bool
	width   int
	height  int
}

// NewTable creates a new Table with the given columns and rows.
// Returns a Table ready for embedding in a host model.
func NewTable(columns []Column, rows [][]string) *Table {
	return &Table{
		columns:      columns,
		rows:         rows,
		selectedRow:  0,
		scrollOffset: 0,
		focused:      false,
	}
}

// Focus marks the table as focused.
func (t *Table) Focus() {
	t.focused = true
}

// Blur marks the table as blurred.
func (t *Table) Blur() {
	t.focused = false
}

// Focused returns whether the table is currently focused.
func (t *Table) Focused() bool {
	return t.focused
}

// SetSize sets the table's display dimensions.
func (t *Table) SetSize(width, height int) {
	t.width = width
	t.height = height
}

// Update handles input events for the table.
func (t *Table) Update(msg tea.Msg) tea.Cmd {
	if !t.focused || len(t.rows) == 0 {
		return nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		return t.handleKey(msg)
	}

	return nil
}

// handleKey processes keyboard input for navigation.
func (t *Table) handleKey(msg tea.KeyMsg) tea.Cmd {
	key := msg.String()

	switch key {
	case "up", "k":
		t.moveSelection(-1)
	case "down", "j":
		t.moveSelection(1)
	case "pgup":
		pageSize := max(t.visibleRows()/2, 1)
		t.moveSelection(-pageSize)
	case "pgdown":
		pageSize := max(t.visibleRows()/2, 1)
		t.moveSelection(pageSize)
	case "home", "g":
		t.selectedRow = 0
		t.ensureVisible()
	case "end", "G":
		if len(t.rows) > 0 {
			t.selectedRow = len(t.rows) - 1
		}
		t.ensureVisible()
	case "enter":
		// Host should handle activation
		return nil
	}

	return nil
}

// moveSelection moves the selected row by delta, clamping to valid range.
func (t *Table) moveSelection(delta int) {
	if len(t.rows) == 0 {
		return
	}
	t.selectedRow = clampInt(t.selectedRow+delta, 0, len(t.rows)-1)
	t.ensureVisible()
}

// ensureVisible adjusts scroll offset to keep the selected row visible.
func (t *Table) ensureVisible() {
	visibleRows := t.visibleRows()
	if visibleRows <= 0 {
		return
	}

	if t.selectedRow < t.scrollOffset {
		t.scrollOffset = t.selectedRow
	} else if t.selectedRow >= t.scrollOffset+visibleRows {
		t.scrollOffset = t.selectedRow - visibleRows + 1
	}

	maxOffset := max(len(t.rows)-visibleRows, 0)
	t.scrollOffset = clampInt(t.scrollOffset, 0, maxOffset)
}

// visibleRows returns the number of rows that can fit in the current height.
func (t *Table) visibleRows() int {
	if t.height <= 2 {
		return 0
	}
	return t.height - 2 // subtract header and separator
}

// SelectedRow returns the index of the currently selected row.
// Returns -1 if no rows are available.
func (t *Table) SelectedRow() int {
	if len(t.rows) == 0 {
		return -1
	}
	return t.selectedRow
}

// SetRows replaces the table's rows and resets selection.
func (t *Table) SetRows(rows [][]string) {
	t.rows = rows
	t.selectedRow = 0
	t.scrollOffset = 0
	if len(rows) > 0 {
		t.selectedRow = 0
	}
}

// Rows returns the table's current rows.
func (t *Table) Rows() [][]string {
	return t.rows
}

// View renders the table.
func (t *Table) View() string {
	styles := style.CurrentStyleSet()

	// Empty state
	if len(t.columns) == 0 || len(t.rows) == 0 {
		emptyMsg := T(MsgNoData)
		if len(t.columns) == 0 {
			emptyMsg = T(MsgNoColumns)
		}
		return styles.Muted.Render(emptyMsg)
	}

	// Compute column widths
	colWidths := t.computeColumnWidths()

	var b strings.Builder

	// Render header
	b.WriteString(t.renderHeader(colWidths, styles))
	b.WriteString("\n")

	// Render separator
	b.WriteString(t.renderSeparator(colWidths, styles))
	b.WriteString("\n")

	// Render visible rows
	visibleRows := t.visibleRows()
	if visibleRows <= 0 {
		return b.String()
	}

	endIdx := min(t.scrollOffset+visibleRows, len(t.rows))
	for i := t.scrollOffset; i < endIdx; i++ {
		b.WriteString(t.renderRow(i, colWidths, styles))
		if i < endIdx-1 {
			b.WriteString("\n")
		}
	}

	// Render scroll indicators if needed
	if t.needsScrollIndicator() {
		return t.addScrollIndicator(b.String(), styles)
	}

	return b.String()
}

// computeColumnWidths calculates display widths for all columns.
func (t *Table) computeColumnWidths() []int {
	if len(t.columns) == 0 {
		return nil
	}

	widths := make([]int, len(t.columns))
	totalFixed := 0
	flexCount := 0

	// First pass: use explicit widths or measure content
	for i, col := range t.columns {
		if col.Width > 0 {
			widths[i] = col.Width
			totalFixed += col.Width
		} else {
			flexCount++
			// Start with title width as minimum
			widths[i] = lipgloss.Width(col.Title)
		}
	}

	// Calculate available width for flex columns
	if flexCount > 0 && t.width > 0 {
		// Account for spacing between columns (1 space each)
		spacing := max(len(t.columns)-1, 0)
		available := t.width - totalFixed - spacing - 2 // -2 for potential scroll indicator

		if available > 0 {
			flexWidth := available / flexCount
			for i := range t.columns {
				if t.columns[i].Width == 0 {
					widths[i] = max(widths[i], flexWidth)
				}
			}
		}
	}

	// Ensure minimum width of 3 for each column
	for i := range widths {
		widths[i] = max(widths[i], 3)
	}

	return widths
}

// renderHeader renders the table header row.
func (t *Table) renderHeader(widths []int, styles style.StyleSet) string {
	var cells []string
	for i, col := range t.columns {
		title := col.Title
		if widths[i] < lipgloss.Width(title) {
			title = ansi.Truncate(title, widths[i], "…")
		}
		// Pad to column width
		cell := lipgloss.NewStyle().Width(widths[i]).Render(title)
		cells = append(cells, styles.Title.Render(cell))
	}
	return styles.AppBackground.Render(strings.Join(cells, styles.AppBackground.Render(" ")))
}

// renderSeparator renders a separator line below the header.
func (t *Table) renderSeparator(widths []int, styles style.StyleSet) string {
	var segments []string
	for _, w := range widths {
		segments = append(segments, strings.Repeat("─", w))
	}
	return styles.Border.Render(styles.AppBackground.Render(strings.Join(segments, styles.AppBackground.Render(" "))))
}

// renderRow renders a single data row.
func (t *Table) renderRow(rowIdx int, widths []int, styles style.StyleSet) string {
	row := t.rows[rowIdx]
	isSelected := rowIdx == t.selectedRow && t.focused

	var cells []string
	for i, width := range widths {
		var cellContent string
		if i < len(row) {
			cellContent = row[i]
		}

		// Truncate if needed
		if lipgloss.Width(cellContent) > width {
			cellContent = ansi.Truncate(cellContent, width, "…")
		}

		// Pad to column width
		cell := lipgloss.NewStyle().Width(width).Render(cellContent)

		// Apply row style
		if isSelected {
			cell = styles.SelectedItem.Render(cell)
		} else {
			cell = styles.MenuItem.Render(cell)
		}

		cells = append(cells, cell)
	}

	return styles.AppBackground.Render(strings.Join(cells, styles.AppBackground.Render(" ")))
}

// needsScrollIndicator returns true if scroll indicators should be shown.
func (t *Table) needsScrollIndicator() bool {
	return len(t.rows) > t.visibleRows()
}

// addScrollIndicator appends scroll indicator to the rendered content.
func (t *Table) addScrollIndicator(content string, styles style.StyleSet) string {
	if !t.needsScrollIndicator() {
		return content
	}

	lines := strings.Split(content, "\n")
	if len(lines) < 2 {
		return content
	}

	// Add indicators to header and last visible row
	maxOffset := max(len(t.rows)-t.visibleRows(), 0)

	// Top indicator (if not at top)
	if t.scrollOffset > 0 && len(lines) > 0 {
		lines[0] = lines[0] + styles.AppBackground.Render(" ") + styles.Muted.Render("▲")
	}

	// Bottom indicator (if not at bottom)
	if t.scrollOffset < maxOffset && len(lines) > 0 {
		lastIdx := len(lines) - 1
		lines[lastIdx] = lines[lastIdx] + styles.AppBackground.Render(" ") + styles.Muted.Render("▾")
	}

	return strings.Join(lines, "\n")
}
