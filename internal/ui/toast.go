package ui

import (
	"github.com/anhoder/foxful-cli/model"
	"github.com/skratchdot/open-golang/open"

	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/utils/notify"
)

// toastLevelToModel 将 notify.ToastLevel 映射为 foxful-cli 的通知级别。
func toastLevelToModel(l notify.ToastLevel) model.NotificationLevel {
	switch l {
	case notify.ToastSuccess:
		return model.NotificationSuccess
	case notify.ToastWarning:
		return model.NotificationWarning
	case notify.ToastError:
		return model.NotificationError
	default:
		return model.NotificationInfo
	}
}

const toastOpenURLActionID = "open-url"

// BuildToastNotificationSpec 将 notify.NotifyContent 转为 foxful-cli 的
// NotificationSpec（含可选的打开链接动作）。导出供外部插件复用（检查更新
// 插件经 ui.BuildToastNotificationSpec 构建「发现新版本」toast）。
func BuildToastNotificationSpec(content notify.NotifyContent, level notify.ToastLevel, openURL func(string) error) model.NotificationSpec {
	spec := model.NotificationSpec{
		Level:   toastLevelToModel(level),
		Title:   content.Title,
		Message: content.Text,
	}
	if content.ActionLabel == "" || content.Url == "" {
		return spec
	}

	url := content.Url
	spec.Actions = []model.NotificationAction{{ID: toastOpenURLActionID, Label: content.ActionLabel}}
	spec.OnAction = func(result model.NotificationActionResult) {
		if result.ActionID == toastOpenURLActionID {
			_ = openURL(url)
		}
	}
	return spec
}

// registerToastHook 将 notify 包的 toast 回调接到 App.Notify，
// 使所有 notify.Notify(...) 调用在启用 inApp 时同时弹出 TUI 内原生 toast。
// 由 InitHook 调用，此时 App.Run 已启动、program 已就绪。
func (n *Netease) registerToastHook() {
	if !configs.AppConfig.Main.Notification.InApp {
		return
	}
	notify.SetToastHook(func(content notify.NotifyContent, level notify.ToastLevel) {
		n.Notify(BuildToastNotificationSpec(content, level, open.Start))
	})
}
