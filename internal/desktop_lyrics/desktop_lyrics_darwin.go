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
	inactiveAlpha        = 0.45
)

// Position indices
const (
	posFirst  = 0 // position A: active when currentIndex is even
	posSecond = 1 // position B: active when currentIndex is odd
)

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
	screenW := float64(cocoa.CGDisplayPixelsWide(cocoa.CGMainDisplayID()))
	screenH := float64(cocoa.CGDisplayPixelsHigh(cocoa.CGMainDisplayID()))

	fontSize := c.origFontSz
	padding := defaultWindowPadding
	lineH := fontSize + defaultLineSpacing

	// Window: wide enough, fits 2 lines + padding
	winW := fontSize * 20
	if winW > screenW*0.9 {
		winW = screenW * 0.9
	}

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

	winX := screenW*xFactor - winW/2
	if winX < 0 {
		winX = 0
	}
	if winX+winW > screenW {
		winX = screenW - winW
	}
	winY := screenH*yFactor - winH/2
	if winY < 4 {
		winY = 4
	}
	if winY+winH > screenH {
		winY = screenH - winH
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

	// Create two text fields with diagonal layout: top-left and bottom-right
	textW := winW * 0.6 // each field takes ~60% of window width
	textH := lineH + defaultLineSpacing

	// posFirst: top-left (y = tall, x = left)
	c.labels[posFirst] = c.makeTextField(
		winH-padding-lineH, padding, textW, textH,
		inactiveColor, shadow,
	)
	// Align text to the left within its frame
	c.labels[posFirst].SetAlignment(0) // NSTextAlignmentLeft

	// posSecond: bottom-right (y = low, x = right edge)
	posSecondAlign := 1 // NSTextAlignmentRight (for two-line mode)
	if c.cfg.OneLineMode {
		posSecondAlign = 2 // NSTextAlignmentCenter
	}
	c.labels[posSecond] = c.makeTextField(
		padding, winW-textW-padding, textW, textH,
		inactiveColor, shadow,
	)
	// Align text to the right (or center in single-line mode)
	c.labels[posSecond].SetAlignment(posSecondAlign)

	if !c.cfg.OneLineMode {
		c.bgView.AddSubview(c.labels[posFirst].NSView)
		c.bgView.AddSubview(c.labels[posSecond].NSView)
	} else {
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
		radius = 3.0
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
	c.pendingMu.Unlock()

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
		}
		if c.labels[posFirst].ID != 0 {
			c.labels[posFirst].SetStringValue("")
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
	}
	if c.labels[nextPos].ID != 0 {
		c.setLabelPlainText(c.labels[nextPos], nextLine, inactiveColor)
	}
}

// setLabelText sets the label content, using word-by-word coloring when
// YRC word data is available, falling back to plain text.
func (c *darwinController) setLabelText(label cocoa.NSTextField, line LyricLine, timeMs int64, activeColor, inactiveColor cocoa.NSColor) {
	if len(line.Words) > 0 {
		attrStr := c.buildAttributedLine(line, timeMs, activeColor, inactiveColor)
		label.SetAttributedStringValue(attrStr)
		attrStr.Release()
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

// buildAttributedLine creates an NSAttributedString with per-word coloring.
// Played words get activeColor, unplayed get inactiveColor, and the
// currently-playing word is interpolated between the two.
func (c *darwinController) buildAttributedLine(line LyricLine, timeMs int64, activeColor, inactiveColor cocoa.NSColor) cocoa.NSMutableAttributedString {
	// Build plain text from words
	var totalText strings.Builder
	for _, w := range line.Words {
		totalText.WriteString(w.Word)
	}
	plainText := totalText.String()

	attrStr := cocoa.NSMutableAttributedString_alloc().InitWithString(plainText)

	offset := 0
	for _, w := range line.Words {
		runeLen := utf8.RuneCountInString(w.Word)
		if runeLen == 0 {
			continue
		}
		rng := cocoa.NSRange{Location: offset, Length: runeLen}

		var color cocoa.NSColor
		switch {
		case timeMs >= w.EndTime:
			// Fully played
			color = activeColor
		case timeMs < w.StartTime:
			// Not yet played
			color = inactiveColor
		default:
			// Currently playing — interpolate
			wordDuration := w.EndTime - w.StartTime
			t := 1.0
			if wordDuration > 0 {
				t = float64(timeMs-w.StartTime) / float64(wordDuration)
			}
			color = c.blendColor(inactiveColor, activeColor, t)
		}

		attrStr.AddAttribute(cocoa.NSForegroundColorAttributeName, color, rng)
		offset += runeLen
	}

	return attrStr
}

// blendColor creates a new NSColor by linearly interpolating between two colors.
// t=0.0 returns a, t=1.0 returns b.
func (c *darwinController) blendColor(a, b cocoa.NSColor, t float64) cocoa.NSColor {
	// Extract RGB from the hex config (we don't have direct NSColor component access)
	aR, aG, aB := parseHexRGB(c.cfg.TextColor)                 // a = textColor @ inactiveAlpha
	bR, bG, bB := parseHexRGB(c.cfg.TextColor)                 // b = textColor @ 1.0
	aAlpha := inactiveAlpha
	bAlpha := 1.0

	return cocoa.NSColor_ColorWithRedGreenBlueAlpha(
		cocoa.CGFloat(lerp(aR*aAlpha, bR*bAlpha, t)),
		cocoa.CGFloat(lerp(aG*aAlpha, bG*bAlpha, t)),
		cocoa.CGFloat(lerp(aB*aAlpha, bB*bAlpha, t)),
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
