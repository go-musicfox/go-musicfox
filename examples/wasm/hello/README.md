# WASM Hello 插件

go-musicfox WASM 插件系统的可运行示例插件。

它是标准 Go `wasip1` reactor：导出函数经 `//go:wasmexport` 编译，模块在运行时由宿主（`internal/wasm`）通过 [wazero](https://wazero.io/) 加载。

## 编译

```sh
GOOS=wasip1 GOARCH=wasm go build -buildmode=c-shared -o main.wasm .
```

产物是 reactor 模块（无 `_start`）：Go 运行时入口为 `_initialize`，宿主在实例化后显式调用。插件导出：

| 导出 | 签名 | 用途 |
|---|---|---|
| `alloc` | `(size: u32) -> u32` | 分配一段 guest 缓冲区供宿主写入 |
| `dealloc` | `(ptr: u32, size: u32)` | 释放先前分配的缓冲区 |
| `run` | `(reqPtr: u32, reqLen: u32) -> u64` | 处理一个 JSON 请求，返回一个 JSON 响应 |
| `hang` | `(reqPtr: u32, reqLen: u32) -> u64` | 死循环；供宿主超时测试使用 |

由于 Go 的 `wasmexport` ABI 只允许一个结果值，`run` 把 `(outPtr << 32) | outLen` 打包进单个 `u64`。

## 目录结构

一个插件是一个目录，包含 manifest 与 wasm 文件：

```
hello/
  manifest.toml   # 插件元数据 + 菜单声明
  main.wasm       # 编译产物 reactor
```

示例 `manifest.toml`：

```toml
id = "hello"
name = "Hello WASM"
version = "0.1.0"
author = "you"
description = "示例 WASM 插件"
sha256 = ""          # 可选：wasm 文件的十六进制 SHA-256；非空时启动校验
wasm = "main.wasm"   # 可选，默认 "main.wasm"

[[menus]]
key = "wasm_hello"   # 全局唯一的菜单注册 key
title = "你好 WASM"   # 主菜单项标题
after = ""           # 可选：主菜单 after-anchor 前驱项 key（空 = 追加在链尾）
export = "run"       # 要调用的 wasm 导出，默认 "run"
args = {}            # 可选：传给插件的静态参数
```

把目录放入配置的插件目录后，宿主在启动时加载它。

## 请求 / 响应协议

宿主以 JSON 请求调用 `run`，并期望收到 JSON 响应：

```jsonc
// 请求
{
  "version": 1,
  "action": "wasm_hello",           // 触发调用的菜单 key
  "args": { "name": "musicfox" },   // manifest 菜单项声明的静态参数
  "context": {
    "userId": 123,
    "userName": "musicfox",
    "playing": true,
    "song": { "id": 1, "name": "Song", "artist": "Artist", "album": "Album" }
  }
}

// 响应
{
  "action": "view",                 // "toast" | "view" | "open_url" | "exec"
  "title": "WASM Hello",
  "message": "Hello, musicfox! ...",
  "level": "info"
}
```

本插件根据 `args["name"]` 与上下文中当前歌曲名拼接问候语。设置 `args["action"] = "toast"` 可将响应 action 切换为 `toast`。
