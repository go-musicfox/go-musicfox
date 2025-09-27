# RPC 插件开发指南

RPC（Remote Procedure Call）插件通过进程间通信的方式与 go-musicfox v2 微内核进行交互。这种插件类型提供了最高的安全性和隔离性，支持多种编程语言开发。

## 目录

- [概述](#概述)
- [技术原理](#技术原理)
- [开发环境准备](#开发环境准备)
- [项目结构](#项目结构)
- [协议定义](#协议定义)
- [服务端实现](#服务端实现)
- [客户端集成](#客户端集成)
- [通信优化](#通信优化)
- [错误处理](#错误处理)
- [部署配置](#部署配置)
- [监控和调试](#监控和调试)
- [最佳实践](#最佳实践)
- [多语言支持](#多语言支持)

## 概述

### 什么是 RPC 插件

RPC 插件是运行在独立进程中的程序，通过网络协议（如 gRPC、HTTP、TCP）与主程序通信。插件进程可以用任何支持相应协议的编程语言开发。

### 优势

- **进程隔离**：插件崩溃不会影响主程序
- **语言无关**：支持多种编程语言开发
- **独立部署**：可以独立更新和部署
- **水平扩展**：支持分布式部署
- **资源控制**：可以精确控制资源使用

### 劣势

- **通信开销**：网络通信带来延迟
- **复杂性**：需要处理网络错误和重连
- **序列化成本**：数据需要序列化/反序列化
- **调试困难**：跨进程调试相对复杂

### 适用场景

- 网络服务集成
- 第三方 API 封装
- 计算密集型任务
- 需要特定语言库的功能
- 安全要求较高的场景

## 技术原理

### gRPC 通信

go-musicfox v2 使用 gRPC 作为主要的 RPC 通信协议：

```protobuf
// plugin.proto
syntax = "proto3";

package plugin;

option go_package = "github.com/go-musicfox/kernel/pkg/plugin/rpc";

// 插件服务定义
service PluginService {
    // 生命周期管理
    rpc Initialize(InitializeRequest) returns (InitializeResponse);
    rpc Start(StartRequest) returns (StartResponse);
    rpc Stop(StopRequest) returns (StopResponse);
    rpc HealthCheck(HealthCheckRequest) returns (HealthCheckResponse);
    
    // 音频处理
    rpc ProcessAudio(ProcessAudioRequest) returns (ProcessAudioResponse);
    rpc ProcessAudioStream(stream ProcessAudioRequest) returns (stream ProcessAudioResponse);
    
    // 配置管理
    rpc GetConfig(GetConfigRequest) returns (GetConfigResponse);
    rpc UpdateConfig(UpdateConfigRequest) returns (UpdateConfigResponse);
    
    // 元数据
    rpc GetMetadata(GetMetadataRequest) returns (GetMetadataResponse);
}

// 消息定义
message InitializeRequest {
    string config_json = 1;
    map<string, string> environment = 2;
}

message InitializeResponse {
    bool success = 1;
    string error_message = 2;
}

message ProcessAudioRequest {
    AudioBuffer input = 1;
    map<string, string> parameters = 2;
}

message ProcessAudioResponse {
    AudioBuffer output = 1;
    string error_message = 2;
    ProcessingStats stats = 3;
}

message AudioBuffer {
    repeated AudioChannel channels = 1;
    int32 sample_rate = 2;
    int32 frames = 3;
    int64 timestamp = 4;
}

message AudioChannel {
    repeated float samples = 1;
}

message ProcessingStats {
    int64 processing_time_ns = 1;
    int64 memory_used = 2;
    int64 cpu_time_ns = 3;
}
```

### 进程管理

主程序负责启动和管理插件进程：

```go
// 插件进程管理器
type RPCPluginManager struct {
    plugins map[string]*RPCPlugin
    mutex   sync.RWMutex
}

type RPCPlugin struct {
    ID       string
    Process  *os.Process
    Client   PluginServiceClient
    Config   *RPCPluginConfig
    Status   PluginStatus
}
```

## 开发环境准备

### 工具安装

```bash
# 安装 Protocol Buffers 编译器
# macOS
brew install protobuf

# Ubuntu/Debian
sudo apt-get install protobuf-compiler

# 安装 Go 插件
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# 安装其他语言插件（可选）
# Python
pip install grpcio-tools

# Node.js
npm install -g grpc-tools
```

### 项目初始化

```bash
# 创建项目目录
mkdir music-source-rpc-plugin
cd music-source-rpc-plugin

# 初始化 Go 模块
go mod init github.com/your-username/music-source-rpc-plugin

# 创建目录结构
mkdir -p {proto,cmd/server,internal/{service,client,config},pkg/api,configs,scripts}
```

## 项目结构

```
music-source-rpc-plugin/
├── proto/
│   ├── plugin.proto             # 插件服务定义
│   ├── music.proto              # 音乐相关消息定义
│   └── common.proto             # 通用消息定义
├── cmd/
│   ├── server/
│   │   └── main.go              # RPC 服务器入口
│   └── client/
│       └── main.go              # 测试客户端
├── internal/
│   ├── service/
│   │   ├── plugin.go            # 插件服务实现
│   │   ├── music.go             # 音乐服务实现
│   │   └── health.go            # 健康检查服务
│   ├── client/
│   │   ├── netease.go           # 网易云音乐客户端
│   │   └── spotify.go           # Spotify 客户端
│   ├── config/
│   │   ├── config.go            # 配置管理
│   │   └── validation.go        # 配置验证
│   └── utils/
│       ├── logger.go            # 日志工具
│       └── metrics.go           # 指标收集
├── pkg/
│   └── api/
│       ├── generated/           # 生成的 gRPC 代码
│       ├── types.go             # 数据类型定义
│       └── client.go            # 客户端封装
├── configs/
│   ├── server.yaml              # 服务器配置
│   └── plugin.json              # 插件元数据
├── scripts/
│   ├── generate.sh              # 代码生成脚本
│   ├── build.sh                 # 构建脚本
│   └── deploy.sh                # 部署脚本
├── docker/
│   ├── Dockerfile               # Docker 镜像
│   └── docker-compose.yml       # 容器编排
├── tests/
│   ├── integration/             # 集成测试
│   └── load/                    # 负载测试
├── .gitignore
├── Makefile
├── go.mod
└── go.sum
```

## 协议定义

### 核心协议文件

```protobuf
// proto/plugin.proto
syntax = "proto3";

package plugin;

option go_package = "github.com/your-username/music-source-rpc-plugin/pkg/api/generated";

import "google/protobuf/timestamp.proto";
import "google/protobuf/duration.proto";
import "google/protobuf/struct.proto";

// 插件核心服务
service PluginService {
    // 生命周期管理
    rpc Initialize(InitializeRequest) returns (InitializeResponse);
    rpc Start(StartRequest) returns (StartResponse);
    rpc Stop(StopRequest) returns (StopResponse);
    rpc Shutdown(ShutdownRequest) returns (ShutdownResponse);
    
    // 状态查询
    rpc GetStatus(GetStatusRequest) returns (GetStatusResponse);
    rpc HealthCheck(HealthCheckRequest) returns (HealthCheckResponse);
    rpc GetMetadata(GetMetadataRequest) returns (GetMetadataResponse);
    
    // 配置管理
    rpc GetConfig(GetConfigRequest) returns (GetConfigResponse);
    rpc UpdateConfig(UpdateConfigRequest) returns (UpdateConfigResponse);
    rpc ValidateConfig(ValidateConfigRequest) returns (ValidateConfigResponse);
}

// 音乐源服务
service MusicSourceService {
    // 搜索功能
    rpc Search(SearchRequest) returns (SearchResponse);
    rpc SearchSongs(SearchSongsRequest) returns (SearchSongsResponse);
    rpc SearchAlbums(SearchAlbumsRequest) returns (SearchAlbumsResponse);
    rpc SearchArtists(SearchArtistsRequest) returns (SearchArtistsResponse);
    rpc SearchPlaylists(SearchPlaylistsRequest) returns (SearchPlaylistsResponse);
    
    // 内容获取
    rpc GetSong(GetSongRequest) returns (GetSongResponse);
    rpc GetAlbum(GetAlbumRequest) returns (GetAlbumResponse);
    rpc GetArtist(GetArtistRequest) returns (GetArtistResponse);
    rpc GetPlaylist(GetPlaylistRequest) returns (GetPlaylistResponse);
    
    // 流媒体
    rpc GetStreamURL(GetStreamURLRequest) returns (GetStreamURLResponse);
    rpc GetLyrics(GetLyricsRequest) returns (GetLyricsResponse);
    
    // 用户功能
    rpc Login(LoginRequest) returns (LoginResponse);
    rpc Logout(LogoutRequest) returns (LogoutResponse);
    rpc GetUserInfo(GetUserInfoRequest) returns (GetUserInfoResponse);
    rpc GetUserPlaylists(GetUserPlaylistsRequest) returns (GetUserPlaylistsResponse);
}

// 基础消息类型
message InitializeRequest {
    string config_json = 1;
    map<string, string> environment = 2;
    string plugin_dir = 3;
    string log_level = 4;
}

message InitializeResponse {
    bool success = 1;
    string error_message = 2;
    PluginMetadata metadata = 3;
}

message PluginMetadata {
    string id = 1;
    string name = 2;
    string version = 3;
    string description = 4;
    string author = 5;
    string license = 6;
    repeated string tags = 7;
    repeated string capabilities = 8;
    google.protobuf.Struct metadata = 9;
}

message SearchRequest {
    string query = 1;
    SearchOptions options = 2;
}

message SearchOptions {
    int32 limit = 1;
    int32 offset = 2;
    string type = 3;  // "song", "album", "artist", "playlist"
    string sort = 4;  // "relevance", "popularity", "date"
    map<string, string> filters = 5;
}

message SearchResponse {
    repeated Song songs = 1;
    repeated Album albums = 2;
    repeated Artist artists = 3;
    repeated Playlist playlists = 4;
    int32 total_count = 5;
    bool has_more = 6;
    string error_message = 7;
}

message Song {
    string id = 1;
    string title = 2;
    string artist = 3;
    string album = 4;
    google.protobuf.Duration duration = 5;
    string genre = 6;
    int32 year = 7;
    int32 track_number = 8;
    string cover_url = 9;
    google.protobuf.Struct metadata = 10;
}

message Album {
    string id = 1;
    string title = 2;
    string artist = 3;
    int32 year = 4;
    string genre = 5;
    string cover_url = 6;
    repeated Song songs = 7;
    google.protobuf.Struct metadata = 8;
}

message Artist {
    string id = 1;
    string name = 2;
    string bio = 3;
    string avatar_url = 4;
    repeated Album albums = 5;
    repeated Song popular_songs = 6;
    google.protobuf.Struct metadata = 7;
}

message Playlist {
    string id = 1;
    string name = 2;
    string description = 3;
    string creator = 4;
    string cover_url = 5;
    repeated Song songs = 6;
    google.protobuf.Timestamp created_at = 7;
    google.protobuf.Timestamp updated_at = 8;
    google.protobuf.Struct metadata = 9;
}

message GetStreamURLRequest {
    string song_id = 1;
    Quality quality = 2;
    map<string, string> options = 3;
}

message GetStreamURLResponse {
    string url = 1;
    Quality actual_quality = 2;
    google.protobuf.Duration expires_in = 3;
    string error_message = 4;
}

enum Quality {
    QUALITY_UNKNOWN = 0;
    QUALITY_LOW = 1;     // 128kbps
    QUALITY_MEDIUM = 2;  // 192kbps
    QUALITY_HIGH = 3;    // 320kbps
    QUALITY_LOSSLESS = 4; // FLAC
}

message LoginRequest {
    oneof credentials {
        UsernamePassword username_password = 1;
        string token = 2;
        string oauth_code = 3;
    }
}

message UsernamePassword {
    string username = 1;
    string password = 2;
}

message LoginResponse {
    bool success = 1;
    string error_message = 2;
    UserInfo user_info = 3;
    string session_token = 4;
    google.protobuf.Timestamp expires_at = 5;
}

message UserInfo {
    string id = 1;
    string username = 2;
    string display_name = 3;
    string email = 4;
    string avatar_url = 5;
    bool is_premium = 6;
    google.protobuf.Struct metadata = 7;
}
```

### 代码生成

```bash
#!/bin/bash
# scripts/generate.sh

set -e

PROTO_DIR="proto"
OUT_DIR="pkg/api/generated"

# 创建输出目录
mkdir -p "$OUT_DIR"

# 生成 Go 代码
protoc \
    --proto_path="$PROTO_DIR" \
    --go_out="$OUT_DIR" \
    --go_opt=paths=source_relative \
    --go-grpc_out="$OUT_DIR" \
    --go-grpc_opt=paths=source_relative \
    "$PROTO_DIR"/*.proto

echo "Protocol buffer code generated successfully"

# 生成其他语言代码（可选）
if command -v python3 &> /dev/null; then
    echo "Generating Python code..."
    python3 -m grpc_tools.protoc \
        --proto_path="$PROTO_DIR" \
        --python_out="pkg/api/python" \
        --grpc_python_out="pkg/api/python" \
        "$PROTO_DIR"/*.proto
fi

if command -v node &> /dev/null; then
    echo "Generating Node.js code..."
    mkdir -p "pkg/api/nodejs"
    grpc_tools_node_protoc \
        --proto_path="$PROTO_DIR" \
        --js_out=import_style=commonjs,binary:"pkg/api/nodejs" \
        --grpc_out=grpc_js:"pkg/api/nodejs" \
        "$PROTO_DIR"/*.proto
fi
```

## 服务端实现

### 主服务器

```go
// cmd/server/main.go
package main

import (
    "context"
    "flag"
    "fmt"
    "log"
    "net"
    "os"
    "os/signal"
    "syscall"
    "time"
    
    "google.golang.org/grpc"
    "google.golang.org/grpc/health"
    "google.golang.org/grpc/health/grpc_health_v1"
    "google.golang.org/grpc/reflection"
    
    "github.com/your-username/music-source-rpc-plugin/internal/config"
    "github.com/your-username/music-source-rpc-plugin/internal/service"
    "github.com/your-username/music-source-rpc-plugin/pkg/api/generated"
)

var (
    port       = flag.Int("port", 50051, "The server port")
    configFile = flag.String("config", "configs/server.yaml", "Configuration file path")
)

func main() {
    flag.Parse()
    
    // 加载配置
    cfg, err := config.LoadConfig(*configFile)
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }
    
    // 创建监听器
    lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
    if err != nil {
        log.Fatalf("Failed to listen: %v", err)
    }
    
    // 创建 gRPC 服务器
    s := grpc.NewServer(
        grpc.MaxRecvMsgSize(cfg.Server.MaxMessageSize),
        grpc.MaxSendMsgSize(cfg.Server.MaxMessageSize),
        grpc.ConnectionTimeout(time.Duration(cfg.Server.ConnectionTimeout)*time.Second),
    )
    
    // 注册服务
    pluginService := service.NewPluginService(cfg)
    musicService := service.NewMusicSourceService(cfg)
    
    generated.RegisterPluginServiceServer(s, pluginService)
    generated.RegisterMusicSourceServiceServer(s, musicService)
    
    // 注册健康检查服务
    healthServer := health.NewServer()
    grpc_health_v1.RegisterHealthServer(s, healthServer)
    healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
    
    // 启用反射（开发环境）
    if cfg.Server.EnableReflection {
        reflection.Register(s)
    }
    
    // 启动服务器
    log.Printf("Starting gRPC server on port %d", *port)
    go func() {
        if err := s.Serve(lis); err != nil {
            log.Fatalf("Failed to serve: %v", err)
        }
    }()
    
    // 等待中断信号
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM)
    <-c
    
    log.Println("Shutting down gRPC server...")
    
    // 优雅关闭
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    done := make(chan struct{})
    go func() {
        s.GracefulStop()
        close(done)
    }()
    
    select {
    case <-done:
        log.Println("Server stopped gracefully")
    case <-ctx.Done():
        log.Println("Server stop timeout, forcing shutdown")
        s.Stop()
    }
}
```

### 插件服务实现

```go
// internal/service/plugin.go
package service

import (
    "context"
    "encoding/json"
    "fmt"
    "sync"
    "time"
    
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
    "google.golang.org/protobuf/types/known/structpb"
    
    "github.com/your-username/music-source-rpc-plugin/internal/config"
    "github.com/your-username/music-source-rpc-plugin/pkg/api/generated"
)

type PluginService struct {
    generated.UnimplementedPluginServiceServer
    
    config     *config.Config
    status     PluginStatus
    metadata   *generated.PluginMetadata
    startTime  time.Time
    mutex      sync.RWMutex
    
    // 服务组件
    musicService *MusicSourceService
}

type PluginStatus int

const (
    StatusUnknown PluginStatus = iota
    StatusInitializing
    StatusRunning
    StatusStopped
    StatusError
)

func NewPluginService(cfg *config.Config) *PluginService {
    return &PluginService{
        config: cfg,
        status: StatusUnknown,
        metadata: &generated.PluginMetadata{
            Id:          "netease-music-source",
            Name:        "NetEase Cloud Music Source",
            Version:     "1.0.0",
            Description: "NetEase Cloud Music integration plugin",
            Author:      "Your Name",
            License:     "MIT",
            Tags:        []string{"music", "streaming", "netease"},
            Capabilities: []string{"search", "stream", "lyrics", "user-auth"},
        },
    }
}

func (s *PluginService) Initialize(ctx context.Context, req *generated.InitializeRequest) (*generated.InitializeResponse, error) {
    s.mutex.Lock()
    defer s.mutex.Unlock()
    
    if s.status != StatusUnknown {
        return &generated.InitializeResponse{
            Success:      false,
            ErrorMessage: "Plugin already initialized",
        }, nil
    }
    
    s.status = StatusInitializing
    
    // 解析配置
    if req.ConfigJson != "" {
        var pluginConfig map[string]interface{}
        if err := json.Unmarshal([]byte(req.ConfigJson), &pluginConfig); err != nil {
            s.status = StatusError
            return &generated.InitializeResponse{
                Success:      false,
                ErrorMessage: fmt.Sprintf("Failed to parse config: %v", err),
            }, nil
        }
        
        // 更新配置
        if err := s.updateConfigFromMap(pluginConfig); err != nil {
            s.status = StatusError
            return &generated.InitializeResponse{
                Success:      false,
                ErrorMessage: fmt.Sprintf("Failed to update config: %v", err),
            }, nil
        }
    }
    
    // 初始化音乐服务
    s.musicService = NewMusicSourceService(s.config)
    if err := s.musicService.Initialize(ctx); err != nil {
        s.status = StatusError
        return &generated.InitializeResponse{
            Success:      false,
            ErrorMessage: fmt.Sprintf("Failed to initialize music service: %v", err),
        }, nil
    }
    
    // 设置元数据
    metadataStruct, _ := structpb.NewStruct(map[string]interface{}{
        "api_version":     "v1",
        "supported_formats": []string{"mp3", "flac", "m4a"},
        "max_quality":     "lossless",
        "requires_auth":   true,
    })
    s.metadata.Metadata = metadataStruct
    
    return &generated.InitializeResponse{
        Success:  true,
        Metadata: s.metadata,
    }, nil
}

func (s *PluginService) Start(ctx context.Context, req *generated.StartRequest) (*generated.StartResponse, error) {
    s.mutex.Lock()
    defer s.mutex.Unlock()
    
    if s.status != StatusInitializing {
        return &generated.StartResponse{
            Success:      false,
            ErrorMessage: "Plugin not initialized",
        }, nil
    }
    
    // 启动音乐服务
    if err := s.musicService.Start(ctx); err != nil {
        s.status = StatusError
        return &generated.StartResponse{
            Success:      false,
            ErrorMessage: fmt.Sprintf("Failed to start music service: %v", err),
        }, nil
    }
    
    s.status = StatusRunning
    s.startTime = time.Now()
    
    return &generated.StartResponse{
        Success: true,
    }, nil
}

func (s *PluginService) Stop(ctx context.Context, req *generated.StopRequest) (*generated.StopResponse, error) {
    s.mutex.Lock()
    defer s.mutex.Unlock()
    
    if s.status != StatusRunning {
        return &generated.StopResponse{
            Success:      false,
            ErrorMessage: "Plugin not running",
        }, nil
    }
    
    // 停止音乐服务
    if err := s.musicService.Stop(ctx); err != nil {
        return &generated.StopResponse{
            Success:      false,
            ErrorMessage: fmt.Sprintf("Failed to stop music service: %v", err),
        }, nil
    }
    
    s.status = StatusStopped
    
    return &generated.StopResponse{
        Success: true,
    }, nil
}

func (s *PluginService) HealthCheck(ctx context.Context, req *generated.HealthCheckRequest) (*generated.HealthCheckResponse, error) {
    s.mutex.RLock()
    defer s.mutex.RUnlock()
    
    healthy := s.status == StatusRunning
    
    // 检查音乐服务健康状态
    if healthy && s.musicService != nil {
        if err := s.musicService.HealthCheck(ctx); err != nil {
            healthy = false
        }
    }
    
    response := &generated.HealthCheckResponse{
        Healthy: healthy,
        Status:  s.getStatusString(),
        Uptime:  int64(time.Since(s.startTime).Seconds()),
    }
    
    if !healthy {
        response.ErrorMessage = "Service is not healthy"
    }
    
    return response, nil
}

func (s *PluginService) GetMetadata(ctx context.Context, req *generated.GetMetadataRequest) (*generated.GetMetadataResponse, error) {
    s.mutex.RLock()
    defer s.mutex.RUnlock()
    
    return &generated.GetMetadataResponse{
        Metadata: s.metadata,
    }, nil
}

func (s *PluginService) GetConfig(ctx context.Context, req *generated.GetConfigRequest) (*generated.GetConfigResponse, error) {
    s.mutex.RLock()
    defer s.mutex.RUnlock()
    
    configBytes, err := json.Marshal(s.config)
    if err != nil {
        return nil, status.Errorf(codes.Internal, "Failed to marshal config: %v", err)
    }
    
    return &generated.GetConfigResponse{
        ConfigJson: string(configBytes),
    }, nil
}

func (s *PluginService) UpdateConfig(ctx context.Context, req *generated.UpdateConfigRequest) (*generated.UpdateConfigResponse, error) {
    s.mutex.Lock()
    defer s.mutex.Unlock()
    
    var configMap map[string]interface{}
    if err := json.Unmarshal([]byte(req.ConfigJson), &configMap); err != nil {
        return &generated.UpdateConfigResponse{
            Success:      false,
            ErrorMessage: fmt.Sprintf("Failed to parse config: %v", err),
        }, nil
    }
    
    if err := s.updateConfigFromMap(configMap); err != nil {
        return &generated.UpdateConfigResponse{
            Success:      false,
            ErrorMessage: fmt.Sprintf("Failed to update config: %v", err),
        }, nil
    }
    
    // 通知音乐服务配置更新
    if s.musicService != nil {
        if err := s.musicService.UpdateConfig(s.config); err != nil {
            return &generated.UpdateConfigResponse{
                Success:      false,
                ErrorMessage: fmt.Sprintf("Failed to update music service config: %v", err),
            }, nil
        }
    }
    
    return &generated.UpdateConfigResponse{
        Success: true,
    }, nil
}

// 私有方法
func (s *PluginService) getStatusString() string {
    switch s.status {
    case StatusUnknown:
        return "unknown"
    case StatusInitializing:
        return "initializing"
    case StatusRunning:
        return "running"
    case StatusStopped:
        return "stopped"
    case StatusError:
        return "error"
    default:
        return "unknown"
    }
}

func (s *PluginService) updateConfigFromMap(configMap map[string]interface{}) error {
    // 这里应该根据实际配置结构进行更新
    // 简化实现
    configBytes, err := json.Marshal(configMap)
    if err != nil {
        return err
    }
    
    var newConfig config.Config
    if err := json.Unmarshal(configBytes, &newConfig); err != nil {
        return err
    }
    
    // 验证配置
    if err := newConfig.Validate(); err != nil {
        return err
    }
    
    s.config = &newConfig
    return nil
}
```

### 音乐服务实现

```go
// internal/service/music.go
package service

import (
    "context"
    "fmt"
    "sync"
    "time"
    
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
    "google.golang.org/protobuf/types/known/durationpb"
    "google.golang.org/protobuf/types/known/timestamppb"
    
    "github.com/your-username/music-source-rpc-plugin/internal/client"
    "github.com/your-username/music-source-rpc-plugin/internal/config"
    "github.com/your-username/music-source-rpc-plugin/pkg/api/generated"
)

type MusicSourceService struct {
    generated.UnimplementedMusicSourceServiceServer
    
    config       *config.Config
    neteaseClient *client.NeteaseClient
    mutex        sync.RWMutex
    
    // 缓存
    searchCache  map[string]*generated.SearchResponse
    streamCache  map[string]*generated.GetStreamURLResponse
    cacheMutex   sync.RWMutex
}

func NewMusicSourceService(cfg *config.Config) *MusicSourceService {
    return &MusicSourceService{
        config:      cfg,
        searchCache: make(map[string]*generated.SearchResponse),
        streamCache: make(map[string]*generated.GetStreamURLResponse),
    }
}

func (s *MusicSourceService) Initialize(ctx context.Context) error {
    s.mutex.Lock()
    defer s.mutex.Unlock()
    
    // 初始化网易云音乐客户端
    s.neteaseClient = client.NewNeteaseClient(s.config.Netease)
    
    return nil
}

func (s *MusicSourceService) Start(ctx context.Context) error {
    s.mutex.Lock()
    defer s.mutex.Unlock()
    
    // 启动客户端
    if err := s.neteaseClient.Start(ctx); err != nil {
        return fmt.Errorf("failed to start netease client: %w", err)
    }
    
    return nil
}

func (s *MusicSourceService) Stop(ctx context.Context) error {
    s.mutex.Lock()
    defer s.mutex.Unlock()
    
    // 停止客户端
    if err := s.neteaseClient.Stop(ctx); err != nil {
        return fmt.Errorf("failed to stop netease client: %w", err)
    }
    
    return nil
}

func (s *MusicSourceService) HealthCheck(ctx context.Context) error {
    s.mutex.RLock()
    defer s.mutex.RUnlock()
    
    if s.neteaseClient == nil {
        return fmt.Errorf("netease client not initialized")
    }
    
    return s.neteaseClient.HealthCheck(ctx)
}

func (s *MusicSourceService) UpdateConfig(cfg *config.Config) error {
    s.mutex.Lock()
    defer s.mutex.Unlock()
    
    s.config = cfg
    
    // 更新客户端配置
    if s.neteaseClient != nil {
        return s.neteaseClient.UpdateConfig(cfg.Netease)
    }
    
    return nil
}

func (s *MusicSourceService) Search(ctx context.Context, req *generated.SearchRequest) (*generated.SearchResponse, error) {
    // 检查缓存
    cacheKey := fmt.Sprintf("%s_%v", req.Query, req.Options)
    if cached := s.getCachedSearch(cacheKey); cached != nil {
        return cached, nil
    }
    
    s.mutex.RLock()
    client := s.neteaseClient
    s.mutex.RUnlock()
    
    if client == nil {
        return nil, status.Error(codes.FailedPrecondition, "Service not initialized")
    }
    
    // 执行搜索
    result, err := client.Search(ctx, req.Query, convertSearchOptions(req.Options))
    if err != nil {
        return &generated.SearchResponse{
            ErrorMessage: err.Error(),
        }, nil
    }
    
    // 转换结果
    response := &generated.SearchResponse{
        Songs:      convertSongs(result.Songs),
        Albums:     convertAlbums(result.Albums),
        Artists:    convertArtists(result.Artists),
        Playlists:  convertPlaylists(result.Playlists),
        TotalCount: int32(result.TotalCount),
        HasMore:    result.HasMore,
    }
    
    // 缓存结果
    s.setCachedSearch(cacheKey, response)
    
    return response, nil
}

func (s *MusicSourceService) GetSong(ctx context.Context, req *generated.GetSongRequest) (*generated.GetSongResponse, error) {
    s.mutex.RLock()
    client := s.neteaseClient
    s.mutex.RUnlock()
    
    if client == nil {
        return nil, status.Error(codes.FailedPrecondition, "Service not initialized")
    }
    
    song, err := client.GetSong(ctx, req.SongId)
    if err != nil {
        return &generated.GetSongResponse{
            ErrorMessage: err.Error(),
        }, nil
    }
    
    return &generated.GetSongResponse{
        Song: convertSong(song),
    }, nil
}

func (s *MusicSourceService) GetStreamURL(ctx context.Context, req *generated.GetStreamURLRequest) (*generated.GetStreamURLResponse, error) {
    // 检查缓存
    cacheKey := fmt.Sprintf("%s_%v", req.SongId, req.Quality)
    if cached := s.getCachedStream(cacheKey); cached != nil {
        return cached, nil
    }
    
    s.mutex.RLock()
    client := s.neteaseClient
    s.mutex.RUnlock()
    
    if client == nil {
        return nil, status.Error(codes.FailedPrecondition, "Service not initialized")
    }
    
    streamInfo, err := client.GetStreamURL(ctx, req.SongId, convertQuality(req.Quality))
    if err != nil {
        return &generated.GetStreamURLResponse{
            ErrorMessage: err.Error(),
        }, nil
    }
    
    response := &generated.GetStreamURLResponse{
        Url:           streamInfo.URL,
        ActualQuality: convertQualityToProto(streamInfo.Quality),
        ExpiresIn:     durationpb.New(streamInfo.ExpiresIn),
    }
    
    // 缓存结果（较短时间）
    s.setCachedStream(cacheKey, response)
    
    return response, nil
}

func (s *MusicSourceService) Login(ctx context.Context, req *generated.LoginRequest) (*generated.LoginResponse, error) {
    s.mutex.RLock()
    client := s.neteaseClient
    s.mutex.RUnlock()
    
    if client == nil {
        return nil, status.Error(codes.FailedPrecondition, "Service not initialized")
    }
    
    var err error
    var userInfo *client.UserInfo
    var sessionToken string
    var expiresAt time.Time
    
    switch cred := req.Credentials.(type) {
    case *generated.LoginRequest_UsernamePassword:
        userInfo, sessionToken, expiresAt, err = client.LoginWithPassword(
            ctx, cred.UsernamePassword.Username, cred.UsernamePassword.Password)
    case *generated.LoginRequest_Token:
        userInfo, sessionToken, expiresAt, err = client.LoginWithToken(ctx, cred.Token)
    case *generated.LoginRequest_OauthCode:
        userInfo, sessionToken, expiresAt, err = client.LoginWithOAuth(ctx, cred.OauthCode)
    default:
        return &generated.LoginResponse{
            Success:      false,
            ErrorMessage: "Invalid credentials type",
        }, nil
    }
    
    if err != nil {
        return &generated.LoginResponse{
            Success:      false,
            ErrorMessage: err.Error(),
        }, nil
    }
    
    return &generated.LoginResponse{
        Success:      true,
        UserInfo:     convertUserInfo(userInfo),
        SessionToken: sessionToken,
        ExpiresAt:    timestamppb.New(expiresAt),
    }, nil
}

// 缓存管理
func (s *MusicSourceService) getCachedSearch(key string) *generated.SearchResponse {
    s.cacheMutex.RLock()
    defer s.cacheMutex.RUnlock()
    
    return s.searchCache[key]
}

func (s *MusicSourceService) setCachedSearch(key string, response *generated.SearchResponse) {
    s.cacheMutex.Lock()
    defer s.cacheMutex.Unlock()
    
    // 简单的 LRU 缓存实现
    if len(s.searchCache) > 100 {
        // 清理一半缓存
        for k := range s.searchCache {
            delete(s.searchCache, k)
            if len(s.searchCache) <= 50 {
                break
            }
        }
    }
    
    s.searchCache[key] = response
}

func (s *MusicSourceService) getCachedStream(key string) *generated.GetStreamURLResponse {
    s.cacheMutex.RLock()
    defer s.cacheMutex.RUnlock()
    
    return s.streamCache[key]
}

func (s *MusicSourceService) setCachedStream(key string, response *generated.GetStreamURLResponse) {
    s.cacheMutex.Lock()
    defer s.cacheMutex.Unlock()
    
    // 流媒体 URL 缓存时间较短
    if len(s.streamCache) > 50 {
        for k := range s.streamCache {
            delete(s.streamCache, k)
            if len(s.streamCache) <= 25 {
                break
            }
        }
    }
    
    s.streamCache[key] = response
    
    // 设置过期时间（简化实现）
    go func() {
        time.Sleep(5 * time.Minute)
        s.cacheMutex.Lock()
        delete(s.streamCache, key)
        s.cacheMutex.Unlock()
    }()
}

// 转换函数
func convertSong(song *client.Song) *generated.Song {
    return &generated.Song{
        Id:          song.ID,
        Title:       song.Title,
        Artist:      song.Artist,
        Album:       song.Album,
        Duration:    durationpb.New(song.Duration),
        Genre:       song.Genre,
        Year:        int32(song.Year),
        TrackNumber: int32(song.TrackNumber),
        CoverUrl:    song.CoverURL,
    }
}

func convertSongs(songs []*client.Song) []*generated.Song {
    result := make([]*generated.Song, len(songs))
    for i, song := range songs {
        result[i] = convertSong(song)
    }
    return result
}

// 其他转换函数...
```

## 客户端集成

### 主程序中的 RPC 客户端

```go
// pkg/api/client.go
package api

import (
    "context"
    "fmt"
    "time"
    
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
    "google.golang.org/grpc/keepalive"
    
    "github.com/your-username/music-source-rpc-plugin/pkg/api/generated"
)

type RPCClient struct {
    conn           *grpc.ClientConn
    pluginClient   generated.PluginServiceClient
    musicClient    generated.MusicSourceServiceClient
    
    address        string
    timeout        time.Duration
}

type ClientConfig struct {
    Address         string
    Timeout         time.Duration
    MaxMessageSize  int
    KeepAlive       KeepAliveConfig
}

type KeepAliveConfig struct {
    Time                time.Duration
    Timeout             time.Duration
    PermitWithoutStream bool
}

func NewRPCClient(config ClientConfig) (*RPCClient, error) {
    // 设置连接选项
    opts := []grpc.DialOption{
        grpc.WithTransportCredentials(insecure.NewCredentials()),
        grpc.WithDefaultCallOptions(
            grpc.MaxCallRecvMsgSize(config.MaxMessageSize),
            grpc.MaxCallSendMsgSize(config.MaxMessageSize),
        ),
        grpc.WithKeepaliveParams(keepalive.ClientParameters{
            Time:                config.KeepAlive.Time,
            Timeout:             config.KeepAlive.Timeout,
            PermitWithoutStream: config.KeepAlive.PermitWithoutStream,
        }),
    }
    
    // 建立连接
    conn, err := grpc.Dial(config.Address, opts...)
    if err != nil {
        return nil, fmt.Errorf("failed to connect to %s: %w", config.Address, err)
    }
    
    return &RPCClient{
        conn:         conn,
        pluginClient: generated.NewPluginServiceClient(conn),
        musicClient:  generated.NewMusicSourceServiceClient(conn),
        address:      config.Address,
        timeout:      config.Timeout,
    }, nil
}

func (c *RPCClient) Close() error {
    return c.conn.Close()
}

// 插件生命周期方法
func (c *RPCClient) Initialize(config map[string]interface{}) error {
    ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
    defer cancel()
    
    configJSON, err := json.Marshal(config)
    if err != nil {
        return fmt.Errorf("failed to marshal config: %w", err)
    }
    
    resp, err := c.pluginClient.Initialize(ctx, &generated.InitializeRequest{
        ConfigJson: string(configJSON),
    })
    if err != nil {
        return fmt.Errorf("initialize failed: %w", err)
    }
    
    if !resp.Success {
        return fmt.Errorf("initialize failed: %s", resp.ErrorMessage)
    }
    
    return nil
}

func (c *RPCClient) Start() error {
    ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
    defer cancel()
    
    resp, err := c.pluginClient.Start(ctx, &generated.StartRequest{})
    if err != nil {
        return fmt.Errorf("start failed: %w", err)
    }
    
    if !resp.Success {
        return fmt.Errorf("start failed: %s", resp.ErrorMessage)
    }
    
    return nil
}

func (c *RPCClient) Stop() error {
    ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
    defer cancel()
    
    resp, err := c.pluginClient.Stop(ctx, &generated.StopRequest{})
    if err != nil {
        return fmt.Errorf("stop failed: %w", err)
    }
    
    if !resp.Success {
        return fmt.Errorf("stop failed: %s", resp.ErrorMessage)
    }
    
    return nil
}

func (c *RPCClient) HealthCheck() error {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    resp, err := c.pluginClient.HealthCheck(ctx, &generated.HealthCheckRequest{})
    if err != nil {
        return fmt.Errorf("health check failed: %w", err)
    }
    
    if !resp.Healthy {
        return fmt.Errorf("service unhealthy: %s", resp.ErrorMessage)
    }
    
    return nil
}

// 音乐服务方法
func (c *RPCClient) Search(query string, options SearchOptions) (*SearchResult, error) {
    ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
    defer cancel()
    
    resp, err := c.musicClient.Search(ctx, &generated.SearchRequest{
        Query: query,
        Options: &generated.SearchOptions{
            Limit:  int32(options.Limit),
            Offset: int32(options.Offset),
            Type:   options.Type,
            Sort:   options.Sort,
        },
    })
    if err != nil {
        return nil, fmt.Errorf("search failed: %w", err)
    }
    
    if resp.ErrorMessage != "" {
        return nil, fmt.Errorf("search error: %s", resp.ErrorMessage)
    }
    
    return &SearchResult{
        Songs:      convertProtoSongs(resp.Songs),
        Albums:     convertProtoAlbums(resp.Albums),
        Artists:    convertProtoArtists(resp.Artists),
        Playlists:  convertProtoPlaylists(resp.Playlists),
        TotalCount: int(resp.TotalCount),
        HasMore:    resp.HasMore,
    }, nil
}

func (c *RPCClient) GetStreamURL(songID string, quality Quality) (*StreamInfo, error) {
    ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
    defer cancel()
    
    resp, err := c.musicClient.GetStreamURL(ctx, &generated.GetStreamURLRequest{
        SongId:  songID,
        Quality: convertQualityToProto(quality),
    })
    if err != nil {
        return nil, fmt.Errorf("get stream URL failed: %w", err)
    }
    
    if resp.ErrorMessage != "" {
        return nil, fmt.Errorf("get stream URL error: %s", resp.ErrorMessage)
    }
    
    return &StreamInfo{
        URL:           resp.Url,
        Quality:       convertProtoQuality(resp.ActualQuality),
        ExpiresIn:     resp.ExpiresIn.AsDuration(),
    }, nil
}

// 连接管理
func (c *RPCClient) IsConnected() bool {
    return c.conn.GetState() == connectivity.Ready
}

func (c *RPCClient) WaitForReady(timeout time.Duration) error {
    ctx, cancel := context.WithTimeout(context.Background(), timeout)
    defer cancel()
    
    return c.conn.WaitForStateChange(ctx, c.conn.GetState())
}
```

## 通信优化

### 连接池管理

```go
// pkg/api/pool.go
package api

import (
    "context"
    "sync"
    "time"
)

type ConnectionPool struct {
    clients    []*RPCClient
    available  chan *RPCClient
    config     ClientConfig
    size       int
    mutex      sync.RWMutex
    closed     bool
}

func NewConnectionPool(config ClientConfig, size int) (*ConnectionPool, error) {
    pool := &ConnectionPool{
        clients:   make([]*RPCClient, 0, size),
        available: make(chan *RPCClient, size),
        config:    config,
        size:      size,
    }
    
    // 创建初始连接
    for i := 0; i < size; i++ {
        client, err := NewRPCClient(config)
        if err != nil {
            pool.Close()
            return nil, err
        }
        
        pool.clients = append(pool.clients, client)
        pool.available <- client
    }
    
    return pool, nil
}

func (p *ConnectionPool) Get(ctx context.Context) (*RPCClient, error) {
    p.mutex.RLock()
    if p.closed {
        p.mutex.RUnlock()
        return nil, fmt.Errorf("connection pool is closed")
    }
    p.mutex.RUnlock()
    
    select {
    case client := <-p.available:
        // 检查连接状态
        if !client.IsConnected() {
            // 重新连接
            if err := client.reconnect(); err != nil {
                // 创建新连接
                newClient, err := NewRPCClient(p.config)
                if err != nil {
                    p.available <- client // 放回原连接
                    return nil, err
                }
                client.Close()
                client = newClient
            }
        }
        return client, nil
    case <-ctx.Done():
        return nil, ctx.Err()
    }
}

func (p *ConnectionPool) Put(client *RPCClient) {
    p.mutex.RLock()
    if p.closed {
        p.mutex.RUnlock()
        client.Close()
        return
    }
    p.mutex.RUnlock()
    
    select {
    case p.available <- client:
    default:
        // 池已满，关闭连接
        client.Close()
    }
}

func (p *ConnectionPool) Close() error {
    p.mutex.Lock()
    defer p.mutex.Unlock()
    
    if p.closed {
        return nil
    }
    
    p.closed = true
    close(p.available)
    
    for _, client := range p.clients {
        client.Close()
    }
    
    return nil
}
```

### 流式处理

```go
// 流式音频处理
func (c *RPCClient) ProcessAudioStream(ctx context.Context, input <-chan *AudioBuffer) (<-chan *AudioBuffer, error) {
    stream, err := c.musicClient.ProcessAudioStream(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to create stream: %w", err)
    }
    
    output := make(chan *AudioBuffer, 10)
    
    // 发送协程
    go func() {
        defer stream.CloseSend()
        
        for buffer := range input {
            req := &generated.ProcessAudioRequest{
                Input: convertAudioBufferToProto(buffer),
            }
            
            if err := stream.Send(req); err != nil {
                log.Printf("Failed to send audio buffer: %v", err)
                return
            }
        }
    }()
    
    // 接收协程
    go func() {
        defer close(output)
        
        for {
            resp, err := stream.Recv()
            if err == io.EOF {
                return
            }
            if err != nil {
                log.Printf("Failed to receive audio buffer: %v", err)
                return
            }
            
            if resp.ErrorMessage != "" {
                log.Printf("Audio processing error: %s", resp.ErrorMessage)
                continue
            }
            
            buffer := convertProtoAudioBuffer(resp.Output)
            
            select {
            case output <- buffer:
            case <-ctx.Done():
                return
            }
        }
    }()
    
    return output, nil
}
```

## 错误处理

### 重试机制

```go
// pkg/api/retry.go
package api

import (
    "context"
    "math"
    "time"
    
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
)

type RetryConfig struct {
    MaxAttempts     int
    InitialDelay    time.Duration
    MaxDelay        time.Duration
    BackoffFactor   float64
    RetryableCodes  []codes.Code
}

func DefaultRetryConfig() RetryConfig {
    return RetryConfig{
        MaxAttempts:   3,
        InitialDelay:  100 * time.Millisecond,
        MaxDelay:      5 * time.Second,
        BackoffFactor: 2.0,
        RetryableCodes: []codes.Code{
            codes.Unavailable,
            codes.DeadlineExceeded,
            codes.ResourceExhausted,
            codes.Aborted,
        },
    }
}

func WithRetry(ctx context.Context, config RetryConfig, fn func() error) error {
    var lastErr error
    delay := config.InitialDelay
    
    for attempt := 0; attempt < config.MaxAttempts; attempt++ {
        if attempt > 0 {
            // 等待重试延迟
            select {
            case <-time.After(delay):
            case <-ctx.Done():
                return ctx.Err()
            }
            
            // 指数退避
            delay = time.Duration(float64(delay) * config.BackoffFactor)
            if delay > config.MaxDelay {
                delay = config.MaxDelay
            }
        }
        
        err := fn()
        if err == nil {
            return nil
        }
        
        lastErr = err
        
        // 检查是否为可重试错误
        if grpcErr, ok := status.FromError(err); ok {
            retryable := false
            for _, code := range config.RetryableCodes {
                if grpcErr.Code() == code {
                    retryable = true
                    break
                }
            }
            if !retryable {
                return err
            }
        }
    }
    
    return lastErr
}
```

### 断路器模式

```go
// pkg/api/circuitbreaker.go
package api

import (
    "fmt"
    "sync"
    "time"
)

type CircuitBreakerState int

const (
    StateClosed CircuitBreakerState = iota
    StateHalfOpen
    StateOpen
)

type CircuitBreaker struct {
    maxFailures     int
    resetTimeout    time.Duration
    
    mutex           sync.RWMutex
    state           CircuitBreakerState
    failures        int
    lastFailureTime time.Time
    nextAttempt     time.Time
}

func NewCircuitBreaker(maxFailures int, resetTimeout time.Duration) *CircuitBreaker {
    return &CircuitBreaker{
        maxFailures:  maxFailures,
        resetTimeout: resetTimeout,
        state:        StateClosed,
    }
}

func (cb *CircuitBreaker) Call(fn func() error) error {
    cb.mutex.Lock()
    defer cb.mutex.Unlock()
    
    now := time.Now()
    
    switch cb.state {
    case StateOpen:
        if now.Before(cb.nextAttempt) {
            return fmt.Errorf("circuit breaker is open")
        }
        cb.state = StateHalfOpen
        
    case StateHalfOpen:
        // 允许一次尝试
        
    case StateClosed:
        // 正常状态
    }
    
    err := fn()
    
    if err != nil {
        cb.onFailure(now)
        return err
    }
    
    cb.onSuccess()
    return nil
}

func (cb *CircuitBreaker) onSuccess() {
    cb.failures = 0
    cb.state = StateClosed
}

func (cb *CircuitBreaker) onFailure(now time.Time) {
    cb.failures++
    cb.lastFailureTime = now
    
    if cb.failures >= cb.maxFailures {
        cb.state = StateOpen
        cb.nextAttempt = now.Add(cb.resetTimeout)
    }
}

func (cb *CircuitBreaker) GetState() CircuitBreakerState {
    cb.mutex.RLock()
    defer cb.mutex.RUnlock()
    return cb.state
}
```

## 部署配置

### Docker 部署

```dockerfile
# docker/Dockerfile
FROM golang:1.21-alpine AS builder

# 安装依赖
RUN apk add --no-cache git ca-certificates tzdata

# 设置工作目录
WORKDIR /app

# 复制 go mod 文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server ./cmd/server

# 运行阶段
FROM alpine:latest

# 安装 ca-certificates
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# 复制二进制文件
COPY --from=builder /app/server .

# 复制配置文件
COPY --from=builder /app/configs ./configs

# 暴露端口
EXPOSE 50051

# 健康检查
HEALTHCHEK --interval=30s --timeout=5s --start-period=5s --retries=3 \
    CMD grpc_health_probe -addr=localhost:50051 || exit 1

# 启动命令
CMD ["./server"]
```

### Docker Compose

```yaml
# docker/docker-compose.yml
version: '3.8'

services:
  music-source-plugin:
    build:
      context: ..
      dockerfile: docker/Dockerfile
    ports:
      - "50051:50051"
    environment:
      - LOG_LEVEL=info
      - CONFIG_FILE=/app/configs/server.yaml
    volumes:
      - ../configs:/app/configs:ro
      - plugin_data:/app/data
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "grpc_health_probe", "-addr=localhost:50051"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s
    networks:
      - plugin_network
    
  # 可选：添加监控
  prometheus:
    image: prom/prometheus:latest
    ports:
      - "9090:9090"
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml:ro
    networks:
      - plugin_network
      
  grafana:
    image: grafana/grafana:latest
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
    volumes:
      - grafana_data:/var/lib/grafana
    networks:
      - plugin_network

volumes:
  plugin_data:
  grafana_data:

networks:
  plugin_network:
    driver: bridge
```

### Kubernetes 部署

```yaml
# k8s/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: music-source-plugin
  labels:
    app: music-source-plugin
spec:
  replicas: 3
  selector:
    matchLabels:
      app: music-source-plugin
  template:
    metadata:
      labels:
        app: music-source-plugin
    spec:
      containers:
      - name: plugin
        image: your-registry/music-source-plugin:latest
        ports:
        - containerPort: 50051
          name: grpc
        env:
        - name: LOG_LEVEL
          value: "info"
        - name: CONFIG_FILE
          value: "/app/configs/server.yaml"
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          exec:
            command:
            - grpc_health_probe
            - -addr=localhost:50051
          initialDelaySeconds: 30
          periodSeconds: 30
        readinessProbe:
          exec:
            command:
            - grpc_health_probe
            - -addr=localhost:50051
          initialDelaySeconds: 5
          periodSeconds: 10
        volumeMounts:
        - name: config
          mountPath: /app/configs
          readOnly: true
      volumes:
      - name: config
        configMap:
          name: plugin-config
---
apiVersion: v1
kind: Service
metadata:
  name: music-source-plugin-service
spec:
  selector:
    app: music-source-plugin
  ports:
  - port: 50051
    targetPort: 50051
    name: grpc
  type: ClusterIP
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: plugin-config
data:
  server.yaml: |
    server:
      port: 50051
      max_message_size: 4194304
      connection_timeout: 30
      enable_reflection: false
    
    netease:
      api_base_url: "https://music.163.com/api"
      timeout: 30
      retry_count: 3
      
    logging:
      level: info
      format: json
```

## 监控和调试

### 指标收集

```go
// internal/utils/metrics.go
package utils

import (
    "context"
    "time"
    
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
    "google.golang.org/grpc"
    "google.golang.org/grpc/status"
)

var (
    // gRPC 请求计数器
    grpcRequestsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "grpc_requests_total",
            Help: "Total number of gRPC requests",
        },
        []string{"method", "status"},
    )
    
    // gRPC 请求延迟
    grpcRequestDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "grpc_request_duration_seconds",
            Help:    "Duration of gRPC requests",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method"},
    )
    
    // 活跃连接数
    activeConnections = promauto.NewGauge(
        prometheus.GaugeOpts{
            Name: "grpc_active_connections",
            Help: "Number of active gRPC connections",
        },
    )
    
    // 音乐搜索请求
    musicSearchRequests = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "music_search_requests_total",
            Help: "Total number of music search requests",
        },
        []string{"source", "status"},
    )
)

// gRPC 指标拦截器
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
    return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
        start := time.Now()
        
        resp, err := handler(ctx, req)
        
        duration := time.Since(start)
        method := info.FullMethod
        
        // 记录延迟
        grpcRequestDuration.WithLabelValues(method).Observe(duration.Seconds())
        
        // 记录请求计数
        statusCode := "OK"
        if err != nil {
            if st, ok := status.FromError(err); ok {
                statusCode = st.Code().String()
            } else {
                statusCode = "Unknown"
            }
        }
        grpcRequestsTotal.WithLabelValues(method, statusCode).Inc()
        
        return resp, err
    }
}

// 流式拦截器
func StreamServerInterceptor() grpc.StreamServerInterceptor {
    return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
        start := time.Now()
        
        err := handler(srv, stream)
        
        duration := time.Since(start)
        method := info.FullMethod
        
        grpcRequestDuration.WithLabelValues(method).Observe(duration.Seconds())
        
        statusCode := "OK"
        if err != nil {
            if st, ok := status.FromError(err); ok {
                statusCode = st.Code().String()
            } else {
                statusCode = "Unknown"
            }
        }
        grpcRequestsTotal.WithLabelValues(method, statusCode).Inc()
        
        return err
    }
}
```

### 日志记录

```go
// internal/utils/logger.go
package utils

import (
    "context"
    "log/slog"
    "os"
    
    "google.golang.org/grpc"
    "google.golang.org/grpc/metadata"
)

type contextKey string

const (
    requestIDKey contextKey = "request_id"
    userIDKey    contextKey = "user_id"
)

func NewLogger() *slog.Logger {
    opts := &slog.HandlerOptions{
        Level: slog.LevelInfo,
        AddSource: true,
    }
    
    handler := slog.NewJSONHandler(os.Stdout, opts)
    return slog.New(handler)
}

// gRPC 日志拦截器
func LoggingUnaryInterceptor(logger *slog.Logger) grpc.UnaryServerInterceptor {
    return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
        // 提取元数据
        md, _ := metadata.FromIncomingContext(ctx)
        requestID := getMetadataValue(md, "request-id")
        userID := getMetadataValue(md, "user-id")
        
        // 添加到上下文
        ctx = context.WithValue(ctx, requestIDKey, requestID)
        ctx = context.WithValue(ctx, userIDKey, userID)
        
        logger.InfoContext(ctx, "gRPC request started",
            "method", info.FullMethod,
            "request_id", requestID,
            "user_id", userID,
        )
        
        resp, err := handler(ctx, req)
        
        if err != nil {
            logger.ErrorContext(ctx, "gRPC request failed",
                "method", info.FullMethod,
                "error", err,
                "request_id", requestID,
                "user_id", userID,
            )
        } else {
            logger.InfoContext(ctx, "gRPC request completed",
                "method", info.FullMethod,
                "request_id", requestID,
                "user_id", userID,
            )
        }
        
        return resp, err
    }
}

func getMetadataValue(md metadata.MD, key string) string {
    values := md.Get(key)
    if len(values) > 0 {
        return values[0]
    }
    return ""
}
```

## 最佳实践

### 1. 协议设计

- 使用版本化的 API
- 设计向后兼容的消息格式
- 合理使用 oneof 和 optional 字段
- 避免嵌套过深的消息结构

### 2. 性能优化

- 使用连接池管理连接
- 实现客户端负载均衡
- 合理设置超时和重试
- 使用流式处理大数据

### 3. 错误处理

- 使用标准的 gRPC 错误码
- 提供详细的错误信息
- 实现重试和断路器
- 记录详细的错误日志

### 4. 安全性

- 使用 TLS 加密通信
- 实现认证和授权
- 验证输入数据
- 限制资源使用

## 多语言支持

### Python 实现

```python
# python/server.py
import asyncio
import logging
from concurrent import futures

import grpc
from grpc_health.v1 import health
from grpc_health.v1 import health_pb2_grpc

from generated import plugin_pb2_grpc, music_pb2_grpc
from services import PluginService, MusicSourceService

class Server:
    def __init__(self, port=50051):
        self.port = port
        self.server = None
        
    async def start(self):
        self.server = grpc.aio.server(futures.ThreadPoolExecutor(max_workers=10))
        
        # 注册服务
        plugin_service = PluginService()
        music_service = MusicSourceService()
        
        plugin_pb2_grpc.add_PluginServiceServicer_to_server(plugin_service, self.server)
        music_pb2_grpc.add_MusicSourceServiceServicer_to_server(music_service, self.server)
        
        # 健康检查
        health_servicer = health.HealthServicer()
        health_pb2_grpc.add_HealthServicer_to_server(health_servicer, self.server)
        health_servicer.set("", health_pb2.HealthCheckResponse.SERVING)
        
        # 启动服务器
        listen_addr = f'[::]:{self.port}'
        self.server.add_insecure_port(listen_addr)
        
        logging.info(f"Starting server on {listen_addr}")
        await self.server.start()
        await self.server.wait_for_termination()
        
    async def stop(self):
        if self.server:
            await self.server.stop(5)

if __name__ == '__main__':
    logging.basicConfig(level=logging.INFO)
    server = Server()
    
    try:
        asyncio.run(server.start())
    except KeyboardInterrupt:
        asyncio.run(server.stop())
```

### Node.js 实现

```javascript
// nodejs/server.js
const grpc = require('@grpc/grpc-js');
const protoLoader = require('@grpc/proto-loader');
const path = require('path');

const PROTO_PATH = path.join(__dirname, '../proto/plugin.proto');

// 加载 proto 文件
const packageDefinition = protoLoader.loadSync(PROTO_PATH, {
    keepCase: true,
    longs: String,
    enums: String,
    defaults: true,
    oneofs: true
});

const pluginProto = grpc.loadPackageDefinition(packageDefinition).plugin;

// 服务实现
const pluginService = {
    initialize: (call, callback) => {
        console.log('Initialize called:', call.request);
        callback(null, {
            success: true,
            metadata: {
                id: 'nodejs-music-plugin',
                name: 'Node.js Music Plugin',
                version: '1.0.0',
                description: 'Music plugin implemented in Node.js'
            }
        });
    },
    
    start: (call, callback) => {
        console.log('Start called');
        callback(null, { success: true });
    },
    
    stop: (call, callback) => {
        console.log('Stop called');
        callback(null, { success: true });
    },
    
    healthCheck: (call, callback) => {
        callback(null, {
            healthy: true,
            status: 'running',
            uptime: process.uptime()
        });
    }
};

// 创建服务器
const server = new grpc.Server();
server.addService(pluginProto.PluginService.service, pluginService);

const port = process.env.PORT || 50051;
server.bindAsync(`0.0.0.0:${port}`, grpc.ServerCredentials.createInsecure(), (err, port) => {
    if (err) {
        console.error('Failed to bind server:', err);
        return;
    }
    
    console.log(`Server running on port ${port}`);
    server.start();
});

// 优雅关闭
process.on('SIGINT', () => {
    console.log('Shutting down server...');
    server.tryShutdown(() => {
        console.log('Server stopped');
        process.exit(0);
    });
});
```

## 相关文档

- [插件开发快速入门](plugin-quickstart.md)
- [动态库插件开发指南](dynamic-library.md)
- [WebAssembly 插件开发指南](webassembly.md)
- [插件测试指南](plugin-testing.md)
- [API 文档](../api/README.md)