package ui

import (
	"github.com/anhoder/foxful-cli/model"
)

// showConfirmPopup shows a modal Yes/No confirmation popup. The 取消 action is
// placed first so it is the default-focused button: a bare Enter cancels, and
// the user must move focus to 确定 (or click it) to confirm — preventing
// accidental destructive actions. Esc and outside-click also cancel.
//
// onConfirm runs only when 确定 is activated. Callables that schedule deferred
// work via NewOperation().ShowLoading().Execute() (which relies on a tickMainMsg
// to flush) must trigger that flush themselves, e.g. by calling app.Rerender(false)
// at the end of onConfirm — the player ticker is not guaranteed to be running.
func showConfirmPopup(app *model.App, title, content string, onConfirm func()) {
	popup, err := model.NewPopup(model.PopupSpec{
		Title:   title,
		Content: content,
		Actions: []model.PopupAction{
			{ID: "cancel", Label: "取消", IsCancel: true},
			{ID: "confirm", Label: "确定"},
		},
		OnResult: func(r model.PopupResult) {
			if r.ActionID == "confirm" && onConfirm != nil {
				onConfirm()
			}
		},
	})
	if err != nil {
		return
	}
	app.ShowPopup(popup)
}

// ShowConfirmPopup is the exported form of showConfirmPopup (Phase 3.9 plugin
// boundary; the Last.fm menu and profile moved into internal/plugins/lastfm
// and show confirmations through it).
func ShowConfirmPopup(app *model.App, title, content string, onConfirm func()) {
	showConfirmPopup(app, title, content, onConfirm)
}
