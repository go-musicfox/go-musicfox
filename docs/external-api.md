# SPlayer-Next 兼容外部 API

go-musicfox 提供了一个可选的本机 HTTP / WebSocket 服务，用于向 OBS、直播挂件和
其他本地工具暴露播放状态。接口字段和消息类型与
[SPlayer-Next 外部 API](https://github.com/SPlayer-Dev/SPlayer-Next/blob/master/docs/api.md)
保持兼容，因此可以直接连接现有的
[SPlayer-Next OBS 歌词挂件](https://github.com/VincentZyu233/splayer-status-html-overlay-for-obs)。

## 启用服务

外部 API 默认关闭。在 `config.toml` 中加入：

```toml
[externalApi]
enable = true
enableWs = true
allowLan = false
port = 14558
```

也可以在启动时使用命令行参数覆盖配置：

```sh
musicfox --external-api --external-api-port 14558
```

如果还需要从局域网访问，显式增加 `--external-api-allow-lan`。该服务没有鉴权，默认只监听
`127.0.0.1`，仅应在可信网络中开放局域网访问。

## HTTP 接口

基础地址：`http://127.0.0.1:14558/api`，时间单位均为毫秒，控制接口成功时返回
`{"ok":true}`。

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| `GET` | `/api/info` | 应用版本与 WebSocket 连接数 |
| `GET` | `/api/status` | 播放状态、进度、时长和音量 |
| `GET` | `/api/volume` | 当前音量（`0` 到 `1`） |
| `GET` | `/api/now-playing` | 当前歌曲轻量快照 |
| `GET` | `/api/lyrics` | 当前歌曲完整结构化歌词 |
| `POST` | `/api/play` | 播放 |
| `POST` | `/api/pause` | 暂停 |
| `POST` | `/api/stop` | 停止 |
| `POST` | `/api/next` | 下一曲 |
| `POST` | `/api/prev` | 上一曲 |
| `POST` | `/api/seek` | 跳转，JSON：`{"positionMs":60000}` |
| `POST` | `/api/volume` | 设置音量，JSON：`{"volume":0.5}` |

示例：

```sh
curl http://127.0.0.1:14558/api/status
curl http://127.0.0.1:14558/api/now-playing
curl -X POST http://127.0.0.1:14558/api/pause
curl -X POST http://127.0.0.1:14558/api/seek \
  -H "Content-Type: application/json" \
  -d '{"positionMs":60000}'
```

## WebSocket 接口

连接地址：`ws://127.0.0.1:14558/ws`。

服务器到客户端的消息格式：

```json
{"kind":"hello","clients":1}
{"kind":"event","type":"track","data":{"track":{}}}
{"kind":"event","type":"lyric","data":{"source":{},"lyric":[]}}
{"kind":"event","type":"status","data":{"state":"playing","position":12000}}
{"kind":"ack","op":"pause"}
{"kind":"error","op":"seek","error":"positionMs (number, >=0) required"}
```

客户端可以发送以下命令：

```json
{"op":"play"}
{"op":"pause"}
{"op":"stop"}
{"op":"next"}
{"op":"prev"}
{"op":"seek","positionMs":60000}
{"op":"setVolume","volume":0.5}
```

## 在 OBS 中使用

1. 启动 go-musicfox，并确认 `http://127.0.0.1:14558/api/status` 可以访问。
2. 在 OBS 中添加“浏览器”来源。
3. 选择 [SPlayer-Next OBS 歌词挂件](https://github.com/VincentZyu233/splayer-status-html-overlay-for-obs)
   中的 `index.html`，或使用它的本地 HTTP 地址。
4. 挂件中的 `apiHost` 使用 `127.0.0.1`，`apiPort` 使用 `14558`。
5. 建议设置自定义帧率为 60 FPS；如果 CPU 占用较高，可改为 30 FPS。

挂件会优先使用 WebSocket 推送，并定期请求 `/api/status` 校准播放进度。若页面能显示
“已连接”但没有歌曲信息，先在 go-musicfox 中播放一首歌，再检查
`/api/now-playing` 和 `/api/lyrics` 的返回值。
