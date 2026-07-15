//go:build darwin

package desktop_lyrics

import (
	"log/slog"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf16"

	"github.com/rivo/uniseg"

	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/macdriver/cocoa"
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
	inactiveAlpha        = 0.42
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
	windowResizeDuration = 0.2
	wordFadeDurationMs   = 160
)

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

type windowResizeState struct {
	active                  bool
	startWidth, targetWidth float64
	centerX, centerY        float64
	startedAt               time.Time
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

	// Dynamic sizing
	screenW     float64    // cached screen width
	screenH     float64    // cached screen height
	minWinW     float64    // minimum window width
	maxWinW     float64    // maximum window width
	currentWinW float64    // current window width
	targetWinW  float64    // requested window width after clamping
	labelBaseY  [2]float64 // stored Y position per label

	// Animation state
	scroll        [2]*scrollState // [0]=posFirst, [1]=posSecond
	wordAnimation wordAnimationState
	windowResize  windowResizeState
	animating     bool
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

	lineCount := 2
	if c.cfg.OneLineMode {
		lineCount = 1
	}

	// Minimum window width: tighter for one-line, wider for two-line to avoid pure vertical stacking
	if c.cfg.OneLineMode {
		c.minWinW = fontSize * 5
	} else {
		c.minWinW = fontSize*12 + padding*2
	}
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
	c.targetWinW = winW

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
	c.labels[posFirst].SetAlignment(nsTextAlignmentLeft)

	// posSecond: bottom-right (same width, positioned lower)
	c.labels[posSecond] = c.makeTextField(
		padding, padding, textW, textH,
		inactiveColor, shadow,
	)

	if !c.cfg.OneLineMode {
		c.labels[posSecond].SetAlignment(nsTextAlignmentRight) // Right-align for two-line mode
		c.labelBaseY[posFirst] = winH - padding - lineH
		c.labelBaseY[posSecond] = padding
		c.bgView.AddSubview(c.labels[posFirst].NSView)
		c.bgView.AddSubview(c.labels[posSecond].NSView)
	} else {
		c.labels[posSecond].SetAlignment(nsTextAlignmentCenter) // Center for one-line mode
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
	c.wordAnimation.active = false
	c.windowResize.active = false
	c.animating = false
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
	targetW := constrainWindowWidth(textW+padding*2, c.minWinW, c.maxWinW)
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

func windowResizeWidth(startWidth, targetWidth, elapsed float64) (float64, bool) {
	if elapsed <= 0 {
		return startWidth, false
	}
	if elapsed >= windowResizeDuration {
		return targetWidth, true
	}
	progress := elapsed / windowResizeDuration
	progress = progress * progress * (3 - 2*progress)
	return lerp(startWidth, targetWidth, progress), false
}

// resizeWindow animates to a new bounded width while preserving the current center.
func (c *darwinController) resizeWindow(newW float64) {
	newW = constrainWindowWidth(newW, c.minWinW, c.maxWinW)
	if newW == c.targetWinW {
		return
	}

	frame := c.window.Frame()
	c.currentWinW = frame.Size.Width
	c.targetWinW = newW
	c.windowResize = windowResizeState{
		active:      c.currentWinW != newW,
		startWidth:  c.currentWinW,
		targetWidth: newW,
		centerX:     frame.Origin.X + frame.Size.Width/2,
		centerY:     frame.Origin.Y + frame.Size.Height/2,
		startedAt:   time.Now(),
	}
	if !c.windowResize.active {
		c.applyWindowWidth(newW, c.windowResize.centerX, c.windowResize.centerY)
		return
	}
	c.startAnimating()
}

func (c *darwinController) applyWindowWidth(newW, centerX, centerY float64) {
	newW = constrainWindowWidth(newW, c.minWinW, c.maxWinW)
	padding := defaultWindowPadding
	lineH := c.origFontSz + defaultLineSpacing
	lineCount := 2
	if c.cfg.OneLineMode {
		lineCount = 1
	}
	newH := float64(lineCount)*lineH + padding*2

	newX := centerX - newW/2
	if newX < 0 {
		newX = 0
	}
	if newX+newW > c.screenW {
		newX = c.screenW - newW
	}
	newY := centerY - newH/2
	if newY < 4 {
		newY = 4
	}
	if newY+newH > c.screenH {
		newY = c.screenH - newH
	}

	c.window.SetFrameDisplayTopLeft(cocoa.NSRect{
		Origin: cocoa.CGPoint{X: newX, Y: newY},
		Size:   cocoa.CGSize{Width: newW, Height: newH},
	}, true)
	c.bgView.SetFrameSize(newW, newH)

	textW := newW - padding*2
	textH := lineH + defaultLineSpacing
	c.labelBaseY[posFirst] = newH - padding - lineH
	c.labelBaseY[posSecond] = padding
	if c.labels[posFirst].ID != 0 {
		c.labels[posFirst].SetFrameSize(textW, textH)
		c.labels[posFirst].SetFrameOrigin(padding-c.scroll[posFirst].offset, c.labelBaseY[posFirst])
	}
	if c.labels[posSecond].ID != 0 {
		c.labels[posSecond].SetFrameSize(textW, textH)
		c.labels[posSecond].SetFrameOrigin(padding-c.scroll[posSecond].offset, c.labelBaseY[posSecond])
	}
	c.currentWinW = newW
}

func (c *darwinController) tickWindowResize(now time.Time) bool {
	state := &c.windowResize
	if !state.active {
		return false
	}
	width, done := windowResizeWidth(state.startWidth, state.targetWidth, now.Sub(state.startedAt).Seconds())
	c.applyWindowWidth(width, state.centerX, state.centerY)
	if done {
		state.active = false
	}
	return !done
}

// measureTextWidth estimates the rendered width of a string in points.
// CJK characters (Chinese, Japanese, Korean) are typically full-width (~1.0x fontSize),
// while ASCII/Latin characters are narrower (~0.5x fontSize).
func (c *darwinController) measureTextWidth(text string) float64 {
	var cjkCount, otherCount int
	for _, r := range text {
		if isCJK(r) {
			cjkCount++
		} else {
			otherCount++
		}
	}
	// CJK characters: ~1.0x fontSize (full-width)
	// Non-CJK characters: ~0.5x fontSize (half-width, averaged for proportional fonts)
	return c.origFontSz * (float64(cjkCount)*1.0 + float64(otherCount)*0.5)
}

// isCJK checks if a rune is a CJK (Chinese, Japanese, Korean) character.
func isCJK(r rune) bool {
	return (r >= 0x4E00 && r <= 0x9FFF) || // CJK Unified Ideographs
		(r >= 0x3400 && r <= 0x4DBF) || // CJK Unified Ideographs Extension A
		(r >= 0x20000 && r <= 0x2A6DF) || // CJK Unified Ideographs Extension B
		(r >= 0x2A700 && r <= 0x2B73F) || // CJK Unified Ideographs Extension C
		(r >= 0x2B740 && r <= 0x2B81F) || // CJK Unified Ideographs Extension D
		(r >= 0x2B820 && r <= 0x2CEAF) || // CJK Unified Ideographs Extension E
		(r >= 0xF900 && r <= 0xFAFF) || // CJK Compatibility Ideographs
		(r >= 0x2F800 && r <= 0x2FA1F) || // CJK Compatibility Ideographs Supplement
		(r >= 0x3040 && r <= 0x309F) || // Hiragana
		(r >= 0x30A0 && r <= 0x30FF) || // Katakana
		(r >= 0xAC00 && r <= 0xD7AF) // Hangul Syllables
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

// startAnimating ensures the shared animation tick is running.
func (c *darwinController) startAnimating() {
	if c.animating {
		return
	}
	c.animating = true
	cancelScheduled(sel_scrollTick)
	scheduleAfter(sel_scrollTick, scrollTickInterval)
}

func (c *darwinController) requestAnimationTick() {
	c.animating = true
	cancelScheduled(sel_scrollTick)
	scheduleAfter(sel_scrollTick, scrollTickInterval)
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

	if c.tickWindowResize(now) {
		hasAnimation = true
	}

	if delay, active := c.tickWordAnimation(now); active {
		if !hasAnimation || delay < nextDelay {
			nextDelay = delay
		}
		hasAnimation = true
	}

	if hasAnimation {
		cancelScheduled(sel_scrollTick)
		scheduleAfter(sel_scrollTick, nextDelay)
	} else {
		c.animating = false
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
	if text != "" {
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
	dispatchSync(sel_closeWindow)
	setDispatchCtrl(nil)
}
