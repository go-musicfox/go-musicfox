package model

import (
	"image/color"
	"math"
	"strconv"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/anhoder/foxful-cli/layout"
	"github.com/anhoder/foxful-cli/style"
	"github.com/anhoder/foxful-cli/util"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"
	"github.com/fogleman/ease"
	"github.com/lucasb-eyer/go-colorful"
)

// StartupAnimation selects the visual treatment for StartupPage.
// All modes use text and ANSI colors only, so they work without terminal image
// protocols. Terminals without color support automatically render the final,
// motionless frame.
type StartupAnimation string

const (
	// StartupAnimationSequence is the default game/IDE-like boot sequence. It
	// combines fade, rainbow sweep, a short glitch transition, and staged status text.
	StartupAnimationSequence StartupAnimation = "sequence"
	// StartupAnimationFadeIn reveals the logo with a terminal-friendly dithered
	// opacity approximation and a dim-to-bright color ramp.
	StartupAnimationFadeIn StartupAnimation = "fade-in"
	// StartupAnimationRainbowWave moves a rainbow hue wave through the logo.
	StartupAnimationRainbowWave StartupAnimation = "rainbow-wave"
	// StartupAnimationSpinner displays the logo with a custom animated spinner.
	StartupAnimationSpinner StartupAnimation = "spinner"
	// StartupAnimationSlideIn slides the logo in from the right with an elastic
	// easing curve.
	StartupAnimationSlideIn StartupAnimation = "slide-in"
	// StartupAnimationGlitch applies deterministic character corruption and RGB
	// color separation before settling on the logo.
	StartupAnimationGlitch StartupAnimation = "glitch"
	// StartupAnimationMatrixRain draws a Matrix-style character rain behind the
	// logo. It is intentionally opt-in because it redraws the full viewport.
	StartupAnimationMatrixRain StartupAnimation = "matrix-rain"
	// StartupAnimationParticleBurst draws particles that converge into the logo.
	// It is intentionally opt-in because it redraws the full viewport.
	StartupAnimationParticleBurst StartupAnimation = "particle-burst"
)

const (
	startupSpinnerFrames = "◐◓◑◒"
	asciiSpinnerFrames   = "|/-\\"
)

type startupLogoEffect uint8

const (
	logoStatic startupLogoEffect = iota
	logoFade
	logoRainbow
	logoGlitch
)

// animationProgress returns the semantic, non-eased progress. It deliberately
// never exposes a value outside [0, 1], including when LoadingDuration is zero.
func (s *StartupPage) animationProgress() float64 {
	if s.options.LoadingDuration <= 0 {
		return 1
	}
	return min(1, max(0, float64(s.loadedDuration)/float64(s.options.LoadingDuration)))
}

func (s *StartupPage) animationEnabled() bool {
	return !s.options.ReducedMotion && util.TermProfile > colorprofile.ASCII
}

func (s *StartupPage) animationFrame() int {
	if s.options.TickDuration <= 0 {
		return 0
	}
	return int(s.loadedDuration / s.options.TickDuration)
}

func (s *StartupPage) startupLogoSource(a *App) string {
	logo := util.GetAlphaAscii(s.options.Welcome)
	if strings.TrimSpace(logo) == "" {
		logo = s.options.Welcome
	}
	// The block logo has no scalable form. A plain name is a predictable compact
	// fallback instead of relying on the renderer to cut glyphs at the edge.
	if layout.Width(logo) > max(1, a.WindowWidth()-4) || lipgloss.Height(logo) > max(1, a.WindowHeight()-7) {
		return s.options.Welcome
	}
	return logo
}

func (s *StartupPage) logoEffect() (startupLogoEffect, float64, bool) {
	p := s.animationProgress()
	if !s.animationEnabled() {
		return logoStatic, 1, false
	}

	switch s.options.Animation {
	case StartupAnimationFadeIn:
		return logoFade, p, false
	case StartupAnimationRainbowWave:
		return logoRainbow, p, false
	case StartupAnimationGlitch:
		return logoGlitch, p, false
	case StartupAnimationSlideIn:
		return logoRainbow, p, true
	case StartupAnimationSpinner:
		return logoRainbow, p, false
	case StartupAnimationMatrixRain, StartupAnimationParticleBurst:
		return logoFade, min(1, p*1.8), false
	case StartupAnimationSequence, "":
		// A short staged boot: color fades in, then a chromatic sweep and a
		// brief glitch hand-off to the main page.
		switch {
		case p < .38:
			return logoFade, p / .38, false
		case p < .82:
			return logoRainbow, (p - .38) / .44, false
		case p < .94:
			return logoGlitch, (p - .82) / .12, false
		default:
			return logoStatic, 1, false
		}
	default:
		return logoStatic, 1, false
	}
}

func (s *StartupPage) animatedLogoView(a *App) string {
	logo := s.startupLogoSource(a)
	effect, progress, slide := s.logoEffect()
	rendered := renderStartupLogo(logo, effect, progress, s.animationFrame())
	return s.positionAnimatedLogo(a, rendered, slide, progress)
}

func (s *StartupPage) positionAnimatedLogo(a *App, rendered string, slide bool, progress float64) string {
	if slide && s.animationEnabled() {

		// EaseOutElastic intentionally overshoots a little. Clamping keeps the
		// content in the terminal while preserving the spring-like arrival.
		arrival := min(1, max(0, ease.OutElastic(progress)))
		distance := max(0, a.WindowWidth()/3)
		padding := int(math.Round(float64(distance) * (1 - arrival)))
		rendered = indentLines(rendered, padding)
	}
	rendered = truncateStartupLines(rendered, a.WindowWidth())

	return lipgloss.NewStyle().
		Align(lipgloss.Center).
		Width(a.WindowWidth()).
		Render(rendered)
}

func renderStartupLogo(logo string, effect startupLogoEffect, progress float64, frame int) string {
	lines := strings.Split(logo, "\n")
	total := 0
	for _, line := range lines {
		for _, r := range line {
			if r != ' ' {
				total++
			}
		}
	}
	if total == 0 {
		return logo
	}

	var b strings.Builder
	seen := 0
	for y, line := range lines {
		for x, r := range line {
			if r == ' ' {
				b.WriteRune(r)
				continue
			}
			seen++
			visible := true
			glyph := r
			var fg color.Color = util.GetPrimaryColor()

			switch effect {
			case logoFade:
				// ANSI has no portable foreground alpha. Dither plus a dim-to-bright
				// foreground produces a fade without terminal image support.
				visible = float64(startupHash(x, y, 0)%1000)/1000 < progress
				fg = fadedColor(fg, progress)
			case logoRainbow:
				fg = rainbowColor(float64(x*11+y*7+frame*5) / 2)
			case logoGlitch:
				fg = glitchColor(x, y, frame)
				if progress < .82 && startupHash(x, y, frame)%13 == 0 {
					// Add a cell before a corrupted glyph. This deliberately shifts
					// the rest of that scan line for a frame, producing a text-safe
					// horizontal displacement instead of relying on cursor controls.
					b.WriteRune(' ')
					glyph = []rune("@#$%&*+-")[startupHash(y, x, frame)%8]
				}
			}

			if !visible {
				b.WriteRune(' ')
				continue
			}
			b.WriteString(startupPaint(string(glyph), fg))
		}
		if y < len(lines)-1 {
			b.WriteByte('\n')
		}
	}
	return b.String()
}

func (s *StartupPage) startupStatusView(a *App) string {
	p := s.animationProgress()
	stage := "Initializing interface"
	if s.options.Animation == StartupAnimationSequence || s.options.Animation == "" {
		switch {
		case p < .26:
			stage = "Loading runtime"
		case p < .56:
			stage = "Building interface"
		case p < .86:
			stage = "Syncing colors"
		default:
			stage = "Ready"
		}
	}

	spinner := ""
	if s.options.Animation == StartupAnimationSpinner || s.options.Animation == StartupAnimationSequence || s.options.Animation == "" {
		frames := []rune(startupSpinnerFrames)
		if util.TermProfile <= colorprofile.ASCII {
			frames = []rune(asciiSpinnerFrames)
		}
		spinner = string(frames[s.animationFrame()%len(frames)]) + " "
	}
	text := spinner + stage + " · " + formatStartupPercent(p)
	text = ansi.TruncateWc(text, max(0, a.WindowWidth()), "")
	return style.CurrentStyleSet().Subtitle.Copy().
		Align(lipgloss.Center).
		Width(a.WindowWidth()).
		Render(text)
}

// truncateStartupLines prevents ANSI-styled animation frames from wrapping on
// narrow terminals before they are centered or composited.
func truncateStartupLines(content string, width int) string {
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		lines[i] = ansi.TruncateWc(line, max(0, width), "")
	}
	return strings.Join(lines, "\n")
}

func formatStartupPercent(progress float64) string {
	return formatFloat(progress*100) + "%"
}

// formatFloat keeps the percentage renderer stable for snapshots. Startup
// progress has no need for decimal precision.
func formatFloat(value float64) string {
	return strconv.Itoa(int(math.Round(value)))
}

func startupPaint(content string, fg color.Color) string {
	if util.TermProfile <= colorprofile.ASCII || fg == nil {
		return content
	}
	return lipgloss.NewStyle().Foreground(util.TermProfile.Convert(fg)).Render(content)
}

func fadedColor(fg color.Color, progress float64) color.Color {
	base, ok := colorful.MakeColor(fg)
	if !ok {
		return fg
	}
	background := colorful.Color{R: .1, G: .1, B: .1}
	if !style.HasDarkBackground() {
		background = colorful.Color{R: 1, G: 1, B: 1}
	}
	return base.BlendLab(background, 1-min(1, max(0, progress))).Clamped()
}

func rainbowColor(hue float64) color.Color {
	hue = math.Mod(hue, 360)
	if hue < 0 {
		hue += 360
	}
	return colorful.Hsv(hue, .78, 1).Clamped()
}

func glitchColor(x, y, frame int) color.Color {
	colors := []color.Color{lipgloss.BrightCyan, lipgloss.BrightMagenta, lipgloss.BrightYellow}
	return colors[startupHash(x, y, frame)%len(colors)]
}

func startupHash(x, y, frame int) int {
	v := uint32(x*73856093) ^ uint32(y*19349663) ^ uint32(frame*83492791)
	v ^= v >> 13
	v *= 0x5bd1e995
	return int(v ^ (v >> 15))
}

func indentLines(content string, amount int) string {
	if amount <= 0 {
		return content
	}
	prefix := strings.Repeat(" ", amount)
	return prefix + strings.ReplaceAll(content, "\n", "\n"+prefix)
}

// startupSpecialView owns full-screen effects. The bool is false for the
// normal logo-based modes, allowing StartupPage.View to keep its regular layout.
func (s *StartupPage) startupSpecialView(a *App) (string, bool) {
	if !s.animationEnabled() {
		return "", false
	}
	switch s.options.Animation {
	case StartupAnimationMatrixRain:
		return s.matrixRainView(a), true
	case StartupAnimationParticleBurst:
		return s.particleBurstView(a), true
	default:
		return "", false
	}
}

func (s *StartupPage) matrixRainView(a *App) string {
	width, height := a.WindowWidth(), a.WindowHeight()
	cells := make([][]string, height)
	// Keep rain glyphs one terminal cell wide so the cell grid remains aligned
	// even in terminals with different East Asian-width settings.
	glyphs := []rune("01ABCDEFGHIJKLMNOPQRSTUVWXYZ<>[]{}#%")
	frame := s.animationFrame()
	for y := 0; y < height; y++ {
		cells[y] = make([]string, width)
		for x := 0; x < width; x++ {
			trail := (x*7 + frame*2) % max(4, height/2)
			delta := (y - trail + height) % height
			if delta < 7 {
				c := color.Color(lipgloss.Green)
				if delta == 0 {
					c = lipgloss.BrightGreen
				}
				cells[y][x] = startupPaint(string(glyphs[startupHash(x, y, frame)%len(glyphs)]), c)
			} else {
				cells[y][x] = " "
			}
		}
	}
	return s.overlayCentered(a, cellLines(cells), s.specialForeground(a, min(1, s.animationProgress()*1.8), frame))
}

func (s *StartupPage) particleBurstView(a *App) string {
	width, height := a.WindowWidth(), a.WindowHeight()
	cells := make([][]string, height)
	for y := range cells {
		cells[y] = make([]string, width)
		for x := range cells[y] {
			cells[y][x] = " "
		}
	}

	p := min(1, s.animationProgress()/.72)
	frame := s.animationFrame()
	cx, cy := width/2, height/2
	// Fixed count and deterministic coordinates avoid visual flicker caused by
	// re-seeding a random generator on every frame.
	for i := 0; i < min(180, max(40, width*height/8)); i++ {
		angle := float64(startupHash(i, 1, 0)%628) / 100
		radius := float64(max(width, height)) / 2 * (1 - p)
		targetX := (startupHash(i, 2, 0) % max(1, width/2)) - width/4
		targetY := (startupHash(i, 3, 0) % max(1, height/2)) - height/4
		x := cx + int(math.Round(math.Cos(angle)*radius+float64(targetX)*p))
		y := cy + int(math.Round(math.Sin(angle)*radius+float64(targetY)*p))
		if x >= 0 && x < width && y >= 0 && y < height {
			cells[y][x] = startupPaint("·", rainbowColor(float64(i*9+frame*8)))
		}
	}
	logoProgress := min(1, max(0, (s.animationProgress()-.3)/.5))
	return s.overlayCentered(a, cellLines(cells), s.specialForeground(a, logoProgress, frame))
}

// specialForeground preserves status and progress feedback for full-screen P2
// effects. The Logo is the only animated layer; the rest remains readable.
func (s *StartupPage) specialForeground(a *App, logoProgress float64, frame int) string {
	logo := truncateStartupLines(
		renderStartupLogo(s.startupLogoSource(a), logoFade, logoProgress, frame),
		a.WindowWidth(),
	)
	return layout.JoinVertical(
		lipgloss.Center,
		logo,
		"",
		s.startupStatusView(a),
		"",
		s.progressView(a),
	)
}

func (s *StartupPage) overlayCentered(a *App, background, foreground string) string {
	x := max(0, (a.WindowWidth()-layout.Width(foreground))/2)
	y := max(0, (a.WindowHeight()-lipgloss.Height(foreground))/2)
	return layout.Overlay(background, foreground, x, y)
}

func cellLines(cells [][]string) string {
	var b strings.Builder
	for y, row := range cells {
		for _, cell := range row {
			b.WriteString(cell)
		}
		if y < len(cells)-1 {
			b.WriteByte('\n')
		}
	}
	return b.String()
}
