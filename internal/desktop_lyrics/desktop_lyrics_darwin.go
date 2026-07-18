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
	"github.com/go-musicfox/go-musicfox/internal/player"
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

	containerView cocoa.NSView  // LyricsDragView subclass — content container
	contentOrigin cocoa.CGPoint // window's screen origin (bottom-left)
	dragActive    bool          // true during custom mouse drag
	isMoving     bool          // true during AppKit window-level drag (SetMovableByWindowBackground)

	dragStartScreenMX float64 // screen mouse X at drag start
	dragStartScreenMY float64 // screen mouse Y at drag start
	dragStartXFactor  float64 // xFactor at drag start
	dragStartYFactor  float64 // yFactor at drag start
	currentContentH   float64 // cached content height for drag calculations

	// LyricsX-style screen-relative center positioning (0–1).
	// These factors survive resizes; absolute origin is derived from them.
	xFactor float64
	yFactor float64

	// Animation state
	scroll        [2]*scrollState // [0]=posFirst, [1]=posSecond
	wordAnimation wordAnimationState
	animating     bool
	nextFrameAt   time.Time // scheduled deadline of the next animation frame

	// GPU spectrum visualization (Core Animation CALayer sublayers of bgView.layer)
	spectrumBars      []cocoa.CALayer      // one CALayer per frequency band (bar/mirror style)
	spectrumLineLayer cocoa.CAShapeLayer   // shape layer for line/curve style
	spectrumLinePath  cocoa.NSBezierPath   // reusable bezier path for building the curve
	spectrumFrame     player.SpectrumFrame // latest FFT frame (immutable snapshot)
	spectrumMu        sync.Mutex           // guards spectrumFrame
	rawSamples        player.RawSampleFrame // latest raw PCM snapshot for waveform style
	rawSampleMu       sync.Mutex            // guards rawSamples
	dragOverlay       cocoa.NSView           // transparent overlay for mouse event capture
	// Fire style: peak dots with acceleration falloff.
	firePeakDots   []cocoa.CALayer       // small circle per band for peak indicator
	firePeaks      [64]float64           // peak height per band (0–1)
	fireFalloff    [64]float64           // falloff speed per band
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

	// Cache screen visible frame for coordinate calculations.
	visibleFrame := cocoa.NSScreen_MainScreen().VisibleFrame()
	c.screenFrame = visibleFrame

	// Minimum window width
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

	// Compute initial content size
	labelH := 2*lineH + 2*defaultLineSpacing // needLargeHeight=true initially
	contentW := c.currentWinW
	contentH := float64(lineCount)*labelH + padding*2
	if lineCount == 2 {
		contentH += labelGap
	}
	// Add spectrum height
	if c.cfg.SpectrumEnabled {
		if sh := c.cfg.SpectrumEffectiveHeight(); sh > 0 {
			contentH += sh + 4.0 // 4px padding between spectrum and lyrics bg
		}
	}
	c.currentContentH = contentH

	// Compute initial window screen position from factors
	originX := c.screenFrame.Origin.X + c.screenFrame.Size.Width*c.xFactor - contentW/2
	originY := c.screenFrame.Origin.Y + c.screenFrame.Size.Height*c.yFactor - contentH/2
	originX, originY = c.constrainWindowOrigin(originX, originY, contentW, contentH)
	c.contentOrigin = cocoa.CGPoint{X: originX, Y: originY}

	// ---- Content-sized borderless window ----
	windowRect := cocoa.NSRect{
		Origin: c.contentOrigin,
		Size:   cocoa.CGSize{Width: contentW, Height: contentH},
	}
	c.window = cocoa.NSWindow_alloc().InitWithContentRectStyleMaskBackingDefer(
		windowRect,
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
	c.window.SetIgnoresMouseEvents(!c.cfg.Draggable)
	c.window.SetMovableByWindowBackground(c.cfg.Draggable)

	// Set the helper instance as window delegate so windowDidMove: fires
	// and syncPositionFactors keeps xFactor/yFactor current during drags.
	selSetDelegate := objc.RegisterName("setDelegate:")
	c.window.Send(selSetDelegate, helperInst.ID)

	// Content view (fills window)
	contentView := c.window.ContentView()
	contentView.SetWantsLayer(true)

	// ---- LyricsDragView container (at 0,0 within window) ----
	dragViewID := objc.ID(dragViewClass).Send(macdriver.SEL_alloc).Send(macdriver.SEL_init)
	c.containerView = cocoa.NSView{NSObject: core.NSObject{ID: dragViewID}}
	c.containerView.SetWantsLayer(true)

	c.containerView.SetFrameOrigin(0, 0)
	c.containerView.SetFrameSize(contentW, contentH)

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

	// ---- bgView inside container (at 0,0, same size) ----
	bgRect := cocoa.NSRect{
		Origin: cocoa.CGPoint{X: 0, Y: 0},
		Size:   cocoa.CGSize{Width: contentW, Height: contentH},
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

	// ---- Spectrum bars (sublayers of bgView's layer, below lyrics) ----
	if c.cfg.SpectrumEnabled {
		if sh := c.cfg.SpectrumEffectiveHeight(); sh > 0 {
			barCount := c.cfg.SpectrumEffectiveBarCount()
			if c.cfg.SpectrumStyle == "line" || c.cfg.SpectrumStyle == "waveform" || c.cfg.SpectrumStyle == "ring_arc" || c.cfg.SpectrumStyle == "ripple" {
				// Shape-based styles: single CAShapeLayer draws curves/arcs/rings.
				c.spectrumLinePath = cocoa.NSBezierPath_New()
				c.spectrumLineLayer = cocoa.CAShapeLayer_New()
				bgLayer.Send(sel_addSublayer, c.spectrumLineLayer.ID)
			} else {
				totalBars := barCount
				if c.cfg.SpectrumStyle == "mirror" {
					totalBars = barCount * 2 // upper + lower bars
				}
				c.spectrumBars = make([]cocoa.CALayer, totalBars)
				for i := 0; i < totalBars; i++ {
					bar := cocoa.CALayer_New()
					c.spectrumBars[i] = bar
					bgLayer.Send(sel_addSublayer, bar.ID)
			}

			// Fire style: create peak indicator dots (small circles above bars).
			if c.cfg.SpectrumStyle == "fire" {
				c.firePeakDots = make([]cocoa.CALayer, barCount)
				for i := 0; i < barCount; i++ {
					dot := cocoa.CALayer_New()
					dot.SetCornerRadius(2.0)
					c.firePeakDots[i] = dot
					bgLayer.Send(sel_addSublayer, dot.ID)
				}
			}
		}
	}
	}

	// ---- Labels ----
	textW := contentW - padding*2
	textH := labelH
	// Labels sit above spectrum area when spectrum is enabled.
	specBottomPad := 0.0
	sh := c.cfg.SpectrumEffectiveHeight()
	if c.cfg.SpectrumEnabled && sh > 0 {
		specBottomPad = sh + 4.0
	}
	if !c.cfg.OneLineMode {
		baseFirst := specBottomPad + padding + textH + labelGap
		baseSecond := specBottomPad + padding
		c.labelBaseY[posFirst] = baseFirst
		c.labelBaseY[posSecond] = baseSecond
		c.labels[posFirst] = c.makeTextField(baseFirst, padding, textW, textH, inactiveColor, shadow)
		c.labels[posFirst].SetAlignment(nsTextAlignmentLeft)
		c.labels[posSecond] = c.makeTextField(specBottomPad+padding, padding, textW, textH, inactiveColor, shadow)
		c.labels[posSecond].SetAlignment(nsTextAlignmentRight)
		c.bgView.AddSubview(c.labels[posFirst].NSView)
		c.bgView.AddSubview(c.labels[posSecond].NSView)
	} else {
		c.labelBaseY[posSecond] = specBottomPad + padding
		c.labels[posSecond] = c.makeTextField(specBottomPad+padding, padding, textW, textH, inactiveColor, shadow)
		c.labels[posSecond].SetAlignment(nsTextAlignmentCenter)
		c.bgView.AddSubview(c.labels[posSecond].NSView)
	}

	contentView.AddSubview(c.containerView)

	// ---- Transparent drag overlay (on top, handles all mouse events) ----
	overlayID := objc.ID(dragViewClass).Send(macdriver.SEL_alloc).Send(macdriver.SEL_init)
	c.dragOverlay = cocoa.NSView{NSObject: core.NSObject{ID: overlayID}}
	c.dragOverlay.SetWantsLayer(true)
	c.dragOverlay.SetFrameSize(contentW, contentH)
	c.dragOverlay.SetFrameOrigin(0, 0)
	c.containerView.AddSubview(c.dragOverlay)

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
	// Use LyricsTextField (NSTextField subclass) with mouseDownCanMoveWindow=YES
	// so AppKit's window-level drag also works from label-covered areas.
	id := objc.ID(textFieldClass).Send(macdriver.SEL_alloc)
	tf := cocoa.NSTextField{
		NSView: cocoa.NSView{
			NSObject: core.NSObject{ID: id},
		},
	}.InitWithFrame(cocoa.NSRect{
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

// constrainWindowOrigin clamps the window's screen origin so the content
// stays within the screen's visible frame.
func (c *darwinController) constrainWindowOrigin(originX, originY, width, height float64) (float64, float64) {
	minX := c.screenFrame.Origin.X
	maxX := c.screenFrame.Origin.X + c.screenFrame.Size.Width - width
	if originX < minX {
		originX = minX
	} else if originX > maxX {
		originX = maxX
	}
	minY := c.screenFrame.Origin.Y + 4
	maxY := c.screenFrame.Origin.Y + c.screenFrame.Size.Height - height
	if originY < minY {
		originY = minY
	} else if originY > maxY {
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
	c.window.OrderFront(0) // was MakeKeyAndOrderFront

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
	if c.dragOverlay.ID != 0 {
		c.dragOverlay.RemoveFromSuperview()
		c.dragOverlay.Release()
		c.dragOverlay.SetObjcID(0)
	}
	if c.bgView.ID != 0 {
		c.bgView.RemoveFromSuperview()
		c.bgView.Release()
		c.bgView.SetObjcID(0)
	}
	// Clean up spectrum layers
	for i := range c.spectrumBars {
		if c.spectrumBars[i].ID != 0 {
			c.spectrumBars[i].Release()
			c.spectrumBars[i].SetObjcID(0)
		}
	}
	c.spectrumBars = nil
	// Fire style peak dots
	for i := range c.firePeakDots {
		if c.firePeakDots[i].ID != 0 {
			c.firePeakDots[i].Release()
			c.firePeakDots[i].SetObjcID(0)
		}
	}
	c.firePeakDots = nil
	if c.spectrumLineLayer.ID != 0 {
		c.spectrumLineLayer.RemoveFromSuperlayer()
		c.spectrumLineLayer.Release()
		c.spectrumLineLayer.SetObjcID(0)
	}
	if c.spectrumLinePath.ID != 0 {
		c.spectrumLinePath.Release()
		c.spectrumLinePath.SetObjcID(0)
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
		if c.cfg.SpectrumEnabled {
			c.renderSpectrum()
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
	if c.cfg.SpectrumEnabled {
		c.renderSpectrum()
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

// resizeWindow clamps the width and triggers an animated content re-layout.
// The window frame is updated during layoutContent to match the new content size.
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

// layoutContent computes the window frame from current factors and content size,
// then moves/resizes the window. If animate=true, uses NSAnimationContext.
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
	// Add spectrum height
	sh := 0.0
	spectrumPad := 0.0
	if c.cfg.SpectrumEnabled {
		sh = c.cfg.SpectrumEffectiveHeight()
		if sh > 0 {
			spectrumPad = 4.0
			contentH += sh + spectrumPad
		}
	}
	c.currentContentH = contentH

	// Compute new origin from the controller's stored factors, adjusting for any
	// content size change. Uses the factor-preserved center position so
	// the window resizes around its current center point.
	originX := c.screenFrame.Origin.X + c.screenFrame.Size.Width*c.xFactor - contentW/2
	originY := c.screenFrame.Origin.Y + c.screenFrame.Size.Height*c.yFactor - contentH/2
	originX, originY = c.constrainWindowOrigin(originX, originY, contentW, contentH)
	c.contentOrigin = cocoa.CGPoint{X: originX, Y: originY}

	windowRect := cocoa.NSRect{
		Origin: c.contentOrigin,
		Size:   cocoa.CGSize{Width: contentW, Height: contentH},
	}

	// Label positions (relative to bgView's own coordinate system).
	// Labels shift up to make room for spectrum bars at the bottom of bgView.
	textW := contentW - padding*2
	textH := labelH
	specBottomPad := spectrumPad + sh
	baseFirst := specBottomPad + padding + textH + labelGap
	baseSecond := specBottomPad + padding
	if lineCount == 2 {
		c.labelBaseY[posFirst] = baseFirst
	}
	c.labelBaseY[posSecond] = baseSecond

	if animate {
		// Animate window frame and subviews together under a single
		// NSAnimationContext so the bgView's cornerRadius stays in sync
		// with the window bounds — avoids right-edge corner going square.
		cocoa.NSAnimationContext_RunAnimationGroup(func(ctx cocoa.NSAnimationContext) {
			ctx.SetDuration(0.3)
			ctx.SetTimingFunction(cocoa.CAMediaTimingFunction_FunctionWithName(cocoa.CAMediaTimingFunctionEaseInEaseOut))
			ctx.SetAllowsImplicitAnimation(true)

			c.window.Animator().SetFrameDisplayAnimated(windowRect, true)

			c.containerView.Animator().SetFrameSize(contentW, contentH)
			c.bgView.Animator().SetFrameSize(contentW, contentH)
			c.bgView.Animator().SetFrameOrigin(0, 0)
			c.dragOverlay.Animator().SetFrameSize(contentW, contentH)
			c.dragOverlay.Animator().SetFrameOrigin(0, 0)
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
		// Immediate reposition (during drag or initial setup)
		c.window.SetFrameDisplayTopLeft(windowRect, true)

		c.containerView.SetFrameSize(contentW, contentH)
		c.containerView.SetFrameOrigin(0, 0)
		c.bgView.SetFrameSize(contentW, contentH)
		c.bgView.SetFrameOrigin(0, 0)
		c.dragOverlay.SetFrameSize(contentW, contentH)
		c.dragOverlay.SetFrameOrigin(0, 0)
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

// ---- Drag handlers (delta-based for content-sized window) ----

// handleDragStart records the starting screen mouse position and factors.
func (c *darwinController) handleDragStart(mouseX, mouseY float64) {
	if !c.cfg.Draggable || c.closed {
		return
	}
	// Convert window-relative mouse to screen coordinates
	c.dragStartScreenMX = c.contentOrigin.X + mouseX
	c.dragStartScreenMY = c.contentOrigin.Y + mouseY
	c.dragStartXFactor = c.xFactor
	c.dragStartYFactor = c.yFactor
	c.dragActive = true
}

// handleDragMove updates factors based on mouse delta from drag start.
func (c *darwinController) handleDragMove(mouseX, mouseY float64) {
	if !c.dragActive {
		return
	}
	screenMouseX := c.contentOrigin.X + mouseX
	screenMouseY := c.contentOrigin.Y + mouseY

	deltaX := screenMouseX - c.dragStartScreenMX
	deltaY := screenMouseY - c.dragStartScreenMY

	c.xFactor = c.dragStartXFactor + deltaX/c.screenFrame.Size.Width
	c.yFactor = c.dragStartYFactor + deltaY/c.screenFrame.Size.Height

	// Clamp to 0–1
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

	// Snap to center if within 8px
	centerScreenX := c.screenFrame.Origin.X + c.screenFrame.Size.Width/2
	centerScreenY := c.screenFrame.Origin.Y + c.screenFrame.Size.Height/2
	contentCenterX := c.contentOrigin.X + c.currentWinW/2
	contentCenterY := c.contentOrigin.Y + c.currentContentH/2
	if math.Abs(contentCenterX-centerScreenX) < 8 {
		c.xFactor = 0.5
	}
	if math.Abs(contentCenterY-centerScreenY) < 8 {
		c.yFactor = 0.5
	}

	c.layoutContent(false)
}

// handleDragEnd persists factors and cleans up.
func (c *darwinController) handleDragEnd() {
	if !c.dragActive {
		return
	}
	c.dragActive = false
	c.persistPositionFactors()
}

// syncPositionFactors recomputes xFactor/yFactor from the window's current
// screen frame. Called when AppKit moves the window (via windowDidMove:),
// keeping factor-based positioning in sync regardless of how the window moves.
// Persistence is debounced to avoid excessive DB writes during drag.
func (c *darwinController) syncPositionFactors() {
	if c.window.ID == 0 {
		return
	}
	// Skip if a custom drag is active — handleDragMove already updates factors.
	if c.dragActive {
		return
	}
	frame := c.window.Frame()
	originX := float64(frame.Origin.X)
	originY := float64(frame.Origin.Y)
	contentW := float64(frame.Size.Width)
	contentH := float64(frame.Size.Height)
	if c.screenFrame.Size.Width == 0 || c.screenFrame.Size.Height == 0 {
		return
	}

	// Reverse the layoutContent calculation:
	//   originX = screenOrigin.X + screenW*xFactor - contentW/2
	//   => xFactor = (originX + contentW/2 - screenOrigin.X) / screenW
	c.xFactor = (originX + contentW/2 - c.screenFrame.Origin.X) / c.screenFrame.Size.Width
	c.yFactor = (originY + contentH/2 - c.screenFrame.Origin.Y) / c.screenFrame.Size.Height
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
	c.contentOrigin = cocoa.CGPoint{X: originX, Y: originY}
	c.currentContentH = contentH

	// Debounce persistence: cancel previous, schedule after 0.5s
	scheduleDebouncedPersist()
}

// persistPositionFactors saves the current screen-relative center factors to DB.
func (c *darwinController) persistPositionFactors() {
	screen := c.window.Screen()
	screenID := screen.DisplayID()
	if screen.ID == 0 || screenID == 0 {
		return
	}
	// Update screen frame cache (contentOrigin is now window's screen origin).
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

// catmullRomToCubicBezier converts Catmull-Rom control points P0,P1,P2,P3 into
// cubic Bezier control points for the segment P1→P2. Uses the standard centripetal
// parameterization that produces C1-continuous curves through all control points.
//   cp1 = P1 + (P2 - P0) / 6
//   cp2 = P2 - (P3 - P1) / 6
func catmullRomToCubicBezier(p0x, p0y, p1x, p1y, p2x, p2y, p3x, p3y float64) (cp1x, cp1y, cp2x, cp2y float64) {
	cp1x = p1x + (p2x-p0x)/6
	cp1y = p1y + (p2y-p0y)/6
	cp2x = p2x - (p3x-p1x)/6
	cp2y = p2y - (p3y-p1y)/6
	return
}

// renderSpectrum updates all CALayer bar frames and colors from the latest spectrum data.
// Must be called on the main thread (from doUpdateText).
func (c *darwinController) renderSpectrum() {
	c.spectrumMu.Lock()
	frame := c.spectrumFrame
	c.spectrumMu.Unlock()

	isLine := c.cfg.SpectrumStyle == "line"
	isWaveform := c.cfg.SpectrumStyle == "waveform"
	isRingArc := c.cfg.SpectrumStyle == "ring_arc"
	isRipple := c.cfg.SpectrumStyle == "ripple"

	// Guard: bars must exist for bar-based styles; shape layer must exist for shape-based styles.
	needsShape := isLine || isWaveform || isRingArc || isRipple
	if needsShape {
		if c.spectrumLineLayer.ID == 0 || c.bgView.ID == 0 {
			return
		}
	} else {
		if len(c.spectrumBars) == 0 || c.bgView.ID == 0 {
			return
		}
	}

	contentW := c.currentWinW
	padding := defaultWindowPadding
	availableW := contentW - padding*2
	effectiveBarCount := c.cfg.SpectrumEffectiveBarCount()
	maxH := c.cfg.SpectrumEffectiveHeight()
	if maxH <= 0 {
		return
	}
	barGap := c.cfg.SpectrumEffectiveBarGap()
	barW := (availableW - barGap*float64(effectiveBarCount-1)) / float64(effectiveBarCount)
	if barW < 1 {
		barW = 1
	}

	// Get colors
	lowR, lowG, lowB := parseHexRGB(c.cfg.SpectrumColorLow)
	midR, midG, midB := parseHexRGB(c.cfg.SpectrumColorMid)
	highR, highG, highB := parseHexRGB(c.cfg.SpectrumColorHigh)
	opacity := c.cfg.SpectrumOpacity
	if opacity <= 0 {
		opacity = 0.8
	}
	opacityCg := cocoa.CGFloat(opacity)

	cocoa.CATransaction_Begin()
	cocoa.CATransaction_SetDisableActions(true)

	// ---- Line/Curve style: CAShapeLayer with bezier curve ----
	if isLine {
		c.renderSpectrumCurve(frame, effectiveBarCount, availableW, maxH,
			float64(padding), barW, barGap,
			lowR, lowG, lowB, midR, midG, midB, highR, highG, highB, opacityCg)
		cocoa.CATransaction_Commit()
		return
	}

	// ---- Waveform style: raw PCM oscilloscope ----
	if isWaveform {
		c.renderWaveform(
			float64(padding), float64(availableW), maxH,
			midR, midG, midB, opacityCg,
		)
		cocoa.CATransaction_Commit()
		return
	}

	// ---- Ring Arc HUD: CAShapeLayer arc segments around a ring ----
	if isRingArc {
		c.renderRingArc(effectiveBarCount, availableW, maxH,
			lowR, lowG, lowB, midR, midG, midB, highR, highG, highB, opacityCg)
		cocoa.CATransaction_Commit()
		return
	}

	// ---- Ripple Rings: concentric pulsing circles with glow ----
	if isRipple {
		c.renderRipple(effectiveBarCount, maxH,
			midR, midG, midB, opacityCg)
		cocoa.CATransaction_Commit()
		return
	}

	// ---- Bar / Mirror / Capsule / Dot / Fire / LED / Circular styles ----
	barBottom := cocoa.CGFloat(padding)
	isMirror := c.cfg.SpectrumStyle == "mirror"
	isCapsule := c.cfg.SpectrumStyle == "capsule"
	isDot := c.cfg.SpectrumStyle == "dot"
	isFire := c.cfg.SpectrumStyle == "fire"
	isLED := c.cfg.SpectrumStyle == "led"
	isCircular := c.cfg.SpectrumStyle == "circular"

	// Dot style: use fewer bands and smaller dots to avoid crowding.
	dotSize := cocoa.CGFloat(4.0)
	halfDot := dotSize / 2
	if isDot {
		if effectiveBarCount > 32 {
			effectiveBarCount = 32
		}
		barW = (availableW - barGap*float64(effectiveBarCount-1)) / float64(effectiveBarCount)
		if barW < 1 {
			barW = 1
		}
	}

	// Thin bar width for deprecated "line" bar style (kept for compat, unused here).
	var lineW float64
	if isLine {
		lineW = 2.0
		if lineW > barW {
			lineW = barW
		}
		lineW = cocoa.CGFloat(lineW)
	}

	// Thin bar width for deprecated "line" bar style (kept for compat, unused here).
	_ = lineW // suppress unused warning

	for i := 0; i < effectiveBarCount; i++ {
		srcIdx := i * player.SpectrumBandCount / effectiveBarCount
		if srcIdx >= player.SpectrumBandCount {
			srcIdx = player.SpectrumBandCount - 1
		}
		level := frame.Levels[srcIdx]
		x := cocoa.CGFloat(padding + float64(i)*(barW+barGap))

		// Compute color based on style
		var r, g, b float64
		if isFire {
			// Fire gradient: bottom red → orange → top yellow/white
			r, g, b = fireColor(level)
		} else if isLED {
			// Retro LED: bright green/amber gradient left→right
			r, g, b = ledColor(float64(srcIdx) / float64(player.SpectrumBandCount-1))
		} else {
			t := float64(srcIdx) / float64(player.SpectrumBandCount-1)
			if t < 0.5 {
				s := t * 2
				r = lerp(lowR, midR, s)
				g = lerp(lowG, midG, s)
				b = lerp(lowB, midB, s)
			} else {
				s := (t - 0.5) * 2
				r = lerp(midR, highR, s)
				g = lerp(midG, highG, s)
				b = lerp(midB, highB, s)
			}
		}
		color := cocoa.NSColor_ColorWithRedGreenBlueAlpha(
			cocoa.CGFloat(r), cocoa.CGFloat(g), cocoa.CGFloat(b), opacityCg,
		)
		cgColor := uintptr(color.CGColorRef())

		if isMirror {
			centreY := barBottom + cocoa.CGFloat(maxH)/2
			halfH := cocoa.CGFloat(level * maxH / 2)
			if halfH < 0.5 {
				halfH = 0.5
			}
			if halfH > cocoa.CGFloat(maxH)/2 {
				halfH = cocoa.CGFloat(maxH) / 2
			}
			bw := cocoa.CGFloat(barW)
			cr := cocoa.CGFloat(0.0)
			if isCapsule {
				cr = bw / 2
			}
			if bar := c.spectrumBars[i]; bar.ID != 0 {
				bar.SetFrame(x, centreY, bw, halfH)
				bar.SetCornerRadius(cr)
				bar.SetBackgroundCGColor(cgColor)
			}
			if bar := c.spectrumBars[effectiveBarCount+i]; bar.ID != 0 {
				bar.SetFrame(x, centreY-halfH, bw, halfH)
				bar.SetCornerRadius(cr)
				bar.SetBackgroundCGColor(cgColor)
			}
		} else if isDot {
			barH := cocoa.CGFloat(level * maxH)
			if barH < 1 {
				barH = 1
			}
			if barH > cocoa.CGFloat(maxH) {
				barH = cocoa.CGFloat(maxH)
			}
			cx := x + cocoa.CGFloat(barW)/2
			dy := barBottom + barH
			if dy > barBottom+cocoa.CGFloat(maxH) {
				dy = barBottom + cocoa.CGFloat(maxH)
			}
			if bar := c.spectrumBars[i]; bar.ID != 0 {
				bar.SetFrame(cx-halfDot, dy-halfDot, dotSize, dotSize)
				bar.SetCornerRadius(halfDot)
				bar.SetBackgroundCGColor(cgColor)
			}
		} else {
			barH := cocoa.CGFloat(level * maxH)
			if isLED {
				// LED: quantize bar height into discrete steps.
				const ledSteps = 8.0
				barH = cocoa.CGFloat(math.Round(float64(barH)/maxH*ledSteps) / ledSteps * maxH)
				if barH > 0 && barH < cocoa.CGFloat(maxH/ledSteps) {
					barH = cocoa.CGFloat(maxH / ledSteps) // always show at least one block
				}
			}
			if barH < 0.5 {
				barH = 0.5
			}
			if barH > cocoa.CGFloat(maxH) {
				barH = cocoa.CGFloat(maxH)
			}
			w := cocoa.CGFloat(barW)
			bx := x
			cr := cocoa.CGFloat(0.0)
			if isCapsule || isLED {
				cr = w / 2
			}
			if isCircular {
				// Circular: bars radiate outward from centre. Each bar is drawn as a
				// thin line from inner to outer radius (no rotation needed).
				centreX := availableW/2 + padding
				centreY := maxH/2 + float64(padding)
				innerR := min(availableW, maxH) / 2 * 0.42
				outerBase := min(availableW, maxH) / 2 * 0.95
				r := innerR + level*(outerBase-innerR)
				angle := float64(i)/float64(effectiveBarCount)*2.0*math.Pi - math.Pi/2
				cosA := math.Cos(angle)
				sinA := math.Sin(angle)
				bw := 3.0
				if barW < 3 {
					bw = barW
				}
				bh := r - innerR
				if bh < 1 {
					bh = 1
				}
				barMidX := centreX + cosA*innerR + cosA*bh/2
				barMidY := centreY + sinA*innerR + sinA*bh/2
				if bar := c.spectrumBars[i]; bar.ID != 0 {
					bar.SetFrame(
						cocoa.CGFloat(barMidX-bw/2),
						cocoa.CGFloat(barMidY-bh/2),
						cocoa.CGFloat(bw),
						cocoa.CGFloat(bh),
					)
					bar.SetCornerRadius(cocoa.CGFloat(bw / 2))
					bar.SetBackgroundCGColor(cgColor)
				}
				continue
			}
			if bar := c.spectrumBars[i]; bar.ID != 0 {
				bar.SetFrame(bx, barBottom, w, barH)
				bar.SetCornerRadius(cr)
				bar.SetBackgroundCGColor(cgColor)
			}
		}
	}

	// Fire peak dots: draw + animate falloff after bars.
	if isFire {
		for i := 0; i < effectiveBarCount; i++ {
			srcIdx := i * player.SpectrumBandCount / effectiveBarCount
			if srcIdx >= player.SpectrumBandCount {
				srcIdx = player.SpectrumBandCount - 1
			}
			level := frame.Levels[srcIdx]
			peakH := level * maxH
			if peakH > c.firePeaks[srcIdx] {
				c.firePeaks[srcIdx] = peakH
				c.fireFalloff[srcIdx] = 2.0
			}
			c.firePeaks[srcIdx] -= c.fireFalloff[srcIdx]
			c.fireFalloff[srcIdx] *= 1.08
			if c.firePeaks[srcIdx] < 0 {
				c.firePeaks[srcIdx] = 0
			}
			if c.fireFalloff[srcIdx] > 20 {
				c.fireFalloff[srcIdx] = 20
			}
			if i < len(c.firePeakDots) {
				dot := c.firePeakDots[i]
				if dot.ID != 0 {
					dotY := barBottom + cocoa.CGFloat(c.firePeaks[srcIdx]) - 2
					dotX := cocoa.CGFloat(padding + float64(i)*(barW+barGap))
					dot.SetFrame(dotX, dotY, 4, 4)
					dot.SetBackgroundCGColor(uintptr(cocoa.NSColor_ColorWithRedGreenBlueAlpha(
						cocoa.CGFloat(1.0), cocoa.CGFloat(1.0), cocoa.CGFloat(1.0), opacityCg,
					).CGColorRef()))
				}
			}
		}
	}

	cocoa.CATransaction_Commit()
}

const (
	waveformTargetPoints = 200 // number of sample points drawn in waveform style
)

// fireColor returns RGB for fire style: base red → orange → yellow/white as level increases.
func fireColor(level float64) (r, g, b float64) {
	if level < 0.25 {
		t := level / 0.25
		return lerp(0.6, 0.9, t), lerp(0.0, 0.25, t), 0.0
	}
	if level < 0.5 {
		t := (level - 0.25) / 0.25
		return lerp(0.9, 1.0, t), lerp(0.25, 0.6, t), 0.0
	}
	if level < 0.75 {
		t := (level - 0.5) / 0.25
		return 1.0, lerp(0.6, 0.9, t), lerp(0.0, 0.1, t)
	}
	t := (level - 0.75) / 0.25
	return 1.0, lerp(0.9, 1.0, t), lerp(0.1, 0.5, t)
}

// ledColor returns retro LED green/amber palette based on position (0=left, 1=right).
func ledColor(t float64) (r, g, b float64) {
	if t < 0.5 {
		s := t * 2
		return lerp(0.0, 0.2, s), lerp(0.6, 1.0, s), lerp(0.1, 0.2, s)
	}
	s := (t - 0.5) * 2
	return lerp(0.2, 1.0, s), lerp(1.0, 0.5, s), lerp(0.2, 0.0, s)
}

// renderRingArc draws spectrum as arc segments around a circle (HUD/radar style).
func (c *darwinController) renderRingArc(
	bandCount int, availableW, maxH float64,
	lowR, lowG, lowB, midR, midG, midB, highR, highG, highB float64,
	opacityCg cocoa.CGFloat,
) {
	shapeLayer := c.spectrumLineLayer
	shapeLayer.SetFrame(0, 0, cocoa.CGFloat(availableW), cocoa.CGFloat(maxH))

	radius := min(availableW, maxH) / 2 * 0.9
	cx := availableW / 2
	cy := maxH / 2

	path := c.spectrumLinePath
	path.RemoveAllPoints()

	for i := 0; i < bandCount; i++ {
		si := i * player.SpectrumBandCount / bandCount
		if si >= player.SpectrumBandCount {
			si = player.SpectrumBandCount - 1
		}
		level := c.spectrumFrame.Levels[si]
		r := radius * (0.45 + level*0.55) // radius expands with level

		segAngle := (2.0 * math.Pi / float64(bandCount)) * 0.8 // 80% of slot (gap)
		startAng := float64(i)/float64(bandCount)*2.0*math.Pi - math.Pi/2
		endAng := startAng + segAngle

		// Simple line from inner to outer radius
		cosS := math.Cos(startAng)
		sinS := math.Sin(startAng)
		cosE := math.Cos(endAng)
		sinE := math.Sin(endAng)

		x1 := cx + radius*0.4*cosS
		y1 := cy + radius*0.4*sinS
		x2 := cx + r*cosE
		y2 := cy + r*sinE

		if i == 0 {
			path.MoveToPoint(cocoa.CGFloat(x1), cocoa.CGFloat(y1))
		}
		path.LineToPoint(cocoa.CGFloat(x2), cocoa.CGFloat(y2))
	}

	t := float64(bandCount/2) / float64(bandCount)
	r, g, b := ledColor(t)
	strokeColor := cocoa.NSColor_ColorWithRedGreenBlueAlpha(
		cocoa.CGFloat(r), cocoa.CGFloat(g), cocoa.CGFloat(b), opacityCg,
	)
	shapeLayer.SetStrokeCGColor(uintptr(strokeColor.CGColorRef()))
	shapeLayer.SetFillCGColor(0)
	shapeLayer.SetLineWidth(2.5)
	shapeLayer.SetPath(path.CGPath())
}

// renderRipple draws concentric pulsing circles with CALayer shadow for glow.
func (c *darwinController) renderRipple(
	bandCount int, maxH float64,
	r, g, b float64, opacityCg cocoa.CGFloat,
) {
	shapeLayer := c.spectrumLineLayer
	shapeLayer.SetFrame(0, 0, cocoa.CGFloat(maxH*2), cocoa.CGFloat(maxH))

	rings := min(bandCount, 12)
	cx := maxH
	cy := maxH / 2
	maxRadius := maxH * 0.7

	path := c.spectrumLinePath
	path.RemoveAllPoints()

	for i := 0; i < rings; i++ {
		si := i * player.SpectrumBandCount / rings
		if si >= player.SpectrumBandCount {
			si = player.SpectrumBandCount - 1
		}
		level := c.spectrumFrame.Levels[si]
		rr := maxRadius * (0.4 + level*0.6)
		rect := cocoa.NSRect{
			Origin: cocoa.CGPoint{X: cocoa.CGFloat(cx - rr), Y: cocoa.CGFloat(cy - rr)},
			Size:   cocoa.CGSize{Width: cocoa.CGFloat(rr * 2), Height: cocoa.CGFloat(rr * 2)},
		}
		path.AppendBezierPathWithOvalInRect(rect)
	}

	strokeColor := cocoa.NSColor_ColorWithRedGreenBlueAlpha(
		cocoa.CGFloat(r), cocoa.CGFloat(g), cocoa.CGFloat(b), opacityCg,
	)
	shapeLayer.SetStrokeCGColor(uintptr(strokeColor.CGColorRef()))
	shapeLayer.SetFillCGColor(0)
	shapeLayer.SetLineWidth(1.5)
	shapeLayer.SetPath(path.CGPath())
}

// renderWaveform draws a raw-PCM oscilloscope waveform using CAShapeLayer.
func (c *darwinController) renderWaveform(
	padding, availableW, maxH float64,
	r, g, b float64, opacityCg cocoa.CGFloat,
) {
	c.rawSampleMu.Lock()
	snap := c.rawSamples
	c.rawSampleMu.Unlock()

	if snap.Count < 2 {
		return
	}

	shapeLayer := c.spectrumLineLayer
	shapeLayer.SetFrame(
		cocoa.CGFloat(padding), cocoa.CGFloat(padding),
		cocoa.CGFloat(availableW), cocoa.CGFloat(maxH),
	)
	shapeLayer.SetLineWidth(1.5)
	strokeColor := cocoa.NSColor_ColorWithRedGreenBlueAlpha(
		cocoa.CGFloat(r), cocoa.CGFloat(g), cocoa.CGFloat(b), opacityCg,
	)
	shapeLayer.SetStrokeCGColor(uintptr(strokeColor.CGColorRef()))
	shapeLayer.SetFillCGColor(0)

	// Downsample to fit target point count.
	step := snap.Count / waveformTargetPoints
	if step < 1 {
		step = 1
	}
	halfH := maxH / 2

	path := c.spectrumLinePath
	path.RemoveAllPoints()
	first := true
	for i := 0; i < snap.Count; i += step {
		// Normalize: map [-1,1] → [0, maxH]. centre = halfH.
		sample := snap.SamplesL[i]
		// Clip extreme values.
		if sample > 1 {
			sample = 1
		} else if sample < -1 {
			sample = -1
		}
		y := sample*halfH + halfH
		x := float64(i) / float64(snap.Count) * availableW

		if first {
			path.MoveToPoint(cocoa.CGFloat(x), cocoa.CGFloat(y))
			first = false
		} else {
			path.LineToPoint(cocoa.CGFloat(x), cocoa.CGFloat(y))
		}
	}

	shapeLayer.SetPath(path.CGPath())
}

// renderSpectrumCurve draws the spectrum as a smooth Catmull-Rom curve using
// a single CAShapeLayer + NSBezierPath. The curve connects frequency band peaks
// with cubic Bezier segments for C1 continuity.
func (c *darwinController) renderSpectrumCurve(
	frame player.SpectrumFrame,
	effectiveBarCount int,
	availableW, maxH, padding, barW, barGap float64,
	lowR, lowG, lowB, midR, midG, midB, highR, highG, highB float64,
	opacityCg cocoa.CGFloat,
) {
	if effectiveBarCount < 2 {
		return
	}

	// Collect band peak points (centre-x of each band slot, peak-y).
	type pt struct{ x, y float64 }
	points := make([]pt, effectiveBarCount)
	for i := 0; i < effectiveBarCount; i++ {
		srcIdx := i * player.SpectrumBandCount / effectiveBarCount
		if srcIdx >= player.SpectrumBandCount {
			srcIdx = player.SpectrumBandCount - 1
		}
		level := frame.Levels[srcIdx]
		points[i].x = padding + float64(i)*(barW+barGap) + barW/2
		points[i].y = padding + level*maxH
	}

	// Use mid-frequency color for the curve stroke.
	strokeR := lerp(lowR, midR, 0.5)
	strokeG := lerp(lowG, midG, 0.5)
	strokeB := lerp(lowB, midB, 0.5)
	strokeColor := cocoa.NSColor_ColorWithRedGreenBlueAlpha(
		cocoa.CGFloat(strokeR), cocoa.CGFloat(strokeG), cocoa.CGFloat(strokeB), opacityCg,
	)
	// Set up the shape layer frame to cover the spectrum area.
	shapeLayer := c.spectrumLineLayer
	shapeLayer.SetFrame(
		cocoa.CGFloat(padding), cocoa.CGFloat(padding),
		cocoa.CGFloat(availableW), cocoa.CGFloat(maxH),
	)
	shapeLayer.SetLineWidth(2.0)
	shapeLayer.SetStrokeCGColor(uintptr(strokeColor.CGColorRef()))
	// Clear fill: curve-only, no filled area beneath.
	shapeLayer.SetFillCGColor(0)

	// Build bezier path in the shape layer's own coordinate space (origin = bottom-left of spectrum area).
	path := c.spectrumLinePath
	path.RemoveAllPoints()

	// Y in layer coords: 0 = bottom of spectrum, maxH = top.
	rel := func(px, py float64) (cocoa.CGFloat, cocoa.CGFloat) {
		return cocoa.CGFloat(px - padding), cocoa.CGFloat(py - padding)
	}

	rx, ry := rel(points[0].x, points[0].y)
	path.MoveToPoint(rx, ry)

	for i := 0; i < effectiveBarCount-1; i++ {
		var p0x, p0y, p1x, p1y, p2x, p2y, p3x, p3y float64

		if i == 0 {
			p0x, p0y = points[0].x, points[0].y
		} else {
			p0x, p0y = points[i-1].x, points[i-1].y
		}
		p1x, p1y = points[i].x, points[i].y
		p2x, p2y = points[i+1].x, points[i+1].y
		if i+2 < effectiveBarCount {
			p3x, p3y = points[i+2].x, points[i+2].y
		} else {
			p3x, p3y = points[i+1].x, points[i+1].y
		}

		cp1x, cp1y, cp2x, cp2y := catmullRomToCubicBezier(p0x, p0y, p1x, p1y, p2x, p2y, p3x, p3y)
		cp1rx, cp1ry := rel(cp1x, cp1y)
		cp2rx, cp2ry := rel(cp2x, cp2y)
		e2rx, e2ry := rel(p2x, p2y)
		path.CurveToPoint(e2rx, e2ry, cp1rx, cp1ry, cp2rx, cp2ry)
	}

	// Curve-only: no bottom fill, path is an open stroke.
	shapeLayer.SetPath(path.CGPath())
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

func (c *darwinController) UpdateSpectrum(frame player.SpectrumFrame) {
	if c == nil {
		return
	}
	c.spectrumMu.Lock()
	c.spectrumFrame = frame
	c.spectrumMu.Unlock()
	// Trigger a main-thread render (reuse the existing updateText path)
	c.pendingMu.Lock()
	if c.closed || c.window.ID == 0 || !c.visible {
		c.pendingMu.Unlock()
		return
	}
	c.pendingMu.Unlock()
	dispatchAsync(sel_updateText)
}

func (c *darwinController) UpdateRawSamples(snap player.RawSampleFrame) {
	if c == nil || c.cfg.SpectrumStyle != "waveform" {
		return
	}
	c.rawSampleMu.Lock()
	c.rawSamples = snap
	c.rawSampleMu.Unlock()
	c.pendingMu.Lock()
	if c.closed || c.window.ID == 0 || !c.visible {
		c.pendingMu.Unlock()
		return
	}
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
