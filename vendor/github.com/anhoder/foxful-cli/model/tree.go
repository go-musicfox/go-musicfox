package model

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/anhoder/foxful-cli/style"
)

// TreeNode represents a node in the tree structure.
// Label is displayed, Children are sub-nodes, Expanded controls visibility of children.
type TreeNode struct {
	Label    string
	Children []*TreeNode
	Expanded bool
}

// Tree renders a vertical tree with expand/collapse and keyboard navigation.
type Tree struct {
	root          *TreeNode
	flat          []flatNode
	selectedIndex int
	focused       bool
	width         int
	height        int
	scrollOffset  int
	needsRebuild  bool
}

// flatNode represents a node in the flattened visible list with its indent level.
type flatNode struct {
	node  *TreeNode
	depth int
}

// NewTree creates a new tree widget with the given root node.
// The root node is initially selected.
func NewTree(root *TreeNode) *Tree {
	t := &Tree{
		root:         root,
		width:        80,
		height:       20,
		needsRebuild: true,
	}
	t.rebuildFlat()
	return t
}

// Focus marks the tree as focused for keyboard input.
func (t *Tree) Focus() {
	t.focused = true
}

// Blur removes focus from the tree.
func (t *Tree) Blur() {
	t.focused = false
}

// Focused returns whether the tree currently has focus.
func (t *Tree) Focused() bool {
	return t.focused
}

// SetSize updates the available width and height for rendering.
func (t *Tree) SetSize(width, height int) {
	t.width = width
	t.height = height
}

// SelectedNode returns the currently selected tree node.
// Never returns nil after construction.
func (t *Tree) SelectedNode() *TreeNode {
	if t.selectedIndex >= 0 && t.selectedIndex < len(t.flat) {
		return t.flat[t.selectedIndex].node
	}
	if len(t.flat) > 0 {
		return t.flat[0].node
	}
	return t.root
}

// Toggle expands or collapses the selected node if it has children.
func (t *Tree) Toggle() {
	node := t.SelectedNode()
	if len(node.Children) == 0 {
		return
	}
	node.Expanded = !node.Expanded
	t.needsRebuild = true
	t.rebuildFlat()
	t.ensureSelectionVisible()
}

// Expand expands the selected node if it has children.
func (t *Tree) Expand() {
	node := t.SelectedNode()
	if len(node.Children) == 0 {
		return
	}
	if !node.Expanded {
		node.Expanded = true
		t.needsRebuild = true
		t.rebuildFlat()
		t.ensureSelectionVisible()
	}
}

// Collapse collapses the selected node if it is expanded.
func (t *Tree) Collapse() {
	node := t.SelectedNode()
	if node.Expanded {
		node.Expanded = false
		t.needsRebuild = true
		t.rebuildFlat()
		t.ensureSelectionVisible()
	}
}

// Update handles keyboard input for tree navigation and expand/collapse.
// Supports: up/k/down/j (navigate), enter/space (toggle), right/l (expand), left/h (collapse),
// home/g (first), end/G (last).
func (t *Tree) Update(msg tea.Msg) tea.Cmd {
	if !t.focused {
		return nil
	}

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "up", "k":
			t.moveUp()
		case "down", "j":
			t.moveDown()
		case "enter", "space", " ":
			t.Toggle()
		case "right", "l":
			t.Expand()
		case "left", "h":
			t.Collapse()
		case "home", "g":
			t.selectedIndex = 0
			t.ensureSelectionVisible()
		case "end", "G":
			if len(t.flat) > 0 {
				t.selectedIndex = len(t.flat) - 1
			}
			t.ensureSelectionVisible()
		case "pgup":
			t.pageUp()
		case "pgdown":
			t.pageDown()
		}
	}

	return nil
}

// View renders the tree structure with indentation and expand/collapse indicators.
// Selection is highlighted using SelectedItem style.
func (t *Tree) View() string {
	if t.needsRebuild {
		t.rebuildFlat()
	}

	if len(t.flat) == 0 {
		return ""
	}

	ss := style.CurrentStyleSet()

	var lines []string
	visibleStart := t.scrollOffset
	visibleEnd := t.scrollOffset + t.height
	if visibleEnd > len(t.flat) {
		visibleEnd = len(t.flat)
	}

	for i := visibleStart; i < visibleEnd; i++ {
		fn := t.flat[i]
		indent := strings.Repeat("  ", fn.depth)

		var prefix string
		if len(fn.node.Children) > 0 {
			if fn.node.Expanded {
				prefix = "▾ "
			} else {
				prefix = "▸ "
			}
		} else {
			prefix = "• "
		}

		line := indent + prefix + fn.node.Label

		if i == t.selectedIndex {
			line = ss.SelectedItem.Render(line)
		} else {
			line = ss.MenuItem.Render(line)
		}

		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

// rebuildFlat reconstructs the flat visible node list from the root.
func (t *Tree) rebuildFlat() {
	t.flat = nil
	t.flattenNode(t.root, 0)
	t.needsRebuild = false

	// Clamp selection to valid range
	if t.selectedIndex >= len(t.flat) {
		t.selectedIndex = len(t.flat) - 1
	}
	if t.selectedIndex < 0 && len(t.flat) > 0 {
		t.selectedIndex = 0
	}
}

// flattenNode recursively adds nodes to the flat list in depth-first order.
func (t *Tree) flattenNode(node *TreeNode, depth int) {
	if node == nil {
		return
	}

	t.flat = append(t.flat, flatNode{node: node, depth: depth})

	if node.Expanded {
		for _, child := range node.Children {
			t.flattenNode(child, depth+1)
		}
	}
}

// moveUp moves selection up by one visible node, wrapping at the top.
func (t *Tree) moveUp() {
	if len(t.flat) == 0 {
		return
	}
	t.selectedIndex--
	if t.selectedIndex < 0 {
		t.selectedIndex = len(t.flat) - 1
	}
	t.ensureSelectionVisible()
}

// moveDown moves selection down by one visible node, wrapping at the bottom.
func (t *Tree) moveDown() {
	if len(t.flat) == 0 {
		return
	}
	t.selectedIndex++
	if t.selectedIndex >= len(t.flat) {
		t.selectedIndex = 0
	}
	t.ensureSelectionVisible()
}

// pageUp moves selection up by one page height.
func (t *Tree) pageUp() {
	if len(t.flat) == 0 {
		return
	}
	t.selectedIndex -= t.height
	if t.selectedIndex < 0 {
		t.selectedIndex = 0
	}
	t.ensureSelectionVisible()
}

// pageDown moves selection down by one page height.
func (t *Tree) pageDown() {
	if len(t.flat) == 0 {
		return
	}
	t.selectedIndex += t.height
	if t.selectedIndex >= len(t.flat) {
		t.selectedIndex = len(t.flat) - 1
	}
	t.ensureSelectionVisible()
}

// ensureSelectionVisible adjusts scroll offset to keep the selected node visible.
func (t *Tree) ensureSelectionVisible() {
	if len(t.flat) == 0 {
		return
	}

	// Scroll up if selection is above viewport
	if t.selectedIndex < t.scrollOffset {
		t.scrollOffset = t.selectedIndex
	}

	// Scroll down if selection is below viewport
	if t.selectedIndex >= t.scrollOffset+t.height {
		t.scrollOffset = t.selectedIndex - t.height + 1
	}

	// Clamp scroll offset
	maxScroll := len(t.flat) - t.height
	if maxScroll < 0 {
		maxScroll = 0
	}
	if t.scrollOffset > maxScroll {
		t.scrollOffset = maxScroll
	}
	if t.scrollOffset < 0 {
		t.scrollOffset = 0
	}
}
