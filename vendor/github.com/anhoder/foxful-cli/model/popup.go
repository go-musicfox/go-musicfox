package model

import (
	"fmt"
	"image/color"
	"os"
	"strings"
	"unicode"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/anhoder/foxful-cli/style"
	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
)

// ResizeHandle identifies which edge or corner of the popup is being resized.
type ResizeHandle int

const (
	ResizeNone ResizeHandle = iota
	// Corners
	ResizeBottomRight
	ResizeBottomLeft
	ResizeTopRight
	ResizeTopLeft
	// Edges
	ResizeRight
	ResizeLeft
	ResizeBottom
	ResizeTop
)

const (
	popupFrameHorizontalOverhead = 4 // rounded border + 1-cell padding on each side
	popupFrameVerticalOverhead   = 2 // rounded border top + bottom (no vertical padding)
	popupFrameInsetX             = 2 // left border plus left padding
	popupFrameInsetY             = 1 // top border only (content has no top padding)
)

// PopupAction is an explicit action rendered in a popup's action area.
type PopupAction struct {
	ID       string
	Label    string
	IsCancel bool
}

// PopupDismissCause describes how a popup was dismissed.
type PopupDismissCause uint8

const (
	PopupDismissAction PopupDismissCause = iota
	PopupDismissEscape
	PopupDismissOutsideClick
	// PopupDismissKey indicates dismissal by a configured close key other than
	// Escape (see PopupSpec.CloseKeys). PopupResult.Key holds the key name.
	PopupDismissKey
)

// PopupResult is passed to PopupSpec.OnResult after the popup is dismissed.
// ActionID is empty when dismissal did not trigger an action.
type PopupResult struct {
	ActionID string
	Cause    PopupDismissCause
	// Key is the key name (tea.Key.String, e.g. "esc", "q") that triggered a
	// key-based dismissal. Empty for action, mouse, or outside-click dismissal.
	Key string
}

// PopupAnchor controls where on the screen a popup appears.
type PopupAnchor int

const (
	AnchorCenter PopupAnchor = iota
	AnchorTopLeft
	AnchorTopCenter
	AnchorTopRight
	AnchorBottomLeft
	AnchorBottomCenter
	AnchorBottomRight
	AnchorCustom
)

// PopupSpec defines a popup before it is shown. Title and action labels are
// plain, single-line text. Content may be ANSI-styled; Popup preserves its
// foreground and text attributes while owning every content cell background.
type PopupSpec struct {
	Title     string
	Content   string
	Actions   []PopupAction
	MaxWidth  int // whole popup width, including border and padding; 0 = unlimited
	MaxHeight int // whole popup height, including border and padding; 0 = unlimited
	Anchor    PopupAnchor
	OffsetX   int
	OffsetY   int
	OnResult  func(PopupResult)

	// DisableResize disables the mouse-driven resize feature (indicator, corner
	// drag, and resize cursor hints). When true, the popup renders without the
	// ◢ indicator and ignores resize-corner mouse events. Default false.
	DisableResize bool

	// CloseKeys lists key names (as reported by tea.Key.String, e.g. "esc",
	// "q", "ctrl+c") that dismiss the popup as a cancel. When nil, it defaults
	// to {"esc"}. Pass a non-nil empty slice to disable key-based dismissal.
	// If "esc" is included, the first Esc press clears an active text selection
	// and the next dismisses.
	CloseKeys []string

	// DisableOutsideClick keeps the popup open when the user clicks outside it.
	// Default false (an outside left-click dismisses the popup).
	DisableOutsideClick bool
}

// Popup is an active modal dialog. Its configuration is immutable after
// construction; rendering and input state are private to the model package.
type Popup struct {
	title     string
	content   string
	actions   []PopupAction
	maxWidth  int
	maxHeight int
	anchor    PopupAnchor
	offsetX   int
	offsetY   int
	onResult  func(PopupResult)

	disableResize bool // when true, hide indicator and ignore resize mouse events

	// closeKeys is the set of key names that dismiss the popup as a cancel.
	closeKeys map[string]struct{}
	// disableOutsideClick keeps the popup open on outside clicks when true.
	disableOutsideClick bool

	// markdownMeta stores markdown rendering metadata for popups created via
	// NewMarkdownPopup. When non-nil, the popup can re-render its content when
	// the terminal background changes (light/dark mode switch).
	markdownMeta *markdownPopupMeta

	focusedAction int
	hoveredAction int
	result        *PopupResult

	scrollOffset        int
	totalContentLines   int
	visibleContentLines int

	dragging      bool
	dragMouseX    int
	dragMouseY    int
	dragStartOffX int
	dragStartOffY int

	// scrollDragging is true while the user drags the scrollbar thumb.
	scrollDragging bool

	// Resize state: true while the user drags a resize handle.
	resizing          bool
	resizeStartW      int // popup width when resize began
	resizeStartH      int // popup height when resize began
	resizeStartMouseX int
	resizeStartMouseY int
	resizeHandle      ResizeHandle

	// Text selection state (selecting/hasSelection/anchor+cursor coords and
	// contentLines) is provided by the embedded textSelection; its fields and
	// methods are promoted onto Popup.
	textSelection

	// pointerShape tracks the currently-set OSC 22 pointer shape ("" = default),
	// so hover changes only emit an escape when the shape actually changes.
	pointerShape string

	// Terminal dimensions, updated by App during render. Used during title-bar
	// drag to detect which boundaries have been reached so the cursor can show
	// the remaining allowable directions (e.g. "e-resize" when at the left edge).
	termWidth  int
	termHeight int

	// Rendering geometry captured on each render(), popup-relative unless the
	// name says otherwise. Used by handleMouse to hit-test the scrollbar and
	// content region. scrollbarRelX is -1 when the content is not scrollable.
	// contentLines is promoted from the embedded textSelection.
	bodyRelX      int
	bodyRelY      int
	visibleRows   int
	contentTextW  int
	scrollbarRelX int
	thumbRelY     int

	bounds       popupRect
	boundsSet    bool
	actionBounds []popupRect
}

// NewPopup validates spec and constructs a popup ready for App.ShowPopup.
func NewPopup(spec PopupSpec) (*Popup, error) {
	if !isPopupPlainText(spec.Title) {
		return nil, fmt.Errorf("popup title must be single-line plain text")
	}
	if spec.MaxWidth < 0 || spec.MaxHeight < 0 {
		return nil, fmt.Errorf("popup dimensions must not be negative")
	}
	if spec.MaxWidth > 0 && spec.MaxWidth <= popupFrameHorizontalOverhead {
		return nil, fmt.Errorf("popup max width must exceed frame overhead")
	}
	if spec.MaxHeight > 0 && spec.MaxHeight <= popupFrameVerticalOverhead {
		return nil, fmt.Errorf("popup max height must exceed frame overhead")
	}

	actions := make([]PopupAction, len(spec.Actions))
	copy(actions, spec.Actions)
	ids := make(map[string]struct{}, len(actions))
	cancelCount := 0
	for _, action := range actions {
		if action.ID == "" {
			return nil, fmt.Errorf("popup action id must not be empty")
		}
		if action.Label == "" {
			return nil, fmt.Errorf("popup action %q label must not be empty", action.ID)
		}
		if !isPopupPlainText(action.Label) {
			return nil, fmt.Errorf("popup action %q label must be single-line plain text", action.ID)
		}
		if _, exists := ids[action.ID]; exists {
			return nil, fmt.Errorf("popup action id %q is duplicated", action.ID)
		}
		ids[action.ID] = struct{}{}
		if action.IsCancel {
			cancelCount++
		}
	}
	if cancelCount > 1 {
		return nil, fmt.Errorf("popup may have at most one cancel action")
	}

	closeKeys := spec.CloseKeys
	if closeKeys == nil {
		closeKeys = []string{"esc"}
	}
	closeKeySet := make(map[string]struct{}, len(closeKeys))
	for _, key := range closeKeys {
		if key == "" {
			return nil, fmt.Errorf("popup close key must not be empty")
		}
		closeKeySet[key] = struct{}{}
	}

	return &Popup{
		title:               spec.Title,
		content:             spec.Content,
		actions:             actions,
		maxWidth:            spec.MaxWidth,
		maxHeight:           spec.MaxHeight,
		anchor:              spec.Anchor,
		offsetX:             spec.OffsetX,
		offsetY:             spec.OffsetY,
		onResult:            spec.OnResult,
		disableResize:       spec.DisableResize,
		closeKeys:           closeKeySet,
		disableOutsideClick: spec.DisableOutsideClick,
		hoveredAction:       -1,
	}, nil
}

// SetTermSize stores the terminal dimensions for boundary detection during drag.
// Called by App on each render cycle.
func (p *Popup) SetTermSize(w, h int) {
	p.termWidth = w
	p.termHeight = h
}

func isPopupPlainText(value string) bool {
	for _, r := range value {
		if r == '\n' || r == '\r' || r == '\x1b' || unicode.IsControl(r) {
			return false
		}
	}
	return true
}

func (p *Popup) update(msg tea.Msg) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return
	}

	if p.isContentScrollable() {
		switch keyMsg.String() {
		case "up", "k":
			p.clearSelection()
			p.scrollOffset = max(p.scrollOffset-1, 0)
			return
		case "down", "j":
			p.clearSelection()
			p.scrollOffset = min(p.scrollOffset+1, p.maxScrollOffset())
			return
		case "pgup":
			p.clearSelection()
			p.scrollOffset = max(p.scrollOffset-max(p.visibleContentLines/2, 1), 0)
			return
		case "pgdn":
			p.clearSelection()
			p.scrollOffset = min(p.scrollOffset+max(p.visibleContentLines/2, 1), p.maxScrollOffset())
			return
		case "home":
			p.clearSelection()
			p.scrollOffset = 0
			return
		case "end":
			p.clearSelection()
			p.scrollOffset = p.maxScrollOffset()
			return
		}
	}

	key := keyMsg.String()
	// Configured close keys dismiss the popup as a cancel. Esc additionally
	// clears an active text selection first (one press to clear, next to close).
	if _, isClose := p.closeKeys[key]; isClose {
		if key == "esc" && p.hasSelection {
			p.clearSelection()
			return
		}
		cause := PopupDismissKey
		if key == "esc" {
			cause = PopupDismissEscape
		}
		p.dismissCancelKey(cause, key)
		return
	}

	switch key {
	case "enter":
		if len(p.actions) > 0 {
			p.dismissAction(p.focusedAction, PopupDismissAction)
		}
	case "tab", "right", "l", "L":
		if len(p.actions) > 1 {
			p.focusedAction = (p.focusedAction + 1) % len(p.actions)
		}
	case "left", "h", "H":
		if len(p.actions) > 1 {
			p.focusedAction = (p.focusedAction - 1 + len(p.actions)) % len(p.actions)
		}
	}
}

func (p *Popup) dismissed() bool {
	return p.result != nil
}

func (p *Popup) consumeResult() *PopupResult {
	result := p.result
	p.result = nil
	return result
}

func (p *Popup) dismissAction(index int, cause PopupDismissCause) {
	p.dismissActionKey(index, cause, "")
}

func (p *Popup) dismissActionKey(index int, cause PopupDismissCause, key string) {
	if p.result != nil || index < 0 || index >= len(p.actions) {
		return
	}
	p.result = &PopupResult{ActionID: p.actions[index].ID, Cause: cause, Key: key}
}

func (p *Popup) dismissCancel(cause PopupDismissCause) {
	p.dismissCancelKey(cause, "")
}

func (p *Popup) dismissCancelKey(cause PopupDismissCause, key string) {
	for i, action := range p.actions {
		if action.IsCancel {
			p.dismissActionKey(i, cause, key)
			return
		}
	}
	if p.result == nil {
		p.result = &PopupResult{Cause: cause, Key: key}
	}
}

// dismissOutside implements Modal.dismissOutside for Popup. It returns false
// without dismissing when outside-click dismissal is disabled.
func (p *Popup) dismissOutside() bool {
	if p.disableOutsideClick {
		return false
	}
	p.dismissCancel(PopupDismissOutsideClick)
	return true
}

// rerenderMarkdown re-renders the popup content if it was created via NewMarkdownPopup
// and has markdown metadata. Used when the terminal background changes to pick up the
// new auto-detected glamour style. Returns true if content was re-rendered.
func (p *Popup) rerenderMarkdown() bool {
	if p.markdownMeta == nil {
		return false
	}

	md := NewMarkdown(
		p.markdownMeta.content,
		WithMarkdownDarkStyle(p.markdownMeta.darkStyle),
		WithMarkdownLightStyle(p.markdownMeta.lightStyle),
		WithMarkdownEmoji(p.markdownMeta.emoji),
		WithMarkdownWordWrap(p.markdownMeta.wrapWidth),
	)

	rendered, err := md.RenderToString(p.markdownMeta.wrapWidth)
	if err != nil {
		return false
	}

	// Update content and reset scroll/selection state
	p.content = rendered
	p.contentLines = nil // force rebuild in renderContent
	p.scrollOffset = 0
	p.clearSelection()

	return true
}

// complete implements Modal.complete for Popup.
// Invokes the onResult callback if present, then returns (nil, nil).
func (p *Popup) complete(app *App) (Page, tea.Cmd) {
	result := p.consumeResult()
	if result != nil && p.onResult != nil {
		p.onResult(*result)
	}
	return nil, nil
}

// allowsRightClickPassthrough implements Modal.allowsRightClickPassthrough for Popup.
// Returns false — traditional modal popups don't allow right-click passthrough.
func (p *Popup) allowsRightClickPassthrough() bool {
	return false
}

type popupRect struct {
	x int
	y int
	w int
	h int
}

func (r popupRect) contains(x, y int) bool {
	return x >= r.x && x < r.x+r.w && y >= r.y && y < r.y+r.h
}

type popupRender struct {
	content      string
	actionBounds []popupRect // relative to the popup's top-left corner
}

func (p *Popup) render(styles style.PopupStyleSet) popupRender {
	maxContentWidth := p.maxContentWidth()

	actions := p.renderActions(styles, maxContentWidth)

	// Build & cache normalized content lines. Rebuild only after invalidation
	// (content or theme change); scrolling and resizing reuse the cache. This
	// eliminates per-frame ANSI parsing (normalizePopupSurface) + string split,
	// which caused scroll lag on long, ANSI-rich (markdown) content.
	if p.contentLines == nil && p.content != "" {
		normalized := normalizePopupSurface(styles.Content.Render(p.content), styles.Surface)
		p.contentLines = strings.Split(normalized, "\n")
	}
	contentLines := p.contentLines
	p.totalContentLines = len(contentLines)

	// The action block, when present, is preceded by one blank spacer row.
	actionsOverhead := actions.height
	if actions.height > 0 {
		actionsOverhead++
	}
	visibleHeight := len(contentLines)
	if len(contentLines) > 0 && p.maxHeight > 0 {
		visibleHeight = max(1, p.maxHeight-popupFrameVerticalOverhead-actionsOverhead)
		visibleHeight = min(visibleHeight, len(contentLines))
	}
	p.visibleContentLines = visibleHeight
	p.scrollOffset = min(p.scrollOffset, p.maxScrollOffset())

	scrolling := len(contentLines) > visibleHeight

	visibleLines := contentLines
	if visibleHeight < len(contentLines) {
		visibleLines = contentLines[p.scrollOffset : p.scrollOffset+visibleHeight]
	}
	if p.hasSelection {
		visibleLines = p.applySelectionHighlight(visibleLines)
	}
	bodyStr, contentWidth, scrollbarLines := renderPopupBody(visibleLines, scrolling, p.scrollOffset, p.maxScrollOffset(), maxContentWidth, styles)

	// Compute bodyWidth: content width + scrollbar overhead when scrolling.
	bodyWidth := contentWidth
	if scrolling {
		bodyWidth = contentWidth + 2 // content + 1-char gap + 1-char scrollbar
	}

	bodyHeight := textHeight(bodyStr)
	// Capture body/scrollbar geometry (popup-relative) for mouse hit-testing.
	p.bodyRelX = popupFrameInsetX
	p.bodyRelY = popupFrameInsetY
	p.visibleRows = visibleHeight
	if scrolling {
		p.contentTextW = contentWidth
		thumbLine := 0
		if p.maxScrollOffset() > 0 && visibleHeight > 0 {
			thumbLine = p.scrollOffset * (visibleHeight - 1) / p.maxScrollOffset()
		}
		p.thumbRelY = popupFrameInsetY + thumbLine
	} else {
		p.contentTextW = contentWidth
		p.scrollbarRelX = -1
		p.thumbRelY = -1
	}

	innerWidth := max(bodyWidth, actions.width)
	if p.maxWidth > 0 {
		innerWidth = max(innerWidth, maxContentWidth)
	}

	// Scrollbar always pinned to the right edge of the body area.
	if scrolling {
		p.scrollbarRelX = popupFrameInsetX + innerWidth - 1
	}

	blocks := make([]string, 0, 3)
	if bodyStr != "" {
		if scrolling {
			// Compose content + gap (surface bg) + scrollbar at right edge.
			bodyLines := strings.Split(bodyStr, "\n")
			gap := innerWidth - contentWidth - 1 // -1 for the scrollbar column
			if gap < 1 {
				gap = 1
			}
			gapStyle := lipgloss.NewStyle().Width(gap).Background(styles.Surface)
			composed := make([]string, len(bodyLines))
			for i := range bodyLines {
				composed[i] = bodyLines[i] + gapStyle.Render("") + scrollbarLines[i]
			}
			bodyStr = lipgloss.NewStyle().Width(innerWidth).Background(styles.Surface).Render(strings.Join(composed, "\n"))
		} else {
			bodyStr = lipgloss.NewStyle().Width(innerWidth).Background(styles.Surface).Render(bodyStr)
		}
		blocks = append(blocks, bodyStr)
	}
	// One blank line separates the content from the action buttons.
	spacerHeight := 0
	if bodyStr != "" && actions.content != "" {
		blocks = append(blocks, lipgloss.NewStyle().Width(innerWidth).Background(styles.Surface).Render(""))
		spacerHeight = 1
	}
	if actions.content != "" {
		blocks = append(blocks, lipgloss.NewStyle().
			Width(innerWidth).
			Align(lipgloss.Center).
			Background(styles.Surface).
			Render(actions.content))
	}
	inner := lipgloss.JoinVertical(lipgloss.Left, blocks...)

	// Pad inner content to fill maxHeight when set (for resize support).
	if p.maxHeight > 0 {
		targetInnerH := p.maxHeight - popupFrameVerticalOverhead
		currentInnerH := lipgloss.Height(inner)
		if currentInnerH < targetInnerH {
			blankLine := lipgloss.NewStyle().Width(innerWidth).Background(styles.Surface).Render("")
			for range targetInnerH - currentInnerH {
				if inner == "" {
					inner = blankLine
				} else {
					inner += "\n" + blankLine
				}
			}
		}
	}

	framed := styles.Frame.Render(inner)

	if p.title != "" {
		framed = embedTitleInTopBorder(framed, p.title, styles)
	}

	// Add resize indicator in bottom-right corner (if resize enabled)
	if !p.disableResize {
		framed = addResizeIndicator(framed, styles.Surface)
	}

	actionY := popupFrameInsetY + bodyHeight + spacerHeight
	actionX := popupFrameInsetX + (innerWidth-actions.width)/2
	actionBounds := make([]popupRect, len(actions.bounds))
	for i, bound := range actions.bounds {
		actionBounds[i] = popupRect{
			x: actionX + bound.x,
			y: actionY + bound.y,
			w: bound.w,
			h: bound.h,
		}
	}

	return popupRender{
		content:      framed,
		actionBounds: actionBounds,
	}
}

func embedTitleInTopBorder(framed, title string, styles style.PopupStyleSet) string {
	screen := popupStyledScreen(framed)
	if len(screen.Lines) == 0 {
		return framed
	}

	titleRunes := []rune(title)
	frameWidth := len(screen.Lines[0])
	availableWidth := frameWidth - 4
	if availableWidth < 1 || len(titleRunes) == 0 {
		return framed
	}

	// Truncate by display width, not rune count: CJK glyphs occupy two columns,
	// so a rune-count budget would let a wide title overrun the top border.
	displayTitle := titleRunes
	titleWidth := ansi.StringWidth(string(displayTitle))
	if titleWidth > availableWidth {
		truncated := make([]rune, 0, len(displayTitle))
		used := 0
		for _, r := range displayTitle {
			rw := ansi.StringWidth(string(r))
			if used+rw > availableWidth-1 { // reserve 1 column for the ellipsis
				break
			}
			truncated = append(truncated, r)
			used += rw
		}
		displayTitle = append(truncated, '…')
		titleWidth = used + 1
	}

	// Center by display width and advance x by each rune's display width so wide
	// glyphs land on their own two columns (Line.Set fills the placeholder cell).
	// Setting Width per rune keeps the top border exactly frameWidth columns wide;
	// hardcoding Width:1 previously pushed the top-right corner out by the sum of
	// the title's extra CJK columns, producing a detached corner / right-edge ghost.
	titleStart := (frameWidth - titleWidth) / 2
	titleStyle := styles.Title.GetForeground()
	x := titleStart
	for _, r := range displayTitle {
		rw := ansi.StringWidth(string(r))
		if rw < 1 {
			rw = 1
		}
		if x >= 1 && x+rw <= frameWidth-1 {
			cell := &uv.Cell{
				Content: string(r),
				Width:   rw,
				Style: uv.Style{
					Fg: titleStyle,
					Bg: styles.Surface,
				},
			}
			screen.Lines[0].Set(x, cell)
		}
		x += rw
	}

	return screen.Render()
}

// addResizeIndicator places a resize indicator near the bottom-right corner.
// It places the indicator inside the popup, not on the border itself.
func addResizeIndicator(framed string, surface color.Color) string {
	screen := popupStyledScreen(framed)
	if len(screen.Lines) == 0 {
		return framed
	}

	lastLine := len(screen.Lines) - 1
	if lastLine < 1 {
		return framed
	}

	// Place indicator on the second-to-last column of the second-to-last line
	// This puts it inside the popup, next to the border
	targetLine := lastLine - 1
	targetCol := len(screen.Lines[targetLine]) - 2

	if targetCol >= 0 && targetLine >= 0 {
		// Use ◢ (lower right triangle) as resize indicator
		// Place it inside the popup with subtle styling
		cell := &uv.Cell{
			Content: "◢",
			Width:   1,
			Style: uv.Style{
				Fg: surface, // Subtle color
				Bg: surface,
			},
		}
		screen.Lines[targetLine].Set(targetCol, cell)
	}

	return screen.Render()
}

func (p *Popup) maxContentWidth() int {
	if p.maxWidth == 0 {
		return 0
	}
	return max(1, p.maxWidth-popupFrameHorizontalOverhead)
}

type popupActionsRender struct {
	content string
	width   int
	height  int
	bounds  []popupRect // relative to the action block
}

func (p *Popup) renderActions(styles style.PopupStyleSet, maxWidth int) popupActionsRender {
	if len(p.actions) == 0 {
		return popupActionsRender{}
	}

	type actionRow struct {
		content strings.Builder
		width   int
		height  int
	}

	rows := []actionRow{{}}
	bounds := make([]popupRect, len(p.actions))
	for i, action := range p.actions {
		buttonStyle := styles.Action
		switch {
		case i == p.focusedAction && i == p.hoveredAction:
			buttonStyle = styles.ActionFocused.Underline(true)
		case i == p.focusedAction:
			buttonStyle = styles.ActionFocused
		case i == p.hoveredAction:
			buttonStyle = styles.ActionHover
		}
		button := buttonStyle.Render(action.Label)
		buttonWidth := lipgloss.Width(button)
		if maxWidth > 0 && buttonWidth > maxWidth {
			button = lipgloss.NewStyle().MaxWidth(maxWidth).Render(button)
			buttonWidth = lipgloss.Width(button)
		}

		row := &rows[len(rows)-1]
		gap := 0
		if row.width > 0 {
			gap = 1
		}
		if maxWidth > 0 && row.width > 0 && row.width+gap+buttonWidth > maxWidth {
			rows = append(rows, actionRow{})
			row = &rows[len(rows)-1]
			gap = 0
		}
		if gap > 0 {
			row.content.WriteString(lipgloss.NewStyle().Background(styles.Surface).Render(" "))
		}
		x := row.width + gap
		row.content.WriteString(button)
		row.width += gap + buttonWidth
		row.height = max(row.height, textHeight(button))
		bounds[i] = popupRect{x: x, y: len(rows) - 1, w: buttonWidth, h: textHeight(button)}
	}

	width := 0
	for _, row := range rows {
		width = max(width, row.width)
	}
	rowTexts := make([]string, len(rows))
	height := 0
	for i, row := range rows {
		rowTexts[i] = lipgloss.NewStyle().Width(width).Background(styles.Surface).Render(row.content.String())
		height += max(row.height, 1)
	}
	return popupActionsRender{
		content: strings.Join(rowTexts, "\n"),
		width:   width,
		height:  height,
		bounds:  bounds,
	}
}

func renderPopupBody(lines []string, scrolling bool, offset, maxOffset, maxWidth int, styles style.PopupStyleSet) (string, int, []string) {
	if len(lines) == 0 {
		return "", 0, nil
	}
	if !scrolling {
		if maxWidth == 0 {
			return strings.Join(lines, "\n"), widestLine(lines), nil
		}
		output := make([]string, len(lines))
		for i, line := range lines {
			output[i] = lipgloss.NewStyle().MaxWidth(maxWidth).Render(line)
		}
		return strings.Join(output, "\n"), widestLine(output), nil
	}

	bodyWidth := widestLine(lines)
	if maxWidth > 0 {
		bodyWidth = min(bodyWidth, max(maxWidth-2, 1))
	}

	thumbLine := 0
	if maxOffset > 0 {
		thumbLine = offset * (len(lines) - 1) / maxOffset
	}
	output := make([]string, len(lines))
	scrollbarLines := make([]string, len(lines))
	for i, line := range lines {
		line = lipgloss.NewStyle().MaxWidth(bodyWidth).Render(line)
		line = lipgloss.NewStyle().Width(bodyWidth).Render(line)
		output[i] = line

		scrollbar := styles.ScrollTrack.Render("│")
		if i == thumbLine {
			scrollbar = styles.ScrollThumb.Render("█")
		}
		scrollbarLines[i] = scrollbar
	}
	return strings.Join(output, "\n"), bodyWidth, scrollbarLines
}

func widestLine(lines []string) int {
	width := 0
	for _, line := range lines {
		width = max(width, lipgloss.Width(line))
	}
	return width
}

func textHeight(content string) int {
	if content == "" {
		return 0
	}
	return lipgloss.Height(content)
}

// normalizePopupSurface preserves rendered content's foreground, text
// attributes, hyperlinks, and grapheme widths while replacing every cell
// background. Reverse video is removed because it is an implicit background.
func normalizePopupSurface(content string, surface color.Color) string {
	if content == "" {
		return ""
	}

	screen := popupStyledScreen(content)
	for y := range screen.Lines {
		for x := range screen.Lines[y] {
			cell := screen.CellAt(x, y)
			if cell == nil || cell.IsZero() {
				continue
			}
			cell.Style.Bg = surface
			cell.Style.Attrs &^= uv.AttrReverse
		}
	}
	return screen.Render()
}

func popupStyledScreen(content string) uv.ScreenBuffer {
	width := max(lipgloss.Width(content), 1)
	height := strings.Count(content, "\n") + 1
	screen := uv.NewScreenBuffer(width, height)
	screen.Method = ansi.GraphemeWidth
	uv.NewStyledString(content).Draw(screen, screen.Bounds())
	return screen
}

func (p *Popup) setBounds(x, y, w, h int, actionBounds []popupRect) {
	p.bounds = popupRect{x: x, y: y, w: w, h: h}
	p.boundsSet = true
	p.actionBounds = make([]popupRect, len(actionBounds))
	for i, bound := range actionBounds {
		p.actionBounds[i] = popupRect{
			x: x + bound.x,
			y: y + bound.y,
			w: bound.w,
			h: bound.h,
		}
	}
}

func (p *Popup) handleMouse(msg tea.MouseMsg) (bool, tea.Cmd) {
	mouse := msg.Mouse()
	_, isClick := msg.(tea.MouseClickMsg)
	_, isRelease := msg.(tea.MouseReleaseMsg)

	// Active drags take priority; they are driven by motion/release and may
	// run outside the popup bounds.
	if p.dragging {
		if isRelease {
			p.dragging = false
			return true, nil
		}
		p.offsetX = p.dragStartOffX + mouse.X - p.dragMouseX
		p.offsetY = p.dragStartOffY + mouse.Y - p.dragMouseY
		return true, p.pointerCmd(p.desiredDragPointer())
	}
	if p.scrollDragging {
		if isRelease {
			p.scrollDragging = false
			return true, nil
		}
		p.scrollToThumbRow(mouse.Y - p.bounds.y)
		return true, nil
	}
	if p.selecting {
		if isRelease {
			p.selecting = false
			return true, p.finalizeSelection()
		}
		p.updateSelectionCursor(mouse)
		return true, nil
	}
	if p.resizing {
		if isRelease {
			p.resizing = false
			return true, nil
		}
		// Calculate deltas from resize start position
		deltaX := mouse.X - p.resizeStartMouseX
		deltaY := mouse.Y - p.resizeStartMouseY
		minW := popupFrameHorizontalOverhead + 10 // enough for buttons
		minH := popupFrameVerticalOverhead + 3    // title + 1 line + buttons

		// Apply resize based on handle type
		switch p.resizeHandle {
		// Corners — change both width and height
		case ResizeBottomRight:
			p.maxWidth = max(p.resizeStartW+deltaX, minW)
			p.maxHeight = max(p.resizeStartH+deltaY, minH)
		case ResizeBottomLeft:
			p.maxWidth = max(p.resizeStartW-deltaX, minW)
			p.maxHeight = max(p.resizeStartH+deltaY, minH)
		case ResizeTopRight:
			p.maxWidth = max(p.resizeStartW+deltaX, minW)
			p.maxHeight = max(p.resizeStartH-deltaY, minH)
		case ResizeTopLeft:
			p.maxWidth = max(p.resizeStartW-deltaX, minW)
			p.maxHeight = max(p.resizeStartH-deltaY, minH)
		// Edges — change only one dimension
		case ResizeRight:
			p.maxWidth = max(p.resizeStartW+deltaX, minW)
		case ResizeLeft:
			p.maxWidth = max(p.resizeStartW-deltaX, minW)
		case ResizeBottom:
			p.maxHeight = max(p.resizeStartH+deltaY, minH)
		case ResizeTop:
			p.maxHeight = max(p.resizeStartH-deltaY, minH)
		}

		return true, nil
	}

	// Update hovered action + desired pointer shape.
	if p.boundsSet && p.bounds.contains(mouse.X, mouse.Y) {
		p.hoveredAction = p.actionAt(mouse.X, mouse.Y)
	} else {
		p.hoveredAction = -1
	}
	hoverCmd := p.pointerCmd(p.desiredPointer(mouse))

	if !p.boundsSet {
		return true, hoverCmd
	}
	if !p.bounds.contains(mouse.X, mouse.Y) {
		return false, hoverCmd
	}

	if isClick && mouse.Button == tea.MouseLeft {
		return p.handleLeftClick(mouse, hoverCmd)
	}

	if mouse.Button == tea.MouseWheelDown {
		if p.isContentScrollable() {
			p.scrollOffset = min(p.scrollOffset+1, p.maxScrollOffset())
			return true, hoverCmd
		}
		if len(p.actions) > 1 {
			p.focusedAction = (p.focusedAction + 1) % len(p.actions)
			return true, hoverCmd
		}
	}
	if mouse.Button == tea.MouseWheelUp {
		if p.isContentScrollable() {
			p.scrollOffset = max(p.scrollOffset-1, 0)
			return true, hoverCmd
		}
		if len(p.actions) > 1 {
			p.focusedAction = (p.focusedAction - 1 + len(p.actions)) % len(p.actions)
			return true, hoverCmd
		}
	}

	return true, hoverCmd
}

// handleLeftClick routes a left mouse-down inside the popup: title-bar drag,
// action activation, scrollbar thumb drag / track jump, or the start of a text
// selection in the content region.
func (p *Popup) handleLeftClick(mouse tea.Mouse, hoverCmd tea.Cmd) (bool, tea.Cmd) {
	relX := mouse.X - p.bounds.x
	relY := mouse.Y - p.bounds.y

	// Check for resize handle (highest priority for edge/corner)
	if h := p.resizeHandleAt(mouse); h != ResizeNone {
		p.clearSelection()
		p.resizing = true
		p.resizeHandle = h
		p.resizeStartW = p.bounds.w
		p.resizeStartH = p.bounds.h
		p.resizeStartMouseX = mouse.X
		p.resizeStartMouseY = mouse.Y
		return true, hoverCmd
	}
	if mouse.Y == p.bounds.y {
		p.clearSelection()
		p.dragging = true
		p.dragMouseX = mouse.X
		p.dragMouseY = mouse.Y
		p.dragStartOffX = p.offsetX
		p.dragStartOffY = p.offsetY
		return true, hoverCmd
	}

	if action := p.actionAt(mouse.X, mouse.Y); action >= 0 {
		p.clearSelection()
		p.focusedAction = action
		p.dismissAction(action, PopupDismissAction)
		return true, hoverCmd
	}

	if p.scrollbarRelX >= 0 && relX == p.scrollbarRelX {
		p.clearSelection()
		if relY == p.thumbRelY {
			p.scrollDragging = true
		} else {
			p.scrollToThumbRow(relY)
		}
		return true, hoverCmd
	}

	if p.pointInContent(relX, relY) {
		line, col := p.contentCoordAt(relX, relY)
		p.selecting = true
		p.hasSelection = true
		p.selAnchorLine, p.selAnchorCol = line, col
		p.selCursorLine, p.selCursorCol = line, col
		return true, hoverCmd
	}

	p.clearSelection()
	return true, hoverCmd
}

// writeMousePointer synchronously emits the OSC 22 escape sequence that sets
// the terminal mouse pointer shape. An empty shape resets to the terminal
// default. This is the single low-level writer shared by the App methods and
// the popup/notification pointer commands.
func writeMousePointer(shape string) {
	os.Stdout.WriteString("\x1b]22;" + shape + "\x1b\\")
}

// setMousePointer returns a tea.Cmd that emits the OSC 22 pointer-shape escape
// sequence when executed by the bubbletea runtime.
func setMousePointer(shape string) tea.Cmd {
	return func() tea.Msg {
		writeMousePointer(shape)
		return nil
	}
}

// pointerCmd returns a command to switch the OSC 22 pointer shape, or nil when
// the shape is unchanged. Empty shape resets to the terminal default.
func (p *Popup) pointerCmd(shape string) tea.Cmd {
	if shape == p.pointerShape {
		return nil
	}
	p.pointerShape = shape
	if shape == "" {
		return setMousePointer("default")
	}
	return setMousePointer(shape)
}

// desiredPointer picks the pointer shape for the current mouse position:
// "pointer" over actions and the entire scrollbar (track + thumb), "text" over
// selectable content, "" (default) elsewhere.
func (p *Popup) desiredPointer(mouse tea.Mouse) string {
	if !p.boundsSet || !p.bounds.contains(mouse.X, mouse.Y) {
		return ""
	}

	relX := mouse.X - p.bounds.x
	relY := mouse.Y - p.bounds.y

	// Check resize handle (corners have highest priority, then edges)
	if h := p.resizeHandleAt(mouse); h != ResizeNone {
		switch h {
		case ResizeBottomRight:
			return "nwse-resize" // ↖↘ diagonal (CSS name, Ghostty/iTerm2)
		case ResizeTopLeft:
			return "nwse-resize" // ↖↘ diagonal
		case ResizeBottomLeft:
			return "nesw-resize" // ↗↙ diagonal (CSS name, Ghostty/iTerm2)
		case ResizeTopRight:
			return "nesw-resize" // ↗↙ diagonal
		case ResizeRight, ResizeLeft:
			return "ew-resize" // ↔
		case ResizeBottom, ResizeTop:
			return "ns-resize" // ↕
		}
	}

	// Title bar area — cursor indicates draggable (grab)
	if relY == 0 {
		return "grab"
	}

	if p.actionAt(mouse.X, mouse.Y) >= 0 {
		return "pointer"
	}

	// Hover over entire scrollbar column (track + thumb) shows pointer
	if p.scrollbarRelX >= 0 && relX == p.scrollbarRelX {
		scrollbarTop := p.bodyRelY
		scrollbarBottom := p.bodyRelY + p.visibleRows
		if relY >= scrollbarTop && relY < scrollbarBottom {
			return "pointer"
		}
	}
	if p.pointInContent(relX, relY) {
		return "text"
	}
	return ""
}

// desiredDragPointer returns the cursor shape during title-bar drag, indicating
// which directions remain movable when the popup reaches terminal boundaries.
// Uses single-direction resize cursors so the cursor arrow points toward the
// direction(s) the user can still drag.
func (p *Popup) desiredDragPointer() string {
	if p.termWidth <= 0 || p.termHeight <= 0 || !p.boundsSet {
		return "grabbing"
	}
	ox, oy := p.anchorOrigin(p.termWidth, p.termHeight, p.bounds.w, p.bounds.h)
	x := ox + p.offsetX
	y := oy + p.offsetY

	atLeft := x <= 0
	atRight := x+p.bounds.w >= p.termWidth
	atTop := y <= 0
	atBottom := y+p.bounds.h >= p.termHeight

	// No boundaries hit — free movement in all directions.
	if !atLeft && !atRight && !atTop && !atBottom {
		return "grabbing"
	}

	switch {
	// Single direction blocked — cursor points to the available side.
	case atLeft && !atRight && !atTop && !atBottom:
		return "e-resize" // only right
	case atRight && !atLeft && !atTop && !atBottom:
		return "w-resize" // only left
	case atTop && !atBottom && !atLeft && !atRight:
		return "s-resize" // only down
	case atBottom && !atTop && !atLeft && !atRight:
		return "n-resize" // only up

	// Corners — two edges blocked, cursor shows diagonal toward the free quadrant.
	case atTop && atLeft && !atRight && !atBottom:
		return "se-resize" // ↘ right + down
	case atTop && atRight && !atLeft && !atBottom:
		return "sw-resize" // ↙ left + down
	case atBottom && atLeft && !atRight && !atTop:
		return "ne-resize" // ↗ right + up
	case atBottom && atRight && !atLeft && !atTop:
		return "nw-resize" // ↖ left + up

	// Horizontal fully blocked — only vertical movement possible.
	case atLeft && atRight && !atTop && !atBottom:
		return "ns-resize"
	// Vertical fully blocked — only horizontal movement possible.
	case atTop && atBottom && !atLeft && !atRight:
		return "ew-resize"

	// Three or four directions blocked (popup pinned or nearly so).
	default:
		return "not-allowed"
	}
}

// resizeHandleAt reports which resize handle (corner or edge) the mouse is over.
func (p *Popup) resizeHandleAt(mouse tea.Mouse) ResizeHandle {
	if !p.boundsSet || p.disableResize {
		return ResizeNone
	}

	x, y := mouse.X, mouse.Y
	relX := x - p.bounds.x
	relY := y - p.bounds.y

	onLeft := relX == 0
	onRight := relX == p.bounds.w-1
	onBottom := relY == p.bounds.h-1

	// Corners (top corners excluded — reserved for title-bar drag)
	if onRight && onBottom {
		return ResizeBottomRight
	}
	if onLeft && onBottom {
		return ResizeBottomLeft
	}

	// Edges (top edge excluded — reserved for title-bar drag).
	// Left/right edge also excludes the top row so top corners trigger drag.
	if onRight && relY != 0 {
		return ResizeRight
	}
	if onLeft && relY != 0 {
		return ResizeLeft
	}
	if onBottom {
		return ResizeBottom
	}

	return ResizeNone
}

// pointInContent reports whether a popup-relative coordinate lands inside the
// selectable text region (excluding the scrollbar and its gap column).
func (p *Popup) pointInContent(relX, relY int) bool {
	if len(p.contentLines) == 0 || p.visibleRows == 0 {
		return false
	}
	if relY < p.bodyRelY || relY >= p.bodyRelY+p.visibleRows {
		return false
	}
	return relX >= p.bodyRelX && relX < p.bodyRelX+p.contentTextW
}

// contentCoordAt maps a popup-relative coordinate to a (line, column) in
// full-content space, clamped to valid ranges.
func (p *Popup) contentCoordAt(relX, relY int) (int, int) {
	row := clampInt(relY-p.bodyRelY, 0, max(p.visibleRows-1, 0))
	line := clampInt(p.scrollOffset+row, 0, max(len(p.contentLines)-1, 0))
	col := clampInt(relX-p.bodyRelX, 0, p.contentTextW)
	return line, col
}

// scrollToThumbRow sets the scroll offset so the thumb sits at the popup-relative
// row derived from relY.
func (p *Popup) scrollToThumbRow(relY int) {
	maxOff := p.maxScrollOffset()
	if maxOff == 0 || p.visibleRows <= 1 {
		return
	}
	row := clampInt(relY-p.bodyRelY, 0, p.visibleRows-1)
	span := p.visibleRows - 1
	offset := (row*maxOff + span/2) / span
	p.scrollOffset = clampInt(offset, 0, maxOff)
}

// updateSelectionCursor extends the active selection to the mouse position,
// auto-scrolling when the drag reaches beyond the visible content edges.
func (p *Popup) updateSelectionCursor(mouse tea.Mouse) {
	relY := mouse.Y - p.bounds.y
	if relY < p.bodyRelY {
		p.scrollOffset = max(p.scrollOffset-1, 0)
	} else if relY >= p.bodyRelY+p.visibleRows {
		p.scrollOffset = min(p.scrollOffset+1, p.maxScrollOffset())
	}
	line, col := p.contentCoordAt(mouse.X-p.bounds.x, relY)
	p.selCursorLine, p.selCursorCol = line, col
}

// applySelectionHighlight returns a copy of the visible lines with the selected
// column ranges rendered in reverse video. visibleLines[k] maps to full-content
// line scrollOffset+k.
func (p *Popup) applySelectionHighlight(visibleLines []string) []string {
	out := make([]string, len(visibleLines))
	for k, line := range visibleLines {
		i := p.scrollOffset + k
		width := lipgloss.Width(line)
		left, right, ok := p.selectionRangeForLine(i, width)
		if !ok {
			out[k] = line
			continue
		}
		out[k] = highlightColumns(line, left, right)
	}
	return out
}

// highlightColumns reverse-videos the display columns [left, right) of a single
// styled line, preserving all other cell styling.
func highlightColumns(line string, left, right int) string {
	screen := popupStyledScreen(line)
	if len(screen.Lines) == 0 {
		return line
	}
	width := len(screen.Lines[0])
	right = min(right, width)
	for x := left; x < right; x++ {
		cell := screen.CellAt(x, 0)
		if cell == nil {
			continue
		}
		cell.Style.Attrs |= uv.AttrReverse
	}
	return screen.Render()
}

func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func (p *Popup) actionAt(x, y int) int {
	for i, bound := range p.actionBounds {
		if bound.contains(x, y) {
			return i
		}
	}
	return -1
}

func (p *Popup) isContentScrollable() bool {
	return p.totalContentLines > p.visibleContentLines
}

func (p *Popup) maxScrollOffset() int {
	return max(p.totalContentLines-p.visibleContentLines, 0)
}

func (p *Popup) computePosition(termW, termH, popupW, popupH int) (int, int) {
	ox, oy := p.anchorOrigin(termW, termH, popupW, popupH)
	x := ox + p.offsetX
	y := oy + p.offsetY

	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}
	if x+popupW > termW {
		x = termW - popupW
	}
	if y+popupH > termH {
		y = termH - popupH
	}
	return x, y
}

func (p *Popup) anchorOrigin(termW, termH, popupW, popupH int) (int, int) {
	switch p.anchor {
	case AnchorTopLeft:
		return 0, 0
	case AnchorTopCenter:
		return (termW - popupW) / 2, 0
	case AnchorTopRight:
		return termW - popupW, 0
	case AnchorBottomLeft:
		return 0, termH - popupH
	case AnchorBottomCenter:
		return (termW - popupW) / 2, termH - popupH
	case AnchorBottomRight:
		return termW - popupW, termH - popupH
	case AnchorCustom:
		return 0, 0
	default:
		return (termW - popupW) / 2, (termH - popupH) / 3
	}
}
