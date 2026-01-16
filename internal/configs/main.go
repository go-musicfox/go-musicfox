package configs

import (
	"github.com/go-musicfox/netease-music/service"
)

type MainOptions struct {
	ShowTitle              bool                     // 主界面是否显示标题
	LoadingText            string                   // 主页面加载中提示
	PlayerSongLevel        service.SongQualityLevel // 歌曲音质级别
	PrimaryColor           string                   // 主题色
	ShowLyric              bool                     // 显示歌词
	LyricOffset            int                      // 偏移:ms
	ShowLyricTrans         bool                     // 显示歌词翻译
	ShowNotify             bool                     // 显示通知
	NotifyIcon             string                   // logo 图片名
	NotifyAlbumCover       bool                     // 通知显示专辑封面
	PProfPort              int                      // pprof端口
	AltScreen              bool                     // AltScreen显示模式
	EnableMouseEvent       bool                     // 启用鼠标事件
	DualColumn             bool                     // 是否双列显示
	DownloadDir            string                   // 指定下载目录
	DownloadFileNameTpl    string                   // 下载文件名模板
	DownloadLyricDir       string                   // 指定歌词文件下载目录
	ShowAllSongsOfPlaylist bool                     // 显示歌单下所有歌曲
	CacheDir               string                   // 指定缓存目录
	CacheLimit             int64                    // 缓存大小（以MB为单位），0为不使用缓存，-1为不限制，默认为0
	DynamicMenuRows        bool                     // 菜单行数动态变更
	UseDefaultKeyBindings  bool                     // 使用默认键绑定
	CenterEverything       bool                     // 界面全部居中
	NeteaseCookie          string                   // 网易云音乐登录cookie
	Debug                  bool                     // 是否启用 Debug
}

// MainConfig 主界面与核心功能配置
type MainConfig struct {
	// AltScreen 显示模式
	AltScreen bool `koanf:"altScreen"`
	// 启用鼠标事件
	EnableMouseEvent bool `koanf:"enableMouseEvent"`
	// 是否启用 Debug
	Debug bool `koanf:"debug"`
	// 播放时 UI 刷新帧率
	FrameRate FrameRate `koanf:"frameRate"`

	Notification NotificationConfig `koanf:"notification"`
	Lyric        LyricConfig        `koanf:"lyric"`
	Pprof        PprofConfig        `koanf:"pprof"`
	Account      AccountConfig      `koanf:"account"`
}

// NotificationConfig 桌面通知相关设置
type NotificationConfig struct {
	// 显示通知
	Enable bool `koanf:"enable"`
	// logo 图片名
	Icon string `koanf:"icon"`
	// 通知显示专辑封面
	AlbumCover bool `koanf:"albumCover"`
}

// LyricConfig 歌词显示相关设置
type LyricConfig struct {
	// 显示歌词
	Show bool `koanf:"show"`
	// 显示歌词翻译
	ShowTranslation bool `koanf:"showTranslation"`
	// 偏移: ms
	Offset int `koanf:"offset"`
	// 忽略歌词解析错误
	SkipParseErr bool `koanf:"skipParseErr"`
	// YRC 歌词渲染模式：simple(简单), smooth(平滑), wave(波浪), glow(发光)
	YrcRenderMode string `koanf:"yrcRenderMode"`
}

// PprofConfig Go 性能分析工具 pprof 的相关设置
type PprofConfig struct {
	// pprof 端口
	Port int `koanf:"port"`
}

// AccountConfig 账号相关配置
type AccountConfig struct {
	// 网易云音乐登录 Cookie
	NeteaseCookie string `koanf:"neteaseCookie"`
}
