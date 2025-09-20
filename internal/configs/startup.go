package configs

import (
	"time"

	"github.com/anhoder/foxful-cli/model"
	"github.com/go-musicfox/go-musicfox/internal/types"
)

type StartupOptions struct {
	model.StartupOptions
	SignIn      bool // 每天启动时自动签到
	CheckUpdate bool // 启动检查更新
}

// StartupConfig 启动页相关配置
type StartupConfig struct {
	// 每天启动时自动签到
	SignIn bool `koanf:"signIn"`
	// 启动检查更新
	CheckUpdate bool `koanf:"checkUpdate"`
	// 是否显示启动页
	Enable bool `koanf:"enable"`
	// 启动页进度条是否有回弹效果
	ProgressOutBounce bool `koanf:"progressOutBounce"`
	// 启动页的持续时长（秒）
	LoadingSeconds int `koanf:"loadingSeconds"`
	// 启动页欢迎语
	Welcome string `koanf:"welcome"`
}

// ToModel 将 StartupConfig 转换为 foxful-cli 所需的 model.StartupOptions。
func (sc StartupConfig) ToModel() model.StartupOptions {
	return model.StartupOptions{
		EnableStartup:     sc.Enable,
		ProgressOutBounce: sc.ProgressOutBounce,
		LoadingDuration:   time.Duration(sc.LoadingSeconds) * time.Second,
		Welcome:           sc.Welcome,
		TickDuration:      types.StartupTickDuration,
	}
}
