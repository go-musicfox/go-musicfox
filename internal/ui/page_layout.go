package ui

import (
	"strings"

	"charm.land/bubbles/v2/textinput"
	"charm.land/lipgloss/v2"
	"github.com/anhoder/foxful-cli/model"
	"github.com/anhoder/foxful-cli/style"
	"github.com/anhoder/foxful-cli/util"
	"github.com/mattn/go-runewidth"
)

// pageTitleView/pageMenuTitleView 为自定义页面（SearchPage、LoginPage 等）提供简化的标题渲染。
// 这些页面不在 Main 菜单栈中，因此默认不使用 foxful-cli Main.MenuTitleView 的返回按钮功能。
// 菜单栈中的正常菜单应继续使用 Main 的原生 MenuTitleView 方法。

const (
	pageBackButtonWidth = 2 // "←" 宽度，与 foxful-cli backButtonWidth 对齐
	pageInputPrompt     = "> "
)

// pageBackButtonIcon 渲染返回图标，样式随主题/hover 状态切换。
func pageBackButtonIcon(hovered bool) string {
	ss := style.CurrentStyleSet()
	if hovered {
		return ss.BackButtonHover.Render("←")
	}
	return ss.BackButton.Render("←")
}

// pageMenuTitleRow 计算标题所在的 0-based 屏幕行，供命中检测使用。
func pageMenuTitleRow(_ *model.App, main *model.Main, topBefore int) int {
	return max(topBefore, max(1, main.MenuStartRow()-2))
}

func fillPageHeight(builder *strings.Builder, height int) {
	missingLines := height - strings.Count(builder.String(), "\n") - 1
	if missingLines > 0 {
		builder.WriteString(strings.Repeat("\n", missingLines))
	}
}

func finishCustomPageView(builder *strings.Builder, a *model.App) string {
	fillPageHeight(builder, a.WindowHeight())
	return model.RenderAppBackground(builder.String(), a.WindowWidth())
}

func pageSubmitText() string {
	return model.T(model.MsgConfirm)
}

func pageSubmitButton(focused bool) string {
	return pageButton(pageSubmitText(), focused)
}

func pageButton(text string, focused bool) string {
	if focused {
		return "[ " + util.GetPrimaryFontStyle(false).Render(text) + " ]"
	}
	return util.GetBlurredButton(text)
}

func pageInputStyles() textinput.Styles {
	styles := textinput.DefaultStyles(style.HasDarkBackground())
	primary := util.GetPrimaryFontStyle(false)
	styles.Focused.Text = primary
	styles.Focused.Prompt = primary
	styles.Focused.Suggestion = primary
	styles.Cursor.Color = util.GetPrimaryColor()
	return styles
}

func focusPageInput(input *textinput.Model) {
	input.Focus()
	input.Prompt = pageInputPrompt
	input.SetStyles(pageInputStyles())
}

func blurPageInput(input *textinput.Model) {
	input.Blur()
	input.Prompt = pageInputPrompt
	input.SetStyles(pageInputStyles())
}

func pageInputView(input textinput.Model, hovered bool) string {
	if !hovered {
		return input.View()
	}

	styles := input.Styles()
	hoverStyle := util.GetPrimaryFontStyle(false)
	styles.Focused.Text = hoverStyle
	styles.Focused.Prompt = hoverStyle
	styles.Blurred.Text = hoverStyle
	styles.Blurred.Prompt = hoverStyle
	input.SetStyles(styles)
	return input.View()
}

func pageButtonHoverView(text string) string {
	return "[ " + util.GetPrimaryFontStyle(true).Underline(true).Render(text) + " ]"
}

func setPageInputCursor(input *textinput.Model, mouseX, inputStartX int) {
	column := max(0, mouseX-inputStartX-runewidth.StringWidth(input.Prompt))
	width := 0
	position := 0
	for _, r := range input.Value() {
		runeWidth := max(1, runewidth.RuneWidth(r))
		if column < width+runeWidth {
			input.SetCursor(position)
			return
		}
		width += runeWidth
		position++
	}
	input.SetCursor(position)
}

func pageTitleView(a *model.App, main *model.Main, top *int) string {
	(*top)++
	// Render status bar at top (if configured), otherwise title bar
	if main.StatusBar() != nil && main.StatusBarPosition() == model.StatusBarTop {
		return main.StatusBar().View(a, main) + "\n"
	}
	return main.TitleView(a) + "\n"
}

// pageHasTopBreadcrumb reports whether the custom page currently renders a
// clickable top status bar (which carries the breadcrumb). Only in this case
// should custom pages delegate breadcrumb hover/click to Main — otherwise the
// status bar is at the bottom and not rendered on custom pages, so delegating
// would mis-hit the bottom row.
func pageHasTopBreadcrumb(main *model.Main) bool {
	return main.StatusBar() != nil && main.StatusBarPosition() == model.StatusBarTop
}

// pageBreadcrumbMotion delegates a mouse-motion event to Main's breadcrumb
// hover tracking when the top status bar is rendered. Returns whether the
// hovered segment changed (caller should re-render) and whether the cursor is
// over a clickable breadcrumb segment (caller should show a pointer).
func pageBreadcrumbMotion(a *model.App, main *model.Main, x, y int) (changed bool, over bool) {
	if !pageHasTopBreadcrumb(main) {
		return false, false
	}
	return main.UpdateBreadcrumbHover(x, y, a)
}

// pageBreadcrumbClick delegates a left-click to Main's breadcrumb handler when
// the top status bar is rendered. Returns the target page (Main navigated back
// to the clicked ancestor) or nil when no clickable segment was hit.
func pageBreadcrumbClick(a *model.App, main *model.Main, x, y int) model.Page {
	if !pageHasTopBreadcrumb(main) {
		return nil
	}
	return main.HandleBreadcrumbClick(x, y, a)
}

func pageMenuTitleView(a *model.App, main *model.Main, top *int, title *model.MenuItem) string {
	if title == nil {
		title = main.MenuTitle()
	}

	targetRow := max(1, main.MenuStartRow()-3)
	var builder strings.Builder
	if targetRow > *top {
		builder.WriteString(strings.Repeat("\n", targetRow-*top))
		*top = targetRow
	}

	startColumn := max(0, main.MenuStartColumn())
	maxWidth := max(0, a.WindowWidth()-startColumn)
	content := title.String()
	if lipgloss.Width(title.OriginString()) > maxWidth {
		content = lipgloss.NewStyle().MaxWidth(maxWidth).Render(content)
	}

	builder.WriteString(style.CurrentStyleSet().AppBackground.Render(strings.Repeat(" ", startColumn)))
	builder.WriteString(style.CurrentStyleSet().MenuTitle.Render(content))
	return builder.String()
}

// pageMenuTitleViewWithBack 渲染带返回按钮的标题（Search/Login 专用）。
func pageMenuTitleViewWithBack(a *model.App, main *model.Main, top *int, title *model.MenuItem, hovered bool) string {
	if title == nil {
		title = main.MenuTitle()
	}

	targetRow := pageMenuTitleRow(a, main, *top)
	var builder strings.Builder
	if targetRow > *top {
		builder.WriteString(strings.Repeat("\n", targetRow-*top))
		*top = targetRow
	}

	startColumn := max(0, main.MenuStartColumn())
	maxWidth := max(0, a.WindowWidth()-startColumn)
	content := title.String()
	if lipgloss.Width(title.OriginString()) > maxWidth {
		content = lipgloss.NewStyle().MaxWidth(maxWidth).Render(content)
	}

	padding := max(0, main.MenuStartColumn()-pageBackButtonWidth)
	builder.WriteString(style.CurrentStyleSet().AppBackground.Render(strings.Repeat(" ", padding)))
	builder.WriteString(pageBackButtonIcon(hovered))
	builder.WriteString(" ")
	builder.WriteString(style.CurrentStyleSet().MenuTitle.Render(content))
	return builder.String()
}
