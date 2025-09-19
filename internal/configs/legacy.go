package configs

import (
	"os"
	"path/filepath"
	"reflect"

	"github.com/go-musicfox/go-musicfox/internal/types"
)

// ConfigFormat 定义了支持的配置文件格式。
type ConfigFormat string

const (
	FormatINI  ConfigFormat = "ini"
	FormatTOML ConfigFormat = "toml"
)

// ResolvedConfig 代表了最终解析出的配置文件信息。
type ResolvedConfig struct {
	Path   string       // 最终应使用的配置文件路径
	Format ConfigFormat // 最终应使用的配置文件格式
	Exists bool         // 该文件当前是否已存在于磁盘上
}

var UseIni = false

// ResolveConfigFile 智能解析应使用的配置文件路径和格式。
// 它接收配置目录作为参数。
// 规则:
// 1. 如果 toml 文件存在，则使用toml。
// 2. 如果 toml 文件不存在，但 ini 文件存在，则使用 ini（兼容模式）。
// 3. 如果两者都不存在，则默认使用 toml 路径（用于全新安装）。
func ResolveConfigFile(configDir string) ResolvedConfig {
	tomlPath := filepath.Join(configDir, types.AppTomlFile)
	iniPath := filepath.Join(configDir, types.AppIniFile)

	// TOML 优先
	if _, err := os.Stat(tomlPath); err == nil {
		return ResolvedConfig{
			Path:   tomlPath,
			Format: FormatTOML,
			Exists: true,
		}
	}

	// INI 回退
	if _, err := os.Stat(iniPath); err == nil {
		UseIni = true
		return ResolvedConfig{
			Path:   iniPath,
			Format: FormatINI,
			Exists: true,
		}
	}

	// 默认使用 TOML
	return ResolvedConfig{
		Path:   tomlPath,
		Format: FormatTOML,
		Exists: false,
	}
}

// MigrateLegacyRegistry 将一个从 INI 加载的旧 Registry 对象，转换为新的 Config 对象。
// 这个函数是兼容性层的核心，它处理所有的结构和数据转换。
func MigrateLegacyRegistry(legacyReg *Registry) *Config {
	newConfig := &Config{} // INI 配置足够完善

	// [startup]
	newConfig.Startup.Enable = legacyReg.Startup.EnableStartup
	newConfig.Startup.ProgressOutBounce = legacyReg.Startup.ProgressOutBounce
	newConfig.Startup.LoadingSeconds = int(legacyReg.Startup.LoadingDuration.Seconds())
	newConfig.Startup.Welcome = legacyReg.Startup.Welcome
	newConfig.Startup.SignIn = legacyReg.Startup.SignIn
	newConfig.Startup.CheckUpdate = legacyReg.Startup.CheckUpdate

	// [main]
	newConfig.Main.AltScreen = legacyReg.Main.AltScreen
	newConfig.Main.EnableMouseEvent = legacyReg.Main.EnableMouseEvent
	newConfig.Main.Debug = legacyReg.Main.Debug

	// [main.notification]
	newConfig.Main.Notification.Enable = legacyReg.Main.ShowNotify
	newConfig.Main.Notification.Icon = legacyReg.Main.NotifyIcon
	newConfig.Main.Notification.AlbumCover = legacyReg.Main.NotifyAlbumCover

	// [main.lyric]
	newConfig.Main.Lyric.Show = legacyReg.Main.ShowLyric
	newConfig.Main.Lyric.ShowTranslation = legacyReg.Main.ShowLyricTrans
	newConfig.Main.Lyric.Offset = legacyReg.Main.LyricOffset

	// [main.pprof]
	newConfig.Main.Pprof.Port = legacyReg.Main.PProfPort

	// [main.account]
	newConfig.Main.Account.NeteaseCookie = legacyReg.Main.NeteaseCookie

	// [theme]
	newConfig.Theme.ShowTitle = legacyReg.Main.ShowTitle
	newConfig.Theme.LoadingText = legacyReg.Main.LoadingText
	newConfig.Theme.DoubleColumn = legacyReg.Main.DualColumn
	newConfig.Theme.DynamicMenuRows = legacyReg.Main.DynamicMenuRows
	newConfig.Theme.CenterEverything = legacyReg.Main.CenterEverything
	newConfig.Theme.PrimaryColor = legacyReg.Main.PrimaryColor

	// [theme.progress]
	vProgressNew := reflect.ValueOf(&newConfig.Theme.Progress).Elem()
	vProgressOld := reflect.ValueOf(legacyReg.Progress.ProgressOptions)
	for i := 0; i < vProgressOld.NumField(); i++ {
		fieldName := vProgressOld.Type().Field(i).Name
		if vProgressNew.FieldByName(fieldName).IsValid() {
			vProgressNew.FieldByName(fieldName).SetString(string(vProgressOld.Field(i).Interface().(rune)))
		}
	}

	// [storage]
	newConfig.Storage.DownloadDir = legacyReg.Storge.DownloadDir
	newConfig.Storage.LyricDir = legacyReg.Storge.DownloadLyricDir
	newConfig.Storage.FileNameTpl = legacyReg.Storge.DownloadFileNameTpl
	newConfig.Storage.DownloadSongWithLyric = legacyReg.Storge.DownloadSongWithLyric
	newConfig.Storage.Cache.Dir = legacyReg.Storge.CacheDir
	newConfig.Storage.Cache.Limit = legacyReg.Storge.CacheLimit

	// [player]
	newConfig.Player.Engine = legacyReg.Player.Engine
	newConfig.Player.MaxPlayErrCount = legacyReg.Player.MaxPlayErrCount
	newConfig.Player.SongLevel = legacyReg.Main.PlayerSongLevel
	newConfig.Player.ShowAllSongsOfPlaylist = legacyReg.Main.ShowAllSongsOfPlaylist

	// [player.beep], [player.mpd], [player.mpv]
	newConfig.Player.Beep.Mp3Decoder = legacyReg.Player.BeepMp3Decoder
	newConfig.Player.Mpd.Bin = legacyReg.Player.MpdBin
	newConfig.Player.Mpd.ConfigFile = legacyReg.Player.MpdConfigFile
	newConfig.Player.Mpd.Network = legacyReg.Player.MpdNetwork
	newConfig.Player.Mpd.Addr = legacyReg.Player.MpdAddr
	newConfig.Player.Mpd.AutoStart = legacyReg.Player.MpdAutoStart
	newConfig.Player.Mpv.Bin = legacyReg.Player.MpvBin

	// [autoplay]
	newConfig.Autoplay.Enable = legacyReg.AutoPlayer.Enable
	newConfig.Autoplay.Playlist = legacyReg.AutoPlayer.Playlist
	newConfig.Autoplay.Offset = legacyReg.AutoPlayer.Offset
	newConfig.Autoplay.Mode = legacyReg.AutoPlayer.Mode

	// [unm]
	newConfig.UNM.Enable = legacyReg.UNM.Enable
	newConfig.UNM.Sources = legacyReg.UNM.Sources
	newConfig.UNM.SearchLimit = legacyReg.UNM.SearchLimit
	newConfig.UNM.EnableLocalVip = legacyReg.UNM.EnableLocalVip
	newConfig.UNM.UnlockSoundEffects = legacyReg.UNM.UnlockSoundEffects
	newConfig.UNM.QQCookieFile = legacyReg.UNM.QQCookieFile
	newConfig.UNM.SkipInvalidTracks = legacyReg.UNM.SkipInvalidTracks

	// [reporter]
	newConfig.Reporter.Netease.Enable = legacyReg.Reporter.Netease.Enable
	newConfig.Reporter.Lastfm.Enable = legacyReg.Reporter.Lastfm.Enable
	newConfig.Reporter.Lastfm.Key = legacyReg.Reporter.Lastfm.Key
	newConfig.Reporter.Lastfm.Secret = legacyReg.Reporter.Lastfm.Secret
	newConfig.Reporter.Lastfm.ScrobblePoint = legacyReg.Reporter.Lastfm.ScrobblePoint
	newConfig.Reporter.Lastfm.OnlyFirstArtist = legacyReg.Reporter.Lastfm.OnlyFirstArtist
	newConfig.Reporter.Lastfm.SkipDjRadio = legacyReg.Reporter.Lastfm.SkipDjRadio

	// [keybindings]
	newConfig.Keybindings.UseDefaultKeyBindings = legacyReg.Main.UseDefaultKeyBindings
	newConfig.Keybindings.Global = legacyReg.GlobalHotkeys
	EffectiveKeybindings = legacyReg.Keybindings
	// newConfig.Keybindings.App 实际不需要

	// [share]
	newConfig.Share = legacyReg.Share

	return newConfig
}
