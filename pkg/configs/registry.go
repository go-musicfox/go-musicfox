package configs

import (
	"time"

	"github.com/gookit/ini/v2"
	"go-musicfox/pkg/constants"
)

var ConfigRegistry *Registry

type Registry struct {
	StartupShow              bool          // 显示启动页
	StartupProgressOutBounce bool          // 是否启动页进度条回弹效果
	StartupLoadingDuration   time.Duration // 启动页加载时长
	StartupWelcome           string        // 启动页欢迎语
	StartupSignIn            bool          // 每天启动时自动签到
	StartupCheckUpdate       bool          // 启动检查更新

	ProgressFullChar  rune // 进度条已加载字符
	ProgressEmptyChar rune // 进度条未加载字符

	MainShowTitle    bool   // 主界面是否显示标题
	MainLoadingText  string // 主页面加载中提示
	MainPlayerSongBr int64  // 歌曲br设置
	MainPrimaryColor string // 主题色
	MainShowLyric    bool   // 显示歌词
	MainShowNotify   bool   // 显示通知
	MainNotifySender string // 通知应用图标
	MainPProfPort    int    // pprof端口
}

func NewRegistryWithDefault() *Registry {
	registry := &Registry{
		StartupShow:              constants.AppShowStartup,
		StartupProgressOutBounce: constants.StartupProgressOutBounce,
		StartupLoadingDuration:   time.Second * constants.StartupLoadingSeconds,
		StartupWelcome:           constants.AppName,
		StartupSignIn:            constants.StartupSignIn,
		StartupCheckUpdate:       constants.StartupCheckUpdate,

		ProgressFullChar:  rune(constants.ProgressFullChar[0]),
		ProgressEmptyChar: rune(constants.ProgressEmptyChar[0]),

		MainShowTitle:    constants.MainShowTitle,
		MainLoadingText:  constants.MainLoadingText,
		MainPlayerSongBr: constants.PlayerSongBr,
		MainPrimaryColor: constants.AppPrimaryRandom,
		MainShowLyric:    constants.MainShowLyric,
		MainShowNotify:   constants.MainShowNotify,
		MainNotifySender: constants.MainNotifySender,
		MainPProfPort:    constants.MainPProfPort,
	}

	return registry
}

func NewRegistryFromIniFile(filepath string) *Registry {
	registry := NewRegistryWithDefault()

	if err := ini.LoadExists(filepath); err != nil {
		return registry
	}

	registry.StartupShow = ini.Bool("startup.show", constants.AppShowStartup)
	registry.StartupProgressOutBounce = ini.Bool("startup.progressOutBounce", constants.AppShowStartup)
	registry.StartupLoadingDuration = time.Second * time.Duration(ini.Int("startup.loadingSeconds", constants.StartupLoadingSeconds))
	registry.StartupWelcome = ini.String("startup.welcome", constants.AppName)
	registry.StartupSignIn = ini.Bool("startup.signIn", constants.StartupSignIn)
	registry.StartupCheckUpdate = ini.Bool("startup.checkUpdate", constants.StartupCheckUpdate)

	fullChar := ini.String("progress.fullChar", constants.ProgressFullChar)
	if len(fullChar) > 0 {
		registry.ProgressFullChar = rune(fullChar[0])
	} else {
		registry.ProgressFullChar = rune(constants.ProgressFullChar[0])
	}
	emptyChar := ini.String("progress.emptyChar", constants.ProgressEmptyChar)
	if len(emptyChar) > 0 {
		registry.ProgressEmptyChar = rune(emptyChar[0])
	} else {
		registry.ProgressEmptyChar = rune(constants.ProgressEmptyChar[0])
	}

	registry.MainShowTitle = ini.Bool("main.showTitle", constants.MainShowTitle)
	registry.MainLoadingText = ini.String("main.loadingText", constants.MainLoadingText)
	registry.MainPlayerSongBr = ini.Int64("main.songBr", constants.PlayerSongBr)
	primaryColor := ini.String("main.primaryColor", constants.AppPrimaryRandom)
	if primaryColor != "" {
		registry.MainPrimaryColor = primaryColor
	} else {
		registry.MainPrimaryColor = constants.AppPrimaryRandom
	}
	registry.MainShowLyric = ini.Bool("main.showLyric", constants.MainShowLyric)
	registry.MainShowNotify = ini.Bool("main.showNotify", constants.MainShowNotify)
	registry.MainNotifySender = ini.String("main.notifySender", constants.MainNotifySender)
	registry.MainPProfPort = ini.Int("main.pprofPort", constants.MainPProfPort)

	return registry
}
