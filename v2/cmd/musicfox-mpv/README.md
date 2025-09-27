# MusicFox MPV - 专用MPV播放器版本

MusicFox MPV 是 go-musicfox v2 的专用MPV播放器版本，基于微内核插件架构实现，专门针对MPV播放器后端进行优化。

## 功能特性

### 核心功能
- **MPV专用后端**：强制使用MPV播放器，提供最佳的音频播放体验
- **音频播放控制**：播放、暂停、恢复、停止、音量控制
- **播放列表管理**：创建、删除、添加歌曲、移除歌曲
- **播放队列管理**：队列操作、打乱、清空
- **播放模式**：顺序播放、随机播放、单曲循环
- **网易云音乐**：集成网易云音乐插件，支持搜索、播放列表、用户功能
- **TUI界面**：基于bubbletea的终端用户界面，提供友好的交互体验
- **事件监听**：实时状态更新和事件通知
- **交互模式**：命令行交互界面
- **配置管理**：MPV专用配置选项

### MPV特性
- **高质量音频**：支持MPV的所有音频格式和编解码器
- **低延迟**：优化的音频管道，减少播放延迟
- **稳定性**：基于成熟的MPV媒体播放器
- **跨平台**：支持Linux、macOS、Windows

## 系统要求

### 基本要求
- **Go**: 1.21 或更高版本
- **MPV**: 必须安装MPV播放器
- **操作系统**: Linux, macOS, Windows
- **内存**: 最少 256MB RAM
- **存储**: 最少 50MB 可用空间

### MPV安装

#### macOS
```bash
# 使用Homebrew安装
brew install mpv

# 或使用MacPorts
sudo port install mpv
```

#### Ubuntu/Debian
```bash
sudo apt update
sudo apt install mpv
```

#### CentOS/RHEL/Fedora
```bash
# CentOS/RHEL
sudo yum install mpv

# Fedora
sudo dnf install mpv
```

#### Windows
1. 从 [MPV官网](https://mpv.io/installation/) 下载Windows版本
2. 解压到任意目录
3. 将MPV可执行文件路径添加到系统PATH环境变量

## 安装和运行

### 编译安装

```bash
# 克隆项目
git clone https://github.com/go-musicfox/go-musicfox.git
cd go-musicfox/v2

# 编译musicfox-mpv
go build -o musicfox-mpv ./cmd/musicfox-mpv

# 运行
./musicfox-mpv --help
```

### 直接运行

```bash
# 进入项目目录
cd go-musicfox/v2

# 直接运行
go run ./cmd/musicfox-mpv
```

## 使用方法

### 命令行参数

```bash
musicfox-mpv [flags] [command]
```

#### 可用标志
- `-c, --config string`: 配置文件路径
- `-l, --log-level string`: 日志级别 (debug, info, warn, error)
- `-v, --version`: 显示版本信息
- `-h, --help`: 显示帮助信息

#### 可用命令
- `play <song_url>`: 播放指定歌曲
- `pause`: 暂停播放
- `resume`: 恢复播放
- `stop`: 停止播放
- `next`: 下一首
- `prev`: 上一首
- `volume <level>`: 设置音量 (0-100)
- `status`: 显示播放状态
- `playlist`: 播放列表管理
- `interactive`: 进入交互模式

### 使用示例

#### 基本播放
```bash
# 播放本地文件
musicfox-mpv play /path/to/song.mp3

# 播放网络流
musicfox-mpv play https://example.com/stream.mp3

# 设置音量
musicfox-mpv volume 80

# 查看状态
musicfox-mpv status
```

#### 播放列表管理
```bash
# 创建播放列表
musicfox-mpv playlist create "我的歌单"

# 列出所有播放列表
musicfox-mpv playlist list

# 添加歌曲到播放列表
musicfox-mpv playlist add playlist-123 /path/to/song.mp3

# 显示播放列表详情
musicfox-mpv playlist show playlist-123
```

#### 交互模式
```bash
# 进入交互模式
musicfox-mpv interactive

# 或者不带参数直接运行
musicfox-mpv
```

在交互模式中，可以使用以下命令：
- `play <url>`: 播放歌曲
- `pause`: 暂停
- `resume`: 恢复
- `stop`: 停止
- `volume <level>`: 设置音量
- `status`: 显示状态
- `help`: 显示帮助
- `quit` 或 `exit`: 退出程序

## 配置文件

### 配置文件位置

配置文件会按以下优先级查找：
1. 命令行指定的路径 (`-c` 参数)
2. `$XDG_CONFIG_HOME/musicfox/mpv-config.yaml`
3. `$HOME/.config/musicfox/mpv-config.yaml`
4. `./config/mpv-config.yaml`

### 配置文件示例

```yaml
app:
  name: "musicfox-mpv"
  version: "2.0.0"
  debug: false

audio:
  backend: "mpv"  # 强制使用MPV
  mpv_path: "/usr/bin/mpv"  # MPV可执行文件路径
  mpv_args:
    - "--no-video"
    - "--quiet"
    - "--really-quiet"
    - "--no-terminal"
    - "--idle"
    - "--force-window=no"
  buffer_size: 4096
  volume: 80

playlist:
  auto_save: true
  default_format: "m3u8"
  save_dir: "./playlists"

netease:
  enabled: true
  cache_dir: "./cache/netease"

tui:
  enabled: true
  theme: "default"
  auto_start: false
  full_screen: false

logging:
  level: "info"
  file: "./logs/musicfox-mpv.log"
  max_size: 100
```

### 配置说明

#### 音频配置 (audio)
- `backend`: 音频后端，固定为"mpv"
- `mpv_path`: MPV可执行文件路径，默认为"mpv"（使用PATH中的mpv）
- `mpv_args`: MPV启动参数，可以根据需要调整
- `buffer_size`: 音频缓冲区大小
- `volume`: 默认音量 (0-100)

#### 播放列表配置 (playlist)
- `auto_save`: 是否自动保存播放列表
- `default_format`: 默认播放列表格式 (m3u, m3u8, pls, xspf)
- `save_dir`: 播放列表保存目录

#### 网易云配置 (netease)
- `enabled`: 是否启用网易云插件
- `cache_dir`: 缓存目录

#### TUI配置 (tui)
- `enabled`: 是否启用TUI插件
- `theme`: TUI主题，默认为"default"
- `auto_start`: 是否自动启动TUI界面
- `full_screen`: 是否使用全屏模式

#### 日志配置 (logging)
- `level`: 日志级别 (debug, info, warn, error)
- `file`: 日志文件路径
- `max_size`: 日志文件最大大小(MB)

## MPV参数优化

### 推荐的MPV参数

```yaml
mpv_args:
  # 基本设置
  - "--no-video"          # 禁用视频输出
  - "--no-terminal"       # 禁用终端控制
  - "--idle"              # 空闲模式
  - "--force-window=no"   # 不强制显示窗口
  
  # 音频优化
  - "--audio-buffer=1"    # 音频缓冲区大小
  - "--audio-latency-hacks=yes"  # 启用音频延迟优化
  
  # 静默设置
  - "--quiet"             # 静默模式
  - "--really-quiet"      # 真正的静默模式
  
  # 网络优化
  - "--cache=yes"         # 启用缓存
  - "--demuxer-max-bytes=50M"  # 最大缓存大小
```

### 高质量音频设置

```yaml
mpv_args:
  - "--no-video"
  - "--audio-format=s32"     # 32位音频格式
  - "--audio-samplerate=96000"  # 高采样率
  - "--audio-channels=stereo"   # 立体声
  - "--volume-max=100"       # 最大音量限制
```

## 故障排除

### 常见问题

#### 1. MPV not found 错误
```
Error: MPV not available: mpv not found in PATH
```

**解决方案**：
- 确保已安装MPV播放器
- 检查MPV是否在系统PATH中
- 在配置文件中指定MPV的完整路径

#### 2. 播放失败
```
Error: failed to play song: ...
```

**解决方案**：
- 检查音频文件格式是否受支持
- 确认文件路径或URL是否正确
- 查看MPV日志获取详细错误信息

#### 3. 配置文件错误
```
Error: failed to parse config file: ...
```

**解决方案**：
- 检查YAML语法是否正确
- 确认配置项名称和值是否有效
- 使用默认配置重新开始

### 调试模式

启用调试模式获取更多信息：

```bash
# 启用调试日志
musicfox-mpv --log-level debug

# 或在配置文件中设置
app:
  debug: true
logging:
  level: "debug"
```

### 日志查看

```bash
# 查看日志文件
tail -f ./logs/musicfox-mpv.log

# 实时查看MPV输出（如果启用）
musicfox-mpv --log-level debug 2>&1 | grep mpv
```

## 架构说明

MusicFox MPV基于go-musicfox v2的微内核插件架构：

- **微内核** (`pkg/kernel`): 提供核心服务和插件管理
- **音频插件** (`plugins/audio`): 处理音频播放，强制使用MPV后端
- **播放列表插件** (`plugins/playlist`): 管理播放列表
- **事件总线** (`pkg/event`): 组件间通信
- **配置管理** (`pkg/config`): 配置加载和验证

这种架构确保了：
- **代码复用**: 最大化复用现有组件
- **模块化**: 清晰的职责分离
- **可扩展性**: 易于添加新功能
- **稳定性**: 基于成熟的插件系统

## 贡献

欢迎贡献代码！请遵循以下步骤：

1. Fork 项目
2. 创建功能分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add some amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 打开 Pull Request

## 许可证

本项目采用 MIT 许可证。详情请参阅 [LICENSE](../../LICENSE) 文件。

## 相关链接

- [go-musicfox 主项目](https://github.com/go-musicfox/go-musicfox)
- [MPV 官网](https://mpv.io/)
- [go-musicfox v2 文档](../../docs/README.md)
- [插件开发指南](../../docs/guides/plugin-development.md)