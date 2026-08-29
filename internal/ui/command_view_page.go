package ui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/anhoder/foxful-cli/model"
	"github.com/anhoder/foxful-cli/style"

	"github.com/go-musicfox/go-musicfox/internal/configs"
)

// PtCommandView is the page type of the "command_view" scrollable text page
// (S1: the TUI consumer of the frontend "view" action).
const PtCommandView model.PageType = "command_view"

// CommandViewOpts is the parameter contract of the "command_view" page provider
// (RegisterPage). The content is produced by a command execution result — the
// page renders it as static text and never mutates it.
type CommandViewOpts struct {
	Title string
	Lines []string // Message split on "\n"; empty lines are preserved
}

// CommandViewPage renders a scrollable text view. It is a pure display page:
// the content is snapshotted at construction and the scroll offset is clamped
// to the rendered body area. Back-button hit coordinates are recorded during
// View, mirroring the Search/Login page convention.
type CommandViewPage struct {
	opts    CommandViewOpts
	scroll  int  // first visible line offset (0-based), always >= 0
	hovered bool // back button hover (pageBackButtonIcon style switch)

	// Back-button mouse hit coordinates (0-based screen row/column), recorded
	// in View from pageMenuTitleViewWithBack's actual render position.
	backBtnRowY   int
	backBtnStartX int
}

// IgnoreQuitKeyMsg keeps the global quit behavior active: this text page has no
// input box, so "q" (and ctrl+c) always quit.
func (p *CommandViewPage) IgnoreQuitKeyMsg(tea.KeyMsg) bool {
	return false
}

// Type returns the custom page type identifier.
func (p *CommandViewPage) Type() model.PageType {
	return PtCommandView
}

// Msg returns nil: the content is fixed at construction, no async loading.
func (p *CommandViewPage) Msg() tea.Msg {
	return nil
}

// Update handles keyboard and mouse input for the scrollable text view.
func (p *CommandViewPage) Update(msg tea.Msg, a *model.App) (model.Page, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
		switch keyMsg.String() {
		case "up":
			p.scrollBy(-1, a)
		case "down":
			p.scrollBy(1, a)
		case "pgup":
			p.scrollBy(-p.pageStep(a), a)
		case "pgdn", "pgdown":
			p.scrollBy(p.pageStep(a), a)
		case "home":
			p.scroll = 0
		case "end":
			p.scroll = p.maxScroll(a)
		case "left", "esc", "backspace":
			return a.Main(), nil
		}
		return p, nil
	}

	if wheelMsg, ok := msg.(tea.MouseWheelMsg); ok {
		switch wheelMsg.Mouse().Button {
		case tea.MouseWheelUp:
			p.scrollBy(-3, a)
		case tea.MouseWheelDown:
			p.scrollBy(3, a)
		}
		return p, nil
	}

	if motionMsg, ok := msg.(tea.MouseMotionMsg); ok {
		mouse := motionMsg.Mouse()
		oldHovered := p.hovered
		p.hovered = p.hitBackButton(mouse.X, mouse.Y)
		if p.hovered != oldHovered {
			return p, a.RerenderCmd(true)
		}
		return p, nil
	}

	if clickMsg, ok := msg.(tea.MouseClickMsg); ok {
		mouse := clickMsg.Mouse()
		if mouse.Button == tea.MouseLeft && p.hitBackButton(mouse.X, mouse.Y) {
			return a.Main(), nil
		}
		return p, nil
	}

	return p, nil
}

// View renders the scrollable text page.
func (p *CommandViewPage) View(a *model.App) string {
	var (
		builder strings.Builder
		top     int
		main    = a.Main()
	)

	// Title bar / top status bar row (mirrors Search/Login page layout).
	if configs.AppConfig.Theme.ShowTitle {
		builder.WriteString(pageTitleView(a, main, &top))
	} else {
		builder.WriteString("\n")
		top++
	}

	// Menu title with back button, then a blank spacer row.
	title := &model.MenuItem{Title: p.opts.Title}
	topBefore := top
	builder.WriteString(pageMenuTitleViewWithBack(a, main, &top, title, p.hovered))
	p.backBtnRowY = pageMenuTitleRow(a, main, topBefore)
	p.backBtnStartX = max(0, main.MenuStartColumn()-pageBackButtonWidth)
	builder.WriteString("\n\n")

	// Body: render the visible line window, truncating over-wide lines.
	startColumn := max(0, main.MenuStartColumn())
	maxWidth := max(0, a.WindowWidth()-startColumn)
	available := p.visibleRows(a)
	for i := p.scroll; i < len(p.opts.Lines) && available > 0; i++ {
		if startColumn > 0 {
			builder.WriteString(style.CurrentStyleSet().AppBackground.Render(strings.Repeat(" ", startColumn)))
		}
		line := p.opts.Lines[i]
		if lipgloss.Width(line) > maxWidth {
			line = lipgloss.NewStyle().MaxWidth(maxWidth).Render(line)
		}
		builder.WriteString(line)
		builder.WriteString("\n")
		available--
	}

	return FinishCustomPageView(&builder, a)
}

// bodyStartRows returns the number of terminal rows consumed by the title
// section (title bar, menu title row and the blank spacer), i.e. where the
// body area starts.
func (p *CommandViewPage) bodyStartRows(a *model.App) int {
	main := a.Main()
	// Title bar / top status bar always occupies one row.
	top := 1
	top = PageMenuTitleRow(a, main, top)
	return top + 2
}

// visibleRows returns how many body lines fit in the window (the scrollable
// viewport). The extra -1 reserves the trailing row so FinishCustomPageView's
// fillPageHeight lands exactly on the window height.
func (p *CommandViewPage) visibleRows(a *model.App) int {
	return max(0, a.Main().EffectiveWindowHeight(a)-p.bodyStartRows(a)-1)
}

// maxScroll returns the largest scroll offset. The body shows at most
// visibleRows lines, so the first-line offset is clamped to
// [0, len(Lines)-visibleRows].
func (p *CommandViewPage) maxScroll(a *model.App) int {
	visible := p.visibleRows(a)
	if len(p.opts.Lines) <= visible {
		return 0
	}
	return len(p.opts.Lines) - visible
}

// pageStep is the half-page scroll step used by pgup/pgdn.
func (p *CommandViewPage) pageStep(a *model.App) int {
	return max(1, a.Main().EffectiveWindowHeight(a)/2)
}

// scrollBy moves the first-line offset by delta, clamped to the valid range.
func (p *CommandViewPage) scrollBy(delta int, a *model.App) {
	p.scroll = min(max(0, p.scroll+delta), p.maxScroll(a))
}

// hitBackButton reports whether (x, y) is over the back button.
func (p *CommandViewPage) hitBackButton(x, y int) bool {
	return y == p.backBtnRowY && x >= p.backBtnStartX && x < p.backBtnStartX+pageBackButtonWidth
}

// ViewPageContent abstracts the data source of a view page (S8 interaction
// protocol extension point: page content could be provided dynamically by the
// command instead of a construction-time snapshot). CommandViewPage holds
// CommandViewOpts directly and does NOT implement this interface yet; S8 will
// extend the page along this shape.
type ViewPageContent interface {
	Title() string
	Lines() []string
}

// ViewPageHooks is the optional interaction-callback extension point (S8: the
// page reports interactions back to the plugin, e.g. line selection or key
// events). Empty interface this stage, kept as an anchor.
type ViewPageHooks interface{}

func init() {
	RegisterPage("command_view", func(opts CommandViewOpts) (model.Page, error) {
		return &CommandViewPage{opts: opts}, nil
	})
}
