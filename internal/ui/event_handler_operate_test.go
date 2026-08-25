package ui

import (
	"context"
	"io"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/anhoder/foxful-cli/model"

	"github.com/go-musicfox/go-musicfox/internal/keybindings"
)

// newOperateTestEventHandler builds an EventHandler backed by a runnable
// model.App (main initialized via app.Run) so handle() can reach the plugin
// dispatch path without panicking on MustMain().
func newOperateTestEventHandler(t *testing.T) *EventHandler {
	t.Helper()

	opts := model.DefaultOptions()
	opts.EnableStartup = false
	opts.MainMenu = &pageLayoutTestMenu{}
	opts.MainMenuTitle = &model.MenuItem{Title: "Main"}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	opts.TeaOptions = []tea.ProgramOption{
		tea.WithContext(ctx),
		tea.WithInput(nil),
		tea.WithOutput(io.Discard),
	}

	app := model.NewApp(opts)
	_ = app.Run()
	if app.Main() == nil {
		t.Fatal("main page was not initialized")
	}
	_, _ = app.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	return &EventHandler{svc: newMenuServices(&Netease{App: app})}
}

// TestHandleDispatchesRegisteredOperate verifies a registered plugin handler
// is invoked with the EventHandler's svc and app, and its result is returned.
func TestHandleDispatchesRegisteredOperate(t *testing.T) {
	h := newOperateTestEventHandler(t)
	op := keybindings.RegisterOperate("test_op_dispatch", "分发测试", []string{"ctrl+t"})

	var (
		called bool
		gotSvc MenuServices
		gotApp *model.App
	)
	RegisterOperateHandler(op, func(svc MenuServices, app *model.App) (model.Page, tea.Cmd) {
		called = true
		gotSvc = svc
		gotApp = app
		return nil, nil
	})

	stop, page, cmd := h.handle(op)
	if !called {
		t.Fatal("registered handler was not called")
	}
	if gotSvc != h.svc {
		t.Fatal("handler received a different svc than the EventHandler's")
	}
	if gotApp != h.svc.App() || gotApp == nil {
		t.Fatal("handler did not receive the EventHandler's app")
	}
	if !stop || page != nil || cmd != nil {
		t.Fatalf("handle(registered op) = (%v, %v, %v), want (true, nil, nil)", stop, page, cmd)
	}
}

// TestHandleUnregisteredOperateFallsBackToDefault verifies an op without a
// registered handler keeps the existing default behavior and does not panic.
func TestHandleUnregisteredOperateFallsBackToDefault(t *testing.T) {
	h := newOperateTestEventHandler(t)
	op := keybindings.RegisterOperate("test_op_unhandled", "未注册处理器", nil)

	stop, page, cmd := h.handle(op)
	if stop || page != nil || cmd != nil {
		t.Fatalf("handle(unregistered op) = (%v, %v, %v), want (false, nil, nil)", stop, page, cmd)
	}
}

// TestRegisterOperateHandlerPanicsOnNil verifies registering a nil handler
// panics (programmer error).
func TestRegisterOperateHandlerPanicsOnNil(t *testing.T) {
	op := keybindings.RegisterOperate("test_op_nil_handler", "nil处理器", nil)
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("RegisterOperateHandler(nil) did not panic")
		}
	}()
	RegisterOperateHandler(op, nil)
}

// TestRegisterOperateHandlerPanicsOnDuplicate verifies registering the same op
// twice panics (programmer error).
func TestRegisterOperateHandlerPanicsOnDuplicate(t *testing.T) {
	op := keybindings.RegisterOperate("test_op_dup_handler", "重复处理器", nil)
	RegisterOperateHandler(op, func(_ MenuServices, _ *model.App) (model.Page, tea.Cmd) { return nil, nil })

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("RegisterOperateHandler(duplicate op) did not panic")
		}
		msg, ok := r.(string)
		if !ok || !strings.Contains(msg, "test_op_dup_handler") {
			t.Fatalf("panic message = %v, want it to mention the duplicate op", r)
		}
	}()
	RegisterOperateHandler(op, func(_ MenuServices, _ *model.App) (model.Page, tea.Cmd) { return nil, nil })
}
