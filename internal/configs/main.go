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
	CenterEverything       bool                     // 界面全部居中
}
