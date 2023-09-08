package configs

import (
	"github.com/anhoder/foxful-cli/model"
)

type StartupOptions struct {
	model.StartupOptions
	SignIn      bool // 每天启动时自动签到
	CheckUpdate bool // 启动检查更新
}
