package model

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/anhoder/foxful-cli/style"
	"github.com/charmbracelet/x/ansi"
)

const (
	contextMenuFrameOverhead  = 2 // rounded border only (no padding inside frame)
	contextMenuMinInnerWidth  = 8
	contextMenuHorizontalPad  = 2 // 1 cell on each side of an item label
	contextMenuScrollbarWidth = 1
)

// ContextMenuItem describes a single entry in a context menu.
type ContextMenuItem struct {
	ID        string
	Label     string
	Disabled  bool
	Separator bool // when true, renders as a separator line; other fields ignored
	Header    bool // 分组标题行：不可选中、加粗、显示 Label
}

// ContextMenu is a vertical list modal anchored at mouse coordinates.
// It appears on right-click and executes Menu.ContextMenuAction when an item is selected.
type ContextMenu struct {
	menu         Menu
	itemIndex    int // the menu list item that was right-clicked
	items        []ContextMenuItem
	mouseX       int
	mouseY       int
	maxWidth     int
	maxHeight    int
	effectiveMaxHeight int // auto-constrained from terminal height when maxHeight==0
	effectiveMaxWidth  int // auto-constrained from terminal width when maxWidth==0
	scrollOffset int

	focused     int // keyboard-focused item index (-1 = none)
	hovered     int // mouse-hovered item index (-1 = none)
	isDismissed bool
	isCanceled  bool
	selected    *ContextMenuItem

	bounds     popupRect
	boundsSet  bool
	itemBounds []popupRect // absolute screen coordinates for each selectable item
}

// NewContextMenu constructs an unlimited context menu anchored at (mouseX, mouseY).
func NewContextMenu(menu Menu, itemIndex int, items []ContextMenuItem, mouseX, mouseY int) *ContextMenu {
	return newContextMenu(menu, itemIndex, items, mouseX, mouseY, ContextMenuOptions{})
}

func newContextMenu(menu Menu, itemIndex int, items []ContextMenuItem, mouseX, mouseY int, options ContextMenuOptions) *ContextMenu {
	return &ContextMenu{
		menu:      menu,
		itemIndex: itemIndex,
		items:     items,
		mouseX:    mouseX,
		mouseY:    mouseY,
		maxWidth:  max(options.MaxWidth, 0),
		maxHeight: max(options.MaxHeight, 0),
		focused:   -1,
		hovered:   -1,
	}
}

func (cm *ContextMenu) isSelectable(index int) bool {
	if index < 0 || index >= len(cm.items) {
		return false
	}
	item := cm.items[index]
	return !item.Disabled && !item.Separator && !item.Header
}

// firstSelectableFrom finds the first selectable item starting at `from`, advancing by `delta`.
// Returns -1 if none found.
func (cm *ContextMenu) firstSelectableFrom(from, delta int) int {
	if delta == 0 {
		return -1
	}
	for i := from; i >= 0 && i < len(cm.items); i += delta {
		if cm.isSelectable(i) {
			return i
		}
	}
	return -1
}

// nextSelectable wraps around to find the next selectable item.
func (cm *ContextMenu) nextSelectable(from int) int {
	if len(cm.items) == 0 {
		return -1
	}
	// Try forward from current+1
	if next := cm.firstSelectableFrom(from+1, 1); next != -1 {
		return next
	}
	// Wrap to start
	return cm.firstSelectableFrom(0, 1)
}

// prevSelectable wraps around to find the previous selectable item.
func (cm *ContextMenu) prevSelectable(from int) int {
	if len(cm.items) == 0 {
		return -1
	}
	// Try backward from current-1
	if prev := cm.firstSelectableFrom(from-1, -1); prev != -1 {
		return prev
	}
	// Wrap to end
	return cm.firstSelectableFrom(len(cm.items)-1, -1)
}

func (cm *ContextMenu) dismissed() bool {
	return cm.isDismissed
}

func (cm *ContextMenu) dismissOutside() bool {
	cm.isDismissed = true
	cm.isCanceled = true
	return true
}

func (cm *ContextMenu) dismissEscape() {
	cm.isDismissed = true
	cm.isCanceled = true
}

// complete is called after dismissal to execute the selected action.
func (cm *ContextMenu) complete(app *App) (Page, tea.Cmd) {
	if cm.isCanceled || cm.selected == nil {
		return nil, nil
	}
	return cm.menu.ContextMenuAction(app, cm.itemIndex, *cm.selected)
}

func (cm *ContextMenu) update(msg tea.Msg) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return
	}

	switch keyMsg.String() {
	case "esc":
		cm.dismissEscape()
	case "enter":
		if cm.focused >= 0 && cm.focused < len(cm.items) && cm.isSelectable(cm.focused) {
			cm.selected = &cm.items[cm.focused]
			cm.isDismissed = true
		}
	case "up", "k":
		if cm.focused == -1 {
			cm.focused = cm.firstSelectableFrom(len(cm.items)-1, -1)
		} else {
			cm.focused = cm.prevSelectable(cm.focused)
		}
		cm.ensureFocusedVisible()
	case "down", "j":
		if cm.focused == -1 {
			cm.focused = cm.firstSelectableFrom(0, 1)
		} else {
			cm.focused = cm.nextSelectable(cm.focused)
		}
		cm.ensureFocusedVisible()
	}
}

func (cm *ContextMenu) visibleItemCount() int {
	visible := len(cm.items)
	if visible == 0 {
		return visible
	}
	maxH := cm.maxHeight
	if maxH == 0 {
		maxH = cm.effectiveMaxHeight
	}
	if maxH == 0 {
		return visible
	}
	return min(visible, max(maxH-contextMenuFrameOverhead, 1))
}

func (cm *ContextMenu) maxScrollOffset() int {
	return max(len(cm.items)-cm.visibleItemCount(), 0)
}

func (cm *ContextMenu) isScrollable() bool {
	return cm.maxScrollOffset() > 0
}

func (cm *ContextMenu) scrollBy(delta int) {
	cm.scrollOffset = min(max(cm.scrollOffset+delta, 0), cm.maxScrollOffset())
	cm.hovered = -1
}

func (cm *ContextMenu) ensureFocusedVisible() {
	if cm.focused < 0 {
		return
	}
	visible := cm.visibleItemCount()
	if cm.focused < cm.scrollOffset {
		cm.scrollOffset = cm.focused
	} else if cm.focused >= cm.scrollOffset+visible {
		cm.scrollOffset = cm.focused - visible + 1
	}
}

func (cm *ContextMenu) handleMouse(msg tea.MouseMsg) (bool, tea.Cmd) {
	mouse := msg.Mouse()
	oldHovered := cm.hovered
	cm.hovered = cm.itemAt(mouse.X, mouse.Y)
	hoverChanged := oldHovered != cm.hovered

	var hoverCmd tea.Cmd
	if hoverChanged {
		if cm.hovered >= 0 {
			hoverCmd = setMousePointer("pointer")
		} else {
			hoverCmd = setMousePointer("default")
		}
	}

	if !cm.boundsSet {
		return true, hoverCmd
	}

	// Check if mouse is inside bounds
	if !cm.bounds.contains(mouse.X, mouse.Y) {
		if cm.hovered != -1 {
			cm.hovered = -1
			hoverCmd = setMousePointer("default")
		}
		return false, hoverCmd
	}

	if cm.isScrollable() {
		switch mouse.Button {
		case tea.MouseWheelUp:
			cm.scrollBy(-1)
			return true, setMousePointer("default")
		case tea.MouseWheelDown:
			cm.scrollBy(1)
			return true, setMousePointer("default")
		}
	}

	// Handle click inside menu
	if _, isClick := msg.(tea.MouseClickMsg); isClick && mouse.Button == tea.MouseLeft {
		if cm.hovered >= 0 && cm.hovered < len(cm.items) && cm.isSelectable(cm.hovered) {
			cm.selected = &cm.items[cm.hovered]
			cm.isDismissed = true
			return true, hoverCmd
		}
	}

	return true, hoverCmd
}

// itemAt returns the index of the item at screen coordinates (x, y), or -1 if none.
func (cm *ContextMenu) itemAt(x, y int) int {
	for i, bound := range cm.itemBounds {
		if bound.contains(x, y) {
			return i
		}
	}
	return -1
}

func (cm *ContextMenu) itemStyle(styles style.StyleSet, index int) lipgloss.Style {
	item := cm.items[index]
	switch {
	case item.Disabled:
		return styles.Popup.ContextMenuItemDisabled
	case index == cm.hovered:
		return styles.Popup.ContextMenuItemHover
	case index == cm.focused:
		return styles.Popup.ContextMenuItemFocused
	default:
		return styles.Popup.ContextMenuItem
	}
}

// renderModal renders the context menu as a vertical list with rounded border.
// termW/termH are the terminal dimensions, used to auto-constrain when no explicit
// MaxWidth/MaxHeight is set in ContextMenuOptions.
func (cm *ContextMenu) renderModal(styles style.StyleSet, termW, termH int) modalRender {
	// Set effective max dimensions: explicit options take precedence, otherwise
	// auto-constrain to terminal bounds so the menu never overflows the screen.
	cm.effectiveMaxHeight = cm.maxHeight
	if cm.effectiveMaxHeight == 0 {
		cm.effectiveMaxHeight = termH
	}
	cm.effectiveMaxWidth = cm.maxWidth
	if cm.effectiveMaxWidth == 0 {
		cm.effectiveMaxWidth = termW
	}

	if len(cm.items) == 0 {
		return modalRender{content: "", itemBounds: nil}
	}

	visibleCount := cm.visibleItemCount()
	cm.scrollOffset = min(cm.scrollOffset, cm.maxScrollOffset())
	visibleStart := cm.scrollOffset
	visibleEnd := visibleStart + visibleCount
	scrolling := visibleCount < len(cm.items)

	maxLabelWidth := 0
	for _, item := range cm.items {
		if !item.Separator {
			maxLabelWidth = max(maxLabelWidth, lipgloss.Width(item.Label))
		}
	}

	scrollbarWidth := 0
	if scrolling {
		scrollbarWidth = contextMenuScrollbarWidth
	}
	effectiveMaxW := cm.maxWidth
	if effectiveMaxW == 0 {
		effectiveMaxW = cm.effectiveMaxWidth
	}
	innerWidth := max(maxLabelWidth+contextMenuHorizontalPad, contextMenuMinInnerWidth) + scrollbarWidth
	if effectiveMaxW > 0 {
		minInnerWidth := contextMenuHorizontalPad + 1 + scrollbarWidth
		maxInnerWidth := max(effectiveMaxW-contextMenuFrameOverhead, minInnerWidth)
		innerWidth = min(innerWidth, maxInnerWidth)
	}
	itemWidth := innerWidth - scrollbarWidth
	labelWidth := max(itemWidth-contextMenuHorizontalPad, 1)

	thumbRow := -1
	if scrolling {
		thumbRow = cm.scrollOffset * (visibleCount - 1) / cm.maxScrollOffset()
	}

	rows := make([]string, 0, visibleCount)
	itemBounds := make([]popupRect, len(cm.items))
	for itemIndex := visibleStart; itemIndex < visibleEnd; itemIndex++ {
		item := cm.items[itemIndex]
		visibleRow := itemIndex - visibleStart
		var row string
		if item.Separator {
			row = styles.Popup.ContextMenuSeparator.
				Width(itemWidth).
				Render(strings.Repeat("─", itemWidth))
		} else {
			label := ansi.Truncate(item.Label, labelWidth, "…")
			if item.Header {
				row = styles.Popup.ContextMenuHeader.
					Width(itemWidth).
					Padding(0, 1).
					Render(label)
			} else {
				row = cm.itemStyle(styles, itemIndex).
					Width(itemWidth).
					Padding(0, 1).
					Render(label)
			}
		}

		if scrolling {
			scrollbar := styles.Popup.ScrollTrack.Render("│")
			if visibleRow == thumbRow {
				scrollbar = styles.Popup.ScrollThumb.Render("█")
			}
			row += scrollbar
		}
		rows = append(rows, row)

		if cm.isSelectable(itemIndex) {
			itemBounds[itemIndex] = popupRect{
				x: 1,
				y: 1 + visibleRow,
				w: itemWidth,
				h: 1,
			}
		}
	}

	inner := lipgloss.JoinVertical(lipgloss.Left, rows...)
	framed := styles.Popup.ContextMenuFrame.
		Padding(0).
		Render(inner)

	return modalRender{
		content:    framed,
		itemBounds: itemBounds,
	}
}

// computePosition calculates the top-left (x, y) for the context menu.
// Applies flip+clamp to keep the menu fully visible.
func (cm *ContextMenu) computePosition(termW, termH, menuW, menuH int) (int, int) {
	// Default: start one row below the clicked item
	x := cm.mouseX
	y := cm.mouseY + 1

	// Flip horizontally if it would overflow right edge
	if x+menuW > termW {
		x = cm.mouseX - menuW
		if x < 0 {
			x = 0
		}
	}

	// Clamp to the screen edges, keeping an overflowing menu as low as possible.
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}
	if x+menuW > termW {
		x = termW - menuW
	}
	if y+menuH > termH {
		y = termH - menuH
	}

	return x, y
}

// Bounds returns the absolute screen rectangle of this context menu as (x, y, width, height).
// Values are meaningful only after the menu has been rendered at least once
// (the app-level compositeModals phase computes and stores bounds).
func (cm *ContextMenu) Bounds() (x, y, w, h int) {
	return cm.bounds.x, cm.bounds.y, cm.bounds.w, cm.bounds.h
}

// setModalBounds stores the absolute screen position and computes absolute item bounds.
func (cm *ContextMenu) setModalBounds(x, y, w, h int, itemBounds []popupRect) {
	cm.bounds = popupRect{x: x, y: y, w: w, h: h}
	cm.boundsSet = true
	cm.itemBounds = make([]popupRect, len(itemBounds))
	for i, bound := range itemBounds {
		cm.itemBounds[i] = popupRect{
			x: x + bound.x,
			y: y + bound.y,
			w: bound.w,
			h: bound.h,
		}
	}
}

// allowsRightClickPassthrough returns true for ContextMenu, allowing right-click outside to reopen.
func (cm *ContextMenu) allowsRightClickPassthrough() bool {
	return true
}

type modalRender struct {
	content    string
	itemBounds []popupRect // relative to modal's top-left corner
}
