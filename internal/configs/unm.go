package configs

type UNMOptions struct {
	Enable             bool     // UNM开关
	Sources            []string // UNM资源
	SearchLimit        int      // UNM其他平台搜索限制
	EnableLocalVip     bool     // UNM修改响应，解除会员限制
	UnlockSoundEffects bool     // UNM修改响应，解除音质限制
	QQCookieFile       string   // UNM QQ音乐cookie文件
	SkipInvalidTracks  bool     // UNM 跳过无效播放连接，例如酷我的无效提示...
}

// UNMConfig UNM (Unlock NetEase Music) 相关配置
type UNMConfig struct {
	// UNM开关
	Enable bool `koanf:"enable"`
	// UNM资源
	Sources []string `koanf:"sources"`
	// UNM其他平台搜索限制
	SearchLimit int `koanf:"searchLimit"`
	// UNM修改响应，解除会员限制
	EnableLocalVip bool `koanf:"enableLocalVip"`
	// UNM修改响应，解除音质限制
	UnlockSoundEffects bool `koanf:"unlockSoundEffects"`
	// UNM QQ音乐cookie文件
	QQCookieFile string `koanf:"qqCookieFile"`
	// UNM 跳过无效播放连接
	SkipInvalidTracks bool `koanf:"skipInvalidTracks"`
}
