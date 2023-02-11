package configs

import (
	"runtime"
	"strings"
	"time"

	"go-musicfox/pkg/constants"

	"github.com/anhoder/netease-music/service"
	"github.com/gookit/ini/v2"
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

	MainShowTitle              bool                     // 主界面是否显示标题
	MainLoadingText            string                   // 主页面加载中提示
	MainPlayerSongLevel        service.SongQualityLevel // 歌曲音质级别
	MainPrimaryColor           string                   // 主题色
	MainShowLyric              bool                     // 显示歌词
	MainLyricOffset            int                      // 偏移:ms
	MainShowLyricTrans         bool                     // 显示歌词翻译
	MainShowNotify             bool                     // 显示通知
	MainPProfPort              int                      // pprof端口
	MainAltScreen              bool                     // AltScreen显示模式
	MainDoubleColumn           bool                     // 是否双列显示
	MainDownloadDir            string                   // 指定下载目录
	MainShowAllSongsOfPlaylist bool                     // 显示歌单下所有歌曲

	UNMSwitch             bool     // UNM开关
	UNMSources            []string // UNM资源
	UNMSearchLimit        int      // UNM其他平台搜索限制
	UNMEnableLocalVip     bool     // UNM修改响应，解除会员限制
	UNMUnlockSoundEffects bool     // UNM修改响应，解除音质限制
	UNMQQCookieFile       string   // UNM QQ音乐cookie文件

	PlayerEngine         string // 播放引擎
	PlayerBeepMp3Decoder string // beep mp3解码器
	PlayerMpdBin         string // mpd路径
	PlayerMpdConfigFile  string // mpd配置文件
	PlayerMpdNetwork     string // mpd网络类型: tcp、unix
	PlayerMpdAddr        string // mpd地址
}

func NewRegistryWithDefault() *Registry {
	registry := &Registry{
		StartupShow:              true,
		StartupProgressOutBounce: true,
		StartupLoadingDuration:   time.Second * constants.StartupLoadingSeconds,
		StartupWelcome:           constants.AppName,
		StartupSignIn:            true,
		StartupCheckUpdate:       true,

		ProgressFullChar:  rune(constants.ProgressFullChar[0]),
		ProgressEmptyChar: rune(constants.ProgressEmptyChar[0]),

		MainShowTitle:        true,
		MainLoadingText:      constants.MainLoadingText,
		MainPlayerSongLevel:  service.Higher,
		MainPrimaryColor:     constants.AppPrimaryColor,
		MainShowLyric:        true,
		MainShowLyricTrans:   true,
		MainShowNotify:       true,
		MainPProfPort:        constants.MainPProfPort,
		MainAltScreen:        true,
		PlayerEngine:         constants.BeepPlayer,
		PlayerBeepMp3Decoder: constants.BeepGoMp3Decoder,

		UNMSwitch:             true,
		UNMSources:            []string{constants.UNMDefaultSources},
		UNMEnableLocalVip:     true,
		UNMUnlockSoundEffects: true,
	}

	if runtime.GOOS == "darwin" {
		registry.PlayerEngine = constants.OsxPlayer
	}

	return registry
}

func NewRegistryFromIniFile(filepath string) *Registry {
	registry := NewRegistryWithDefault()

	if err := ini.LoadExists(filepath); err != nil {
		return registry
	}

	registry.StartupShow = ini.Bool("startup.show", true)
	registry.StartupProgressOutBounce = ini.Bool("startup.progressOutBounce", true)
	registry.StartupLoadingDuration = time.Second * time.Duration(ini.Int("startup.loadingSeconds", constants.StartupLoadingSeconds))
	registry.StartupWelcome = ini.String("startup.welcome", constants.AppName)
	registry.StartupSignIn = ini.Bool("startup.signIn", true)
	registry.StartupCheckUpdate = ini.Bool("startup.checkUpdate", true)

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

	registry.MainShowTitle = ini.Bool("main.showTitle", true)
	registry.MainLoadingText = ini.String("main.loadingText", constants.MainLoadingText)
	songLevel := service.SongQualityLevel(ini.String("main.songLevel", string(service.Higher)))
	if songLevel.IsValid() {
		registry.MainPlayerSongLevel = songLevel
	}
	primaryColor := ini.String("main.primaryColor", constants.AppPrimaryColor)
	if primaryColor != "" {
		registry.MainPrimaryColor = primaryColor
	} else {
		registry.MainPrimaryColor = constants.AppPrimaryColor
	}
	registry.MainShowLyric = ini.Bool("main.showLyric", true)
	registry.MainLyricOffset = ini.Int("main.lyricOffset", 0)
	registry.MainShowLyricTrans = ini.Bool("main.showLyricTrans", true)
	registry.MainShowNotify = ini.Bool("main.showNotify", true)
	registry.MainPProfPort = ini.Int("main.pprofPort", constants.MainPProfPort)
	registry.MainAltScreen = ini.Bool("main.altScreen", true)
	registry.MainDoubleColumn = ini.Bool("main.doubleColumn", true)
	registry.MainDownloadDir = ini.String("main.downloadDir", "")
	registry.MainShowAllSongsOfPlaylist = ini.Bool("main.showAllSongsOfPlaylist", false)

	defaultPlayer := constants.BeepPlayer
	if runtime.GOOS == "darwin" {
		defaultPlayer = constants.OsxPlayer
	}
	registry.PlayerEngine = ini.String("player.engine", defaultPlayer)
	registry.PlayerBeepMp3Decoder = ini.String("player.beepMp3Decoder", constants.BeepGoMp3Decoder)
	registry.PlayerMpdBin = ini.String("player.mpdBin", "")
	registry.PlayerMpdConfigFile = ini.String("player.mpdConfigFile", "")
	registry.PlayerMpdNetwork = ini.String("player.mpdNetwork", "")
	registry.PlayerMpdAddr = ini.String("player.mpdAddr", "")

	// UNM
	registry.UNMSwitch = ini.Bool("unm.switch", true)

	sourceStr := ini.String("unm.sources", "kuwo")
	if sourceStr != "" {
		var sources []string
		for _, source := range strings.Split(sourceStr, ",") {
			sources = append(sources, strings.TrimSpace(source))
		}
		registry.UNMSources = sources
	}

	registry.UNMSearchLimit = ini.Int("unm.searchLimit", 0)
	registry.UNMEnableLocalVip = ini.Bool("unm.enableLocalVip", true)
	registry.UNMUnlockSoundEffects = ini.Bool("unm.unlockSoundEffects", true)
	registry.UNMQQCookieFile = ini.String("unm.qqCookieFile", "")

	return registry
}
