package configs

type PlayerOptions struct {
	Engine         string // 播放引擎
	BeepMp3Decoder string // beep mp3解码器
	MpdBin         string // mpd路径
	MpdConfigFile  string // mpd配置文件
	MpdNetwork     string // mpd网络类型: tcp、unix
	MpdAddr        string // mpd地址
}
