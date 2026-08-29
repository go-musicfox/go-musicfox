package checkupdate

import (
	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/types"
	"github.com/go-musicfox/go-musicfox/utils/notify"
	"github.com/go-musicfox/go-musicfox/utils/version"
)

// startupCheck 是注册为启动钩子的自动检查任务：受 [startup] checkUpdate
// 配置门控，行为与提取前 shell 内联的启动检查一致（仅在有新版本时弹通知）。
// 由 shell 的 InitHook 在用户/登录就绪后经 ui.RegisterStartupHook 注册的
// 钩子列表调用（带 panic 隔离）。
func startupCheck() {
	if !configs.AppConfig.Startup.CheckUpdate {
		return
	}
	if ok, newVersion := version.CheckUpdate(); ok {
		notify.Notify(notify.NotifyContent{
			Title:       "发现新版本: " + newVersion,
			Text:        "去看看呗",
			Url:         types.AppGithubUrl + "/releases/tag/" + newVersion,
			ActionLabel: "前往 GitHub",
		})
	}
}
