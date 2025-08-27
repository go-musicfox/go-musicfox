package configs

// StorgeOptions 下载、缓存等相关配置
type StorgeOptions struct {
	DownloadDir           string // 指定下载目录
	DownloadFileNameTpl   string // 下载文件名模板
	DownloadLyricDir      string // 指定歌词文件下载目录
	DownloadSongWithLyric bool   // 下载歌曲的同时下载歌词，若歌曲已下载则下载歌词，仅当歌曲下载成功才下载歌词
	CacheDir              string // 指定缓存目录
	CacheLimit            int64  // 缓存大小（以MB为单位），0为不使用缓存，-1为不限制，默认为0
}
