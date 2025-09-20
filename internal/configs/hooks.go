package configs

import (
	"reflect"

	"github.com/go-musicfox/go-musicfox/internal/types"
	"github.com/go-musicfox/netease-music/service"
	"github.com/go-viper/mapstructure/v2"
)

// newDecodeHooks 返回一个为koanf配置的复合解码钩子。
// 它包含了所有自定义类型的转换逻辑。
func newDecodeHooks() mapstructure.DecodeHookFunc {
	return mapstructure.ComposeDecodeHookFunc(
		stringToAutoPlayerPlaylistHook(),
		stringToPlayerModeHook(),
		stringToSongQualityLevelHook(),
	)
}

// stringToAutoPlayerPlaylistHook 将字符串转换为 AutoPlayerPlaylist 类型。
func stringToAutoPlayerPlaylistHook() mapstructure.DecodeHookFunc {
	return func(f reflect.Type, t reflect.Type, data any) (any, error) {
		if f.Kind() != reflect.String || t != reflect.TypeOf(AutoPlayerPlaylist("")) {
			return data, nil
		}
		return AutoPlayerPlaylistFromString(data.(string)), nil
	}
}

// stringToPlayerModeHook 将字符串转换为 types.Mode 类型。
func stringToPlayerModeHook() mapstructure.DecodeHookFunc {
	return func(f reflect.Type, t reflect.Type, data any) (any, error) {
		if f.Kind() != reflect.String || t != reflect.TypeOf(types.Mode(0)) {
			return data, nil
		}
		return PlayerModeFromAutoPlayModeString(data.(string)), nil
	}
}

// stringToSongQualityLevelHook 将字符串转换为 service.SongQualityLevel 类型。
func stringToSongQualityLevelHook() mapstructure.DecodeHookFunc {
	return func(f reflect.Type, t reflect.Type, data any) (any, error) {
		if f.Kind() != reflect.String || t != reflect.TypeOf(service.SongQualityLevel("")) {
			return data, nil
		}

		level := service.SongQualityLevel(data.(string))
		if !level.IsValid() {
			return service.Higher, nil
		}
		return level, nil
	}
}
