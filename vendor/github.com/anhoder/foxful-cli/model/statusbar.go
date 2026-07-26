package model

import (
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/anhoder/foxful-cli/style"
	"github.com/charmbracelet/x/ansi"
)

// StatusBarPosition defines where the status bar is rendered.
type StatusBarPosition int

const (
	// StatusBarBottom renders the status bar at the bottom of the screen (default).
	StatusBarBottom StatusBarPosition = iota
	// StatusBarTop renders the status bar at the top, replacing the TitleView.
	StatusBarTop
)

// StatusBar is the interface for the bottom status bar.
// Downstream apps implement this to show playback status, progress, etc.
type StatusBar interface {
	View(a *App, m *Main) string
}

// DefaultStatusBar shows a "PATH" nugget, the breadcrumb path on bar background,
// and the current time on the right, in a lipgloss nugget-style status bar.
// If Center is set, its output is inserted between the breadcrumb and time.
type DefaultStatusBar struct {
	// Center is an optional callback that returns text to display in the center.
	// If nil, the center area remains empty (filled with StatusBarText background).
	Center func(a *App, m *Main) string
}

const maxBreadcrumbSegmentWidth = 32

func (d *DefaultStatusBar) View(a *App, m *Main) string {
	w := a.WindowWidth()
	if w <= 0 {
		return ""
	}

	ss := style.CurrentStyleSet()

	// Left: "PATH" label nugget
	pathLabel := ss.StatusBarNuggetLabel.Render(" » ")
	labelW := lipgloss.Width(pathLabel)

	// Right: time nugget
	now := time.Now().Format("15:04")
	timeNugget := ss.StatusBarTime.Render(" ⏱ " + now + " ")
	timeW := lipgloss.Width(timeNugget)

	// Breadcrumb: constrain to available width so bar stays single-line.
	// Calculate max width optimistically (without time) in case time gets hidden.
	breadcrumbPadding := 2 // Padding(0,1) on each side of breadcrumbBlock
	maxBreadcrumbW := w - labelW - breadcrumbPadding
	if d.Center != nil {
		// Breadcrumb must leave room for the center module to sit roughly
		// in the middle. Limit breadcrumb to the left half of the terminal
		// (minus label, padding) so center fits in the right half.
		centerHalfWidth := w / 2
		maxBreadcrumbW = centerHalfWidth - labelW - breadcrumbPadding
	}
	if maxBreadcrumbW < 10 {
		maxBreadcrumbW = 10
	}
	path := buildBreadcrumbPath(m, ss, maxBreadcrumbW)

	var breadcrumbBlock string
	if path != "" {
		breadcrumbBlock = lipgloss.NewStyle().
			Inherit(ss.StatusBarBreadcrumbBg).
			Padding(0, 1).
			Render(path)
	}
	bcrumbW := lipgloss.Width(breadcrumbBlock)

	// Center: optional callback content
	var centerBlock string
	if d.Center != nil {
		centerContent := d.Center(a, m)
		if centerContent != "" {
			centerBlock = lipgloss.NewStyle().
				Inherit(ss.StatusBarText).
				Padding(0, 1).
				Render(centerContent)
		}
	}
	centerW := lipgloss.Width(centerBlock)

	// If total content exceeds window width, progressively hide time then center.
	if labelW+bcrumbW+centerW+timeW > w {
		if timeW > 0 {
			timeNugget = ""
			timeW = 0
		}
	}
	if labelW+bcrumbW+centerW+timeW > w {
		if centerW > 0 {
			centerBlock = ""
			centerW = 0
		}
	}

	// Layout: label + breadcrumb + leftFiller + center + rightFiller + time
	leftUsed := labelW + bcrumbW
	rightUsed := timeW

	leftFillerW := (w-centerW)/2 - leftUsed
	if leftFillerW < 0 {
		leftFillerW = 0
	}
	rightFillerW := w - leftUsed - leftFillerW - centerW - rightUsed
	if rightFillerW < 0 {
		rightFillerW = 0
		leftFillerW = w - leftUsed - centerW - rightUsed
		if leftFillerW < 0 {
			leftFillerW = 0
		}
	}

	leftFiller := lipgloss.NewStyle().
		Inherit(ss.StatusBarText).
		Width(leftFillerW).
		Render("")
	rightFiller := lipgloss.NewStyle().
		Inherit(ss.StatusBarText).
		Width(rightFillerW).
		Render("")

	bar := lipgloss.JoinHorizontal(lipgloss.Top, pathLabel, breadcrumbBlock, leftFiller, centerBlock, rightFiller, timeNugget)
	return ss.StatusBar.Width(w).Render(bar)
}

// breadcrumbSegmentInfo describes a single segment in the breadcrumb display.
type breadcrumbSegmentInfo struct {
	// DisplayTitle is the title after truncation (may include ellipsis suffix).
	DisplayTitle string
	// DisplayWidth is the visual width of the title segment.
	DisplayWidth int
	// DepthIndex is the index into the full breadcrumb path (0 = root).
	DepthIndex int
	// IsLast is true for the current (deepest) menu level.
	IsLast bool
	// IsEllipsis is true when this segment is the "..." truncation placeholder.
	IsEllipsis bool
}

// computeBreadcrumbSegments returns the display-ready breadcrumb segments
// without rendering. Shared between the status bar view and mouse hit-testing.
func computeBreadcrumbSegments(m *Main) []breadcrumbSegmentInfo {
	stackItems := m.menuStack.ToSlice()
	var fullPath []string
	for _, item := range stackItems {
		if stackItem, ok := item.(*menuStackItem); ok {
			fullPath = append(fullPath, stackItem.menuTitle.Title)
		}
	}

	// Avoid duplicating the current title when the stack was just pushed but
	// the deferred tick hasn't updated m.menuTitle yet (enterMenuWithLoading).
	lastIdx := len(fullPath) - 1
	if lastIdx < 0 || fullPath[lastIdx] != m.menuTitle.Title {
		fullPath = append(fullPath, m.menuTitle.Title)
	}

	if len(fullPath) <= 0 {
		return nil
	}

	// Limit to last 3 levels
	var display []string
	if len(fullPath) > 3 {
		display = append(display, "...")
		display = append(display, fullPath[len(fullPath)-3:]...)
	} else {
		display = fullPath
	}

	// Compute segments with truncated titles
	truncated := len(fullPath) > 3
	segments := make([]breadcrumbSegmentInfo, 0, len(display))
	for i, title := range display {
		isLast := i == len(display)-1
		isDots := title == "..."

		var displayTitle string
		if isDots {
			displayTitle = title
		} else {
			displayTitle = title
			if lipgloss.Width(displayTitle) > maxBreadcrumbSegmentWidth {
				displayTitle = ansi.Truncate(displayTitle, maxBreadcrumbSegmentWidth, "…")
			}
		}

		// Map display index back to full-path index
		var depthIdx int
		if isDots {
			depthIdx = 0 // unused — ellipsis is not clickable
		} else if truncated {
			// After "...", segments align to the last N entries of fullPath
			depthIdx = len(fullPath) - (len(display) - i)
		} else {
			depthIdx = i
		}

		segments = append(segments, breadcrumbSegmentInfo{
			DisplayTitle: displayTitle,
			DisplayWidth: lipgloss.Width(displayTitle),
			DepthIndex:   depthIdx,
			IsLast:       isLast,
			IsEllipsis:   isDots,
		})
	}

	return segments
}

// buildBreadcrumbPath builds the styled breadcrumb path string for the status bar.
// Applies hover/click effects based on m.hoveredBreadcrumbIdx.
// maxWidth is the maximum visual width allowed; 0 means no constraint.
// When constrained, leftmost segments are dropped and replaced with "..." to keep
// the rightmost (current) segment visible.
func buildBreadcrumbPath(m *Main, ss style.StyleSet, maxWidth int) string {
	segments := computeBreadcrumbSegments(m)
	if len(segments) == 0 {
		return ""
	}

	separator := lipgloss.NewStyle().
		Inherit(ss.StatusBarBreadcrumbSep).
		Foreground(ss.StatusBarBreadcrumbSep.GetForeground()).
		Render(" / ")
	breadcrumbBase := ss.StatusBarBreadcrumb

	if maxWidth > 0 {
		// Compute total width and trim left segments if needed
		sepW := lipgloss.Width(separator)
		for len(segments) > 1 {
			totalW := 0
			for i, seg := range segments {
				if i > 0 {
					totalW += sepW
				}
				totalW += seg.DisplayWidth
			}
			if totalW <= maxWidth {
				break
			}
			// Drop leftmost non-ellipsis segment
			if segments[0].IsEllipsis && len(segments) > 1 {
				// drop the segment after the ellipsis, keep ellipsis
				segments = append(segments[:1], segments[2:]...)
			} else {
				segments = segments[1:]
			}
			// Ensure "..." is at the front
			if !segments[0].IsEllipsis {
				segments = append([]breadcrumbSegmentInfo{{
					DisplayTitle: "...",
					DisplayWidth: 3,
					IsEllipsis:   true,
				}}, segments...)
			}
		}
	}

	parts := make([]string, 0, len(segments))
	for i, seg := range segments {
		isHovered := !seg.IsLast && !seg.IsEllipsis && i == m.hoveredBreadcrumbIdx

		var styled string
		switch {
		case seg.IsEllipsis:
			styled = breadcrumbBase.Render(seg.DisplayTitle)
		case seg.IsLast:
			styled = breadcrumbBase.Bold(true).Render(seg.DisplayTitle)
		case isHovered:
			styled = ss.StatusBarBreadcrumbHover.Render(seg.DisplayTitle)
		default:
			styled = breadcrumbBase.Render(seg.DisplayTitle)
		}
		parts = append(parts, styled)
	}

	return strings.Join(parts, separator)
}
