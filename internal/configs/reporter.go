package configs

type ReporterOptions struct {
	Lastfm  ReporterLastfmOptions
	Netease ReporterNeteaseOptions
}

type ReporterLastfmOptions struct {
	Enable          bool   // 是否启用 Last.fm 上报
	Key             string // Last.fm API Key
	Secret          string // Last.fm API Shared Secret
	ScrobblePoint   int    // Last.fm 上报百分比
	OnlyFirstArtist bool   // Last.fm 只上报一位艺术家
	SkipDjRadio     bool   // Last.fm 上报跳过电台节目
}

type ReporterNeteaseOptions struct {
	Enable bool // 是否启用 Netease 上报
}

// ReporterConfig 播放状态上报配置
type ReporterConfig struct {
	Netease NeteaseReporterConfig `koanf:"netease"`
	Lastfm  LastfmReporterConfig  `koanf:"lastfm"`
}

// NeteaseReporterConfig 上报至网易云音乐的配置
type NeteaseReporterConfig struct {
	// 是否启用 Netease 上报
	Enable bool `koanf:"enable"`
}

// LastfmReporterConfig 上报至 Last.fm 的配置
type LastfmReporterConfig struct {
	// 是否启用 Last.fm 上报
	Enable bool `koanf:"enable"`
	// Last.fm API Key
	Key string `koanf:"key"`
	// Last.fm API Shared Secret
	Secret string `koanf:"secret"`
	// Last.fm 上报百分比
	ScrobblePoint int `koanf:"scrobblePoint"`
	// Last.fm 只上报一位艺术家
	OnlyFirstArtist bool `koanf:"onlyFirstArtist"`
	// Last.fm 上报跳过电台节目
	SkipDjRadio bool `koanf:"skipDjRadio"`
}
