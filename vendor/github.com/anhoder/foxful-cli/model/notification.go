package model

import (
	"time"

	tea "charm.land/bubbletea/v2"
)

// NotificationLevel defines the semantic level of a notification.
type NotificationLevel uint8

const (
	NotificationInfo NotificationLevel = iota
	NotificationSuccess
	NotificationWarning
	NotificationError
)

// NotificationID uniquely identifies an active notification.
type NotificationID uint64

// NotificationAction is an explicit action rendered in a notification's action area.
// ID and Label must be non-empty; Label must be plain, single-line text.
type NotificationAction struct {
	ID    string
	Label string
}

// NotificationActionResult identifies the notification action selected by the user.
type NotificationActionResult struct {
	NotificationID NotificationID
	ActionID       string
}

// NotificationSpec defines the content and behavior of a notification.
// Message may contain ANSI-styled text. Title and action labels are plain,
// single-line text. Timeout of 0 uses the configured default for Info/Success,
// except notifications with actions remain visible for user interaction.
// Warning/Error notifications with a zero Timeout also remain visible.
type NotificationSpec struct {
	Level    NotificationLevel
	Title    string
	Message  string
	Timeout  time.Duration
	Actions  []NotificationAction
	OnAction func(NotificationActionResult)
}

func cloneNotificationSpec(spec NotificationSpec) NotificationSpec {
	spec.Actions = append([]NotificationAction(nil), spec.Actions...)
	return spec
}

// notificationRect is the screen-absolute bounding box of a rendered notification.
type notificationRect struct {
	x, y, w, h int
}

func (r notificationRect) contains(x, y int) bool {
	return x >= r.x && x < r.x+r.w && y >= r.y && y < r.y+r.h
}

// Notification is an active notification instance managed by the App.
type Notification struct {
	id        NotificationID
	spec      NotificationSpec
	createdAt time.Time
	expireAt  time.Time // zero means no auto-expire

	hoveredAction int
	actionBounds  []notificationRect
	actionArea    notificationRect

	bounds    notificationRect
	boundsSet bool

	// Selection state (selecting/hasSelection/anchor+cursor coords and
	// contentLines) is provided by the embedded textSelection; its fields and
	// methods are promoted onto Notification.
	textSelection

	// Cached rendering state for hit-testing (populated during render)
	contentWidth int    // width in columns
	titleHeight  int    // 1 if title present, 0 if no title
	titleText    string // raw title text (icon + title) for clipboard copy
}

func (n *Notification) clearBounds() {
	n.bounds = notificationRect{}
	n.boundsSet = false
	n.actionBounds = nil
	n.actionArea = notificationRect{}
}

func (n *Notification) setBounds(x, y, w, h int, actionBounds []notificationRect, actionArea notificationRect) {
	n.bounds = notificationRect{x: x, y: y, w: w, h: h}
	n.boundsSet = true
	n.actionBounds = make([]notificationRect, len(actionBounds))
	for i, bound := range actionBounds {
		n.actionBounds[i] = notificationRect{x: x + bound.x, y: y + bound.y, w: bound.w, h: bound.h}
	}
	if actionArea.w > 0 && actionArea.h > 0 {
		n.actionArea = notificationRect{x: x + actionArea.x, y: y + actionArea.y, w: actionArea.w, h: actionArea.h}
	} else {
		n.actionArea = notificationRect{}
	}
}

func (n *Notification) resetInteraction() {
	n.clearSelection()
	n.contentLines = nil
	n.hoveredAction = -1
	n.clearBounds()
}

func (n *Notification) setContentGeometry(contentLines []string, cw, th int, titleText string) {
	n.contentLines = append(n.contentLines[:0], contentLines...)
	n.contentWidth = cw
	n.titleHeight = th
	n.titleText = titleText
}

// pointInTitle returns true if (x, y) — screen absolute coords — falls on the title area,
// excluding the close button area (last 2 chars).
func (n *Notification) pointInTitle(x, y int) bool {
	if n.titleHeight == 0 {
		return false
	}
	relX := x - n.bounds.x
	relY := y - n.bounds.y
	// Title text area excludes the last 2 chars (close button)
	return relY == 1 && relX >= 2 && relX < 2+n.contentWidth-2
}

// pointInCloseButton returns true if (x, y) falls on the close button (✕).
func (n *Notification) pointInCloseButton(x, y int) bool {
	if n.titleHeight == 0 {
		return false
	}
	relX := x - n.bounds.x
	relY := y - n.bounds.y
	// Close button is at the rightmost 2 chars of the title line
	return relY == 1 && relX >= 2+n.contentWidth-2 && relX < 2+n.contentWidth
}

// pointInContent returns true if (x, y) — screen absolute coords — falls on the message body area.
func (n *Notification) pointInContent(x, y int) bool {
	if len(n.contentLines) == 0 {
		return false
	}
	relX := x - n.bounds.x
	relY := y - n.bounds.y
	// Content starts at relY=1 (below top border), spans all content lines
	bodyRelY := 1
	maxY := bodyRelY + len(n.contentLines)
	return relY >= bodyRelY && relY < maxY && relX >= 2 && relX < 2+n.contentWidth
}

// contentCoordAt maps screen-absolute (x, y) to content line and display column.
func (n *Notification) contentCoordAt(x, y int) (int, int) {
	relX := x - n.bounds.x
	relY := y - n.bounds.y
	contentRelY := 1 // content (title + body) starts at relY=1
	row := clampInt(relY-contentRelY, 0, max(len(n.contentLines)-1, 0))
	col := clampInt(relX-2, 0, n.contentWidth)
	// If on the title line, clamp column to exclude the close button area
	if n.titleHeight > 0 && row == 0 {
		col = clampInt(col, 0, n.contentWidth-2)
	}
	return row, col
}

func (n *Notification) actionAt(x, y int) int {
	for i, bound := range n.actionBounds {
		if bound.contains(x, y) {
			return i
		}
	}
	return -1
}

func (n *Notification) pointInActionArea(x, y int) bool {
	return n.actionArea.w > 0 && n.actionArea.h > 0 && n.actionArea.contains(x, y)
}

type notificationMouseResult struct {
	consumed bool
	dismiss  bool
	rerender bool
	action   *NotificationActionResult
	cmd      tea.Cmd
}

// handleMouse processes mouse events for the notification.
func (n *Notification) handleMouse(msg tea.MouseMsg) notificationMouseResult {
	mouse := msg.Mouse()
	switch msg.(type) {
	case tea.MouseClickMsg:
		if n.pointInCloseButton(mouse.X, mouse.Y) {
			return notificationMouseResult{consumed: true, dismiss: true}
		}
		if mouse.Button == tea.MouseLeft {
			if actionIndex := n.actionAt(mouse.X, mouse.Y); actionIndex >= 0 {
				n.clearSelection()
				n.hoveredAction = actionIndex
				action := n.spec.Actions[actionIndex]
				return notificationMouseResult{
					consumed: true,
					dismiss:  true,
					rerender: true,
					action: &NotificationActionResult{
						NotificationID: n.id,
						ActionID:       action.ID,
					},
				}
			}
		}
		if n.pointInActionArea(mouse.X, mouse.Y) {
			return notificationMouseResult{consumed: true}
		}
		if n.pointInContent(mouse.X, mouse.Y) {
			line, col := n.contentCoordAt(mouse.X, mouse.Y)
			n.hoveredAction = -1
			n.selecting = true
			n.hasSelection = true
			n.selAnchorLine = line
			n.selAnchorCol = col
			n.selCursorLine = line
			n.selCursorCol = col
			return notificationMouseResult{consumed: true}
		}
		return notificationMouseResult{consumed: true, dismiss: true}

	case tea.MouseMotionMsg:
		if n.selecting {
			line, col := n.contentCoordAt(mouse.X, mouse.Y)
			n.selCursorLine = line
			n.selCursorCol = col
			return notificationMouseResult{consumed: true}
		}
		previousHovered := n.hoveredAction
		n.hoveredAction = n.actionAt(mouse.X, mouse.Y)
		result := notificationMouseResult{consumed: true, rerender: previousHovered != n.hoveredAction}
		switch {
		case n.hoveredAction >= 0, n.pointInCloseButton(mouse.X, mouse.Y):
			result.cmd = setMousePointer("pointer")
		case n.pointInContent(mouse.X, mouse.Y):
			result.cmd = setMousePointer("text")
		default:
			result.cmd = setMousePointer("default")
		}
		return result

	case tea.MouseReleaseMsg:
		if n.selecting {
			n.selecting = false
			return notificationMouseResult{consumed: true, cmd: n.finalizeSelection()}
		}
	}
	return notificationMouseResult{}
}

func (n *Notification) applySelectionHighlight(visibleLines []string, width int) []string {
	out := make([]string, len(visibleLines))
	for i, line := range visibleLines {
		left, right, ok := n.selectionRangeForLine(i, width)
		if !ok {
			out[i] = line
			continue
		}
		out[i] = highlightColumns(line, left, right)
	}
	return out
}

// ---- messages ----

// ShowNotificationMsg triggers displaying a notification. It can be sent from a
// goroutine via program.Send to remain race-free in the Update loop.
type ShowNotificationMsg struct {
	Spec NotificationSpec
}

// notificationExpireMsg signals that a notification's timeout elapsed.
type notificationExpireMsg struct {
	id NotificationID
}

// updateNotificationMsg updates the content of an existing notification.
type updateNotificationMsg struct {
	id   NotificationID
	spec NotificationSpec
}

// dismissNotificationMsg dismisses a specific notification early.
type dismissNotificationMsg struct {
	id NotificationID
}

// clearAllNotificationsMsg dismisses all visible notifications.
type clearAllNotificationsMsg struct{}
