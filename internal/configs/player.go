package configs

import "github.com/go-musicfox/netease-music/service"

type PlayerOptions struct {
	Engine          string // 播放引擎
	BeepMp3Decoder  string // beep mp3解码器
	MpdBin          string // mpd路径
	MpdConfigFile   string // mpd配置文件
	MpdNetwork      string // mpd网络类型: tcp、unix
	MpdAddr         string // mpd地址
	MpdAutoStart    bool   // mpd自动启动
	MaxPlayErrCount int    // 最大错误重试次数
	MpvBin          string // mpv路径
}

// PlayerConfig 播放器引擎与行为配置
type PlayerConfig struct {
	// 播放引擎
	Engine string `koanf:"engine"`
	// 最大错误重试次数
	MaxPlayErrCount int `koanf:"maxPlayErrCount"`
	// 歌曲音质级别
	SongLevel service.SongQualityLevel `koanf:"songLevel"`
	// 显示歌单下所有歌曲
	ShowAllSongsOfPlaylist bool `koanf:"showAllSongsOfPlaylist"`

	Beep BeepConfig `koanf:"beep"`
	Mpd  MpdConfig  `koanf:"mpd"`
	Mpv  MpvConfig  `koanf:"mpv"`
}

// BeepConfig `beep` 引擎专属配置
type BeepConfig struct {
	// beep mp3解码器
	Mp3Decoder string `koanf:"mp3Decoder"`
}

// MpdConfig `mpd` 引擎专属配置
type MpdConfig struct {
	// mpd路径
	Bin string `koanf:"bin"`
	// mpd配置文件
	ConfigFile string `koanf:"configFile"`
	// mpd网络类型: tcp、unix
	Network string `koanf:"network"`
	// mpd地址
	Addr string `koanf:"addr"`
	// mpd自动启动
	AutoStart bool `koanf:"autoStart"`
}

// MpvConfig `mpv` 引擎专属配置
type MpvConfig struct {
	// mpv路径
	Bin string `koanf:"bin"`
}
