package configs

import (
	"runtime"
	"strings"
	"time"

	"github.com/go-musicfox/go-musicfox/pkg/constants"

	"github.com/go-musicfox/netease-music/service"
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

	ProgressFirstEmptyChar rune // 进度条第一个未加载字符
	ProgressEmptyChar      rune // 进度条未加载字符
	ProgressLastEmptyChar  rune // 进度条最后一个未加载字符
	ProgressFirstFullChar  rune // 进度条第一个已加载字符
	ProgressFullChar       rune // 进度条已加载字符
	ProgressLastFullChar   rune // 进度条最后一个已加载字符

	MainShowTitle              bool                     // 主界面是否显示标题
	MainLoadingText            string                   // 主页面加载中提示
	MainPlayerSongLevel        service.SongQualityLevel // 歌曲音质级别
	MainPrimaryColor           string                   // 主题色
	MainShowLyric              bool                     // 显示歌词
	MainLyricOffset            int                      // 偏移:ms
	MainShowLyricTrans         bool                     // 显示歌词翻译
	MainShowNotify             bool                     // 显示通知
	MainNotifyIcon             string                   // logo 图片名
	MainPProfPort              int                      // pprof端口
	MainAltScreen              bool                     // AltScreen显示模式
	MainEnableMouseEvent       bool                     // 启用鼠标事件
	MainDoubleColumn           bool                     // 是否双列显示
	MainDownloadDir            string                   // 指定下载目录
	MainShowAllSongsOfPlaylist bool                     // 显示歌单下所有歌曲

	AutoPlay       bool   // 是否自动开始播放
	AutoPlayList   string // 自动播放列表：dailyReco（每日推荐）、like（我喜欢的音乐）、name:[歌单名]
	AutoPlayOffset int    // 播放偏移：0为歌单第一项，-1为歌单最后一项
	AutoPlayRandom bool   // 是否随机选择歌曲
	AutoPlayMode   string /* 播放模式，如果不为last,将覆盖autoPlayList为true时的播放模式:
	listLoop, order, singleLoop, random, intelligent（心动）, last（上次退出时的模式），默认为last */

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

		ProgressFirstEmptyChar: []rune(constants.ProgressEmptyChar)[0],
		ProgressEmptyChar:      []rune(constants.ProgressEmptyChar)[0],
		ProgressLastEmptyChar:  []rune(constants.ProgressEmptyChar)[0],
		ProgressFirstFullChar:  []rune(constants.ProgressFullChar)[0],
		ProgressFullChar:       []rune(constants.ProgressFullChar)[0],
		ProgressLastFullChar:   []rune(constants.ProgressFullChar)[0],

		MainShowTitle:        true,
		MainLoadingText:      constants.MainLoadingText,
		MainPlayerSongLevel:  service.Higher,
		MainPrimaryColor:     constants.AppPrimaryColor,
		MainShowLyric:        true,
		MainShowLyricTrans:   true,
		MainShowNotify:       true,
		MainNotifyIcon:       constants.DefaultNotifyIcon,
		MainPProfPort:        constants.MainPProfPort,
		MainAltScreen:        true,
		MainEnableMouseEvent: true,
		PlayerEngine:         constants.BeepPlayer,
		PlayerBeepMp3Decoder: constants.BeepGoMp3Decoder,

		AutoPlay:       false,
		AutoPlayList:   "dailyReco",
		AutoPlayOffset: 0,
		AutoPlayMode:   "last",

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

	emptyChar := ini.String("progress.emptyChar", constants.ProgressEmptyChar)
	registry.ProgressEmptyChar = firstCharOrDefault(emptyChar, constants.ProgressEmptyChar)
	firstEmptyChar := ini.String("progress.firstEmptyChar", constants.ProgressEmptyChar)
	registry.ProgressFirstEmptyChar = firstCharOrDefault(firstEmptyChar, constants.ProgressEmptyChar)
	lastEmptyChar := ini.String("progress.lastEmptyChar", constants.ProgressEmptyChar)
	registry.ProgressLastEmptyChar = firstCharOrDefault(lastEmptyChar, constants.ProgressEmptyChar)

	fullChar := ini.String("progress.fullChar", constants.ProgressFullChar)
	registry.ProgressFullChar = firstCharOrDefault(fullChar, constants.ProgressFullChar)
	firstFullChar := ini.String("progress.firstFullChar", constants.ProgressFullChar)
	registry.ProgressFirstFullChar = firstCharOrDefault(firstFullChar, constants.ProgressFullChar)
	lastFullChar := ini.String("progress.lastFullChar", constants.ProgressFullChar)
	registry.ProgressLastFullChar = firstCharOrDefault(lastFullChar, constants.ProgressFullChar)

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
	registry.MainNotifyIcon = ini.String("main.notifyIcon", constants.DefaultNotifyIcon)
	registry.MainPProfPort = ini.Int("main.pprofPort", constants.MainPProfPort)
	registry.MainAltScreen = ini.Bool("main.altScreen", true)
	registry.MainEnableMouseEvent = ini.Bool("main.enableMouseEvent", true)
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

	// Auto play
	registry.AutoPlay = ini.Bool("autoplay.autoPlay")
	registry.AutoPlayList = ini.String("autoplay.autoPlayList")
	registry.AutoPlayOffset = ini.Int("autoplay.offset")
	registry.AutoPlayRandom = ini.Bool("autoplay.random")
	registry.AutoPlayMode = ini.String("autoplay.playMode")

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

func firstCharOrDefault(s, defaultStr string) rune {
	if len(s) > 0 {
		return []rune(s)[0]
	}
	return []rune(defaultStr)[0]
}
