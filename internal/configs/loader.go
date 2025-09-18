package configs

import (
	"fmt"
	"os"

	"github.com/go-musicfox/go-musicfox/internal/keybindings"
	"github.com/go-viper/mapstructure/v2"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/structs"
	"github.com/knadh/koanf/v2"
)

// AppConfig 全局配置
var AppConfig *Config

// EffectiveKeybindings 存储最终生效的、经过处理的键绑定。
var EffectiveKeybindings map[keybindings.OperateType][]string

// NewConfigFromTomlFile 加载 TOML 配置文件
func NewConfigFromTomlFile(tomlPath string) (*Config, error) {
	k := koanf.New(".")

	// 加载默认值
	if err := k.Load(structs.Provider(NewDefaultConfig(), "koanf"), nil); err != nil {
		return nil, fmt.Errorf("error loading default config: %w", err)
	}

	if err := k.Load(file.Provider(tomlPath), toml.Parser()); err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("error loading TOML config file '%s': %w", tomlPath, err)
		}
	}

	finalConfig := &Config{}
	unmarshalConf := koanf.UnmarshalConf{
		DecoderConfig: &mapstructure.DecoderConfig{
			DecodeHook: newDecodeHooks(),
			Result:     finalConfig,
		},
	}
	if err := k.UnmarshalWithConf("", finalConfig, unmarshalConf); err != nil {
		return nil, fmt.Errorf("error unmarshalling config: %w", err)
	}

	// 处理键绑定
	EffectiveKeybindings = keybindings.BuildEffectiveBindings(
		keybindings.ProcessUserBindings(finalConfig.Keybindings.App),
		finalConfig.Keybindings.UseDefaultKeyBindings,
	)

	return finalConfig, nil
}
