# go-musicfox v5.0.0

## 🎨 主题系统全面配置化

- **TOML 主题文件**：支持 dark/light 双变体，内置网易云红主题（`neteasered`）
- **用户自定义主题**：放置在 `~/.config/go-musicfox/themes/` 下，同名覆盖内置主题
- **ANSI + 十六进制双模颜色**：支持终端原生 ANSI 颜色名（如 `BrightMagenta`、`Red`）和十六进制颜色
- **完整配色覆盖**：歌词色（活跃/过渡/未播放/高亮）、播放栏色（模式/音量/状态/喜欢/艺术家）、菜单标题
- **运行时主题切换**：通过快捷键绑定 `switchTheme` 或右键菜单「切换主题」即时切换
- **亮暗色自动切换**：跟随 macOS 系统外观变化（LightColorSchemeEvent/DarkColorSchemeEvent）自动切换 dark/light 变体

## 🔧 UI 优化

- **歌词行最大宽度限制**：超长 YRC（逐字高亮）歌词自动截断，超宽时剥离 ANSI 码回退纯色渲染
- **歌曲信息行智能截断**：歌曲名优先显示，歌手名用剩余空间截断
- **截断宽度精确计算**：基于实际 segment 渲染宽度而非估算值
- **状态栏自适应**：面包屑宽度约束为左半屏，终端过窄时渐进隐藏时间 → 中间模块

## 🚀 Foxful-cli 升级 (v1.0.1)

- 新增 Theme 配置化支持（Primary/Secondary 等语义色、Highlight 组件样式、Popup 主题）
- 新增 `ThemeList` + `ThemeSwitchKey` 运行时主题切换机制
- 新增 Markdown 弹窗支持
- 改进 `onBackgroundChanged` 终端背景检测
- 状态栏面包屑宽度约束优化

## 📋 新功能

- **更新日志弹窗**：首次启动新版本时自动弹窗显示 Markdown 格式更新日志，版本记录存储在 BoltDB
- **右键菜单切换主题**：「切换主题」选项添加到右键上下文菜单

## ⚙️ 配置变更

- `config.toml` 新增 `activeTheme` 字段指定启动主题
- 新增 `switchTheme` 键绑定操作
