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
