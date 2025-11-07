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

// StorageConfig 下载、缓存等文件存储相关配置
type StorageConfig struct {
	// 指定下载目录
	DownloadDir string `koanf:"downloadDir"`
	// 指定歌词文件下载目录
	LyricDir string `koanf:"lyricDir"`
	// 下载歌曲的同时下载歌词
	DownloadSongWithLyric bool `koanf:"downloadSongWithLyric"`
	// 下载文件名模板
	FileNameTpl string `koanf:"fileNameTpl"`

	Cache CacheConfig `koanf:"cache"`
}

// CacheConfig 音乐播放缓存相关设置
type CacheConfig struct {
	// 指定缓存目录
	Dir string `koanf:"dir"`
	// 音乐缓存路径，相对于 Dir
	MusicDir string `koanf:"musicDir"`
	// 缓存大小（以MB为单位），0为不使用缓存，-1为不限制
	Limit int64 `koanf:"limit"`
}
