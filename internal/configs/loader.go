package configs

import (
	"os"
	"runtime"

	"github.com/go-musicfox/go-musicfox/internal/keybindings"
	"github.com/go-musicfox/go-musicfox/internal/types"
	"github.com/go-musicfox/go-musicfox/utils/filex"
	"github.com/go-viper/mapstructure/v2"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/knadh/koanf/v2"
	"github.com/pkg/errors"
)

// AppConfig 全局配置
var AppConfig *Config

// EffectiveKeybindings 存储最终生效的、经过处理的键绑定。
var EffectiveKeybindings map[keybindings.OperateType][]string

// NewConfigFromTomlFile 加载 TOML 配置文件
func NewConfigFromTomlFile(tomlPath string) (*Config, error) {
	k := koanf.New(".")

	// 加载内嵌的 TOML 文件作为默认值
	defaultTomlBytes, err := filex.ReadFileFromEmbed("embed/" + types.AppTomlFile)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read embedded default config")
	}
	if err := k.Load(rawbytes.Provider(defaultTomlBytes), toml.Parser()); err != nil {
		return nil, errors.Wrap(err, "failed to parse embedded default config")
	}

	if err := k.Load(file.Provider(tomlPath), toml.Parser()); err != nil {
		if !os.IsNotExist(err) {
			return nil, errors.Wrapf(err, "error loading TOML config file '%s'", tomlPath)
		}
	}

	// 处理动态默认值
	applyDynamicDefaults(k)

	finalConfig := &Config{}
	unmarshalConf := koanf.UnmarshalConf{
		DecoderConfig: &mapstructure.DecoderConfig{
			DecodeHook: newDecodeHooks(),
			Result:     finalConfig,
		},
	}
	if err := k.UnmarshalWithConf("", finalConfig, unmarshalConf); err != nil {
		return nil, errors.Wrap(err, "error unmarshalling config")
	}

	// 处理键绑定
	EffectiveKeybindings = keybindings.BuildEffectiveBindings(
		keybindings.ProcessUserBindings(finalConfig.Keybindings.App),
		finalConfig.Keybindings.UseDefaultKeyBindings,
	)

	return finalConfig, nil
}

// applyDynamicDefaults 处理无法在静态TOML中定义的默认值
func applyDynamicDefaults(k *koanf.Koanf) {
	if k.String("player.engine") == "auto" {
		var defaultEngine string
		switch runtime.GOOS {
		case "darwin":
			defaultEngine = types.OsxPlayer
		case "windows":
			defaultEngine = types.WinMediaPlayer
		default:
			defaultEngine = types.BeepPlayer
		}
		k.Set("player.engine", defaultEngine)
	}
}
