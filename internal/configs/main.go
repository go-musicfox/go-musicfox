package configs

import "github.com/go-musicfox/netease-music/service"

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
	FrameRate  FrameRate        `koanf:"frameRate"`
	Visualizer VisualizerConfig `koanf:"visualizer"`

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

// VisualizerConfig controls live spectrum rendering.
// Character fields for each style accept a single Unicode character; the first rune is used.
type VisualizerConfig struct {
	Enable      bool   `koanf:"enable"`
	MaxHeight   int    `koanf:"maxHeight"`
	Style       string `koanf:"style"`       // "bar", "line", "mirror_bar", "dot", "oscilloscope", "vectorscope", "spectrogram"
	ChannelMode string `koanf:"channelMode"` // "dual" (default) or "mono"

	// --- bar style ---
	BarHalfBlock          string `koanf:"barHalfBlock"`
	BarFullBlock          string `koanf:"barFullBlock"`
	BarEmptyBlock         string `koanf:"barEmptyBlock"`
	BarVerticalGradient   bool   `koanf:"barVerticalGradient"`
	BarHorizontalGradient bool   `koanf:"barHorizontalGradient"` // enable per-bar horizontal color gradient
	ShowIdleBarHeads      bool   `koanf:"showIdleBarHeads"`      // show small caps when bars are nearly silent (cava-style)
	BarOrientation        string `koanf:"barOrientation"`        // "bottom" (default), "top", "left", "right", "horizontal", "vertical"

	// --- mirror_bar style (falls back to bar values when left empty) ---
	MirrorBarHalfBlock  string `koanf:"mirrorBarHalfBlock"`
	MirrorBarFullBlock  string `koanf:"mirrorBarFullBlock"`
	MirrorBarEmptyBlock string `koanf:"mirrorBarEmptyBlock"`

	// --- line style ---
	LineMode       string `koanf:"lineMode"`       // "braille" (default) or "block"
	LineFullBlock  string `koanf:"lineFullBlock"`  // block mode: band-position character
	LineHalfBlock  string `koanf:"lineHalfBlock"`  // block mode: interpolated segments (falls back to full)
	LineEmptyBlock string `koanf:"lineEmptyBlock"` // block mode: empty cell (default space)

	// --- dot style ---
	DotMode       string `koanf:"dotMode"`       // "braille" (default) or "block"
	DotFullBlock  string `koanf:"dotFullBlock"`  // block mode: dot character
	DotHalfBlock  string `koanf:"dotHalfBlock"`  // block mode: reserved
	DotEmptyBlock string `koanf:"dotEmptyBlock"` // block mode: empty cell (default space)

	// --- oscilloscope style (time-domain waveform) ---
	OscilloscopeMode       string `koanf:"oscilloscopeMode"`       // "braille" (default) or "block"
	OscilloscopeScatter    bool   `koanf:"oscilloscopeScatter"`    // scatter dots vs connected line
	OscilloscopeFullBlock  string `koanf:"oscilloscopeFullBlock"`
	OscilloscopeHalfBlock  string `koanf:"oscilloscopeHalfBlock"`
	OscilloscopeEmptyBlock string `koanf:"oscilloscopeEmptyBlock"`

	// --- vectorscope style (Lissajous L×R scatter) ---
	VectorscopeMode       string `koanf:"vectorscopeMode"`       // "braille" (default) or "block"
	VectorscopeFullBlock  string `koanf:"vectorscopeFullBlock"`
	VectorscopeHalfBlock  string `koanf:"vectorscopeHalfBlock"`
	VectorscopeEmptyBlock string `koanf:"vectorscopeEmptyBlock"`

	// --- spectrogram style ---
	SpectrogramSpeed int `koanf:"spectrogramSpeed"` // scrolling speed (1=slow, 10=fast, default 4)

	// --- spectrum-wide settings (line, dot styles) ---
	SpectrumAverage  int  `koanf:"spectrumAverage"`  // FFT frame averaging count (1 = off, higher = smoother)
	SpectrumLogScale bool `koanf:"spectrumLogScale"` // use dB (true) or linear amplitude (false) for Y axis
	SpectrumPhaseDiff bool `koanf:"spectrumPhaseDiff"` // overlay channel phase correlation

	// --- cava-inspired smoothing (applied after spring/EMA, affects bar/mirror_bar) ---
	Monstercat float64 `koanf:"monstercat"` // cava monstercat smoothing (0=off, >0=enabled, typical 1.0-3.0)
	Waves      bool    `koanf:"waves"`      // cava waves smoothing mode
	Overshoot  float64 `koanf:"overshoot"`  // visual overshoot percentage (0-100, default 0)
}

// BarCharacters returns bar style glyphs with defaults.
func (c VisualizerConfig) BarCharacters() (halfBlock, fullBlock, emptyBlock rune) {
	return firstCharOrDefault(c.BarHalfBlock, "▌"),
		firstCharOrDefault(c.BarFullBlock, "█"),
		firstCharOrDefault(c.BarEmptyBlock, " ")
}

// MirrorBarCharacters returns mirror_bar glyphs, falling back to bar values when empty.
func (c VisualizerConfig) MirrorBarCharacters() (halfBlock, fullBlock, emptyBlock rune) {
	h := c.MirrorBarHalfBlock
	f := c.MirrorBarFullBlock
	e := c.MirrorBarEmptyBlock
	if h == "" {
		h = c.BarHalfBlock
	}
	if f == "" {
		f = c.BarFullBlock
	}
	if e == "" {
		e = c.BarEmptyBlock
	}
	return firstCharOrDefault(h, "▌"),
		firstCharOrDefault(f, "█"),
		firstCharOrDefault(e, " ")
}

// MaxBarHeight returns zero for an unlimited spectrum height.
func (c VisualizerConfig) MaxBarHeight() int {
	return max(0, c.MaxHeight)
}

// IsMono returns true when the visualizer should render a single combined channel
// instead of separate left/right stereo channels.
func (c VisualizerConfig) IsMono() bool {
	return c.ChannelMode == "mono"
}

// EffectiveBarOrientation returns the bar growth direction, defaulting to "bottom".
func (c VisualizerConfig) EffectiveBarOrientation() string {
	if c.BarOrientation == "" {
		return "bottom"
	}
	return c.BarOrientation
}

// EffectiveSpectrogramSpeed returns the spectrogram scrolling speed, defaulting to 4.
func (c VisualizerConfig) EffectiveSpectrogramSpeed() int {
	if c.SpectrogramSpeed <= 0 {
		return 4
	}
	return c.SpectrogramSpeed
}

// EffectiveOvershoot returns the overshoot multiplier (0 = no overshoot).
func (c VisualizerConfig) EffectiveOvershoot() float64 {
	if c.Overshoot <= 0 {
		return 0
	}
	return c.Overshoot / 100.0
}

// IsSpectrogram returns true when the style is "spectrogram".
func (c VisualizerConfig) IsSpectrogram() bool {
	return c.Style == "spectrogram"
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
	// 歌词渲染模式：smooth(平滑), wave(波浪), glow(发光)
	RenderMode string `koanf:"renderMode"`
	// 封面图设置
	Cover CoverConfig `koanf:"cover"`
	// 桌面歌词设置
	DesktopLyrics DesktopLyricsConfig `koanf:"desktopLyrics"`
}

// DesktopLyricsConfig 桌面歌词设置
type DesktopLyricsConfig struct {
	// 启用桌面歌词
	Enable bool `koanf:"enable"`
	// X 位置系数（0.0-1.0，0.5 表示居中）
	XPositionFactor float64 `koanf:"xPositionFactor"`
	// Y 位置系数（0.0-1.0，0 表示底部，1 表示顶部）
	YPositionFactor float64 `koanf:"yPositionFactor"`
	// 字体大小
	FontSize float64 `koanf:"fontSize"`
	// 字体名称
	FontName string `koanf:"fontName"`
	// 文字颜色（十六进制）
	TextColor string `koanf:"textColor"`
	// 文字阴影颜色（十六进制）
	ShadowColor string `koanf:"shadowColor"`
	// 阴影模糊半径
	ShadowRadius float64 `koanf:"shadowRadius"`
	// 背景颜色（十六进制）
	BackgroundColor string `koanf:"backgroundColor"`
	// 背景透明度（0.0-1.0）
	BackgroundAlpha float64 `koanf:"backgroundAlpha"`
	// 背景圆角半径（像素，0 表示无圆角）
	CornerRadius float64 `koanf:"cornerRadius"`
	// 窗口整体透明度（0.0-1.0）
	WindowAlpha float64 `koanf:"windowAlpha"`
	// 单行模式（只显示当前行）
	OneLineMode bool `koanf:"oneLineMode"`
	// 暂停时隐藏歌词
	HideOnPause bool `koanf:"hideOnPause"`
	// 允许拖拽移动歌词窗口
	Draggable bool `koanf:"draggable"`
	// 窗口最大宽度占屏幕比例（0.3-0.9，默认 0.7）
	MaxWindowWidth float64 `koanf:"maxWindowWidth"`

	// 桌面歌词频谱可视化
	SpectrumEnabled    bool    `koanf:"spectrumEnabled"`    // 启用频谱
	SpectrumHeight     float64 `koanf:"spectrumHeight"`     // 频谱区域高度（像素）
	SpectrumBarCount   int     `koanf:"spectrumBarCount"`   // 频段数量（≤64）
	SpectrumBarGap     float64 `koanf:"spectrumBarGap"`     // 频段间距（像素）
	SpectrumFPS        int     `koanf:"spectrumFPS"`        // 刷新帧率
	SpectrumOpacity    float64 `koanf:"spectrumOpacity"`    // 频谱整体透明度
	SpectrumStyle      string  `koanf:"spectrumStyle"`      // 样式："bar"(默认) / "mirror"
	SpectrumColorLow   string  `koanf:"spectrumColorLow"`   // 低频颜色（hex）
	SpectrumColorMid   string  `koanf:"spectrumColorMid"`   // 中频颜色（hex）
	SpectrumColorHigh  string  `koanf:"spectrumColorHigh"`  // 高频颜色（hex）
}

// CoverConfig 封面图显示设置
type CoverConfig struct {
	// 是否显示封面图（需要支持Kitty图形协议的终端）
	Show bool `koanf:"show"`
	// 封面图宽度占窗口宽度的比例（取值范围 0.1-0.8）
	WidthRatio float64 `koanf:"widthRatio"`
	// 封面图圆角半径百分比（取值范围 0-100，默认 8 即 8%）
	CornerRadius int `koanf:"cornerRadius"`
	// 是否启用旋转
	Spin bool `koanf:"spin"`
	// 旋转帧率（取值范围 1-60，默认 30）
	SpinFPS int `koanf:"spinFPS"`
	// 旋转一圈的时长（秒，取值范围 1-30，默认 6）
	SpinDuration int `koanf:"spinDuration"`
}

// PprofConfig Go 性能分析工具 pprof 的相关设置
type PprofConfig struct {
	// pprof 端口
	Port int `koanf:"port"`
}

// SpectrumEffectiveHeight 返回有效频谱高度（启用时 > 0）
func (c DesktopLyricsConfig) SpectrumEffectiveHeight() float64 {
	if !c.SpectrumEnabled {
		return 0
	}
	if c.SpectrumHeight <= 0 {
		return 60 // default 60px
	}
	return c.SpectrumHeight
}

// SpectrumEffectiveBarCount 返回有效频段数（2-64）
func (c DesktopLyricsConfig) SpectrumEffectiveBarCount() int {
	if c.SpectrumBarCount <= 0 || c.SpectrumBarCount > 64 {
		return 64
	}
	return c.SpectrumBarCount
}

// SpectrumEffectiveBarGap 返回频段间距，默认1px
func (c DesktopLyricsConfig) SpectrumEffectiveBarGap() float64 {
	if c.SpectrumBarGap <= 0 {
		return 1
	}
	return c.SpectrumBarGap
}

// AccountConfig 账号相关配置
type AccountConfig struct {
	// 网易云音乐登录 Cookie
	NeteaseCookie string `koanf:"neteaseCookie"`
}
