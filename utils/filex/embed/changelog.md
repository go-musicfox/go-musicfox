# go-musicfox v5.0.0

## 1. 启动页全面升级

启动页已从简单的进度条升级为完整的动画启动体验，并提供丰富的配置选项。

### 动画效果

内置 **8 种启动动画**，通过 `[startup]` → `animation` 配置切换：

| 动画模式 | 说明 |
|---------|------|
| `sequence`（默认） | 经典游戏/IDE 风格启动序列：渐显 → 彩虹色扫描 → 短暂 glitch 过渡 → 稳定 logo |
| `fade-in` | Logo 从暗到亮渐显（使用抖动模拟终端友好的透明度） |
| `rainbow-wave` | 彩虹色波扫过 logo |
| `spinner` | Logo + 自定义动画旋转器（`◐◓◑◒`） |
| `slide-in` | Logo 从右侧弹性滑入（EaseOutElastic 缓动曲线） |
| `glitch` | 确定性字符损坏 + RGB 颜色分离效果 |
| `matrix-rain` | 全屏《黑客帝国》风格绿色字符雨背景 + logo 覆盖 |
| `particle-burst` | 粒子从四周汇聚成 logo |

### 配置选项

```toml
[startup]
enable = true            # 是否显示启动页
welcome = "go-musicfox"  # 启动页欢迎语/logo 文字
loadingSeconds = 2        # 启动页持续时长（秒）
animation = "sequence"    # 动画效果
progressOutBounce = true  # 进度条回弹效果
reducedMotion = false     # 停用动画以减少动态效果
signIn = true             # 每天启动时自动签到
checkUpdate = true        # 启动时检查更新
```

### 修复

- **修复启动时长不准的问题**：`loadingSeconds` 现在精确控制启动页持续时间，不再受初始化耗时影响

---

## 2. macOS GUI 桌面歌词 + 频谱显示

全新的原生 macOS 桌面歌词窗口，支持丰富的视觉效果与频谱可视化。

### 桌面歌词功能

- **原生 NSWindow 实现**：基于 objc 的 borderless 浮动窗口，支持所有 Spaces、始终置顶
- **YRC 逐字高亮**：支持 YRC 格式歌词的逐字渐亮效果，使用 NSMutableAttributedString 实时更新颜色
- **动画滚动**：超长歌词自动水平滚动，带初始延迟和末尾停顿
- **拖拽定位**：鼠标拖拽移动窗口，位置自动按屏幕百分比因子保存到 BoltDB
- **文本阴影**：可配置的文字阴影（颜色、模糊半径）
- **单行/双行模式**：`oneLineMode` 控制显示一行还是两行歌词
- **暂停隐藏**：`hideOnPause` 暂停时自动隐藏

### 视觉定制

```toml
[main.lyric.desktopLyrics]
enable = true
fontSize = 24.0           # 字体大小
fontName = ""             # 字体名称（空=系统默认）
textColor = "#FFFFFF"     # 文字颜色（十六进制）
shadowColor = "#000000"   # 阴影颜色
shadowRadius = 2.0        # 阴影模糊半径
backgroundColor = "#000000"   # 背景颜色
backgroundAlpha = 0.6     # 背景透明度
cornerRadius = 12.0       # 背景圆角
windowAlpha = 0.9         # 窗口整体透明度
oneLineMode = false       # 单行模式
hideOnPause = false       # 暂停时隐藏
draggable = true          # 允许拖拽移动
maxWindowWidth = 0.7      # 最大窗口宽度占比
```

### 桌面频谱可视化

桌面歌词窗口内直接渲染频谱，与音频同步：

- **频谱样式**：`bar`（柱状）/ `mirror`（镜像）/ `capsule`（胶囊）/ `dot`（圆点）/ `fire`（火焰渐变色）/ `led`（复古 LED）/ `circular`（圆形）/ `line`（曲线）/ `waveform`（波形）/ `ring_arc`（环弧）/ `ripple`（涟漪圈）

```toml
spectrumEnabled = true
spectrumHeight = 60       # 频谱区域高度（像素）
spectrumBarCount = 64     # 频段数量
spectrumBarGap = 1        # 频段间距（像素）
spectrumFPS = 30          # 刷新帧率
spectrumOpacity = 0.8     # 透明度
spectrumStyle = "bar"     # 样式
spectrumColorLow = "#00FF00"   # 低频颜色
spectrumColorMid = "#FFFF00"   # 中频颜色
spectrumColorHigh = "#FF0000"  # 高频颜色
```

---

## 3. 终端 TUI 频谱图显示

终端内实时音频频谱可视化，支持多种渲染风格和高度配置。**当前仅支持 macOS `osx` 和 `beep` 播放引擎**（需要通过 `MTAudioProcessingTap` 或内置 PCM 流水线获取原始音频数据）。

### 频谱分析引擎

基于 FFT 的实时频谱分析：
- **1024 点 FFT**，汉宁窗处理，64 频段输出
- **临界阻尼弹簧**（harmonica）平滑每个频段，避免帧间跳跃
- **EMA 加权平均** 支持自定义 `spectrumAverage` 帧叠加
- **Monstercat / Waves 后处理**（cava 风格平滑）
- **可选对数刻度**（dB）或线性振幅缩放
- **立体声双通道**：支持独立 L/R 渲染或 `mono` 合并模式

### 频谱渲染风格

通过 `[main.visualizer]` → `style` 配置：

| 样式 | 说明 |
|------|------|
| `bar`（默认） | 横向进度条，支持 6 种方向（bottom/top/left/right/horizontal/vertical） |
| `line` | 点阵盲文连曲线，支持 braille/block 双模式 |
| `mirror_bar` | 镜像柱状，支持独立 L/R 通道 |
| `dot` | 点阵盲文散点图，支持 braille/block 双模式 |
| `oscilloscope` | 时域原始波形（示波器），支持 braille/block 双模式 |
| `vectorscope` | L×R 李萨如散点图，支持 braille/block 双模式 |

### 颜色配置

- **水平进度色渐变**：主题色 start → end 渐变
- **垂直渐变**：顶部偏暗，底部明亮
- **水平渐变**：每行独立颜色插值
- **相位差可视化**（`spectrumPhaseDiff`）：左右声道相位不同步时向橙色偏移
- 字符三态渲染：满单元 / 半单元（前景+背景双色渐变） / 空白单元

### 配置选项

```toml
[main.visualizer]
enable = false
maxHeight = 0             # 最大频谱行数（0=不限制，填满可用空间）
style = "bar"             # 渲染样式
channelMode = "dual"      # "dual"（立体声）或 "mono"（单声道）
barOrientation = "bottom" # 柱状方向
barVerticalGradient = false   # 垂直渐变色
barHorizontalGradient = false # 水平渐变色
showIdleBarHeads = true       # 微弱信号显示小帽
monstercat = 0.0              # cava 风格平滑强度（0=关闭）
spectrumAverage = 1           # FFT 帧平均数量
spectrumLogScale = false      # 对数刻度
spectrumPhaseDiff = false     # 相位差可视化
```

---

## 4. 接入 Lipgloss 布局体系

foxful-cli v1.0.1 全面集成 lipgloss 布局引擎，提供丰富的 UI 组件能力：

### 弹窗系统（Popup）

- **Markdown 弹窗**：支持 Markdown 内容渲染，可滚动、可缩放
  - 帮助页弹窗：动态生成快捷键 Markdown 表格（内建键位 + 自定义键位两大类分组展示）
  - 更新日志弹窗：首次启动新版本自动弹出，Markdown 格式
- **Anchor 定位**：弹窗支持 Center/Top/Bottom/Left/Right 等锚点定位
- **可配置尺寸**：弹窗宽高按终端比例自适应，支持 resize
- **关闭快捷键**：支持自定义键位（Esc/q）

### 通知栏（Notification）

- 原生 TUI 内鹅卵石（toast）通知
- 支持 4 种级别：Info（蓝）/ Success（绿）/ Warning（黄）/ Error（红）
- 支持操作按钮（如「打开链接」跳转浏览器）
- 自动消失时长可配置（`inAppTimeout`）

### Tab 组件

- 登录页集成 foxful-cli Tab 切换（手机号 / 二维码 / Cookie 登录）
- 支持鼠标点击切换和键盘导航

### 状态栏（StatusBar）

- 面包屑导航路径显示
- 自适应宽度约束（终端过窄时渐进隐藏时间→中间模块）

### 主题集成

- Theme 配置化支持（Primary/Secondary 等语义色、Highlight 组件样式、Popup 主题）
- ThemeList + ThemeSwitchKey 运行时主题切换机制

---

## 5. 完整的鼠标能力支持

全面支持 bubbletea 鼠标事件，覆盖 click、wheel、motion 三种消息类型。

### 鼠标点击（MouseClickMsg）

| 区域 | 左键行为 |
|------|---------|
| **播放模式图标** | 切换播放模式（列表循环/单曲/随机等） |
| **播放状态指示器**（`♫ ♪ ♫ ♪`） | 播放/暂停切换 |
| **喜欢按钮**（`♥`） | 切换喜欢/取消喜欢 |
| **歌曲名** | 弹出歌曲操作菜单 |
| **歌手名** | 跳转到歌手详情页 |
| **进度条** | 按点击位置 seek 跳转 |

### 鼠标滚轮（MouseWheelMsg）

- **菜单区**：滚动列表（委托给 foxful-cli 处理）
- **非菜单区**：上下滚轮调节音量
- **Ctrl + 滚轮**：细粒度音量调节（步长由 `mouseVolumeStep` 配置，默认 1，最大 20）

### 鼠标移动（MouseMotionMsg）

- **播放栏元素 hover**：模式/状态/喜欢/歌曲名/歌手名 — 自动切换指针为 `pointer`
- **进度条 hover**：切换指针为 `pointer`
- **返回按钮 hover**：页面级 breadcrumb hover 检测
- **Tab hover**：登录页 Tab 按钮 hover
- **输入框 hover**：登录页 / 搜索页输入框 hover 切换指针为 `text`
- **菜单项 hover**：委托 foxful-cli 处理菜单项高亮

### 右键菜单

参见第 7 节「右键菜单功能」。

### 桌面歌词拖拽

- 鼠标拖拽移动桌面歌词窗口
- 窗口位置按屏幕百分比因子保存到 BoltDB

---

## 6. TUI 内部通知功能

### 架构设计

双通道通知系统：桌面系统通知 + TUI 内原生 toast。

```
notify.Notify(NotifyContent{...})
    ├── emitToast() → UI toast hook → App.Notify(NotificationSpec)
    └── 桌面通知（macOS / Linux D-Bus / Windows beeep）
```

### TUI 内 Toast

- **语义级别**：Info（蓝色边框）/ Success（绿色）/ Warning（黄色）/ Error（红色）
- **操作按钮**：支持显示操作按钮（如「查看详情」，可配置跳转 URL）
- **超时自动消失**：默认 4 秒，可通过 `inAppTimeout` 配置
- **可控开关**：通过 `[main.notification]` → `inApp` 配置

### 桌面系统通知

- **macOS**：优先使用内置 `musicfox-notifier.app`，fallback 到 `osascript`
- **Linux**：通过 D-Bus 协议调用 org.freedesktop.Notifications
- **Windows**：使用 beeep 库
- 支持专辑封面图片、点击跳转 URL

### 配置

```toml
[main.notification]
enable = true      # 桌面通知开关
inApp = true       # TUI 内 toast 通知开关
inAppTimeout = 4   # TUI toast 消失时长（秒）
albumCover = true  # 通知中显示专辑封面
```

### 通知场景

操作反馈通知覆盖：点赞/取消点赞、下载歌曲、收藏/取消收藏歌单/专辑/歌手、添加到播放列表、自动签到、更新检测等。

---

## 7. 右键菜单功能

基于 foxful-cli ContextMenu 组件的完整右键上下文菜单系统。

### 菜单结构

右键菜单按当前页面类型动态构建，分为三个区域：

```
┌─ 当前选中项操作 ─────────────────┐
│  ▲ 所属专辑                       │
│  ▲ 所属歌手                       │
│  ▲ 收藏/取消收藏 歌曲/专辑/歌手   │
│  ▲ 下载 / 添加到播放列表          │
│  ————————————————————————————————  │
│  ▶ 当前播放                       │
│  ▶ 相似歌曲 / 相似歌单            │
│  ▶ 在网页中打开                   │
│  ————————————————————————————————  │
│  ♪ 播放控制                       │
│    ▶ 播放/暂停                    │
│    ▶ 上一首 / 下一首              │
│  ————————————————————————————————  │
│  ↻ 刷新当前列表                   │
│  ⚙ 切换主题                       │
└──────────────────────────────────┘
```

### 智能分类

菜单根据页面实现的接口自动识别类型，选择对应图标：
- **歌曲页面**（SongsMenu）：🎵 当前选中 + 歌曲相关操作
- **歌单页面**（PlaylistsMenu）：📋 收藏/取消收藏
- **专辑页面**（AlbumsMenu）：💿 专辑操作
- **歌手页面**（ArtistsMenu）：🎤 歌手操作

### 全局菜单项

- **刷新当前列表**：重新加载当前页面数据
- **切换主题**：快捷键快捷主题切换入口（也可通过 `switchTheme` 键位触发）

### 系统集成

- 右键菜单由 foxful-cli 的 MouseController 处理，在鼠标点击时定位到合适位置
- 支持 Nerd Font 图标美化
- 分隔线 / 分组标题区分不同操作区域

---

## 8. 主题配置系统

完全可配置的 TOML 主题系统，支持深色/浅色双变体，内置 \*\*6 种主题\*\*（Default / Gruvbox / GitHub / Dracula / Nord / Transparent），允许用户自定义覆盖。

### 主题文件结构

```toml
name = "Default"
description = "网易云音乐经典红配色"

[dark]
# ── 基础调色板 ──────────────────────────────────────
primary = "#EA403F"       # 主色调
secondary = "#808080"     # 辅色调
accent = "#0000FF"        # 强调色
success = "#00FF00"       # 成功色
warning = "#FFFF00"       # 警告色
error = "#FF0000"         # 错误色
info = "#00FFFF"          # 信息色
muted = "#808080"         # 弱化色
hintKey = "#6E6E6E"       # 快捷键提示色
background = "#1A1A1A"    # 终端背景色
foreground = "#FFFFFF"    # 终端前景色
border = "#333333"        # 边框色
surface = "#242424"       # 表面色（卡片/弹窗背景）

# ── 通用高亮 ──────────────────────────────────────
[dark.highlights]
menuTitle = "BrightGreen"       # 菜单标题
# menuItem = ""                  # 菜单项文字
# selectedItem = ""              # 选中项前景
# menuItemHover = ""             # 鼠标悬停
# selectedItemHover = ""         # 选中+悬停
# title = ""                     # 页面标题
# prompt = ""                    # 输入提示
# backButton = ""                # 返回按钮
# backButtonHover = ""           # 返回按钮悬停
# subtitle = ""                  # 副标题
# progressEmpty = ""             # 进度条空槽
# scrollTrack = ""               # 滚动条轨道
# scrollThumb = ""               # 滚动条滑块
# button = ""                    # 按钮常态
# buttonBlurred = ""             # 按钮失焦

# ── 弹窗配色 ──────────────────────────────────────
[dark.popup]
surface = "#383838"       # 弹窗背景
border = "#5C5C5C"        # 弹窗边框

# ── 通知组件配色 ──────────────────────────────────
[dark.notification]
surface = "#2D2D2D"       # 通知背景
# infoBorder = ""          # 信息边框
# successBorder = ""       # 成功边框
# warningBorder = ""       # 警告边框
# errorBorder = ""         # 错误边框

# ── 状态栏配色 ────────────────────────────────────
[dark.statusBar]
# bar = ""                 # 状态栏基底背景
# text = ""                # 文字颜色
# breadcrumb = ""          # 面包屑路径
# breadcrumbSep = ""       # 分隔符 ">"
# breadcrumbHover = ""     # 面包屑悬停
# breadcrumbClick = ""     # 面包屑点击
# time = ""                # 时间显示
# nugget = ""              # 标签块文字
# nuggetLabelFg = ""       # 面包屑角标前景
# nuggetLabelBg = ""       # 面包屑角标背景

# ── Markdown 渲染主题 ─────────────────────────────
[dark.markdown]
dark = "dark"             # 暗色 glamour 主题名/文件路径
light = "light"           # 亮色 glamour 主题名/文件路径

# ── 应用自定义颜色 ─────────────────────────────────
[dark.app]
# 歌词颜色
lyricActive = "#7EC8E3"          # 歌词活跃行
lyricTransition = "#C9B1D4"      # 歌词过渡行
lyricInactive = "#6B6B6B"        # 歌词未播放行
lyricHighlight = "#E8E8E8"       # 歌词高亮词

# 播放栏颜色
playbarMode = "BrightMagenta"    # 播放模式
playbarVolume = "BrightBlue"     # 音量
playbarPlaying = "BrightYellow"  # 播放中
playbarPaused = "Yellow"         # 暂停
playbarHeartLiked = "Red"        # 喜欢心形
playbarHeartUnliked = "White"    # 未喜欢心形
playbarArtist = "BrightBlack"    # 歌手名字

# 频谱颜色
visualizerColorStart = "random"  # 频谱起始色（random=每次随机）
visualizerColorEnd = "random"    # 频谱结束色

# 应用背景
# background = ""                 # 整个应用背景色（默认透明）
# contextMenuBg = ""              # 右键菜单背景
# contextMenuBorder = ""          # 右键菜单边框
# spectrumVerticalBg = ""         # 频谱纵向渐变色
# desktopLyricsBg = ""            # 桌面歌词背景
# loginError = ""                 # 登录错误状态色
# loginSuccess = ""               # 登录成功状态色
# authStatus = ""                 # 认证状态色
# configTableBorder = ""          # CLI 配置表边框
# configTableSelectedFg = ""      # CLI 配置表选中前景
# configTableSelectedBg = ""      # CLI 配置表选中背景

# [light] 结构相同，颜色方案适配浅色背景
```

### 内置主题

| 主题 | 文件 | Dark Primary | 风格 |
|------|------|-------------|------|
| Default | `default.toml` | `#EA403F` | 网易云经典红色 |
| Gruvbox | `gruvbox.toml` | `#d79921` | 复古温暖色系 |
| GitHub | `github.toml` | `#58a6ff` | GitHub 官方配色 |
| Dracula | `dracula.toml` | `#bd93f9` | 经典 Dracula 紫色 |
| Nord | `nord.toml` | `#88c0d0` | 北极冰蓝风格 |
| Transparent | `transparent.toml` | `#EA403F` | 透明背景（无背景色） |

### 颜色格式

- **十六进制颜色**：`#EA403F`（推荐，精确控制）
- **ANSI 颜色名**：`BrightMagenta`、`Red`、`BrightBlue` 等 16 种终端原生颜色
- **random**：在 `[app]` 视觉器颜色中使用，每次随机生成亮色

### 主题加载

```
内置主题（embed/themes/*.toml）
    →   合并    →  ThemeRegistry
用户主题（~/.config/go-musicfox/themes/*.toml）
```

同名主题**用户版本覆盖内置版本**。

### 运行时主题切换

- **快捷键**：`switchTheme` 键位绑定
- **右键菜单**：「切换主题」菜单项
- **自动切换**：跟随 macOS 系统外观变化（Light/Dark ColorSchemeEvent）自动切换亮/暗变体
- 切换时在当前亮度类别（深色/浅色）的主题列表中循环

### 配置

```toml
[theme]
activeTheme = "Default"  # 启动时默认主题名称
```

---

## 9. 更新日志弹窗

### 功能

- **首次启动新版本**：自动弹出 Markdown 格式更新日志
- **Debug 模式**：每次启动都弹出（便于开发者验证）
- **版本记录**：通过 BoltDB 存储 `ChangelogSeen`，记录用户已查看的版本号
- **版本比较**：使用语义化版本比较，新版本 > 已见版本时弹窗
- 弹窗尺寸按终端自适应（宽度 70%，最大 120 列；高度 80%）

### 技术实现

- 更新日志内容存储在 `embed/changelog.md`（嵌入二进制）
- 通过 foxful-cli 的 `MarkdownPopup` 渲染
- 使用 `time.AfterFunc` 延迟 1.5 秒弹出确保启动页完成

---

## 配置结构速查

```toml
[startup]                    # 启动页配置
[main]                       # 主界面配置
  [main.visualizer]          # TUI 频谱配置
  [main.lyric]
    [main.lyric.desktopLyrics] # 桌面歌词配置
  [main.notification]        # 通知配置
[theme]                      # 主题配置
[player]                     # 播放器配置
  mouseVolumeStep = 1        # 鼠标滚轮音量步长
```
