//go:build darwin

package desktop_lyrics

import (
	"log/slog"
	"strconv"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/macdriver/cocoa"
)

// lerp performs linear interpolation between a and b by factor t (0.0-1.0).
func lerp(a, b, t float64) float64 {
	return a + (b-a)*t
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
	inactiveAlpha        = 0.65
)

// Position indices
const (
	posFirst  = 0 // position A: active when currentIndex is even
	posSecond = 1 // position B: active when currentIndex is odd
)

const (
	scrollInitialDelay = 1.5  // seconds before starting scroll
	scrollEndPause     = 1.0  // seconds to pause at the end before resetting
	scrollSpeed        = 30.0 // points per second
)

// scrollState tracks horizontal scrolling for one text label.
type scrollState struct {
	active     bool    // whether scrolling is currently active
	offset     float64 // current scroll offset (positive = scrolled left)
	maxOffset  float64 // total overflow width to scroll through
	pauseTimer float64 // countdown before starting/resuming scroll
}

type darwinController struct {
	cfg       configs.DesktopLyricsConfig
	window    cocoa.NSWindow
	bgView    cocoa.NSView
	labels    [2]cocoa.NSTextField // [0]=posFirst, [1]=posSecond
	visible   bool
	closed    bool

	pendingMu       sync.Mutex
	pendingCurLine  LyricLine
	pendingNextLine LyricLine
	pendingIndex    int
	pendingTimeMs   int64

	font       cocoa.NSFont
	origFontSz float64

	// Dynamic sizing
	screenW     float64 // cached screen width
	screenH     float64 // cached screen height (for Y coord calc)
	minWinW     float64 // minimum window width
	maxWinW     float64 // maximum window width
	currentWinW float64 // last set window width
	labelBaseY  [2]float64 // stored Y position per label

	// Scroll animation
	scroll   [2]*scrollState // [0]=posFirst, [1]=posSecond
	scrolling bool           // whether any label is currently scrolling
}

func newController(cfg configs.DesktopLyricsConfig) Controller {
	if !cfg.Enable {
		return nil
	}

	c := &darwinController{
		cfg:        cfg,
		origFontSz: cfg.FontSize,
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
	c.screenW = float64(cocoa.CGDisplayPixelsWide(cocoa.CGMainDisplayID()))
	c.screenH = float64(cocoa.CGDisplayPixelsHigh(cocoa.CGMainDisplayID()))

	fontSize := c.origFontSz
	padding := defaultWindowPadding
	lineH := fontSize + defaultLineSpacing

	// Base window width (wide enough for typical 10-char line)
	c.minWinW = fontSize * 20
	if c.minWinW > c.screenW*0.9 {
		c.minWinW = c.screenW * 0.9
	}

	// Max window width
	maxFactor := c.cfg.MaxWindowWidth
	if maxFactor <= 0 || maxFactor > 0.9 {
		maxFactor = 0.7
	}
	c.maxWinW = c.screenW * maxFactor
	if c.maxWinW < c.minWinW {
		c.maxWinW = c.minWinW
	}

	winW := c.minWinW
	c.currentWinW = winW

	lineCount := 2
	if c.cfg.OneLineMode {
		lineCount = 1
	}
	winH := float64(lineCount)*(lineH) + padding*2

	// Center-based positioning
	xFactor := c.cfg.XPositionFactor
	if xFactor <= 0 {
		xFactor = 0.5
	}
	yFactor := c.cfg.YPositionFactor

	winX := c.screenW*xFactor - winW/2
	if winX < 0 {
		winX = 0
	}
	if winX+winW > c.screenW {
		winX = c.screenW - winW
	}
	winY := c.screenH*yFactor - winH/2
	if winY < 4 {
		winY = 4
	}
	if winY+winH > c.screenH {
		winY = c.screenH - winH
	}

	rect := cocoa.NSRect{
		Origin: cocoa.CGPoint{X: winX, Y: winY},
		Size:   cocoa.CGSize{Width: winW, Height: winH},
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
	c.applyMouseBehavior()
	c.window.SetBackgroundColor(cocoa.NSColor_ClearColor())

	// Content view
	contentView := c.window.ContentView()
	contentView.SetWantsLayer(true)

	// Background with corner radius
	bgRect := cocoa.NSRect{
		Origin: cocoa.CGPoint{X: 0, Y: 0},
		Size:   cocoa.CGSize{Width: winW, Height: winH},
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

	contentView.AddSubview(c.bgView)

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

	// Create two text fields filling most of the window width
	textW := winW - padding*2
	textH := lineH + defaultLineSpacing

	// posFirst: top-left
	c.labels[posFirst] = c.makeTextField(
		winH-padding-lineH, padding, textW, textH,
		inactiveColor, shadow,
	)
	c.labels[posFirst].SetAlignment(0) // NSTextAlignmentLeft

	// posSecond: bottom-right (same width, positioned lower)
	c.labels[posSecond] = c.makeTextField(
		padding, padding, textW, textH,
		inactiveColor, shadow,
	)

	if !c.cfg.OneLineMode {
		c.labels[posSecond].SetAlignment(1) // NSTextAlignmentRight for two-line mode
		c.labelBaseY[posFirst] = winH - padding - lineH
		c.labelBaseY[posSecond] = padding
		c.bgView.AddSubview(c.labels[posFirst].NSView)
		c.bgView.AddSubview(c.labels[posSecond].NSView)
	} else {
		c.labels[posSecond].SetAlignment(2) // NSTextAlignmentCenter
		c.labelBaseY[posSecond] = padding
		c.bgView.AddSubview(c.labels[posSecond].NSView)
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
	tf.SetAlignment(2)
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

func (c *darwinController) applyMouseBehavior() {
	if c.cfg.Draggable {
		c.window.SetIgnoresMouseEvents(false)
		c.window.SetMovableByWindowBackground(true)
	} else {
		c.window.SetIgnoresMouseEvents(true)
	}
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
	c.scrolling = false
	c.pendingMu.Unlock()
	cancelScheduled(sel_scrollTick)

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
	c.pendingMu.Unlock()

	fgR, fgG, fgB := parseHexRGB(c.cfg.TextColor)
	activeColor := cocoa.NSColor_ColorWithRedGreenBlueAlpha(
		cocoa.CGFloat(fgR), cocoa.CGFloat(fgG), cocoa.CGFloat(fgB), 1.0,
	)
	inactiveColor := cocoa.NSColor_ColorWithRedGreenBlueAlpha(
		cocoa.CGFloat(fgR), cocoa.CGFloat(fgG), cocoa.CGFloat(fgB), cocoa.CGFloat(inactiveAlpha),
	)

	if c.cfg.OneLineMode {
		if c.labels[posSecond].ID != 0 {
			c.setLabelText(c.labels[posSecond], curLine, timeMs, activeColor, inactiveColor)
			c.updateScrollNeed(posSecond, curLine)
		}
		if c.labels[posFirst].ID != 0 {
			c.labels[posFirst].SetStringValue("")
			c.resetScroll(posFirst)
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

	if c.labels[activePos].ID != 0 {
		c.setLabelText(c.labels[activePos], curLine, timeMs, activeColor, inactiveColor)
		c.updateScrollNeed(activePos, curLine)
	}
	if c.labels[nextPos].ID != 0 {
		c.setLabelPlainText(c.labels[nextPos], nextLine, inactiveColor)
		c.resetScroll(nextPos)
	}

	// Enforce position-based alignment — AttributedString may reset it
	c.labels[posFirst].SetAlignment(0)  // top row: left
	c.labels[posSecond].SetAlignment(1) // bottom row: right
}

// updateScrollNeed checks if the text needs horizontal scrolling and adjusts
// the window width to accommodate it up to the max.
func (c *darwinController) updateScrollNeed(pos int, line LyricLine) {
	text := c.getLinePlainText(line)
	textW := c.measureTextWidth(text)
	padding := defaultWindowPadding
	neededWinW := textW + padding*2

	ss := c.scroll[pos]
	ss.active = false
	ss.pauseTimer = scrollInitialDelay
	ss.offset = 0

	if neededWinW > c.currentWinW {
		// Expand window if within max
		targetW := min(neededWinW, c.maxWinW)
		if targetW > c.currentWinW {
			c.resizeWindow(targetW)
		}
	}

	// After resize, check if text still overflows
	labelW := c.currentWinW - padding*2
	if textW > labelW && c.currentWinW >= c.maxWinW {
		ss.active = true
		ss.maxOffset = textW - labelW + padding
		c.startScrolling()
	} else {
		c.resetScroll(pos)
	}
}

// resizeWindow resizes the window to the given width while preserving
// the origin, and updates all subviews.
func (c *darwinController) resizeWindow(newW float64) {
	fontSize := c.origFontSz
	padding := defaultWindowPadding
	lineH := fontSize + defaultLineSpacing

	lineCount := 2
	if c.cfg.OneLineMode {
		lineCount = 1
	}
	newH := float64(lineCount)*(lineH) + padding*2

	xFactor := c.cfg.XPositionFactor
	if xFactor <= 0 {
		xFactor = 0.5
	}
	yFactor := c.cfg.YPositionFactor

	newX := c.screenW*xFactor - newW/2
	if newX < 0 {
		newX = 0
	}
	if newX+newW > c.screenW {
		newX = c.screenW - newW
	}
	newY := c.screenH*yFactor - newH/2
	if newY < 4 {
		newY = 4
	}
	if newY+newH > c.screenH {
		newY = c.screenH - newH
	}

	// Set window frame
	c.window.SetFrameDisplayTopLeft(cocoa.NSRect{
		Origin: cocoa.CGPoint{X: newX, Y: newY},
		Size:   cocoa.CGSize{Width: newW, Height: newH},
	}, true)

	// Resize background view
	c.bgView.SetFrameSize(newW, newH)

	// Resize text fields
	textW := newW - padding*2
	textH := lineH + defaultLineSpacing

	c.labelBaseY[posFirst] = newH - padding - lineH
	c.labelBaseY[posSecond] = padding

	if c.labels[posFirst].ID != 0 {
		c.labels[posFirst].SetFrameSize(textW, textH)
		c.labels[posFirst].SetFrameOrigin(padding, c.labelBaseY[posFirst])
	}
	if c.labels[posSecond].ID != 0 {
		c.labels[posSecond].SetFrameSize(textW, textH)
		c.labels[posSecond].SetFrameOrigin(padding, c.labelBaseY[posSecond])
	}

	c.currentWinW = newW
}

// measureTextWidth estimates the rendered width of a string in points.
func (c *darwinController) measureTextWidth(text string) float64 {
	runes := utf8.RuneCountInString(text)
	return c.origFontSz * float64(runes) * 0.7
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
	if c.scroll[pos].active {
		c.scroll[pos].active = false
		c.scroll[pos].offset = 0
		if c.labels[pos].ID != 0 {
			c.labels[pos].SetFrameOrigin(defaultWindowPadding, c.labelBaseY[pos])
		}
	}
}

// startScrolling ensures the scroll tick is running.
func (c *darwinController) startScrolling() {
	if c.scrolling {
		return
	}
	c.scrolling = true
	cancelScheduled(sel_scrollTick)
	scheduleAfter(sel_scrollTick, 0.016)
}

// doTick advances the scroll animation and re-schedules.
func (c *darwinController) doTick() {
	c.pendingMu.Lock()
	closed := c.closed
	c.pendingMu.Unlock()
	if closed {
		return
	}

	anyActive := false
	tickSec := 0.016 // ~60fps

	for i := range c.scroll {
		ss := c.scroll[i]
		if !ss.active {
			continue
		}
		anyActive = true

		if ss.pauseTimer > 0 {
			ss.pauseTimer -= tickSec
			continue
		}

		// Advance scroll
		if ss.offset < ss.maxOffset {
			ss.offset += scrollSpeed * tickSec
			if ss.offset > ss.maxOffset {
				ss.offset = ss.maxOffset
				ss.pauseTimer = scrollEndPause
			}
		} else {
			// At end, pause then reset
			if ss.pauseTimer <= 0 {
				ss.offset = 0
				ss.pauseTimer = scrollInitialDelay
			}
		}

		// Apply offset to text field (negative x = scroll left)
		padding := defaultWindowPadding
		if c.labels[i].ID != 0 {
			c.labels[i].SetFrameOrigin(padding-ss.offset, c.labelBaseY[i])
		}
	}

	if anyActive {
		cancelScheduled(sel_scrollTick)
		scheduleAfter(sel_scrollTick, tickSec)
	} else {
		c.scrolling = false
	}
}

// setLabelText sets the label content, using word-by-word coloring when
// YRC word data is available, falling back to plain text.
func (c *darwinController) setLabelText(label cocoa.NSTextField, line LyricLine, timeMs int64, activeColor, inactiveColor cocoa.NSColor) {
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

		// Apply word-by-word coloring with fade-in only
		const fadeInMs int64 = 100

		offset := 0
		for _, w := range line.Words {
			runeLen := utf8.RuneCountInString(w.Word)
			if runeLen == 0 {
				continue
			}
			rng := cocoa.NSRange{Location: offset, Length: runeLen}

			// Calculate progress: 0.0 = inactive (unplayed), 1.0 = active (played/playing)
			// Once a word starts, it stays bright — no fade-out after
			var progress float64
			switch {
			case timeMs < w.StartTime-fadeInMs:
				progress = 0.0 // not yet: gray
			case timeMs < w.StartTime:
				// Fade in: gray → bright
				progress = float64(timeMs-(w.StartTime-fadeInMs)) / float64(fadeInMs)
			default:
				// Already playing or played: stay bright
				progress = 1.0
			}
			color := c.blendColor(progress)

			testAttr.AddAttribute(cocoa.NSForegroundColorAttributeName, color, rng)
			offset += runeLen
		}

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
func (c *darwinController) setLabelPlainText(label cocoa.NSTextField, line LyricLine, inactiveColor cocoa.NSColor) {
	text := line.Text
	if text == "" && len(line.Words) > 0 {
		var sb strings.Builder
		for _, w := range line.Words {
			sb.WriteString(w.Word)
		}
		text = sb.String()
	}
	if text != "" {
		label.SetStringValue(text)
		label.SetTextColor(inactiveColor)
	}
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

func (c *darwinController) Update(curLine, nextLine LyricLine, currentIndex int, currentTimeMs int64) {
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
	c.pendingMu.Unlock()

	dispatchAsync(sel_updateText)
}

func (c *darwinController) Close() {
	if c == nil {
		return
	}
	dispatchSync(sel_closeWindow)
	setDispatchCtrl(nil)
}
