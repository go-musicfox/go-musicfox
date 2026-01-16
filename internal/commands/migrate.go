package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/go-musicfox/go-musicfox/internal/keybindings"
	"github.com/go-musicfox/go-musicfox/internal/types"
	"github.com/go-musicfox/go-musicfox/utils/app"
	"github.com/gookit/gcli/v2"
	"github.com/gookit/ini/v2"
)

// Migrate options
var migrateOpts struct {
	dryRun bool
	force  bool
}

// NewMigrateCommand creates the migration command
func NewMigrateCommand() *gcli.Command {
	cmd := &gcli.Command{
		Name:   "migrate",
		UseFor: "Migrate legacy INI configuration to TOML format",
		Examples: "{$binName} {$cmd}          # Perform actual migration\n" +
			"  {$binName} {$cmd} -n      # Dry run to preview\n" +
			"  {$binName} {$cmd} -f      # Force overwrite existing TOML",
		Config: func(c *gcli.Command) {
			c.Flags.BoolOpt(&migrateOpts.dryRun, "dry-run", "n", false, "Preview migration without making changes")
			c.Flags.BoolOpt(&migrateOpts.force, "force", "f", false, "Overwrite existing TOML config if it exists")
		},
		Func: runMigrate,
	}
	return cmd
}

func runMigrate(_ *gcli.Command, _ []string) error {
	const legacyIniFile = "go-musicfox.ini"

	// 获取配置目录
	configDir := app.ConfigDir()

	// 检测旧版 INI 配置文件
	iniPath := filepath.Join(configDir, legacyIniFile)
	tomlPath := filepath.Join(configDir, types.AppTomlFile)

	iniExists := false
	if _, err := os.Stat(iniPath); err == nil {
		iniExists = true
	}

	if !iniExists {
		fmt.Println("✓ 未检测到旧版 INI 配置文件，无需迁移。")
		return nil
	}

	fmt.Println("检测到旧版 INI 配置文件：")
	fmt.Printf("  %s\n\n", iniPath)

	// 读取旧版 INI 配置
	fmt.Println("正在读取旧版 INI 配置...")
	registry, err := loadLegacyRegistry(iniPath)
	if err != nil {
		return fmt.Errorf("读取 INI 配置失败: %w", err)
	}

	// 转换为 TOML
	tomlContent, err := convertToToml(registry)
	if err != nil {
		return fmt.Errorf("转换为 TOML 格式失败: %w", err)
	}

	// 检查 TOML 文件是否已存在
	tomlExists := false
	if _, err := os.Stat(tomlPath); err == nil {
		tomlExists = true
	}

	if tomlExists && !migrateOpts.force {
		fmt.Printf("⚠ TOML 配置文件已存在：\n")
		fmt.Printf("  %s\n\n", tomlPath)
		fmt.Println("请使用 -f/--force 参数覆盖，或手动删除现有文件后重试。")
		return nil
	}

	if migrateOpts.dryRun {
		fmt.Println("=== 预览：TOML 配置文件内容 ===")
		fmt.Println(tomlContent)
		fmt.Println("=== 预览结束 ===")
		return nil
	}

	// 执行迁移
	fmt.Printf("正在写入新的 TOML 配置文件：\n")
	fmt.Printf("  %s\n\n", tomlPath)

	if err := os.WriteFile(tomlPath, []byte(tomlContent), 0644); err != nil {
		return fmt.Errorf("写入 TOML 配置失败: %w", err)
	}

	// 备份或删除 INI 文件
	backupPath := iniPath + ".backup"
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		fmt.Printf("正在备份 INI 文件：\n")
		fmt.Printf("  %s\n\n", backupPath)
		if err := os.Rename(iniPath, backupPath); err != nil {
			fmt.Printf("⚠ 警告：备份 INI 文件失败: %v\n", err)
		}
	} else {
		fmt.Printf("正在删除旧版 INI 配置文件：\n")
		fmt.Printf("  %s\n\n", iniPath)
		if err := os.Remove(iniPath); err != nil {
			fmt.Printf("⚠ 警告：删除 INI 文件失败: %v\n", err)
		}
	}

	fmt.Println("✓ 迁移完成！")
	fmt.Println("\nINI 配置文件已迁移为 TOML 格式。")
	fmt.Println("如有自定义设置，请在新的 TOML 配置文件中确认。")

	return nil
}

// Legacy structures (mirrored from the old registry.go for migration)
type legacyRegistry struct {
	Startup       legacyStartupOptions
	Progress      legacyProgressOptions
	Main          legacyMainOptions
	AutoPlayer    legacyAutoPlayerOptions
	UNM           legacyUNMOptions
	Player        legacyPlayerOptions
	Reporter      legacyReporterOptions
	GlobalHotkeys map[string]string
	Keybindings   map[keybindings.OperateType][]string
	Share         map[string]string
	Storge        legacyStorageOptions
}

type legacyStartupOptions struct {
	EnableStartup     bool
	ProgressOutBounce bool
	LoadingDuration   time.Duration
	Welcome           string
	SignIn            bool
	CheckUpdate       bool
}

type legacyProgressOptions struct {
	EmptyChar          rune
	EmptyCharWhenFirst rune
	EmptyCharWhenLast  rune
	FirstEmptyChar     rune
	FullChar           rune
	FullCharWhenFirst  rune
	FullCharWhenLast   rune
	LastFullChar       rune
}

type legacyMainOptions struct {
	ShowTitle              bool
	LoadingText            string
	PlayerSongLevel        interface{}
	PrimaryColor           string
	ShowLyric              bool
	ShowLyricTrans         bool
	ShowNotify             bool
	NotifyIcon             string
	NotifyAlbumCover       bool
	PProfPort              int
	AltScreen              bool
	EnableMouseEvent       bool
	DualColumn             bool
	DynamicMenuRows        bool
	UseDefaultKeyBindings  bool
	CenterEverything       bool
	NeteaseCookie          string
	Debug                  bool
	LyricOffset            int
	PlayerEngine           string
	ShowAllSongsOfPlaylist bool
}

type legacyStorageOptions struct {
	DownloadDir           string
	DownloadLyricDir      string
	DownloadSongWithLyric bool
	DownloadFileNameTpl   string
	CacheDir              string
	CacheLimit            int64
}

type legacyPlayerOptions struct {
	Engine          string
	BeepMp3Decoder  string
	MpdBin          string
	MpdConfigFile   string
	MpdNetwork      string
	MpdAddr         string
	MpdAutoStart    bool
	MaxPlayErrCount int
	MpvBin          string
}

type legacyAutoPlayerOptions struct {
	Enable   bool
	Playlist string
	Offset   int
	Mode     string
}

type legacyUNMOptions struct {
	Enable             bool
	Sources            []string
	SearchLimit        int
	EnableLocalVip     bool
	UnlockSoundEffects bool
	QQCookieFile       string
	SkipInvalidTracks  bool
}

type legacyReporterOptions struct {
	Lastfm  legacyReporterLastfmOptions
	Netease legacyReporterNeteaseOptions
}

type legacyReporterLastfmOptions struct {
	Key             string
	Secret          string
	Enable          bool
	ScrobblePoint   int
	OnlyFirstArtist bool
	SkipDjRadio     bool
}

type legacyReporterNeteaseOptions struct {
	Enable bool
}

// loadLegacyRegistry loads configuration from legacy INI file
func loadLegacyRegistry(filepath string) (*legacyRegistry, error) {
	r := &legacyRegistry{
		Startup: legacyStartupOptions{
			EnableStartup:     true,
			ProgressOutBounce: true,
			LoadingDuration:   time.Second * types.StartupLoadingSeconds,
			Welcome:           types.AppName,
			SignIn:            false,
			CheckUpdate:       true,
		},
		Progress: legacyProgressOptions{
			EmptyChar:          []rune(types.ProgressEmptyChar)[0],
			EmptyCharWhenFirst: []rune(types.ProgressEmptyChar)[0],
			EmptyCharWhenLast:  []rune(types.ProgressEmptyChar)[0],
			FirstEmptyChar:     []rune(types.ProgressEmptyChar)[0],
			FullChar:           []rune(types.ProgressFullChar)[0],
			FullCharWhenFirst:  []rune(types.ProgressFullChar)[0],
			FullCharWhenLast:   []rune(types.ProgressFullChar)[0],
			LastFullChar:       []rune(types.ProgressEmptyChar)[0],
		},
		Main: legacyMainOptions{
			ShowTitle:              true,
			LoadingText:            types.MainLoadingText,
			PrimaryColor:           types.AppPrimaryColor,
			ShowLyric:              true,
			ShowLyricTrans:         true,
			ShowNotify:             true,
			NotifyIcon:             types.DefaultNotifyIcon,
			NotifyAlbumCover:       false,
			PProfPort:              types.MainPProfPort,
			AltScreen:              true,
			EnableMouseEvent:       true,
			DualColumn:             true,
			DynamicMenuRows:        false,
			UseDefaultKeyBindings:  true,
			CenterEverything:       false,
			NeteaseCookie:          "",
			Debug:                  false,
			LyricOffset:            0,
			ShowAllSongsOfPlaylist: false,
		},
		Storge: legacyStorageOptions{
			DownloadDir:           "",
			DownloadLyricDir:      "",
			DownloadSongWithLyric: false,
			DownloadFileNameTpl:   "{{.SongName}}-{{.SongArtists}}.{{.FileExt}}",
			CacheLimit:            0,
			CacheDir:              "",
		},
		Player: legacyPlayerOptions{
			Engine:         types.BeepPlayer,
			BeepMp3Decoder: types.BeepGoMp3Decoder,
		},
		UNM: legacyUNMOptions{
			Enable:             false,
			Sources:            []string{types.UNMDefaultSources},
			EnableLocalVip:     true,
			UnlockSoundEffects: true,
			SkipInvalidTracks:  false,
		},
		Reporter: legacyReporterOptions{
			Lastfm: legacyReporterLastfmOptions{
				Key:             "",
				Secret:          "",
				Enable:          false,
				ScrobblePoint:   50,
				OnlyFirstArtist: false,
				SkipDjRadio:     false,
			},
			Netease: legacyReporterNeteaseOptions{
				Enable: false,
			},
		},
		Keybindings: getDefaultBindingsMap(),
		Share:       nil,
	}

	// Set platform-specific defaults
	switch runtime.GOOS {
	case "darwin":
		r.Player.Engine = types.OsxPlayer
	case "windows":
		r.Player.Engine = types.WinMediaPlayer
	}

	if err := ini.LoadExists(filepath); err != nil {
		return r, err
	}

	// Parse INI configuration
	r.Startup.EnableStartup = ini.Bool("startup.show", true)
	r.Startup.ProgressOutBounce = ini.Bool("startup.progressOutBounce", true)
	r.Startup.LoadingDuration = time.Second * time.Duration(ini.Int("startup.loadingSeconds", types.StartupLoadingSeconds))
	r.Startup.Welcome = ini.String("startup.welcome", types.AppName)
	r.Startup.SignIn = ini.Bool("startup.signIn", false)
	r.Startup.CheckUpdate = ini.Bool("startup.checkUpdate", true)

	// Progress chars
	r.Progress.EmptyChar = firstCharOrDefault(ini.String("progress.emptyChar", types.ProgressEmptyChar), types.ProgressEmptyChar)
	r.Progress.EmptyCharWhenFirst = firstCharOrDefault(ini.String("progress.emptyCharWhenFirst", types.ProgressEmptyChar), types.ProgressEmptyChar)
	r.Progress.EmptyCharWhenLast = firstCharOrDefault(ini.String("progress.emptyCharWhenLast", types.ProgressEmptyChar), types.ProgressEmptyChar)
	r.Progress.FirstEmptyChar = firstCharOrDefault(ini.String("progress.firstEmptyChar", types.ProgressEmptyChar), types.ProgressEmptyChar)
	r.Progress.FullChar = firstCharOrDefault(ini.String("progress.fullChar", types.ProgressFullChar), types.ProgressFullChar)
	r.Progress.FullCharWhenFirst = firstCharOrDefault(ini.String("progress.fullCharWhenFirst", types.ProgressFullChar), types.ProgressFullChar)
	r.Progress.FullCharWhenLast = firstCharOrDefault(ini.String("progress.fullCharWhenLast", types.ProgressFullChar), types.ProgressFullChar)
	r.Progress.LastFullChar = firstCharOrDefault(ini.String("progress.lastFullChar", types.ProgressEmptyChar), types.ProgressEmptyChar)

	r.GlobalHotkeys = ini.StringMap("global_hotkey")

	// Main options
	r.Main.ShowTitle = ini.Bool("main.showTitle", true)
	r.Main.LoadingText = ini.String("main.loadingText", types.MainLoadingText)
	r.Main.PrimaryColor = ini.String("main.primaryColor", types.AppPrimaryColor)
	if r.Main.PrimaryColor == "" {
		r.Main.PrimaryColor = types.AppPrimaryColor
	}
	r.Main.ShowLyric = ini.Bool("main.showLyric", true)
	r.Main.LyricOffset = ini.Int("main.lyricOffset", 0)
	r.Main.ShowLyricTrans = ini.Bool("main.showLyricTrans", true)
	r.Main.ShowNotify = ini.Bool("main.showNotify", true)
	r.Main.NotifyIcon = ini.String("main.notifyIcon", types.DefaultNotifyIcon)
	r.Main.NotifyAlbumCover = ini.Bool("main.notifyAlbumCover", false)
	r.Main.PProfPort = ini.Int("main.pprofPort", types.MainPProfPort)
	r.Main.AltScreen = ini.Bool("main.altScreen", true)
	r.Main.EnableMouseEvent = ini.Bool("main.enableMouseEvent", true)
	r.Main.DualColumn = ini.Bool("main.doubleColumn", true)
	r.Main.ShowAllSongsOfPlaylist = ini.Bool("main.showAllSongsOfPlaylist", false)
	r.Main.DynamicMenuRows = ini.Bool("main.dynamicMenuRows", false)
	r.Main.UseDefaultKeyBindings = ini.Bool("main.useDefaultKeyBindings", true)
	r.Main.CenterEverything = ini.Bool("main.centerEverything", false)
	r.Main.NeteaseCookie = ini.String("main.neteaseCookie", "")
	r.Main.Debug = ini.Bool("main.debug", false)

	// Storage options
	r.Storge.DownloadDir = ini.String("storage.downloadDir", "")
	r.Storge.DownloadLyricDir = ini.String("storage.downloadLyricDir", "")
	r.Storge.DownloadSongWithLyric = ini.Bool("storage.downloadSongWithLyric", false)
	r.Storge.DownloadFileNameTpl = ini.String("storage.downloadFileNameTpl", "{{.SongName}}-{{.SongArtists}}.{{.FileExt}}")
	r.Storge.CacheDir = ini.String("storage.cacheDir", "")
	r.Storge.CacheLimit = ini.Int64("storage.cacheLimit", 0)

	// Player options
	defaultPlayer := types.BeepPlayer
	switch runtime.GOOS {
	case "darwin":
		defaultPlayer = types.OsxPlayer
	case "windows":
		defaultPlayer = types.WinMediaPlayer
	}
	r.Player.Engine = ini.String("player.engine", defaultPlayer)
	r.Player.BeepMp3Decoder = ini.String("player.beepMp3Decoder", types.BeepGoMp3Decoder)
	r.Player.MpdBin = ini.String("player.mpdBin", "")
	r.Player.MpdConfigFile = ini.String("player.mpdConfigFile", "")
	r.Player.MpdNetwork = ini.String("player.mpdNetwork", "")
	r.Player.MpdAddr = ini.String("player.mpdAddr", "")
	r.Player.MpdAutoStart = ini.Bool("player.mpdAutoStart", true)
	r.Player.MaxPlayErrCount = ini.Int("player.maxPlayErrCount", types.MaxPlayErrCount)
	r.Player.MpvBin = ini.String("player.mpvBin", "")

	// Autoplay
	r.AutoPlayer.Enable = ini.Bool("autoplay.autoPlay", false)
	r.AutoPlayer.Playlist = ini.String("autoplay.autoPlayList", "no")
	r.AutoPlayer.Offset = ini.Int("autoplay.offset", 0)
	r.AutoPlayer.Mode = ini.String("autoplay.playMode", "listLoop")

	// UNM
	r.UNM.Enable = ini.Bool("unm.switch", false)
	sourceStr := ini.String("unm.sources", "kuwo")
	if sourceStr != "" {
		var sources []string
		for _, source := range strings.Split(sourceStr, ",") {
			sources = append(sources, strings.TrimSpace(source))
		}
		r.UNM.Sources = sources
	}
	r.UNM.SearchLimit = ini.Int("unm.searchLimit", 0)
	r.UNM.EnableLocalVip = ini.Bool("unm.enableLocalVip", true)
	r.UNM.UnlockSoundEffects = ini.Bool("unm.unlockSoundEffects", true)
	r.UNM.QQCookieFile = ini.String("unm.qqCookieFile", "")
	r.UNM.SkipInvalidTracks = ini.Bool("unm.skipInvalidTracks", false)

	// Reporter
	r.Reporter.Lastfm.Key = ini.String("reporter.lastfmKey", "")
	r.Reporter.Lastfm.Secret = ini.String("reporter.lastfmSecret", "")
	r.Reporter.Lastfm.Enable = ini.Bool("reporter.lastfmEnable", false)
	r.Reporter.Lastfm.ScrobblePoint = ini.Int("reporter.lastfmScrobblePoint", 50)
	r.Reporter.Lastfm.OnlyFirstArtist = ini.Bool("reporter.lastfmOnlyFirstArtist", false)
	r.Reporter.Lastfm.SkipDjRadio = ini.Bool("reporter.lastfmSkipDjRadio", false)
	r.Reporter.Netease.Enable = ini.Bool("reporter.neteaseEnable", false)

	// Keybindings (for reference, not directly migrated)
	_ = ini.StringMap("keybindings")

	// Share
	r.Share = ini.StringMap("share")

	return r, nil
}

func firstCharOrDefault(s, defaultStr string) rune {
	if len(s) > 0 {
		return []rune(s)[0]
	}
	return []rune(defaultStr)[0]
}

func getDefaultBindingsMap() map[keybindings.OperateType][]string {
	ops := keybindings.InitDefaults(true)
	defaultMap := make(map[keybindings.OperateType][]string, len(ops))
	for op := range ops {
		defaultMap[op] = op.Keys()
	}
	return defaultMap
}

// convertToToml converts the legacy registry to TOML format
func convertToToml(r *legacyRegistry) (string, error) {
	var sb strings.Builder

	// Header
	sb.WriteString("# ---------------------------------------------------------------- #\n")
	sb.WriteString("#                  Go-Musicfox Configuration File                  #\n")
	sb.WriteString("# ---------------------------------------------------------------- #\n")
	sb.WriteString("# Migrated from legacy INI format\n")
	sb.WriteString("# ---------------------------------------------------------------- #\n\n")

	// Startup section
	sb.WriteString("# 启动页相关配置\n[startup]\n")
	sb.WriteString("# 是否显示启动页\nenable = " + boolToToml(r.Startup.EnableStartup) + "\n")
	sb.WriteString("# 启动页进度条是否具有回弹效果\nprogressOutBounce = " + boolToToml(r.Startup.ProgressOutBounce) + "\n")
	sb.WriteString("# 启动页的持续时长（秒）\nloadingSeconds = " + fmt.Sprintf("%d", int(r.Startup.LoadingDuration.Seconds())) + "\n")
	sb.WriteString("# 启动页欢迎语\nwelcome = " + fmt.Sprintf("%q", r.Startup.Welcome) + "\n")
	sb.WriteString("# 启动时是否自动签到\nsignIn = " + boolToToml(r.Startup.SignIn) + "\n")
	sb.WriteString("# 启动时是否检查应用更新\ncheckUpdate = " + boolToToml(r.Startup.CheckUpdate) + "\n")
	sb.WriteString("\n")

	// Main section
	sb.WriteString("# 主界面与核心功能配置\n[main]\n")
	sb.WriteString("# altScreen 显示模式\naltScreen = " + boolToToml(r.Main.AltScreen) + "\n")
	sb.WriteString("# 是否在界面中启用鼠标事件\nenableMouseEvent = " + boolToToml(r.Main.EnableMouseEvent) + "\n")
	sb.WriteString("# 是否开启 Debug 模式\ndebug = " + boolToToml(r.Main.Debug) + "\n")
	sb.WriteString("\n")

	// Notification subsection
	sb.WriteString("# 桌面通知相关设置\n[main.notification]\n")
	sb.WriteString("# 是否启用桌面通知\nenable = " + boolToToml(r.Main.ShowNotify) + "\n")
	sb.WriteString("# 默认通知图标\nicon = " + fmt.Sprintf("%q", r.Main.NotifyIcon) + "\n")
	sb.WriteString("# 是否使用歌曲的专辑封面作为通知图标\nalbumCover = " + boolToToml(r.Main.NotifyAlbumCover) + "\n")
	sb.WriteString("\n")

	// Lyric subsection
	sb.WriteString("# 歌词显示相关设置\n[main.lyric]\n")
	sb.WriteString("# 是否显示歌词\nshow = " + boolToToml(r.Main.ShowLyric) + "\n")
	sb.WriteString("# 是否显示翻译歌词\nshowTranslation = " + boolToToml(r.Main.ShowLyricTrans) + "\n")
	sb.WriteString("# 歌词显示时间的全局偏移量（毫秒）\noffset = " + fmt.Sprintf("%d", r.Main.LyricOffset) + "\n")
	sb.WriteString("\n")

	// Pprof subsection
	sb.WriteString("# Go性能分析工具pprof的相关设置\n[main.pprof]\n")
	sb.WriteString("# pprof 服务监听的端口\nport = " + fmt.Sprintf("%d", r.Main.PProfPort) + "\n")
	sb.WriteString("\n")

	// Account subsection
	sb.WriteString("# 账号相关配置\n[main.account]\n")
	sb.WriteString("# 网易云音乐的登录 Cookie\nneteaseCookie = " + fmt.Sprintf("%q", r.Main.NeteaseCookie) + "\n")
	sb.WriteString("\n")

	// Theme section
	sb.WriteString("# 主题设置\n[theme]\n")
	sb.WriteString("# 是否在界面顶部显示标题\nshowTitle = " + boolToToml(r.Main.ShowTitle) + "\n")
	sb.WriteString("# 菜单加载时显示的提示文字\nloadingText = " + fmt.Sprintf("%q", r.Main.LoadingText) + "\n")
	sb.WriteString("# 是否使用双列布局\ndoubleColumn = " + boolToToml(r.Main.DualColumn) + "\n")
	sb.WriteString("# 菜单行数是否根据终端高度动态变化\ndynamicMenuRows = " + boolToToml(r.Main.DynamicMenuRows) + "\n")
	sb.WriteString("# 界面所有内容居中\ncenterEverything = " + boolToToml(r.Main.CenterEverything) + "\n")
	sb.WriteString("# 主题颜色\nprimaryColor = " + fmt.Sprintf("%q", r.Main.PrimaryColor) + "\n")
	sb.WriteString("\n")

	// Progress subsection
	sb.WriteString("# 进度条字符样式配置\n[theme.progress]\n")
	sb.WriteString("# 进度条已加载字符\nfullChar = " + fmt.Sprintf("%q", string(r.Progress.FullChar)) + "\n")
	sb.WriteString("# fullCharWhenFirst\ngivenFirst = " + fmt.Sprintf("%q", string(r.Progress.FullCharWhenFirst)) + "\n")
	sb.WriteString("# fullCharWhenLast\ngivenLast = " + fmt.Sprintf("%q", string(r.Progress.FullCharWhenLast)) + "\n")
	sb.WriteString("# lastFullChar\nlastFullChar = " + fmt.Sprintf("%q", string(r.Progress.LastFullChar)) + "\n")
	sb.WriteString("# 进度条未加载字符\nemptyChar = " + fmt.Sprintf("%q", string(r.Progress.EmptyChar)) + "\n")
	sb.WriteString("# emptyCharWhenFirst\nemptyCharWhenFirst = " + fmt.Sprintf("%q", string(r.Progress.EmptyCharWhenFirst)) + "\n")
	sb.WriteString("# emptyCharWhenLast\nemptyCharWhenLast = " + fmt.Sprintf("%q", string(r.Progress.EmptyCharWhenLast)) + "\n")
	sb.WriteString("# firstEmptyChar\nfirstEmptyChar = " + fmt.Sprintf("%q", string(r.Progress.FirstEmptyChar)) + "\n")
	sb.WriteString("\n")

	// Storage section
	sb.WriteString("# 下载、缓存等文件存储相关配置\n[storage]\n")
	sb.WriteString("# 下载目录\ndownloadDir = " + fmt.Sprintf("%q", r.Storge.DownloadDir) + "\n")
	sb.WriteString("# 歌词文件的默认下载目录\nlyricDir = " + fmt.Sprintf("%q", r.Storge.DownloadLyricDir) + "\n")
	sb.WriteString("# 下载歌曲时是否同时下载歌词文件\ndownloadSongWithLyric = " + boolToToml(r.Storge.DownloadSongWithLyric) + "\n")
	sb.WriteString("# 下载文件的命名模板\n# fileNameTpl = " + fmt.Sprintf("%q", r.Storge.DownloadFileNameTpl) + "\n")
	sb.WriteString("\n")
	sb.WriteString("# 音乐播放缓存相关设置\n[storage.cache]\n")
	sb.WriteString("# 缓存目录\ndir = " + fmt.Sprintf("%q", r.Storge.CacheDir) + "\n")
	sb.WriteString("# 音乐缓存文件的总大小限制（单位：MB）\nlimit = " + fmt.Sprintf("%d", r.Storge.CacheLimit) + "\n")
	sb.WriteString("\n")

	// Player section
	sb.WriteString("# 播放器引擎与行为配置\n[player]\n")
	sb.WriteString("# 播放引擎\nengine = " + fmt.Sprintf("%q", r.Player.Engine) + "\n")
	sb.WriteString("# 允许的最大连续失败重试次数\nmaxPlayErrCount = " + fmt.Sprintf("%d", r.Player.MaxPlayErrCount) + "\n")
	sb.WriteString("# 是否获取并显示歌单下的所有歌曲\nshowAllSongsOfPlaylist = " + boolToToml(r.Main.ShowAllSongsOfPlaylist) + "\n")
	sb.WriteString("\n")

	// Beep subsection
	sb.WriteString("# `beep` 引擎专属配置\n[player.beep]\n")
	sb.WriteString("# MP3解码器\nmp3Decoder = " + fmt.Sprintf("%q", r.Player.BeepMp3Decoder) + "\n")
	sb.WriteString("\n")

	// MPD subsection
	sb.WriteString("# `mpd` 引擎专属配置\n[player.mpd]\n")
	sb.WriteString("# mpd 可执行文件的路径\nbin = " + fmt.Sprintf("%q", r.Player.MpdBin) + "\n")
	sb.WriteString("# mpd配置文件的路径\nconfigFile = " + fmt.Sprintf("%q", r.Player.MpdConfigFile) + "\n")
	sb.WriteString("# 与mpd服务的连接方式\nnetwork = " + fmt.Sprintf("%q", r.Player.MpdNetwork) + "\n")
	sb.WriteString("# 连接地址\naddr = " + fmt.Sprintf("%q", r.Player.MpdAddr) + "\n")
	sb.WriteString("# 是否在需要时自动启动 mpd 服务\nautoStart = " + boolToToml(r.Player.MpdAutoStart) + "\n")
	sb.WriteString("\n")

	// MPV subsection
	sb.WriteString("# `mpv` 引擎专属配置\n[player.mpv]\n")
	sb.WriteString("# mpv 可执行文件的路径\nbin = " + fmt.Sprintf("%q", r.Player.MpvBin) + "\n")
	sb.WriteString("\n")

	// Autoplay section
	sb.WriteString("# 启动时自动播放相关配置\n[autoplay]\n")
	sb.WriteString("# 是否在启动后自动开始播放\nenable = " + boolToToml(r.AutoPlayer.Enable) + "\n")
	sb.WriteString("# 自动播放的歌单\nplaylist = " + fmt.Sprintf("%q", r.AutoPlayer.Playlist) + "\n")
	sb.WriteString("# 播放列表的起始偏移量\noffset = " + fmt.Sprintf("%d", r.AutoPlayer.Offset) + "\n")
	sb.WriteString("# 播放模式\nmode = " + fmt.Sprintf("%q", r.AutoPlayer.Mode) + "\n")
	sb.WriteString("\n")

	// UNM section
	sb.WriteString("# UNM (Unlock NetEase Music) 相关配置\n[unm]\n")
	sb.WriteString("# 是否启用 UNM 功能\nenable = " + boolToToml(r.UNM.Enable) + "\n")
	sb.WriteString("# 音源匹配来源，可配置多个\nsources = [" + formatStringSlice(r.UNM.Sources) + "]\n")
	sb.WriteString("# UNM搜索其他平台限制\nsearchLimit = " + fmt.Sprintf("%d", r.UNM.SearchLimit) + "\n")
	sb.WriteString("# 解除会员限制\nenableLocalVip = " + boolToToml(r.UNM.EnableLocalVip) + "\n")
	sb.WriteString("# 解除音质限制\nunlockSoundEffects = " + boolToToml(r.UNM.UnlockSoundEffects) + "\n")
	sb.WriteString("# 用于获取QQ音乐音源的Cookie文件路径\nqqCookieFile = " + fmt.Sprintf("%q", r.UNM.QQCookieFile) + "\n")
	sb.WriteString("# 检测到无效的歌时跳过播放\nskipInvalidTracks = " + boolToToml(r.UNM.SkipInvalidTracks) + "\n")
	sb.WriteString("\n")

	// Reporter section
	sb.WriteString("# 播放状态上报配置\n[reporter]\n")
	sb.WriteString("# 是否将播放状态上报回网易云音乐\n[reporter.netease]\n")
	sb.WriteString("enable = " + boolToToml(r.Reporter.Netease.Enable) + "\n")
	sb.WriteString("\n")
	sb.WriteString("# 上报至 Last.fm\n[reporter.lastfm]\n")
	sb.WriteString("enable = " + boolToToml(r.Reporter.Lastfm.Enable) + "\n")
	sb.WriteString("# Last.fm API Key\nkey = " + fmt.Sprintf("%q", r.Reporter.Lastfm.Key) + "\n")
	sb.WriteString("# Last.fm API Shared Secret\nsecret = " + fmt.Sprintf("%q", r.Reporter.Lastfm.Secret) + "\n")
	sb.WriteString("# 播放一首歌的百分比达到多少时，才进行上报\nscrobblePoint = " + fmt.Sprintf("%d", r.Reporter.Lastfm.ScrobblePoint) + "\n")
	sb.WriteString("# 是否只上报歌曲的第一位艺术家\nonlyFirstArtist = " + boolToToml(r.Reporter.Lastfm.OnlyFirstArtist) + "\n")
	sb.WriteString("# 是否跳过电台节目的上报\nskipDjRadio = " + boolToToml(r.Reporter.Lastfm.SkipDjRadio) + "\n")
	sb.WriteString("\n")

	// Keybindings section
	sb.WriteString("# 快捷键绑定配置\n[keybindings]\n")
	sb.WriteString("# 是否使用应用内置的默认快捷键作为基础\nuseDefaultKeyBindings = " + boolToToml(r.Main.UseDefaultKeyBindings) + "\n")
	sb.WriteString("\n")

	// Global hotkeys
	if len(r.GlobalHotkeys) > 0 {
		sb.WriteString("# 全局快捷键\n[keybindings.global]\n")
		for key, value := range r.GlobalHotkeys {
			sb.WriteString(key + " = " + fmt.Sprintf("%q", value) + "\n")
		}
		sb.WriteString("\n")
	}

	// Share section
	if len(r.Share) > 0 {
		sb.WriteString("# 自定义分享模板\n[share]\n")
		for key, value := range r.Share {
			sb.WriteString(key + " = " + fmt.Sprintf("%q", value) + "\n")
		}
	}

	return sb.String(), nil
}

func boolToToml(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

func formatStringSlice(slice []string) string {
	var parts []string
	for _, s := range slice {
		parts = append(parts, fmt.Sprintf("%q", s))
	}
	return strings.Join(parts, ", ")
}
