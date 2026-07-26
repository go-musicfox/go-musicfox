package ui

import (
	tea "charm.land/bubbletea/v2"
	"github.com/anhoder/foxful-cli/model"
	"github.com/skratchdot/open-golang/open"

	"github.com/go-musicfox/go-musicfox/internal/types"
	"github.com/go-musicfox/go-musicfox/utils/notify"
	"github.com/go-musicfox/go-musicfox/utils/version"
)

func newVersionNotifyContent(newVersion string) notify.NotifyContent {
	return notify.NotifyContent{
		Title:       "发现新版本: " + newVersion,
		Text:        "去看看呗",
		Url:         types.AppGithubUrl + "/releases/tag/" + newVersion,
		ActionLabel: "前往 GitHub",
	}
}

func checkUpdateCmd() tea.Cmd {
	return func() tea.Msg {
		hasUpdate, latestVersion := version.CheckUpdate()
		return checkUpdateNotificationMsg(hasUpdate, latestVersion)
	}
}

func checkUpdateNotificationMsg(hasUpdate bool, latestVersion string) model.ShowNotificationMsg {
	var spec model.NotificationSpec
	switch {
	case hasUpdate:
		spec = buildToastNotificationSpec(newVersionNotifyContent(latestVersion), notify.ToastInfo, open.Start)
	case latestVersion != "":
		spec = model.NotificationSpec{
			Level:   model.NotificationSuccess,
			Title:   "检查更新",
			Message: types.AppVersion + " 已是最新版本",
		}
	default:
		spec = model.NotificationSpec{
			Level:   model.NotificationError,
			Title:   "检查更新失败",
			Message: "无法获取最新版本信息，请稍后重试",
		}
	}
	return model.ShowNotificationMsg{Spec: spec}
}
