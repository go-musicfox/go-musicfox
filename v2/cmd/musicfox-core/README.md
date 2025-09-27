# MusicFox Core - 核心功能主程序

MusicFox Core 是 go-musicfox v2 的核心功能版本，基于微内核插件架构实现，提供基础的音频播放和播放列表管理功能。

## 功能特性

### 核心功能
- **音频播放控制**：播放、暂停、恢复、停止、音量控制
- **播放列表管理**：创建、删除、添加歌曲、移除歌曲
- **播放队列管理**：队列操作、打乱、清空
- **播放模式**：顺序播放、随机播放、单曲循环
- **事件监听**：实时状态更新和事件通知
- **交互模式**：命令行交互界面

### 技术特性
- **微内核架构**：基于插件的可扩展架构
- **事件驱动**：完整的事件总线系统
- **多播放器后端**：支持多种音频播放后端
- **实时监控**：播放状态和系统状态监控
- **错误处理**：完善的错误处理和恢复机制

## 安装和构建

### 前置要求
- Go 1.23.0 或更高版本
- 支持的操作系统：macOS、Linux、Windows

### 构建

```bash
# 进入项目目录
cd v2/cmd/musicfox-core

# 下载依赖
go mod tidy

# 构建
go build -o musicfox-core

# 或者使用 make（如果有 Makefile）
make build
```

### 安装

```bash
# 安装到 GOPATH/bin
go install

# 或者复制到系统路径
sudo cp musicfox-core /usr/local/bin/
```

## 使用方法

### 命令行参数

```bash
musicfox-core [flags] [command]
```

#### 全局参数
- `-c, --config <path>`：指定配置文件路径
- `-l, --log-level <level>`：设置日志级别 (debug, info, warn, error)
- `-v, --version`：显示版本信息
- `-h, --help`：显示帮助信息

### 基础命令

#### 播放控制
```bash
# 播放指定歌曲
musicfox-core play /path/to/song.mp3
musicfox-core play http://example.com/song.mp3

# 暂停播放
musicfox-core pause

# 恢复播放
musicfox-core resume

# 停止播放
musicfox-core stop

# 下一首
musicfox-core next

# 上一首
musicfox-core prev

# 音量控制
musicfox-core volume 50        # 设置音量为50%
musicfox-core volume           # 显示当前音量

# 显示播放状态
musicfox-core status
```

#### 播放列表管理
```bash
# 创建播放列表
musicfox-core playlist create "My Playlist"

# 列出所有播放列表
musicfox-core playlist list

# 显示播放列表详情
musicfox-core playlist show <playlist-id>

# 添加歌曲到播放列表
musicfox-core playlist add <playlist-id> /path/to/song.mp3

# 从播放列表移除歌曲
musicfox-core playlist remove <playlist-id> <song-id>

# 删除播放列表
musicfox-core playlist delete <playlist-id>
```

#### 播放队列管理
```bash
# 显示当前队列
musicfox-core playlist queue

# 添加歌曲到队列
musicfox-core playlist queue add /path/to/song.mp3

# 打乱队列
musicfox-core playlist queue shuffle

# 清空队列
musicfox-core playlist queue clear
```

### 交互模式

```bash
# 进入交互模式
musicfox-core interactive

# 或者不带任何参数启动
musicfox-core
```

在交互模式中，可以使用以下命令：

```
> play /path/to/song.mp3     # 播放歌曲
> pause                      # 暂停
> resume                     # 恢复
> stop                       # 停止
> volume 75                  # 设置音量
> status                     # 显示状态
> playlist create "My List"  # 创建播放列表
> playlist list              # 列出播放列表
> help                       # 显示帮助
> quit                       # 退出
```

## 配置文件

可以使用 YAML 格式的配置文件来自定义设置：

```yaml
# config.yaml
kernel:
  log_level: "info"
  
audio:
  default_backend: "beep"  # 默认音频后端
  volume: 80               # 默认音量
  buffer_size: 4096        # 缓冲区大小
  sample_rate: 44100       # 采样率
  channels: 2              # 声道数
  
playlist:
  max_history: 100         # 最大历史记录数
  default_mode: "sequential" # 默认播放模式
```

使用配置文件：
```bash
musicfox-core --config config.yaml interactive
```

## 支持的音频格式

根据所选的播放器后端，支持以下音频格式：

- **MP3**：所有后端
- **FLAC**：beep、mpv 后端
- **WAV**：所有后端
- **OGG**：beep、mpv 后端
- **M4A/AAC**：mpv、osx 后端

## 播放器后端

- **beep**：Go 原生音频库，跨平台支持
- **mpv**：基于 mpv 媒体播放器（需要安装 mpv）
- **osx**：macOS 原生 AVAudioPlayer（仅 macOS）
- **windows**：Windows Media Player API（仅 Windows）

## 事件系统

应用支持以下事件类型：

### 音频事件
- `audio.play.start`：播放开始
- `audio.play.pause`：播放暂停
- `audio.play.resume`：播放恢复
- `audio.play.stop`：播放停止
- `audio.volume.change`：音量变化
- `audio.position.update`：播放位置更新

### 播放列表事件
- `playlist.created`：播放列表创建
- `playlist.updated`：播放列表更新
- `queue.changed`：队列变化
- `playmode.changed`：播放模式变化

## 开发和测试

### 运行测试

```bash
# 运行所有测试
go test -v

# 运行测试并显示覆盖率
go test -v -cover

# 运行特定测试
go test -v -run TestCommandHandler_HandlePlay
```

### 代码检查

```bash
# 代码格式化
go fmt ./...

# 代码检查
go vet ./...

# 使用 golangci-lint（如果安装）
golangci-lint run
```

## 故障排除

### 常见问题

1. **音频播放失败**
   - 检查音频文件格式是否支持
   - 确认音频文件路径正确
   - 尝试切换不同的播放器后端

2. **插件加载失败**
   - 检查依赖是否正确安装
   - 查看日志输出获取详细错误信息
   - 确认配置文件格式正确

3. **权限问题**
   - 确保对音频文件有读取权限
   - 在某些系统上可能需要音频设备访问权限

### 调试模式

```bash
# 启用调试日志
musicfox-core --log-level debug interactive
```

## 许可证

本项目采用 MIT 许可证，详见 LICENSE 文件。

## 贡献

欢迎提交 Issue 和 Pull Request！

## 更多信息

- [项目主页](https://github.com/go-musicfox/go-musicfox)
- [API 文档](../../docs/api/)
- [架构文档](../../docs/architecture/)
- [插件开发指南](../../docs/guides/plugin-quickstart.md)