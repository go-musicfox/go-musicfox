package configs

import (
	"github.com/anhoder/foxful-cli/model"
	"github.com/go-musicfox/go-musicfox/internal/types"
)

type ProgressOptions struct {
	model.ProgressOptions
}

// ThemeConfig 主题设置
type ThemeConfig struct {
	// 主界面是否显示标题
	ShowTitle bool `koanf:"showTitle"`
	// 主页面加载中提示
	LoadingText string `koanf:"loadingText"`
	// 是否双列显示
	DoubleColumn bool `koanf:"doubleColumn"`
	// 菜单行数动态变更
	DynamicMenuRows bool `koanf:"dynamicMenuRows"`
	// 界面全部居中
	CenterEverything bool `koanf:"centerEverything"`
	// 主题色
	PrimaryColor string `koanf:"primaryColor"`

	Progress ProgressConfig `koanf:"progress"`
}

// ProgressConfig 进度条字符样式配置
type ProgressConfig struct {
	FullChar           string `koanf:"fullChar"`
	FullCharWhenFirst  string `koanf:"fullCharWhenFirst"`
	FullCharWhenLast   string `koanf:"fullCharWhenLast"`
	LastFullChar       string `koanf:"lastFullChar"`
	EmptyChar          string `koanf:"emptyChar"`
	EmptyCharWhenFirst string `koanf:"emptyCharWhenFirst"`
	EmptyCharWhenLast  string `koanf:"emptyCharWhenLast"`
	FirstEmptyChar     string `koanf:"firstEmptyChar"`
}

// ToModel 将 ProgressConfig 转换为 foxful-cli 所需的 model.ProgressOptions。
func (pc ProgressConfig) ToModel() model.ProgressOptions {
	return model.ProgressOptions{
		FullChar:           firstCharOrDefault(pc.FullChar, types.ProgressFullChar),
		FullCharWhenFirst:  firstCharOrDefault(pc.FullCharWhenFirst, types.ProgressFullChar),
		FullCharWhenLast:   firstCharOrDefault(pc.FullCharWhenLast, types.ProgressFullChar),
		LastFullChar:       firstCharOrDefault(pc.LastFullChar, types.ProgressFullChar),
		EmptyChar:          firstCharOrDefault(pc.EmptyChar, types.ProgressEmptyChar),
		EmptyCharWhenFirst: firstCharOrDefault(pc.EmptyCharWhenFirst, types.ProgressEmptyChar),
		EmptyCharWhenLast:  firstCharOrDefault(pc.EmptyCharWhenLast, types.ProgressEmptyChar),
		FirstEmptyChar:     firstCharOrDefault(pc.FirstEmptyChar, types.ProgressEmptyChar),
	}
}
