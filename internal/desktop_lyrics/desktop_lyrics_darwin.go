//go:build darwin

package desktop_lyrics

import (
	"encoding/json"
	"log/slog"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf16"

	"github.com/ebitengine/purego/objc"
	"github.com/rivo/uniseg"

	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/macdriver"
	"github.com/go-musicfox/go-musicfox/internal/macdriver/cocoa"
	"github.com/go-musicfox/go-musicfox/internal/macdriver/core"
	"github.com/go-musicfox/go-musicfox/internal/storage"
)

// lerp performs linear interpolation between a and b by factor t (0.0-1.0).
func lerp(a, b, t float64) float64 {
	return a + (b-a)*t
}

// utf16Length returns the number of UTF-16 code units in text for NSRange.
func utf16Length(text string) int {
	units := 0
	for _, r := range text {
		if n := utf16.RuneLen(r); n > 0 {
			units += n
		} else {
			units++
		}
	}
	return units
}

// parseHexRGB parses a 6-char hex string into RGB (0.0-1.0).
func parseHexRGB(hex string) (r, g, b float64) {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) < 6 {
		return 1, 1, 1
	}
	rVal, _ := strconv.ParseInt(hex[0:2], 16, 64)
	gVal, _ := strconv.ParseInt(hex[2:4], 16, 64)
	bVal, _ := strconv.ParseInt(hex[4:6], 16, 64)
	return float64(rVal) / 255.0, float64(gVal) / 255.0, float64(bVal) / 255.0
}

const (
	defaultFontSize      = 24.0
	defaultWindowPadding = 16.0
	defaultLineSpacing   = 4.0
	labelGap             = 8.0 // vertical gap between the two labels in two-line mode
	inactiveAlpha        = 0.58
	// dynamicWidthFactor scales the measured lyric width when sizing the
	// window, giving text extra horizontal breathing room (still clamped to
	// maxWinW). Scroll detection continues to use the raw measured width.
	dynamicWidthFactor = 1.5
)

// NSTextAlignment constants (modern macOS 10.12+ / unified with UIKit)
const (
	nsTextAlignmentLeft   = 0
	nsTextAlignmentCenter = 1
	nsTextAlignmentRight  = 2
)

// Position indices
const (
	posFirst  = 0 // position A: active when currentIndex is even
	posSecond = 1 // position B: active when currentIndex is odd
)

const (
	scrollInitialDelay   = 0.8
	scrollEndPause       = 0.5
	scrollSpeed          = 50.0
	scrollTickInterval   = 1.0 / 60.0
	scrollMaxDelta       = 0.1
	windowResizeDuration = 0.5
	wordFadeDurationMs   = 160
)

// clampFactor clamps a position factor to the valid 0–1 range, using fallback
// when the value is <= 0 (unset or invalid).
func clampFactor(v, fallback float64) float64 {
	if v <= 0 || v > 1 {
		return fallback
	}
	return v
}

// scrollState tracks horizontal scrolling for one text label.
type scrollState struct {
	active     bool
	text       string
	offset     float64
	maxOffset  float64
	pauseTimer float64
	lastTick   time.Time
}

// wordAnimationState predicts lyric time between player updates.
type wordAnimationState struct {
	active     bool
	line       LyricLine
	pos        int
	alignment  int
	baseTimeMs int64
	baseAt     time.Time
}
type darwinController struct {
	cfg     configs.DesktopLyricsConfig
	window  cocoa.NSWindow
	bgView  cocoa.NSView
	labels  [2]cocoa.NSTextField // [0]=posFirst, [1]=posSecond
	visible bool
	closed  bool

	pendingMu       sync.Mutex
	pendingCurLine  LyricLine
	pendingNextLine LyricLine
	pendingIndex    int
	pendingTimeMs   int64
	pendingPlaying  bool
	font            cocoa.NSFont
	origFontSz      float64
	measureCacheKey string  // last measured text
	measureCacheW   float64 // cached width for measureCacheKey

	// Dynamic sizing
	screenFrame     cocoa.NSRect // cached bounds of the selected screen
	minWinW         float64      // minimum window width
	maxWinW         float64      // maximum window width
	currentWinW     float64      // current window width
	targetWinW      float64      // requested window width after clamping
	labelBaseY      [2]float64   // stored Y position per label
	needLargeHeight bool         // true when any visible line has \n (translation)

	// Fullscreen + subview drag architecture (LyricsX-style)
	containerView cocoa.NSView  // LyricsDragView subclass — content container
	contentOrigin cocoa.CGPoint // current origin of containerView within window
	dragActive    bool          // true during mouse drag
	dragOffsetX   float64       // mouse offset from container center at drag start (window coords)
	dragOffsetY   float64

	// LyricsX-style screen-relative center positioning (0–1).
	// These factors survive resizes; absolute origin is derived from them.
	xFactor float64
	yFactor float64

	// Animation state
	scroll        [2]*scrollState // [0]=posFirst, [1]=posSecond
	wordAnimation wordAnimationState
	animating     bool
	nextFrameAt   time.Time // scheduled deadline of the next animation frame
}

func newController(cfg configs.DesktopLyricsConfig) Controller {
	if !cfg.Enable {
		return nil
	}

	c := &darwinController{
		cfg:             cfg,
		origFontSz:      cfg.FontSize,
		needLargeHeight: true, // Match initial large window height; shrinks on first update if no translation.
		xFactor:         clampFactor(cfg.XPositionFactor, 0.5),
		yFactor:         clampFactor(cfg.YPositionFactor, 0.5),
	}
	if c.origFontSz <= 0 {
		c.origFontSz = defaultFontSize
	}
	c.scroll[posFirst] = &scrollState{}
	c.scroll[posSecond] = &scrollState{}

	setDispatchCtrl(c)
	dispatchSync(sel_createWindow)

	if c.window.ID == 0 {
		slog.Error("Failed to create desktop lyrics window")
		setDispatchCtrl(nil)
		return nil
	}

	return c
}

// ---- Main-thread operations ----

func (c *darwinController) createWindow() {
	fontSize := c.origFontSz
	padding := defaultWindowPadding
	lineH := fontSize + defaultLineSpacing

	lineCount := 2
	if c.cfg.OneLineMode {
		lineCount = 1
	}

	// Use visible frame (excludes menu bar and Dock) for the fullscreen window.
	visibleFrame := cocoa.NSScreen_MainScreen().VisibleFrame()
	c.screenFrame = visibleFrame

	// Minimum window width: tighter for one-line, wider for two-line.
	if c.cfg.OneLineMode {
		c.minWinW = fontSize * 5
	} else {
		c.minWinW = fontSize*12 + padding*2
	}
	if c.minWinW > c.screenFrame.Size.Width*0.9 {
		c.minWinW = c.screenFrame.Size.Width * 0.9
	}

	// Max window width
	maxFactor := c.cfg.MaxWindowWidth
	if maxFactor <= 0 || maxFactor > 0.9 {
		maxFactor = 0.7
	}
	c.maxWinW = c.screenFrame.Size.Width * maxFactor
	if c.maxWinW < c.minWinW {
		c.maxWinW = c.minWinW
	}

	c.currentWinW = c.minWinW
	c.targetWinW = c.currentWinW

	// ---- Fullscreen borderless window ----
	rect := cocoa.NSRect{
		Origin: visibleFrame.Origin,
		Size:   visibleFrame.Size,
	}
	c.window = cocoa.NSWindow_alloc().InitWithContentRectStyleMaskBackingDefer(
		rect,
		cocoa.NSWindowStyleMaskBorderless|cocoa.NSWindowStyleMaskFullSizeContentView,
		cocoa.NSBackingStoreBuffered,
		false,
	)
	if c.window.ID == 0 {
		return
	}

	c.window.SetTitlebarAppearsTransparent(true)
	c.window.SetOpaque(false)
	c.window.SetHasShadow(false)
	c.window.SetAlphaValue(cocoa.CGFloat(c.cfg.WindowAlpha))
	c.window.SetLevel(cocoa.NSFloatingWindowLevel)
	c.window.SetCollectionBehavior(
		cocoa.NSWindowCollectionBehaviorCanJoinAllSpaces |
			cocoa.NSWindowCollectionBehaviorStationary,
	)
	c.window.SetBackgroundColor(cocoa.NSColor_ClearColor())
	// When not draggable, let mouse events pass through to apps behind.
	c.window.SetIgnoresMouseEvents(!c.cfg.Draggable)

	// Content view (fills window)
	contentView := c.window.ContentView()
	contentView.SetWantsLayer(true)

	// ---- LyricsDragView container ----
	dragViewID := objc.ID(dragViewClass).Send(macdriver.SEL_alloc).Send(macdriver.SEL_init)
	c.containerView = cocoa.NSView{NSObject: core.NSObject{ID: dragViewID}}
	c.containerView.SetWantsLayer(true)

	// Initial content size (large height as safe default for translations).
	labelH := 2*lineH + 2*defaultLineSpacing
	textH := labelH
	textW := c.currentWinW - padding*2
	contentH := float64(lineCount)*labelH + padding*2
	if lineCount == 2 {
		contentH += labelGap
	}

	// Position from factors
	originX := c.screenFrame.Size.Width*c.xFactor - c.currentWinW/2
	originY := c.screenFrame.Size.Height*c.yFactor - contentH/2
	originX, originY = c.constrainWindowOrigin(originX, originY, c.currentWinW, contentH)
	c.contentOrigin = cocoa.CGPoint{X: originX, Y: originY}

	c.containerView.SetFrameOrigin(originX, originY)
	c.containerView.SetFrameSize(c.currentWinW, contentH)

	// Font
	c.font = cocoa.NSFont_SystemFontOfSize(cocoa.CGFloat(fontSize))
	if c.cfg.FontName != "" {
		namedFont := cocoa.NSFont_FontWithNameSize(c.cfg.FontName, cocoa.CGFloat(fontSize))
		if namedFont.ID != 0 {
			c.font = namedFont
		}
	}

	// Colors
	fgR, fgG, fgB := parseHexRGB(c.cfg.TextColor)
	inactiveColor := cocoa.NSColor_ColorWithRedGreenBlueAlpha(
		cocoa.CGFloat(fgR), cocoa.CGFloat(fgG), cocoa.CGFloat(fgB), cocoa.CGFloat(inactiveAlpha),
	)
	shadow := c.makeShadow()

	// ---- bgView inside container ----
	bgRect := cocoa.NSRect{
		Origin: cocoa.CGPoint{X: 0, Y: 0},
		Size:   cocoa.CGSize{Width: c.currentWinW, Height: contentH},
	}
	c.bgView = cocoa.NSView_alloc().InitWithFrame(bgRect)
	c.bgView.SetWantsLayer(true)
	cornerRadius := c.cfg.CornerRadius
	if cornerRadius <= 0 {
		cornerRadius = fontSize / 2
	}
	bgLayer := c.bgView.Layer()
	bgLayer.SetCornerRadius(cocoa.CGFloat(cornerRadius))
	bgLayer.SetMasksToBounds(true)
	bgHex := c.cfg.BackgroundColor
	if bgHex == "" {
		bgHex = "#000000"
	}
	bgR, bgG, bgB := parseHexRGB(bgHex)
	bgLayer.SetBackgroundCGColor(uintptr(cocoa.NSColor_ColorWithRedGreenBlueAlpha(
		cocoa.CGFloat(bgR), cocoa.CGFloat(bgG), cocoa.CGFloat(bgB), cocoa.CGFloat(c.cfg.BackgroundAlpha),
	).CGColorRef()))
	c.containerView.AddSubview(c.bgView)

	// ---- Labels ----
	if !c.cfg.OneLineMode {
		// Two-line: posFirst at top (left-aligned), posSecond at bottom (right-aligned).
		baseFirst := padding + textH + labelGap
		baseSecond := padding
		c.labelBaseY[posFirst] = baseFirst
		c.labelBaseY[posSecond] = baseSecond

		c.labels[posFirst] = c.makeTextField(baseFirst, padding, textW, textH, inactiveColor, shadow)
		c.labels[posFirst].SetAlignment(nsTextAlignmentLeft)

		c.labels[posSecond] = c.makeTextField(padding, padding, textW, textH, inactiveColor, shadow)
		c.labels[posSecond].SetAlignment(nsTextAlignmentRight)

		c.bgView.AddSubview(c.labels[posFirst].NSView)
		c.bgView.AddSubview(c.labels[posSecond].NSView)
	} else {
		// One-line: only posSecond centered.
		c.labelBaseY[posSecond] = padding
		c.labels[posSecond] = c.makeTextField(padding, padding, textW, textH, inactiveColor, shadow)
		c.labels[posSecond].SetAlignment(nsTextAlignmentCenter)
		c.bgView.AddSubview(c.labels[posSecond].NSView)
	}

	contentView.AddSubview(c.containerView)

	// Override persisted position from DB.
	if position, ok := loadDesktopLyricsPosition(); ok {
		if position.XFactor > 0 || position.YFactor > 0 {
			c.xFactor = position.XFactor
			c.yFactor = position.YFactor
		}
		c.layoutContent(false)
	}
}

func (c *darwinController) makeTextField(yOff, xOff, w, h float64, color cocoa.NSColor, shadow cocoa.NSShadow) cocoa.NSTextField {
	tf := cocoa.NSTextField_alloc().InitWithFrame(cocoa.NSRect{
		Origin: cocoa.CGPoint{X: xOff, Y: yOff},
		Size:   cocoa.CGSize{Width: w, Height: h},
	})
	tf.SetBezeled(false)
	tf.SetBordered(false)
	tf.SetDrawsBackground(false)
	tf.SetEditable(false)
	tf.SetSelectable(false)
	tf.SetAlignment(nsTextAlignmentCenter) // Default to center, overridden later per position
	tf.SetMaximumNumberOfLines(2)
	tf.SetFont(c.font)
	tf.SetTextColor(color)
	cocoa.SetViewShadow(tf.NSView, shadow)
	// Note: -[NSView setShadow:] is deprecated and does not retain the NSShadow,
	// so we must NOT release it here. The shadow lives with the window lifetime.
	return tf
}

func (c *darwinController) makeShadow() cocoa.NSShadow {
	shadowR, shadowG, shadowB := parseHexRGB(c.cfg.ShadowColor)
	radius := c.cfg.ShadowRadius
	if radius <= 0 {
		radius = 1.0
	}
	s := cocoa.NSShadow_alloc()
	s.SetShadowBlurRadius(cocoa.CGFloat(radius))
	s.SetShadowColor(cocoa.NSColor_ColorWithRedGreenBlueAlpha(
		cocoa.CGFloat(shadowR), cocoa.CGFloat(shadowG), cocoa.CGFloat(shadowB), 0.8,
	))
	s.SetShadowOffset(0, 0)
	return s
}

func loadDesktopLyricsPosition() (storage.DesktopLyricsPosition, bool) {
	data, err := storage.NewTable().GetByKVModel(storage.DesktopLyricsPosition{})
	if err != nil || len(data) == 0 {
		return storage.DesktopLyricsPosition{}, false
	}
	var position storage.DesktopLyricsPosition
	if err := json.Unmarshal(data, &position); err != nil || position.ScreenID == 0 {
		return storage.DesktopLyricsPosition{}, false
	}
	return position, true
}

// constrainWindowOrigin clamps origin (window-relative, i.e. relative to
// screenFrame.Origin) so the content stays within the fullscreen window bounds.
func (c *darwinController) constrainWindowOrigin(originX, originY, width, height float64) (float64, float64) {
	if originX < 0 {
		originX = 0
	}
	if maxX := c.screenFrame.Size.Width - width; originX > maxX {
		originX = maxX
	}
	if originY < 4 {
		originY = 4
	}
	if maxY := c.screenFrame.Size.Height - height; originY > maxY {
		originY = maxY
	}
	return originX, originY
}

// ---- Handlers (main thread) ----

func (c *darwinController) doShow() {
	c.pendingMu.Lock()
	if c.closed || c.window.ID == 0 || c.visible {
		c.pendingMu.Unlock()
		return
	}
	c.pendingMu.Unlock()

	c.window.SetIsVisible(true)
	c.window.MakeKeyAndOrderFront(0)

	c.pendingMu.Lock()
	c.visible = true
	c.pendingMu.Unlock()
}

func (c *darwinController) doHide() {
	c.pendingMu.Lock()
	if c.closed || c.window.ID == 0 || !c.visible {
		c.pendingMu.Unlock()
		return
	}
	c.pendingMu.Unlock()
	c.window.OrderOut(0)
	c.pendingMu.Lock()
	c.visible = false
	c.pendingMu.Unlock()
}

func (c *darwinController) doClose() {
	c.pendingMu.Lock()
	if c.closed {
		c.pendingMu.Unlock()
		return
	}
	c.closed = true
	c.wordAnimation.active = false
	c.animating = false
	c.pendingMu.Unlock()
	cancelScheduled(sel_scrollTick)
	cancelScheduled(sel_persistWindowPosition)

	for i := range c.labels {
		if c.labels[i].ID != 0 {
			c.labels[i].RemoveFromSuperview()
			c.labels[i].Release()
			c.labels[i].SetObjcID(0)
		}
	}
	if c.bgView.ID != 0 {
		c.bgView.RemoveFromSuperview()
		c.bgView.Release()
		c.bgView.SetObjcID(0)
	}
	if c.containerView.ID != 0 {
		c.containerView.RemoveFromSuperview()
		c.containerView.Release()
		c.containerView.SetObjcID(0)
	}
	if c.window.ID != 0 {
		c.window.Close()
		c.window.Release()
		c.window.SetObjcID(0)
	}
	c.pendingMu.Lock()
	c.visible = false
	c.pendingMu.Unlock()
}

func (c *darwinController) doUpdateText() {
	c.pendingMu.Lock()
	if c.closed || c.window.ID == 0 {
		c.pendingMu.Unlock()
		return
	}
	curLine := c.pendingCurLine
	nextLine := c.pendingNextLine
	idx := c.pendingIndex
	timeMs := c.pendingTimeMs
	playing := c.pendingPlaying
	c.pendingMu.Unlock()

	activeColor, inactiveColor := c.lyricColors()

	if c.cfg.OneLineMode {
		if c.labels[posSecond].ID != 0 {
			c.setLabelText(c.labels[posSecond], curLine, timeMs, activeColor, inactiveColor, nsTextAlignmentCenter)
			c.updateScrollNeed(posSecond, curLine)
		}
		c.updateWordAnimation(posSecond, curLine, timeMs, nsTextAlignmentCenter, playing)
		if c.labels[posFirst].ID != 0 {
			c.labels[posFirst].SetStringValue("")
			c.resetScroll(posFirst)
		}

		// Adjust content height: large when translation (\n) is present.
		hasTrans := strings.Contains(curLine.Text, "\n")
		if hasTrans != c.needLargeHeight {
			c.needLargeHeight = hasTrans
			c.layoutContent(true)
		}
		return
	}

	// Two-line alternating mode
	var activePos, nextPos int
	if idx%2 == 0 {
		activePos = posFirst
		nextPos = posSecond
	} else {
		activePos = posSecond
		nextPos = posFirst
	}

	align := nsTextAlignmentLeft // posFirst (top) → left
	if activePos == posSecond {
		align = nsTextAlignmentRight // posSecond (bottom) → right
	}
	if c.labels[activePos].ID != 0 {
		c.setLabelText(c.labels[activePos], curLine, timeMs, activeColor, inactiveColor, align)
		c.updateScrollNeed(activePos, curLine)
	}
	c.updateWordAnimation(activePos, curLine, timeMs, align, playing)
	if c.labels[nextPos].ID != 0 {
		nextAlign := nsTextAlignmentLeft // posFirst (top) → left
		if nextPos == posSecond {
			nextAlign = nsTextAlignmentRight // posSecond (bottom) → right
		}
		c.setLabelPlainText(c.labels[nextPos], nextLine, inactiveColor, nextAlign)
		c.resetScroll(nextPos)
	}

	// Adjust content height: large when any visible line has translation (\n).
	hasTrans := strings.Contains(curLine.Text, "\n") || strings.Contains(nextLine.Text, "\n")
	if hasTrans != c.needLargeHeight {
		c.needLargeHeight = hasTrans
		c.layoutContent(true)
	}
}

func (c *darwinController) lyricColors() (activeColor, inactiveColor cocoa.NSColor) {
	fgR, fgG, fgB := parseHexRGB(c.cfg.TextColor)
	activeColor = cocoa.NSColor_ColorWithRedGreenBlueAlpha(
		cocoa.CGFloat(fgR), cocoa.CGFloat(fgG), cocoa.CGFloat(fgB), 1,
	)
	inactiveColor = cocoa.NSColor_ColorWithRedGreenBlueAlpha(
		cocoa.CGFloat(fgR), cocoa.CGFloat(fgG), cocoa.CGFloat(fgB), cocoa.CGFloat(inactiveAlpha),
	)
	return activeColor, inactiveColor
}

// updateScrollNeed checks if the text needs horizontal scrolling and adjusts
// the window width to accommodate it up to the max.
func (c *darwinController) updateScrollNeed(pos int, line LyricLine) {
	text := c.getLinePlainText(line)
	textW := c.measureTextWidth(text)
	padding := defaultWindowPadding
	// Widen the window beyond the raw text extent so lyrics breathe.
	targetW := constrainWindowWidth(textW*dynamicWidthFactor+padding*2, c.minWinW, c.maxWinW)
	c.resizeWindow(targetW)

	labelW := targetW - padding*2
	if textW <= labelW || targetW < c.maxWinW {
		c.resetScroll(pos)
		return
	}

	ss := c.scroll[pos]
	maxOffset := textW - labelW
	if !beginScroll(ss, text, maxOffset) {
		return
	}
	c.startAnimating()
}

// beginScroll resets the scroll only when its rendered line changes.
func beginScroll(ss *scrollState, text string, maxOffset float64) bool {
	if ss.active && ss.text == text && ss.maxOffset == maxOffset {
		return false
	}
	ss.active = true
	ss.text = text
	ss.offset = 0
	ss.maxOffset = maxOffset
	ss.pauseTimer = scrollInitialDelay
	ss.lastTick = time.Time{}
	return true
}

func constrainWindowWidth(width, minWidth, maxWidth float64) float64 {
	return min(max(width, minWidth), maxWidth)
}

// resizeWindow is a thin wrapper: it clamps the width and triggers an animated
// content re-layout. The fullscreen window itself never changes size.
func (c *darwinController) resizeWindow(newW float64) {
	if c.closed || c.window.ID == 0 {
		return
	}
	newW = constrainWindowWidth(newW, c.minWinW, c.maxWinW)
	if newW == c.currentWinW {
		return
	}
	c.currentWinW = newW
	c.targetWinW = newW
	c.layoutContent(true)
}

// layoutContent computes the containerView's frame from current factors and
// content size, then applies it. If animate=true, uses NSAnimationContext.
func (c *darwinController) layoutContent(animate bool) {
	padding := defaultWindowPadding
	lineH := c.origFontSz + defaultLineSpacing
	labelH := lineH + defaultLineSpacing
	if c.needLargeHeight {
		labelH = 2*lineH + 2*defaultLineSpacing
	}
	lineCount := 2
	if c.cfg.OneLineMode {
		lineCount = 1
	}
	contentW := c.currentWinW
	contentH := float64(lineCount)*labelH + padding*2
	if lineCount == 2 {
		contentH += labelGap
	}

	// Compute container origin from screen-relative center factors.
	originX := c.screenFrame.Size.Width*c.xFactor - contentW/2
	originY := c.screenFrame.Size.Height*c.yFactor - contentH/2
	originX, originY = c.constrainWindowOrigin(originX, originY, contentW, contentH)
	c.contentOrigin = cocoa.CGPoint{X: originX, Y: originY}

	// Compute label positions (relative to bgView).
	textW := contentW - padding*2
	textH := labelH
	baseFirst := padding + textH + labelGap // only used in two-line mode
	baseSecond := padding
	if lineCount == 2 {
		c.labelBaseY[posFirst] = baseFirst
	}
	c.labelBaseY[posSecond] = baseSecond

	if animate {
		cocoa.NSAnimationContext_RunAnimationGroup(func(ctx cocoa.NSAnimationContext) {
			ctx.SetDuration(0.3)
			ctx.SetTimingFunction(cocoa.CAMediaTimingFunction_FunctionWithName(cocoa.CAMediaTimingFunctionEaseInEaseOut))
			ctx.SetAllowsImplicitAnimation(true)

			c.containerView.Animator().SetFrameSize(contentW, contentH)
			c.containerView.Animator().SetFrameOrigin(originX, originY)
			c.bgView.Animator().SetFrameSize(contentW, contentH)
			if c.labels[posFirst].ID != 0 {
				la := c.labels[posFirst].Animator()
				la.SetFrameSize(textW, textH)
				la.SetFrameOrigin(defaultWindowPadding-c.scroll[posFirst].offset, c.labelBaseY[posFirst])
			}
			if c.labels[posSecond].ID != 0 {
				la := c.labels[posSecond].Animator()
				la.SetFrameSize(textW, textH)
				la.SetFrameOrigin(defaultWindowPadding-c.scroll[posSecond].offset, c.labelBaseY[posSecond])
			}
		}, nil)
	} else {
		c.containerView.SetFrameSize(contentW, contentH)
		c.containerView.SetFrameOrigin(originX, originY)
		c.bgView.SetFrameSize(contentW, contentH)
		if c.labels[posFirst].ID != 0 {
			c.labels[posFirst].SetFrameSize(textW, textH)
			c.labels[posFirst].SetFrameOrigin(defaultWindowPadding-c.scroll[posFirst].offset, c.labelBaseY[posFirst])
		}
		if c.labels[posSecond].ID != 0 {
			c.labels[posSecond].SetFrameSize(textW, textH)
			c.labels[posSecond].SetFrameOrigin(defaultWindowPadding-c.scroll[posSecond].offset, c.labelBaseY[posSecond])
		}
	}
}

// ---- Drag handlers (LyricsX-style: subview drag via LyricsDragView) ----

// handleDragStart records the offset from the mouse position to the container center.
func (c *darwinController) handleDragStart(mouseX, mouseY float64) {
	if !c.cfg.Draggable || c.closed {
		return
	}
	// Content height: compute from current needLargeHeight for center calc.
	lineH := c.origFontSz + defaultLineSpacing
	labelH := lineH + defaultLineSpacing
	if c.needLargeHeight {
		labelH = 2*lineH + 2*defaultLineSpacing
	}
	lineCount := 2
	if c.cfg.OneLineMode {
		lineCount = 1
	}
	contentH := float64(lineCount)*labelH + defaultWindowPadding*2
	if lineCount == 2 {
		contentH += labelGap
	}
	cx := c.contentOrigin.X + c.currentWinW/2
	cy := c.contentOrigin.Y + contentH/2
	c.dragOffsetX = cx - mouseX
	c.dragOffsetY = cy - mouseY
	c.dragActive = true
}

// handleDragMove updates the position factors based on mouse movement.
func (c *darwinController) handleDragMove(mouseX, mouseY float64) {
	if !c.dragActive {
		return
	}
	// Compute new center in window coordinates.
	newCenterX := mouseX + c.dragOffsetX
	newCenterY := mouseY + c.dragOffsetY

	// Convert to factors (0–1 range within screen).
	c.xFactor = newCenterX / c.screenFrame.Size.Width
	c.yFactor = newCenterY / c.screenFrame.Size.Height

	// Clamp to valid range.
	if c.xFactor < 0 {
		c.xFactor = 0
	}
	if c.xFactor > 1 {
		c.xFactor = 1
	}
	if c.yFactor < 0 {
		c.yFactor = 0
	}
	if c.yFactor > 1 {
		c.yFactor = 1
	}

	// Snap to center if within 8px.
	centerSnapX := c.screenFrame.Size.Width / 2
	centerSnapY := c.screenFrame.Size.Height / 2
	if math.Abs(newCenterX-centerSnapX) < 8 {
		c.xFactor = 0.5
	}
	if math.Abs(newCenterY-centerSnapY) < 8 {
		c.yFactor = 0.5
	}

	// Reposition immediately (no animation during drag).
	c.layoutContent(false)
}

// handleDragEnd persists the current factors and cleans up.
func (c *darwinController) handleDragEnd() {
	if !c.dragActive {
		return
	}
	c.dragActive = false
	c.persistPositionFactors()
}

// persistPositionFactors saves the current screen-relative center factors to DB.
func (c *darwinController) persistPositionFactors() {
	screen := c.window.Screen()
	screenID := screen.DisplayID()
	if screen.ID == 0 || screenID == 0 {
		return
	}
	// Update screen frame cache (use visibleFrame since the window covers it).
	c.screenFrame = screen.VisibleFrame()

	position := storage.DesktopLyricsPosition{
		ScreenID: screenID,
		X:        c.contentOrigin.X,
		Y:        c.contentOrigin.Y,
		XFactor:  c.xFactor,
		YFactor:  c.yFactor,
	}
	if err := storage.NewTable().SetByKVModel(storage.DesktopLyricsPosition{}, position); err != nil {
		slog.Error("Failed to persist desktop lyrics position", "error", err)
	}
}

// measureTextWidth returns the rendered width of text in points, measured with
// the real display font via AppKit. This accounts for proportional Latin glyph
// advances and full-width CJK / punctuation, eliminating the systematic
// under-/over-estimation of the previous per-character heuristic.
//
// Called once per playback tick with a stable line, so a single-entry cache
// avoids repeated Objective-C round-trips while the lyric line is unchanged.
func (c *darwinController) measureTextWidth(text string) float64 {
	if text == "" {
		return 0
	}
	if text == c.measureCacheKey {
		return c.measureCacheW
	}
	w := cocoa.MeasureTextSize(text, c.font).Width
	c.measureCacheKey = text
	c.measureCacheW = w
	return w
}

// getLinePlainText extracts the display text from a LyricLine.
func (c *darwinController) getLinePlainText(line LyricLine) string {
	if line.Text != "" {
		return line.Text
	}
	var sb strings.Builder
	for _, w := range line.Words {
		sb.WriteString(w.Word)
	}
	return sb.String()
}

func (c *darwinController) resetScroll(pos int) {
	ss := c.scroll[pos]
	ss.active = false
	ss.text = ""
	ss.offset = 0
	ss.maxOffset = 0
	ss.pauseTimer = 0
	ss.lastTick = time.Time{}
	if c.labels[pos].ID != 0 {
		c.labels[pos].SetFrameOrigin(defaultWindowPadding, c.labelBaseY[pos])
	}
}

// scheduleFrame arms the next animation tick, compensating for the latency of
// the current handler and NSRunLoop so the cadence stays on the target grid
// instead of drifting slower with every frame. delay is the desired inter-frame
// interval; the actual sleep is shortened by however long we already overran
// the previous deadline.
func (c *darwinController) scheduleFrame(now time.Time, delay float64) {
	sleep := delay
	if !c.nextFrameAt.IsZero() {
		// Overshoot of the previous deadline eats into this frame's budget.
		if overshoot := now.Sub(c.nextFrameAt).Seconds(); overshoot > 0 {
			sleep -= overshoot
		}
	}
	if sleep < 0 {
		sleep = 0
	}
	c.nextFrameAt = now.Add(time.Duration(sleep * float64(time.Second)))
	cancelScheduled(sel_scrollTick)
	scheduleAfter(sel_scrollTick, sleep)
}

// startAnimating ensures the shared animation tick is running.
func (c *darwinController) startAnimating() {
	if c.animating {
		return
	}
	c.animating = true
	c.nextFrameAt = time.Time{}
	c.scheduleFrame(time.Now(), scrollTickInterval)
}

func (c *darwinController) requestAnimationTick() {
	c.animating = true
	c.nextFrameAt = time.Time{}
	c.scheduleFrame(time.Now(), scrollTickInterval)
}

// advanceScroll advances one scroll state by elapsed seconds.
func advanceScroll(ss *scrollState, elapsed float64) {
	if elapsed <= 0 {
		return
	}
	if ss.pauseTimer > 0 {
		ss.pauseTimer -= elapsed
		if ss.pauseTimer < 0 {
			ss.pauseTimer = 0
		}
		return
	}
	if ss.offset < ss.maxOffset {
		ss.offset += scrollSpeed * elapsed
		if ss.offset >= ss.maxOffset {
			ss.offset = ss.maxOffset
			ss.pauseTimer = scrollEndPause
		}
		return
	}
	ss.offset = 0
	ss.pauseTimer = scrollInitialDelay
}

func (c *darwinController) updateWordAnimation(pos int, line LyricLine, timeMs int64, alignment int, playing bool) {
	if !playing || len(line.Words) == 0 {
		c.wordAnimation.active = false
		c.requestAnimationTick()
		return
	}
	c.wordAnimation = wordAnimationState{
		active:     true,
		line:       line,
		pos:        pos,
		alignment:  alignment,
		baseTimeMs: timeMs,
		baseAt:     time.Now(),
	}
	c.requestAnimationTick()
}

func nextWordAnimationDelay(line LyricLine, timeMs int64) (float64, bool) {
	for _, word := range line.Words {
		if timeMs < word.StartTime {
			return float64(word.StartTime-timeMs) / 1000, true
		}
		if timeMs < wordHighlightEndTime(word) {
			return scrollTickInterval, true
		}
	}
	return 0, false
}

func wordHighlightEndTime(word LyricWord) int64 {
	if word.EndTime > word.StartTime {
		return word.EndTime
	}
	return word.StartTime + wordFadeDurationMs
}

func wordHighlightProgress(word LyricWord, currentTimeMs int64) float64 {
	if currentTimeMs <= word.StartTime {
		return 0
	}
	endTimeMs := wordHighlightEndTime(word)
	if currentTimeMs >= endTimeMs {
		return 1
	}
	return float64(currentTimeMs-word.StartTime) / float64(endTimeMs-word.StartTime)
}

type wordHighlightRange struct {
	highlightedLength   int
	transitioningLength int
	transitionProgress  float64
}

func wordHighlightRangeForProgress(word string, progress float64) wordHighlightRange {
	if progress <= 0 {
		return wordHighlightRange{}
	}
	if progress >= 1 {
		return wordHighlightRange{highlightedLength: utf16Length(word)}
	}

	graphemeCount := uniseg.GraphemeClusterCount(word)
	if graphemeCount == 0 {
		return wordHighlightRange{}
	}
	position := progress * float64(graphemeCount)
	completed := int(position)
	transitionProgress := position - float64(completed)

	result := wordHighlightRange{transitionProgress: transitionProgress}
	graphemes := uniseg.NewGraphemes(word)
	for graphemes.Next() {
		length := utf16Length(graphemes.Str())
		if completed > 0 {
			result.highlightedLength += length
			completed--
			continue
		}
		if transitionProgress > 0 {
			result.transitioningLength = length
		}
		break
	}
	return result
}

func (c *darwinController) tickWordAnimation(now time.Time) (float64, bool) {
	state := &c.wordAnimation
	if !state.active {
		return 0, false
	}
	timeMs := state.baseTimeMs + now.Sub(state.baseAt).Milliseconds()
	delay, active := nextWordAnimationDelay(state.line, timeMs)
	if !active {
		state.active = false
		return 0, false
	}
	if delay <= scrollTickInterval && c.labels[state.pos].ID != 0 {
		activeColor, inactiveColor := c.lyricColors()
		c.setLabelText(c.labels[state.pos], state.line, timeMs, activeColor, inactiveColor, state.alignment)
	}
	return delay, true
}

// doTick advances active animations and re-schedules the next frame.
func (c *darwinController) doTick() {
	c.pendingMu.Lock()
	closed := c.closed
	c.pendingMu.Unlock()
	if closed {
		return
	}

	now := time.Now()
	hasAnimation := false
	nextDelay := scrollTickInterval
	for i := range c.scroll {
		ss := c.scroll[i]
		if !ss.active {
			continue
		}
		hasAnimation = true

		elapsed := scrollTickInterval
		if !ss.lastTick.IsZero() {
			elapsed = now.Sub(ss.lastTick).Seconds()
			if elapsed <= 0 {
				elapsed = scrollTickInterval
			} else if elapsed > scrollMaxDelta {
				elapsed = scrollMaxDelta
			}
		}
		ss.lastTick = now
		advanceScroll(ss, elapsed)

		if c.labels[i].ID != 0 {
			c.labels[i].SetFrameOrigin(defaultWindowPadding-ss.offset, c.labelBaseY[i])
		}
	}
	if delay, active := c.tickWordAnimation(now); active {
		if !hasAnimation || delay < nextDelay {
			nextDelay = delay
		}
		hasAnimation = true
	}

	if hasAnimation {
		c.scheduleFrame(now, nextDelay)
	} else {
		c.animating = false
		c.nextFrameAt = time.Time{}
	}
}

// setLabelText sets the label content, using word-by-word coloring when
// YRC word data is available, falling back to plain text.
// alignment: 0=left, 1=center, 2=right (NSTextAlignment constants, macOS 10.12+)
func (c *darwinController) setLabelText(label cocoa.NSTextField, line LyricLine, timeMs int64, activeColor, inactiveColor cocoa.NSColor, alignment int) {
	if len(line.Words) > 0 {
		// Diagnostic: try simple attributed string first to isolate the issue
		plainText := line.Text
		if plainText == "" {
			var sb strings.Builder
			for _, w := range line.Words {
				sb.WriteString(w.Word)
			}
			plainText = sb.String()
		}

		// Try creating a simple attributed string with one color for all text
		testAttr := cocoa.NSMutableAttributedString_alloc()
		if testAttr.ID == 0 {
			slog.Error("desktop_lyrics: NSMutableAttributedString_alloc returned nil ID")
			label.SetStringValue(plainText)
			label.SetTextColor(activeColor)
			return
		}
		testAttr.InitWithString(plainText)
		if testAttr.ID == 0 {
			slog.Error("desktop_lyrics: InitWithString returned nil ID", "textLen", len(plainText))
			label.SetStringValue(plainText)
			label.SetTextColor(activeColor)
			return
		}

		// Apply KTV-style word coloring from left to right.
		offset := 0
		for _, w := range line.Words {
			wordLen := utf16Length(w.Word)
			if wordLen == 0 {
				continue
			}
			wordRange := cocoa.NSRange{Location: offset, Length: wordLen}
			testAttr.AddAttribute(cocoa.NSForegroundColorAttributeName, inactiveColor, wordRange)

			highlight := wordHighlightRangeForProgress(w.Word, wordHighlightProgress(w, timeMs))
			if highlight.highlightedLength > 0 {
				highlightedRange := cocoa.NSRange{Location: offset, Length: highlight.highlightedLength}
				testAttr.AddAttribute(cocoa.NSForegroundColorAttributeName, activeColor, highlightedRange)
			}
			if highlight.transitioningLength > 0 {
				progress := highlight.transitionProgress
				progress = progress * progress * (3 - 2*progress)
				transitioningRange := cocoa.NSRange{
					Location: offset + highlight.highlightedLength,
					Length:   highlight.transitioningLength,
				}
				testAttr.AddAttribute(cocoa.NSForegroundColorAttributeName, c.blendColor(progress), transitioningRange)
			}
			offset += wordLen
		}

		// Apply paragraph style for correct alignment
		paraRange := cocoa.NSRange{Location: 0, Length: offset}

		// Apply paragraph style for correct alignment
		paraStyle := cocoa.NewParagraphStyle()
		paraStyle.SetAlignment(alignment)
		testAttr.AddParagraphStyle(paraStyle, paraRange)
		paraStyle.Release()

		label.SetAttributedStringValue(testAttr)
		slog.Debug("desktop_lyrics: word-by-word attributed string applied",
			"wordCount", len(line.Words), "timeMs", timeMs, "textLen", len(plainText))
		testAttr.Release()
	} else {
		label.SetStringValue(line.Text)
		label.SetTextColor(activeColor)
	}
}

// setLabelPlainText sets the next-line label content (no word-level coloring).
// alignment: 0=left, 1=center, 2=right (NSTextAlignment constants, macOS 10.12+)
func (c *darwinController) setLabelPlainText(label cocoa.NSTextField, line LyricLine, inactiveColor cocoa.NSColor, alignment int) {
	text := line.Text
	if text == "" && len(line.Words) > 0 {
		var sb strings.Builder
		for _, w := range line.Words {
			sb.WriteString(w.Word)
		}
		text = sb.String()
	}
	if text == "" {
		label.SetStringValue("")
		return
	}
	// Create attributed string to apply alignment
	attrStr := cocoa.NSMutableAttributedString_alloc()
	if attrStr.ID == 0 {
		// Fallback to plain text if attributed string creation fails
		label.SetStringValue(text)
		label.SetTextColor(inactiveColor)
		return
	}
	attrStr.InitWithString(text)
	if attrStr.ID == 0 {
		label.SetStringValue(text)
		label.SetTextColor(inactiveColor)
		return
	}

	// Apply color
	textLen := utf16Length(text)
	colorRange := cocoa.NSRange{Location: 0, Length: textLen}
	attrStr.AddAttribute(cocoa.NSForegroundColorAttributeName, inactiveColor, colorRange)

	// Apply alignment via paragraph style
	paraStyle := cocoa.NewParagraphStyle()
	paraStyle.SetAlignment(alignment)
	attrStr.AddParagraphStyle(paraStyle, colorRange)
	paraStyle.Release()

	label.SetAttributedStringValue(attrStr)
	attrStr.Release()
}

// blendColor creates a new NSColor by interpolating only the alpha channel.
// Both colors share the same RGB (textColor); we only fade opacity.
// t=0.0 returns inactiveColor equivalent, t=1.0 returns fully opaque.
func (c *darwinController) blendColor(t float64) cocoa.NSColor {
	aAlpha := inactiveAlpha
	bAlpha := 1.0

	r, g, b := parseHexRGB(c.cfg.TextColor)
	return cocoa.NSColor_ColorWithRedGreenBlueAlpha(
		cocoa.CGFloat(r),
		cocoa.CGFloat(g),
		cocoa.CGFloat(b),
		cocoa.CGFloat(lerp(aAlpha, bAlpha, t)),
	)
}

// ---- Public interface ----

func (c *darwinController) Show() {
	if c == nil {
		return
	}
	c.pendingMu.Lock()
	closed := c.closed
	c.pendingMu.Unlock()
	if closed {
		return
	}
	dispatchAsync(sel_showWindow)
}

func (c *darwinController) Hide() {
	if c == nil {
		return
	}
	c.pendingMu.Lock()
	closed := c.closed
	c.pendingMu.Unlock()
	if closed {
		return
	}
	dispatchAsync(sel_hideWindow)
}

func (c *darwinController) IsVisible() bool {
	if c == nil {
		return false
	}
	c.pendingMu.Lock()
	defer c.pendingMu.Unlock()
	return c.visible
}

func (c *darwinController) Update(curLine, nextLine LyricLine, currentIndex int, currentTimeMs int64, playing bool) {
	if c == nil {
		return
	}
	c.pendingMu.Lock()
	if c.closed || c.window.ID == 0 {
		c.pendingMu.Unlock()
		return
	}
	c.pendingMu.Unlock()

	c.pendingMu.Lock()
	c.pendingCurLine = curLine
	c.pendingNextLine = nextLine
	c.pendingIndex = currentIndex
	c.pendingTimeMs = currentTimeMs
	c.pendingPlaying = playing
	c.pendingMu.Unlock()

	dispatchAsync(sel_updateText)
}

func (c *darwinController) Close() {
	if c == nil {
		return
	}
	// dispatchSync(sel_closeWindow)
	setDispatchCtrl(nil)
}
