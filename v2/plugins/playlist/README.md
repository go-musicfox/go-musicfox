# Playlist Plugin

播放列表插件是go-musicfox v2的核心插件之一，提供完整的播放列表管理、播放队列管理、播放历史记录和播放模式控制功能。

## 功能特性

### 播放列表管理
- 创建、删除、更新播放列表
- 获取播放列表信息和列表
- 向播放列表添加、移除、移动歌曲
- 清空播放列表

### 播放队列管理
- 设置和获取当前播放队列
- 添加歌曲到队列
- 从队列移除歌曲
- 清空和打乱播放队列

### 播放历史管理
- 自动记录播放历史
- 获取播放历史（支持限制数量）
- 获取最近播放的歌曲（去重）
- 获取播放次数最多的歌曲
- 清空播放历史
- 播放历史统计信息

### 播放模式管理
- 支持四种播放模式：
  - 顺序播放（Sequential）
  - 单曲循环（RepeatOne）
  - 列表循环（RepeatAll）
  - 随机播放（Shuffle）
- 智能的下一首/上一首歌曲选择
- 播放模式切换

### 事件系统集成
- 发送播放列表变化事件
- 发送队列变化事件
- 发送播放模式变化事件
- 监听播放器事件并自动更新历史记录

## 接口定义

### PlaylistPlugin 接口

```go
type PlaylistPlugin interface {
    plugin.Plugin

    // 播放列表管理
    CreatePlaylist(ctx context.Context, name, description string) (*model.Playlist, error)
    DeletePlaylist(ctx context.Context, playlistID string) error
    UpdatePlaylist(ctx context.Context, playlistID string, updates map[string]interface{}) error
    GetPlaylist(ctx context.Context, playlistID string) (*model.Playlist, error)
    ListPlaylists(ctx context.Context) ([]*model.Playlist, error)

    // 播放列表歌曲管理
    AddSong(ctx context.Context, playlistID string, song *model.Song) error
    RemoveSong(ctx context.Context, playlistID string, songID string) error
    MoveSong(ctx context.Context, playlistID string, songID string, newIndex int) error
    ClearPlaylist(ctx context.Context, playlistID string) error

    // 播放队列管理
    SetCurrentQueue(ctx context.Context, songs []*model.Song) error
    GetCurrentQueue(ctx context.Context) ([]*model.Song, error)
    AddToQueue(ctx context.Context, song *model.Song) error
    RemoveFromQueue(ctx context.Context, songID string) error
    ClearQueue(ctx context.Context) error
    ShuffleQueue(ctx context.Context) error

    // 播放历史管理
    AddToHistory(ctx context.Context, song *model.Song) error
    GetHistory(ctx context.Context, limit int) ([]*model.Song, error)
    ClearHistory(ctx context.Context) error

    // 播放模式管理
    SetPlayMode(ctx context.Context, mode model.PlayMode) error
    GetPlayMode(ctx context.Context) model.PlayMode
    GetNextSong(ctx context.Context, currentSong *model.Song) (*model.Song, error)
    GetPreviousSong(ctx context.Context, currentSong *model.Song) (*model.Song, error)
}
```

## 使用示例

### 创建插件实例

```go
playlistPlugin := NewPlaylistPlugin()
```

### 初始化插件

```go
ctx := &plugin.PluginContext{
    Services: map[string]interface{}{
        "event_bus": eventBus,
    },
}

err := playlistPlugin.Initialize(ctx)
if err != nil {
    log.Fatal("Failed to initialize playlist plugin:", err)
}
```

### 创建播放列表

```go
playlist, err := playlistPlugin.CreatePlaylist(context.Background(), "我的播放列表", "我最喜欢的歌曲")
if err != nil {
    log.Error("Failed to create playlist:", err)
    return
}
```

### 添加歌曲到播放列表

```go
song := &model.Song{
    ID:     "song123",
    Title:  "示例歌曲",
    Artist: "示例艺术家",
    Source: "local",
}

err := playlistPlugin.AddSong(context.Background(), playlist.ID, song)
if err != nil {
    log.Error("Failed to add song to playlist:", err)
    return
}
```

### 设置播放队列

```go
songs := []*model.Song{song1, song2, song3}
err := playlistPlugin.SetCurrentQueue(context.Background(), songs)
if err != nil {
    log.Error("Failed to set queue:", err)
    return
}
```

### 设置播放模式

```go
err := playlistPlugin.SetPlayMode(context.Background(), model.PlayModeShuffle)
if err != nil {
    log.Error("Failed to set play mode:", err)
    return
}
```

### 获取下一首歌曲

```go
nextSong, err := playlistPlugin.GetNextSong(context.Background(), currentSong)
if err != nil {
    log.Error("Failed to get next song:", err)
    return
}
```

## 事件

插件会发送以下事件：

### 播放列表事件
- `playlist.created` - 播放列表创建
- `playlist.updated` - 播放列表更新
- `playlist.deleted` - 播放列表删除
- `playlist.cleared` - 播放列表清空
- `playlist.song_added` - 歌曲添加到播放列表
- `playlist.song_removed` - 歌曲从播放列表移除
- `playlist.song_moved` - 播放列表中歌曲位置移动

### 队列事件
- `queue.updated` - 队列更新
- `queue.song_added` - 歌曲添加到队列
- `queue.song_removed` - 歌曲从队列移除
- `queue.cleared` - 队列清空
- `queue.shuffled` - 队列打乱

### 历史事件
- `history.song_added` - 歌曲添加到历史
- `history.cleared` - 历史清空

### 播放模式事件
- `player.mode_changed` - 播放模式变化

## 配置选项

- `max_history`: 最大历史记录数量（默认：100）

## 依赖

- `event_bus`: 事件总线服务

## 文件结构

```
playlist/
├── plugin.go              # 主插件文件和接口实现
├── playlist_manager.go     # 播放列表管理功能
├── queue_manager.go        # 播放队列管理功能
├── history_manager.go      # 播放历史管理功能
├── playmode_manager.go     # 播放模式管理功能
├── go.mod                  # Go模块定义
└── README.md              # 文档
```

## 线程安全

所有公共方法都是线程安全的，使用读写锁保护内部数据结构。

## 性能特性

- 内存中存储，快速访问
- 智能的随机播放算法
- 高效的歌曲查找和操作
- 事件驱动的异步通信

## 扩展性

插件设计为可扩展的，可以轻松添加新的播放模式、历史分析功能或与外部服务的集成。