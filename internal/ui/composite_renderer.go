package ui

import (
	"strings"

	"github.com/anhoder/foxful-cli/model"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mattn/go-runewidth"
)

// CompositeRenderer combines multiple components horizontally with percentage-based width allocation.
// This is used to display the cover image alongside lyrics.
type CompositeRenderer struct {
	netease  *Netease
	columns  []CompositeColumn
	maxLines int // Maximum number of lines from all columns
}

// CompositeColumn represents a column in the composite layout.
type CompositeColumn struct {
	Component  model.Component
	WidthRatio float64 // Width percentage (0.0-1.0)
}

// NewCompositeRenderer creates a new composite renderer with the given columns.
func NewCompositeRenderer(netease *Netease, columns []CompositeColumn) *CompositeRenderer {
	return &CompositeRenderer{
		netease: netease,
		columns: columns,
	}
}

// Update handles UI messages and forwards them to all child components.
func (r *CompositeRenderer) Update(msg tea.Msg, a *model.App) {
	for _, col := range r.columns {
		if col.Component != nil {
			col.Component.Update(msg, a)
		}
	}
}

// View renders all columns side by side.
func (r *CompositeRenderer) View(a *model.App, main *model.Main) (view string, lines int) {
	if len(r.columns) == 0 {
		return "", 0
	}

	windowWidth := r.netease.WindowWidth()

	// Calculate actual widths for each column
	widths := make([]int, len(r.columns))
	totalRatio := 0.0
	for _, col := range r.columns {
		totalRatio += col.WidthRatio
	}

	// Normalize ratios and calculate widths
	remainingWidth := windowWidth
	for i, col := range r.columns {
		if i == len(r.columns)-1 {
			// Last column gets remaining width to avoid rounding issues
			widths[i] = remainingWidth
		} else {
			ratio := col.WidthRatio / totalRatio
			widths[i] = int(float64(windowWidth) * ratio)
			remainingWidth -= widths[i]
		}
	}

	// Get views from each column
	columnViews := make([][]string, len(r.columns))
	columnLines := make([]int, len(r.columns))
	r.maxLines = 0

	for i, col := range r.columns {
		if col.Component == nil {
			columnViews[i] = []string{}
			columnLines[i] = 0
			continue
		}

		viewStr, numLines := col.Component.View(a, main)
		columnViews[i] = splitViewIntoLines(viewStr)
		columnLines[i] = numLines

		if numLines > r.maxLines {
			r.maxLines = numLines
		}
	}

	// If no lines, return empty
	if r.maxLines == 0 {
		return "", 0
	}

	// Combine columns line by line
	var result strings.Builder
	for lineIdx := 0; lineIdx < r.maxLines; lineIdx++ {
		for colIdx, colLines := range columnViews {
			colWidth := widths[colIdx]

			var lineContent string
			if lineIdx < len(colLines) {
				lineContent = colLines[lineIdx]
			}

			// Calculate visible width (accounting for ANSI codes)
			visibleWidth := visibleStringWidth(lineContent)

			// Write the line content
			result.WriteString(lineContent)

			// Pad to column width if needed (except for last column)
			if colIdx < len(r.columns)-1 {
				padding := colWidth - visibleWidth
				if padding > 0 {
					result.WriteString(strings.Repeat(" ", padding))
				}
			}
		}
		if lineIdx < r.maxLines-1 {
			result.WriteString("\n")
		}
	}

	return result.String(), r.maxLines
}

// splitViewIntoLines splits a view string into individual lines.
func splitViewIntoLines(view string) []string {
	if view == "" {
		return []string{}
	}
	lines := strings.Split(view, "\n")
	// Remove trailing empty line if present
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

// visibleStringWidth calculates the visible width of a string, ignoring ANSI escape codes.
func visibleStringWidth(s string) int {
	// Strip ANSI codes to get visible content
	visible := stripAnsiCodesForComposite(s)
	return runewidth.StringWidth(visible)
}

// stripAnsiCodesForComposite removes ANSI escape sequences from a string.
func stripAnsiCodesForComposite(s string) string {
	var result strings.Builder
	inEscape := false
	for i := 0; i < len(s); i++ {
		if s[i] == '\033' && i+1 < len(s) && s[i+1] == '[' {
			inEscape = true
			i++ // skip '['
			continue
		}
		if inEscape {
			if (s[i] >= 'A' && s[i] <= 'Z') || (s[i] >= 'a' && s[i] <= 'z') {
				inEscape = false
			}
			continue
		}
		result.WriteByte(s[i])
	}
	return result.String()
}
