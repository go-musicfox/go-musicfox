package ui

// Layout constants for the terminal UI, shared across renderers.
// All values are in terminal rows or columns.
// 终端 UI 布局常量，跨渲染器共享。所有数值单位均为终端行或列。

const (
	// ---- Vertical layout (rows) / 垂直布局（行） ----

	// SongInfoLines is the number of rows consumed by SongInfoRenderer:
	// 1 content line + 1 blank separator line.
	// 歌曲信息渲染器占用的行数：1 行内容 + 1 行空行分隔。
	SongInfoLines = 2

	// ProgressBarLines is the number of rows consumed by ProgressRenderer.
	// 进度条渲染器占用的行数。
	ProgressBarLines = 1

	// FixedTopBottomRows is the total number of fixed rows at the bottom of
	// the terminal after the menu area: SongInfo (2) + Progress (1) + bottom
	// margin (2).  Used to calculate available space for lyrics/cover/spectrum.
	// 终端底部固定区域总行数：歌曲信息(2) + 进度条(1) + 底部边距(2)。
	// 用于计算歌词/封面/频谱的可用空间。
	FixedTopBottomRows = SongInfoLines + ProgressBarLines + 2

	// EndRowMargin is the number of rows reserved from the terminal bottom.
	// Lyrics and other content should not extend beyond WindowHeight - EndRowMargin.
	// 从终端底部预留的行数。歌词等内容不应超过 WindowHeight - EndRowMargin。
	EndRowMargin = 4

	// MinSpaceHeight is the minimum height of available space for lyrics/cover
	// to be displayed at all.
	// 歌词/封面可显示的最小可用空间高度。
	MinSpaceHeight = 3

	// FullLyricLines is the maximum number of lyric lines to display when
	// sufficient vertical space is available.
	// 垂直空间充裕时显示的最大歌词行数。
	FullLyricLines = 5

	// CompactLyricLines is the reduced number of lyric lines to display when
	// vertical space is tight.
	// 垂直空间紧张时显示的缩减歌词行数。
	CompactLyricLines = 3

	// ---- Lyric horizontal layout (columns) / 歌词水平布局（列） ----

	// LyricHorizontalMargin is the horizontal margin (columns) for lyrics.
	// Used as left offset from MenuStartColumn in non-centered mode and
	// as right margin from window edge for width calculations.
	// 歌词的水平边距（列）。非居中模式下作为距 MenuStartColumn 的左偏移量，
	// 同时也作为宽度计算中距窗口右边缘的边距。
	LyricHorizontalMargin = 4

	// DualColumnLyricPadding is the additional right padding for lyrics in
	// dual-column menu mode.
	// 双列菜单模式下歌词的额外右内边距。
	DualColumnLyricPadding = 3

	// CoverRightPadding is the horizontal gap (columns) between the cover
	// image and the lyrics.
	// 封面图片与歌词之间的水平间距（列）。
	CoverRightPadding = 2

	// MinLyricWidth is the minimum rendered width for lyric lines.
	// 歌词行的最小渲染宽度。
	MinLyricWidth = 20

	// LyricBaseWidth is the reference width used for calculating
	// dynamic extra padding in centered lyric mode.
	// 居中歌词模式下计算动态额外边距的基准宽度。
	LyricBaseWidth = 40

	// LyricPaddingDivisor is the divisor for dynamic extra padding
	// calculation in centered lyric mode.
	// 居中歌词模式下动态额外边距计算的分母。
	LyricPaddingDivisor = 5

	// MinLyricExtraPadding is the minimum extra padding for centered
	// lyric rendering.
	// 居中歌词渲染的最小额外边距。
	MinLyricExtraPadding = 8

	// ---- Cover image constants / 封面图片常量 ----

	// DefaultCoverWidthRatio is the fraction of window width to use for
	// the cover image when the config value is invalid.
	// 当配置值无效时，封面图片占窗口宽度的默认比例。
	DefaultCoverWidthRatio = 0.3

	// MinCoverCols is the minimum cover width in columns.
	// 封面的最小列宽。
	MinCoverCols = 10

	// MinCoverRows is the minimum cover height in rows.
	// 封面的最小行高。
	MinCoverRows = 3

	// TerminalCellAspectRatio is the typical width:height ratio of a
	// terminal character cell.  Cells are roughly twice as tall as wide
	// (e.g. 8x16 pixels), so dividing cols by this ratio yields a
	// visually square image.
	// 终端字符格的典型宽高比。字符格高度大约是宽度的两倍
	// （如 8x16 像素），因此 cols 除以该比值可得到视觉上的正方形图像。
	TerminalCellAspectRatio = 2

	// CoverEndRowMargin is added to the reserved bottom area when
	// calculating cover positioning without lyrics.
	// 在没有歌词时计算封面位置时，在底部预留区域上增加的额外行边距。
	CoverEndRowMargin = 2

	// ---- Spectrum constants / 频谱常量 ----

	// SpectrumVerticalPadding is the number of blank rows above and
	// below the spectrum bars.
	// 频谱条上方和下方的空白行数。
	SpectrumVerticalPadding = 1

	// SpectrumReservedLines is the number of lines reserved above the
	// spectrum (for song info, progress bar, etc).
	// 频谱上方预留的行数（用于歌曲信息、进度条等）。
	SpectrumReservedLines = 2

	// ---- Progress bar constants / 进度条常量 ----

	// ProgressTimeDisplayWidth is the width in columns reserved for the
	// time display (e.g. "03:45/04:30") and surrounding padding.
	// 时间显示区域（如 "03:45/04:30"）及其周围内边距的列宽。
	ProgressTimeDisplayWidth = 14

	// ProgressLongDurationThreshold is the threshold (in minutes) above
	// which the time display uses 3-digit minute format (e.g. "123:45").
	// 时长阈值（分钟）：超过此值时间显示使用三位数分钟格式（如 "123:45"）。
	ProgressLongDurationThreshold = 100

	// ---- Song info constants / 歌曲信息常量 ----

	// SongInfoPrefixBaseWidth is the base character width of the prefix
	// area (mode indicator + volume + playing status icon) before the
	// song name.
	// 歌名前缀区域（播放模式指示器 + 音量 + 播放状态图标）的基础字符宽度。
	SongInfoPrefixBaseWidth = 10

	// MenuArrowWidth is the estimated width of the menu selection arrow
	// (and its surrounding padding) that marks the currently selected item.
	// 菜单选中项箭头（含周围内边距）的估算宽度。
	MenuArrowWidth = 4

	// SongInfoPrefixExtraWidth is the additional characters added to the
	// prefix when there is sufficient menu start column space (for mode
	// and volume display).
	// 当菜单起始列空间充足时，额外添加到前缀的字符宽度（用于显示模式和音量）。
	SongInfoPrefixExtraWidth = 12

	// SongInfoHorizontalPadding is the total horizontal padding (2 on each
	// side) used in centered-mode song info width calculations.
	// 居中模式下歌曲信息宽度计算的总水平内边距（每边 2 列）。
	SongInfoHorizontalPadding = 4

	// ---- Menu/event handler layout constants / 菜单和事件处理布局常量 ----

	// PlayModeRowOffset is the offset from the window bottom for the play
	// mode indicator row (within the fixed bottom area).
	// 播放模式指示器行距窗口底部的偏移量（位于底部固定区域内）。
	PlayModeRowOffset = 3

	// PlayModeClickWidth is the width (columns) of the clickable play mode
	// indicator area from MenuStartColumn.
	// 可点击的播放模式指示器区域距 MenuStartColumn 的宽度（列）。
	PlayModeClickWidth = 5

	// DualColumnWindowThreshold is the minimum window width for the
	// dual-column layout to use a max-width column (44 cols).
	// 双列布局使用最大宽度列（44 列）的最小窗口宽度阈值。
	DualColumnWindowThreshold = 88

	// MaxLeftColumnWidth is the maximum width of the left column in
	// dual-column layout mode.
	// 双列布局模式下左列的最大宽度。
	MaxLeftColumnWidth = 44

	// ---- Dynamic menu mode constants / 动态菜单模式常量 ----

	// DynamicMenuOverhead is the row offset from menu bottom to the start
	// of rendered components in dynamic menu mode (includes search input
	// area and spacing above lyrics).
	// 动态菜单模式下从菜单底部到渲染组件起始位置的行偏移量
	// （包含搜索输入区域和歌词上方的间距）。
	DynamicMenuOverhead = 4

	// DynamicMenuLyricLines is the fixed number of lyric rows in dynamic
	// menu mode: FullLyricLines (5) + 1 spacer line.
	// 动态菜单模式下歌词的固定行数：FullLyricLines(5) + 1 行间距。
	DynamicMenuLyricLines = FullLyricLines + 1

	// DynamicMenuSpectrumLines is the estimated maximum number of spectrum
	// rows in dynamic menu mode: 8 bars + top/bottom padding = 10.
	// Used only for the BottomHeight estimate; the actual spectrum height
	// is determined adaptively by SpectrumRenderer.layout() based on
	// available space.
	// 动态菜单模式下频谱的预估最大行数（8 条 + 上下各 1 行空白 = 10）。
	// 仅用于 BottomHeight 预估值；实际频谱高度由 SpectrumRenderer.layout()
	// 根据可用空间自适应决定。
	DynamicMenuSpectrumLines = 8 + 2*SpectrumVerticalPadding

	// DynamicMenuBottomLines is the number of rows at the bottom after the
	// spectrum in dynamic menu mode: song info + progress + extra spacing.
	// 动态菜单模式下频谱下方的底部行数：歌曲信息 + 进度条 + 额外间距。
	DynamicMenuBottomLines = SongInfoLines + ProgressBarLines + 3

	// ---- Form page constants / 表单页面常量 ----

	// FormBottomReservedRows is the number of rows reserved at the bottom
	// of form pages (search, login, etc.) for the player bar area.
	// 表单页面（搜索、登录等）底部为播放器栏区域预留的行数。
	FormBottomReservedRows = 3
)
