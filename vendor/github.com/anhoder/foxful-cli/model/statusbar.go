package model

import (
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
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

// StatusBarComponent is a renderable module injected into DefaultStatusBar's
// centered content area.
type StatusBarComponent interface {
	View(a *App, m *Main) string
}

// InteractiveStatusBarComponent is an optional extension for components that
// own mouse behavior. Coordinates are relative to the component's rendered text.
type InteractiveStatusBarComponent interface {
	StatusBarComponent
	HandleMouse(mouse tea.Mouse, x, y int) (handled bool, cmd tea.Cmd)
	IsMouseOver(x, y int) bool
}

// DefaultStatusBar shows a "PATH" nugget, the breadcrumb path on bar background,
// injected centered components, and the current time on the right.
type DefaultStatusBar struct {
	Components []StatusBarComponent

	componentBounds []statusBarComponentBounds
}

type statusBarComponentBounds struct {
	component StatusBarComponent
	start     int
	end       int
}

const maxBreadcrumbSegmentWidth = 32

func (d *DefaultStatusBar) View(a *App, m *Main) string {
	w := a.WindowWidth()
	if w <= 0 {
		return ""
	}
	d.componentBounds = d.componentBounds[:0]

	ss := style.CurrentStyleSet()

	// Left: "PATH" label nugget
	pathLabel := ss.StatusBarNuggetLabel.Render(" » ")
	labelW := lipgloss.Width(pathLabel)

	// Breadcrumb right-end separator
	breadcrumbBg := ss.StatusBarBreadcrumbBg.GetBackground()
	breadcrumbEndSepStyle := lipgloss.NewStyle().Foreground(breadcrumbBg).Background(ss.StatusBar.GetBackground())

	// Right: time nugget with Powerline separators
	now := time.Now().Format("15:04")
	timeBg := ss.StatusBarTime.GetBackground()
	timeStyle := lipgloss.NewStyle().
		Foreground(ss.StatusBarTime.GetForeground()).
		Background(timeBg)
	sepStyle := lipgloss.NewStyle().Foreground(timeBg).Background(ss.StatusBar.GetBackground())
	// Fall back to space if no Powerline character configured (e.g., Transparent theme).
	timeSepLeft := ss.StatusBarTimeSepLeft
	if timeSepLeft == "" {
		timeSepLeft = " "
	}
	timeNugget := sepStyle.Render(timeSepLeft) + timeStyle.Render("⏱ "+now+" ")
	timeW := lipgloss.Width(timeNugget)

	// Breadcrumb: constrain to available width so bar stays single-line.
	// Calculate max width optimistically (without time) in case time gets hidden.
	breadcrumbPadding := 1 // PaddingLeft(1)
	maxBreadcrumbW := w - labelW - breadcrumbPadding
	if len(d.Components) > 0 {
		// Breadcrumb must leave room for centered components to sit roughly
		// in the middle. Limit breadcrumb to the left half of the terminal
		// (minus label, padding) so components fit in the right half.
		centerHalfWidth := w / 2
		maxBreadcrumbW = centerHalfWidth - labelW - breadcrumbPadding
	}
	if maxBreadcrumbW < 10 {
		maxBreadcrumbW = 10
	}
	path := buildBreadcrumbPath(m, ss, maxBreadcrumbW)

	var breadcrumbBlock string
	if path != "" {
		bcSepRight := ss.StatusBarBreadcrumbSepRight
		if bcSepRight == "" {
			bcSepRight = " "
		}
		breadcrumbBlock = lipgloss.NewStyle().
			Inherit(ss.StatusBarBreadcrumbBg).
			PaddingLeft(1).
			Render(path) + breadcrumbEndSepStyle.Render(bcSepRight)
	}
	bcrumbW := lipgloss.Width(breadcrumbBlock)

	// Center: injected components.
	type renderedComponent struct {
		component StatusBarComponent
		width     int
	}
	renderedComponents := make([]renderedComponent, 0, len(d.Components))
	centerBlocks := make([]string, 0, len(d.Components))
	for _, component := range d.Components {
		if component == nil {
			continue
		}
		content := component.View(a, m)
		if content == "" {
			continue
		}
		renderedComponents = append(renderedComponents, renderedComponent{
			component: component,
			width:     lipgloss.Width(content),
		})
		centerBlocks = append(centerBlocks, lipgloss.NewStyle().
			Inherit(ss.StatusBarText).
			Padding(0, 1).
			Render(content))
	}
	centerBlock := lipgloss.JoinHorizontal(lipgloss.Top, centerBlocks...)
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
	if centerW > 0 {
		componentStart := leftUsed + leftFillerW
		for _, component := range renderedComponents {
			componentStart++ // left padding
			d.componentBounds = append(d.componentBounds, statusBarComponentBounds{
				component: component.component,
				start:     componentStart,
				end:       componentStart + component.width,
			})
			componentStart += component.width + 1 // content + right padding
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

func (d *DefaultStatusBar) handleComponentClick(mouse tea.Mouse, a *App, m *Main) (tea.Cmd, bool) {
	if mouse.Y != m.statusBarRowY(a) {
		return nil, false
	}
	for _, bounds := range d.componentBounds {
		if mouse.X < bounds.start || mouse.X >= bounds.end {
			continue
		}
		if component, ok := bounds.component.(InteractiveStatusBarComponent); ok {
			handled, cmd := component.HandleMouse(mouse, mouse.X-bounds.start, 0)
			return cmd, handled
		}
		return nil, false
	}
	return nil, false
}

func (d *DefaultStatusBar) isOverComponent(x, y int, a *App, m *Main) bool {
	if y != m.statusBarRowY(a) {
		return false
	}
	for _, bounds := range d.componentBounds {
		if x < bounds.start || x >= bounds.end {
			continue
		}
		component, ok := bounds.component.(InteractiveStatusBarComponent)
		return ok && component.IsMouseOver(x-bounds.start, 0)
	}
	return false
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
