package ui

import (
	"fmt"
	"sync"

	tea "charm.land/bubbletea/v2"
	"github.com/anhoder/foxful-cli/model"

	"github.com/go-musicfox/go-musicfox/internal/keybindings"
)

// operateHandlerFunc 是插件操作的处理函数：在对应快捷键触发时经 svc
// （MenuServices）执行，返回页面与命令。
type operateHandlerFunc func(svc MenuServices, app *model.App) (model.Page, tea.Cmd)

// operateHandlers 是插件操作 -> 处理函数的注册表。插件在 init() 期注册
// （先于事件循环），EventHandler.handle 的 default 分支经 handlePluginOperate
// 查询分发。RWMutex 防御性保护读写，避免未来运行时注册引入数据竞争。
var (
	operateHandlersMu sync.RWMutex
	operateHandlers   = make(map[keybindings.OperateType]operateHandlerFunc)
)

// RegisterOperateHandler 注册插件操作的处理函数：在对应快捷键触发时经
// svc（MenuServices）执行，返回页面与命令。重复注册同一 op 或注册 nil
// handler 时 panic（程序错误）。
func RegisterOperateHandler(op keybindings.OperateType, handler func(svc MenuServices, app *model.App) (model.Page, tea.Cmd)) {
	if handler == nil {
		panic("ui: RegisterOperateHandler requires a non-nil handler")
	}

	operateHandlersMu.Lock()
	defer operateHandlersMu.Unlock()
	if _, ok := operateHandlers[op]; ok {
		panic(fmt.Sprintf("ui: duplicate operate handler for op %s (%d)", op.Name(), op))
	}
	operateHandlers[op] = handler
}

// handlePluginOperate 将未被内置 switch 分支处理的插件操作分发给注册的
// 处理器。未命中注册表时保持 default 分支语义，返回 (false, nil, nil)。
func (h *EventHandler) handlePluginOperate(op keybindings.OperateType) (bool, model.Page, tea.Cmd) {
	operateHandlersMu.RLock()
	handler, ok := operateHandlers[op]
	operateHandlersMu.RUnlock()
	if !ok {
		return false, nil, nil
	}

	page, cmd := handler(h.svc, h.svc.App())
	return true, page, cmd
}
