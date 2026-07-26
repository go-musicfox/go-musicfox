package model

import (
	"image/color"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/anhoder/foxful-cli/style"
)

// Tabs renders a horizontal tab bar with keyboard navigation and optional content area.
// Supports bordered tab bar and content display for a complete TUI panel system.
// Active tab is highlighted, inactive tabs use the default menu item style.
type Tabs struct {
	titles      []string
	active      int
	hoveredIdx  int
	focused     bool
	width       int
	height      int
	content     string
	showBorder  bool
	borderStyle lipgloss.Border
}

// NewTabs creates a new tab bar with the given titles.
// The first tab (index 0) is initially active.
// By default, borders are enabled with rounded border style.
func NewTabs(titles []string) *Tabs {
	return &Tabs{
		titles:      titles,
		active:      0,
		hoveredIdx:  -1,
		width:       80,
		height:      1,
		showBorder:  true,
		borderStyle: lipgloss.RoundedBorder(),
	}
}

// Focus marks the tabs as focused for keyboard input.
func (t *Tabs) Focus() {
	t.focused = true
}

// Blur removes focus from the tabs.
func (t *Tabs) Blur() {
	t.focused = false
}

// Focused returns whether the tabs currently have focus.
func (t *Tabs) Focused() bool {
	return t.focused
}

// SetSize updates the available width and height for rendering.
// Tabs will truncate titles if width is constrained.
// When content is set, height determines the content area size.
func (t *Tabs) SetSize(width, height int) {
	t.width = width
	t.height = height
}

// SetContent sets the content to be displayed below the tab bar.
// Content will be rendered in a bordered box when showBorder is true.
// Pass empty string to clear content.
func (t *Tabs) SetContent(content string) {
	t.content = content
}

// SetBorder enables or disables border rendering for tabs and content.
func (t *Tabs) SetBorder(show bool) {
	t.showBorder = show
}

// SetBorderStyle sets the border style for tabs and content areas.
// Common styles: lipgloss.RoundedBorder(), lipgloss.NormalBorder(), lipgloss.ThickBorder().
func (t *Tabs) SetBorderStyle(border lipgloss.Border) {
	t.borderStyle = border
}

// Active returns the index of the currently active tab.
func (t *Tabs) Active() int {
	return t.active
}

// SetActive sets the active tab by index.
// The index is clamped to the valid range [0, len(titles)-1].
func (t *Tabs) SetActive(index int) {
	if len(t.titles) == 0 {
		t.active = 0
		return
	}
	t.active = clampInt(index, 0, len(t.titles)-1)
}

// SetHovered sets the hovered tab index (-1 for none).
func (t *Tabs) SetHovered(index int) {
	t.hoveredIdx = index
}

// Next moves to the next tab, wrapping to the first tab after the last.
func (t *Tabs) Next() {
	if len(t.titles) == 0 {
		return
	}
	t.active = (t.active + 1) % len(t.titles)
}

// Prev moves to the previous tab, wrapping to the last tab before the first.
func (t *Tabs) Prev() {
	if len(t.titles) == 0 {
		return
	}
	t.active--
	if t.active < 0 {
		t.active = len(t.titles) - 1
	}
}

// Update handles keyboard input for tab navigation.
// Supports: left/h (prev), right/l (next), home/g (first), end/G (last), 1-9 (jump to tab).
func (t *Tabs) Update(msg tea.Msg) tea.Cmd {
	if !t.focused {
		return nil
	}

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "left", "h":
			t.Prev()
		case "right", "l":
			t.Next()
		case "home", "g":
			t.SetActive(0)
		case "end", "G":
			t.SetActive(len(t.titles) - 1)
		case "1", "2", "3", "4", "5", "6", "7", "8", "9":
			index := int(msg.String()[0] - '1')
			if index < len(t.titles) {
				t.SetActive(index)
			}
		}
	}

	return nil
}

// View renders the complete tab system: tab bar and optional content area.
// Active tab uses SelectedItem style, inactive tabs use MenuItem style.
// When content is set, renders a bordered content area below the tabs.
func (t *Tabs) View() string {
	if len(t.titles) == 0 {
		return ""
	}

	ss := style.CurrentStyleSet()

	// Render tab bar with per-tab borders
	tabBar := t.renderTabBar(ss, t.hoveredIdx)

	// If no content, return just the tab bar
	if t.content == "" {
		return tabBar
	}

	// Render tab bar with content area
	return t.renderWithContent(ss, tabBar)
}

// renderTabBar renders tabs as individual bordered boxes.
// The active tab has a notch cut in its bottom border to connect to content below.
func (t *Tabs) renderTabBar(ss style.StyleSet, hoveredIdx int) string {
	if !t.showBorder {
		// Simple rendering without borders
		var parts []string
		for i, title := range t.titles {
			if i == t.active {
				parts = append(parts, ss.SelectedItem.Render(title))
			} else {
				parts = append(parts, ss.MenuItem.Render(title))
			}
		}
		return strings.Join(parts, ss.AppBackground.Render(" "))
	}

	// Border definitions matching lipgloss layout example
	// Active tab has space at bottom to create notch/opening
	activeTabBorder := lipgloss.Border{
		Top:         "─",
		Bottom:      " ", // Space creates opening beneath active tab
		Left:        "│",
		Right:       "│",
		TopLeft:     "╭",
		TopRight:    "╮",
		BottomLeft:  "┘", // Notch corners
		BottomRight: "└",
	}

	// Inactive tabs have T-junction at bottom to connect to bar
	tabBorder := lipgloss.Border{
		Top:         "─",
		Bottom:      "─",
		Left:        "│",
		Right:       "│",
		TopLeft:     "╭",
		TopRight:    "╮",
		BottomLeft:  "┴", // T-junction into bar
		BottomRight: "┴",
	}

	// 只有获得键盘焦点的活动 Tab 使用主色边框；其他边框不指定颜色。
	inactiveBorderColor := color.Color(lipgloss.NoColor{})
	activeBorderColor := inactiveBorderColor
	if t.focused {
		activeBorderColor = ss.SelectedItem.GetForeground()
	}

	tab := lipgloss.NewStyle().
		Border(tabBorder, true).
		BorderForeground(inactiveBorderColor).
		Padding(0, 1)

	activeTab := lipgloss.NewStyle().
		Border(activeTabBorder, true).
		BorderForeground(activeBorderColor).
		Padding(0, 1)

	// Render each tab
	var renderedTabs []string
	for i, title := range t.titles {
		titleStyle := ss.MenuItem
		if (i == t.active && t.focused) || i == hoveredIdx {
			// 聚焦和 hover 仅使用主色前景，避免 SelectedItem 的背景色。
			titleStyle = lipgloss.NewStyle().Foreground(ss.SelectedItem.GetForeground())
		}
		tabContent := titleStyle.Render(title)
		if i == t.active {
			renderedTabs = append(renderedTabs, activeTab.Render(tabContent))
		} else {
			renderedTabs = append(renderedTabs, tab.Render(tabContent))
		}
	}

	// Join tabs horizontally (aligned at top)
	row := lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)

	// Create gap filler to extend the bottom bar across remaining width
	rowWidth := lipgloss.Width(row)
	gapWidth := max(0, t.width-rowWidth-2)
	gapBorderColor := inactiveBorderColor
	if t.focused && t.active == len(t.titles)-1 {
		// 最右侧 Tab 聚焦时，高亮延伸到第三行最右边界。
		gapBorderColor = activeBorderColor
	}

	if gapWidth > 0 {
		// Gap extends the bottom border line
		tabGap := lipgloss.NewStyle().
			Inherit(ss.AppBackground).
			Border(tabBorder, true).
			BorderForeground(gapBorderColor).
			BorderTop(false).
			BorderLeft(false).
			BorderRight(false)

		gap := tabGap.Render(strings.Repeat(" ", gapWidth))
		row = lipgloss.JoinHorizontal(lipgloss.Bottom, row, gap)
	}

	return row
}

// renderWithContent renders tab bar with content area in a complete bordered panel.
func (t *Tabs) renderWithContent(ss style.StyleSet, tabBar string) string {
	if !t.showBorder {
		// Simple rendering without borders
		return lipgloss.JoinVertical(lipgloss.Left, tabBar, t.content)
	}

	// Calculate content height (total height minus tab bar)
	// Tab bar is ~3 lines (top border + content + bottom border/gap)
	contentHeight := t.height - 4
	if contentHeight < 1 {
		contentHeight = 1
	}

	// Content box with full borders - top border aligns with tab bar bottom
	borderColor := t.getBorderColor(ss)
	contentBox := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(borderColor).
		Padding(1, 2).
		Width(t.width - 4).
		Height(contentHeight).
		Render(t.content)

	// Join tab bar and content area vertically
	return lipgloss.JoinVertical(lipgloss.Left, tabBar, contentBox)
}

// getBorderColor returns the appropriate border color based on focus state.
func (t *Tabs) getBorderColor(ss style.StyleSet) color.Color {
	if t.focused {
		return ss.SelectedItem.GetForeground()
	}
	return ss.Border.GetForeground()
}
