# 网易云音乐插件 (Netease Music Plugin)

这是一个为 go-musicfox v2 开发的网易云音乐插件，提供完整的网易云音乐服务支持。

## 功能特性

### 登录功能
- ✅ 手机号+密码登录
- ✅ 邮箱+密码登录
- ✅ Cookie配置登录
- ✅ 二维码登录
- ✅ 登录状态管理和刷新

### 搜索功能
- ✅ 歌曲搜索
- ✅ 专辑搜索
- ✅ 艺术家搜索
- ✅ 播放列表搜索
- ✅ 支持分页和过滤

### 播放列表管理
- ✅ 获取播放列表信息
- ✅ 获取播放列表歌曲
- ✅ 创建播放列表
- ✅ 更新播放列表
- ✅ 删除播放列表
- ✅ 获取用户播放列表

### 歌曲信息
- ✅ 获取歌曲播放URL（支持多种音质）
- ✅ 获取歌曲歌词（支持时间轴歌词）
- ✅ 获取歌曲详细信息
- ✅ 获取歌曲评论

### 用户功能
- ✅ 获取用户信息
- ✅ 关注/取消关注用户
- ✅ 获取用户喜欢的歌曲
- ✅ 获取用户关注列表
- ✅ 获取用户粉丝列表
- ✅ 获取用户播放历史
- ✅ 获取用户云盘音乐

### 推荐功能
- ✅ 每日推荐歌曲
- ✅ 个性化推荐
- ✅ 推荐歌单
- ✅ 相似歌曲推荐
- 🚧 新歌速递（待完善）
- 🚧 热门歌手（待完善）
- ✅ 新碟上架

### 排行榜功能
- ✅ 获取排行榜列表
- ✅ 获取排行榜歌曲
- ✅ 热搜关键词
- ✅ 根据分类获取热门歌单
- ✅ 获取歌单分类

### 艺术家和专辑
- ✅ 获取艺术家信息
- ✅ 获取艺术家热门歌曲
- ✅ 获取艺术家专辑
- ✅ 获取专辑信息
- ✅ 获取专辑歌曲

### 电台功能
- ✅ 获取电台列表
- ✅ 获取电台节目

## 安装和使用

### 构建插件

```bash
cd v2/plugins/netease
go build -o netease-plugin .
```

### 运行测试

```bash
go test -v
```

### 配置说明

插件支持通过配置文件设置Cookie进行自动登录：

```json
{
  "netease": {
    "cookie": "your_netease_cookie_here"
  }
}
```

### 登录方式

#### 1. 手机号登录
```go
credentials := map[string]string{
    "type": "phone",
    "phone": "13800138000",
    "password": "your_password",
    "country_code": "86", // 可选，默认86
}
err := plugin.Login(ctx, credentials)
```

#### 2. 邮箱登录
```go
credentials := map[string]string{
    "type": "email",
    "email": "your_email@example.com",
    "password": "your_password",
}
err := plugin.Login(ctx, credentials)
```

#### 3. Cookie登录
```go
credentials := map[string]string{
    "type": "cookie",
    "cookie": "your_cookie_string",
}
err := plugin.Login(ctx, credentials)
```

#### 4. 二维码登录
```go
// 开始二维码登录
qrKey, qrURL, err := plugin.StartQRLogin(ctx)
if err != nil {
    // 处理错误
}

// 显示二维码给用户扫描
// ...

// 轮询检查登录状态
for {
    status, err := plugin.CheckQRLogin(ctx, qrKey)
    if err != nil {
        // 处理错误
        break
    }
    
    switch status {
    case "success":
        // 登录成功
        return
    case "expired":
        // 二维码过期
        return
    case "waiting":
        // 等待扫描
        time.Sleep(2 * time.Second)
    case "scanned":
        // 已扫描，等待确认
        time.Sleep(2 * time.Second)
    }
}
```

## API 使用示例

### 搜索歌曲
```go
options := plugin.SearchOptions{
    Query:  "周杰伦",
    Type:   plugin.SearchTypeTrack,
    Limit:  20,
    Offset: 0,
}

result, err := plugin.Search(ctx, "周杰伦", options)
if err != nil {
    // 处理错误
}

for _, track := range result.Tracks {
    fmt.Printf("歌曲: %s - %s\n", track.Title, track.Artist)
}
```

### 获取歌曲播放URL
```go
url, err := plugin.GetTrackURL(ctx, "trackID", plugin.AudioQualityHigh)
if err != nil {
    // 处理错误
}
fmt.Printf("播放URL: %s\n", url)
```

### 获取歌词
```go
lyrics, err := plugin.GetTrackLyrics(ctx, "trackID")
if err != nil {
    // 处理错误
}

fmt.Printf("歌词: %s\n", lyrics.Content)
for _, timed := range lyrics.TimedLyrics {
    fmt.Printf("[%v] %s\n", timed.Time, timed.Content)
}
```

## 技术架构

- **基础框架**: 基于 go-musicfox v2 插件系统
- **网易云API**: 使用 github.com/go-musicfox/netease-music 库
- **二维码生成**: 使用 github.com/skip2/go-qrcode 库
- **Cookie管理**: 使用 github.com/telanflow/cookiejar 库
- **JSON解析**: 使用 github.com/buger/jsonparser 库

## 文件结构

```
v2/plugins/netease/
├── plugin.go          # 主插件文件
├── auth.go            # 登录认证功能
├── search.go          # 搜索功能
├── playlist.go        # 播放列表管理
├── track.go           # 歌曲信息获取
├── user.go            # 用户相关功能
├── features.go        # 推荐和其他功能
├── recommend.go       # 推荐功能
├── chart.go           # 排行榜功能
├── plugin_test.go     # 测试文件
├── go.mod             # Go模块文件
└── README.md          # 说明文档
```

## 开发状态

- ✅ 核心功能已完成
- ✅ 登录功能已完成
- ✅ 搜索功能已完成
- ✅ 播放列表管理已完成
- ✅ 歌曲信息获取已完成
- ✅ 用户功能已完成
- ✅ 推荐功能基本完成
- ✅ 排行榜功能已完成
- ✅ 基本测试已完成
- 🚧 部分高级功能待完善

## 注意事项

1. **网络依赖**: 插件需要网络连接才能正常工作
2. **API限制**: 网易云音乐API可能有频率限制
3. **版权限制**: 某些歌曲可能因版权问题无法播放
4. **登录状态**: 建议定期刷新登录状态以保持会话有效
5. **Cookie安全**: 请妥善保管Cookie信息，避免泄露

## 贡献

欢迎提交Issue和Pull Request来改进这个插件。

## 许可证

MIT License