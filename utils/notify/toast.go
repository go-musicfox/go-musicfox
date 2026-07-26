package notify

// ToastLevel 表示 TUI 内 toast 的语义级别，与 foxful-cli 的 NotificationLevel 对应。
// 由 UI 层映射为具体的 model.NotificationLevel，避免 notify 包依赖 foxful-cli。
type ToastLevel uint8

const (
	ToastInfo ToastLevel = iota
	ToastSuccess
	ToastWarning
	ToastError
)

// ToastFunc 由 UI 层注册，用于在 TUI 内弹出原生 toast 通知。
// content 为原始通知内容，level 为语义级别。
type ToastFunc func(content NotifyContent, level ToastLevel)

// toastHook 保存 UI 层注册的 toast 回调，未注册时为 nil。
var toastHook ToastFunc

// SetToastHook 注册 TUI 内 toast 回调。传入 nil 可解除注册。
func SetToastHook(fn ToastFunc) {
	toastHook = fn
}

// emitToast 在已注册回调时触发 TUI 内 toast。level 默认为 ToastInfo。
func emitToast(content NotifyContent, level ToastLevel) {
	if toastHook != nil {
		toastHook(content, level)
	}
}
