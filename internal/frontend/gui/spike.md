# GUI-1 spike：Wails v2 + manual build 关键不确定点验证

> 目标：验证 D-GUI-1（方式 B：同 module 共存 + manual build，跳过 wails CLI 直接
> `go build -tags desktop,production`）与 D-GUI-2（窗口加载路径：GUI-B Navigate 外部
> http URL 优先，GUI-A AssetServer 回退）在 go-musicfox 架构下的可行性。
>
> 环境：Go 1.26.6 darwin/arm64，CGO_ENABLED=1，Xcode CLT 就位。
> Wails 版本：**v2.15.0**（模块 `github.com/wailsapp/wails/v2`，go.mod 要求 go 1.25.0，
> 与项目 go 1.26 兼容）。
>
> 证据引用格式：wails v2.15.0 模块源码（`$GOMODCACHE/github.com/wailsapp/wails/v2@v2.15.0/`）+ 行号；
> 仓库文件 + 行号；外部 URL。

---

## 1. 依赖引入与编译（spike 清单 1）—— ✅ 通过

**结论**：Wails v2.15.0 在本机（Go 1.26.6）引入后可编译，`go build ./...` 全绿；依赖保留手段为
`wails_spike` build tag 文件。

**执行记录**：

| 步骤 | 结果 |
|------|------|
| `go get github.com/wailsapp/wails/v2@v2.15.0` | 成功（经 `goproxy.cn`，本次未遇 504；若遇 504 用 `GOPROXY=direct` 或 `GONOSUMDB=github.com/wailsapp`，本项目拉 foxful 先例） |
| `go mod tidy` | 成功，wails 保留在 go.mod（见「依赖保留机制」） |
| `go mod vendor` + `make vendor` | 成功；modvendor 拷贝了 wails 的 CGo 头文件（darwin `*.h`、linux `window.c`）；vendor 增至 63M（预期） |
| `go build -tags wails_spike ./internal/frontend/gui/` | 通过 = wails v2 模块（cgo + 平台绑定）在本机可编译 |
| `go build ./...` | 全绿（gui 包以默认 tag 编译，仅 register.go 参与） |
| `go test ./internal/frontend/...` | 全绿；实机验证 `frontend.Registered() = [wails headless tui webui]` |
| `go vet ./internal/frontend/...` | 通过 |

**依赖保留机制**：`go mod tidy` 按「所有 build tag 均启用」求包闭包，因此只要存在一个导入 wails
的包，tidy 就不会删掉 wails 依赖。本票新增 `internal/frontend/gui/wails_spike.go`
（`//go:build wails_spike`，不参与正常构建），既把 wails 钉进 go.mod/go.sum，又是 spike 的
编译验证目标（`go build -tags wails_spike ./internal/frontend/gui/`）。**GUI-2 的 run.go 直接
import wails 后必须删除该文件**。

**对 GUI-2 的影响**：依赖已在 go.mod/vendor 就位，GUI-2 无需再动依赖；正常构建不含 wails
编译负担（`wails_spike` tag 文件默认排除）。

---

## 2. `wails.Run` 无 AssetServer 行为 + Navigate API + localhost 访问（spike 清单 2）—— 关键裁决：**GUI-B 可行**

### 2.1 `Assets:nil, Handler:nil, AssetServer:nil` → 启动期 `log.Fatal` 退出（非 panic、非白屏）

- `pkg/options/options.go:51-56`：`Assets`/`AssetsHandler` 已标记 Deprecated，推荐 `AssetServer`。
- `pkg/assetserver/common.go:16-32` `BuildAssetServerConfig`：三者皆 nil 时得到空的
  `assetserver.Options`，随后 `options.Validate()`（`pkg/options/assetserver/options.go`）报错
  **"AssetServer options invalid: either Assets, Handler or Middleware must be set"**。
- `internal/frontend/desktop/darwin/frontend.go:124`（linux 同构 :215；windows 同构）：
  `NewAssetServerMainPage` 返回 error 后 **`log.Fatal(err)` → 进程直接退出**。

**结论**：GUI-B 不能 `Assets:nil`。必须提供 `AssetServer: &assetserver.Options{Handler: <x>}`。
对 GUI-B，`<x>` 只需一个极小 handler（如返回 meta-refresh/JS 跳转 HTML），初始页
（`wails://wails/` 或 windows `http://wails.localhost/`）加载它后立即跳转到外部
`http://127.0.0.1:<port>/...`。GUI-A 回退路径则让 `<x>` = webui server mux
（`ServerOptions{Auth:false}`，S5-2 已留口）。

### 2.2 v2.15 没有公开的 `WebviewWindow.Navigate` API；导航手段 = `OnDomReady` + `runtime.WindowExecJS`

- 全模块 grep `Navigate` 仅一处：`internal/frontend/desktop/windows/frontend.go:614`
  `chromium.Navigate(f.startURL.String())` —— 内部初始加载，非公开 API。
- 公开导航手段（官方同款）：`runtime.WindowExecJS(ctx, js)`（`pkg/runtime/window.go:167`），
  经 frontend `ExecJS` → webview `evaluateJavaScript` 执行，**跨 origin 有效**（外部页面无
  wails runtime 也能执行）。
- wails 自身 reload/reroute 即用 `ExecJS("window.location.href = '<url>'")`：
  darwin `frontend.go:215`、linux `frontend.go:352`、windows `frontend.go:298`。
- 触发时机：`options.App.OnDomReady`（`pkg/options/options.go:62`）在初始页
  `didFinishNavigation` 时触发 —— darwin `WailsContext.m:509` `didFinishNavigation` →
  `processMessage("DomReady")` → `darwin/frontend.go:400-401` 调 `OnDomReady`。
  OnStartup/OnDomReady 均在独立 goroutine 执行（`darwin/frontend.go:249-254` 起 goroutine）。
- `starturl`/`assetserverport`/`assetdir` 三个 context 值仅 **dev 构建**注入
  （`internal/app/app_dev.go:123,166`），生产 manual build 不可用；生产窗口恒加载默认
  startURL（darwin/linux `wails://wails/`，windows `http://wails.localhost/`）。

**GUI-B 窗口装配（GUI-2 落实）**：

```go
wails.Run(&options.App{
    Title: "...", Width: ..., Height: ...,
    AssetServer: &assetserver.Options{Handler: redirectHandler}, // 极小跳转页
    OnDomReady: func(ctx context.Context) {
        navigateOnce.Do(func() { runtime.WindowExecJS(ctx, "window.location.href='http://127.0.0.1:PORT/token?token=...'") })
    },
    OnShutdown: func(ctx context.Context) { server.Close(); engine.Close() },
})
```

**注意（GUI-2 陷阱）**：外部页 `didFinishNavigation` 会再次触发 `processMessage("DomReady")`，
即 **OnDomReady 在导航后二次触发** —— 导航动作必须「仅执行一次」（上例 `navigateOnce`）。
否则会陷入「导航 → DomReady → 再导航」循环。

### 2.3 localhost（`http://127.0.0.1:<port>`）三平台加载可行性 —— ✅ 全平台可行

- **macOS WKWebView**：wails 生成的 Info.plist 已含 `NSAppTransportSecurity →
  NSAllowsLocalNetworking = true`（`pkg/buildassets/build/darwin/Info.plist` 与
  `Info.dev.plist`），ATS 放行本地网络/回环 http；且 ATS 对 loopback/IP 地址连接本就豁免。
  无导航拦截 delegate（darwin 侧无 `decidePolicyForNavigationAction`），WKWebView 标准导航
  自由。`http://127.0.0.1:<port>` 可正常加载。
- **Windows WebView2**：startURL 本身就是 http（`windows/frontend.go:40`
  `http://wails.localhost/`），无 ATS 类限制；WebResourceRequested 过滤器对**非 startURL
  scheme/host 的请求显式放行默认处理**（`windows/frontend.go:656-661`
  "Let the WebView2 handle the request with its default handler"）——`http://127.0.0.1:<port>`
  走真实网络栈，不受 wails 拦截。
- **Linux WebKitGTK**：结构同 darwin（`linux/frontend.go:138`），无 http/localhost 限制。

### 2.4 与 issue #4686 的关系（非障碍）

- https://github.com/wailsapp/wails/issues/4686 "Opening external urls inside wails"
  （Enhancement，已 closed）：外部 URL 页面**拿不到 Wails runtime/IPC**（runtime.js/ipc.js
  仅注入 asset-server 服务的页面，`pkg/assetserver/assetserver.go:126-141,200-206`）。
- 与本项目**非 Bind 模式**设计一致：页面（webui 资产）独立走 HTTP/WS 与后端通信，不依赖
  wails runtime。该限制不构成障碍；反之是 GUI-B 直接可行的依据。

**裁决：GUI-B 可行** —— 窗口经「minimal AssetServer.Handler + OnDomReady +
`runtime.WindowExecJS`」导航到外部 `http://127.0.0.1:<port>/token?token=...`，
认证链路零改动（`ServerOptions{Auth: true}`）。GUI-A（AssetServer.Handler = webui mux +
`Auth:false`）保留为回退，S5-2 的 `ServerOptions.Auth` 已就位。

---

## 3. `wails.json` 在 manual build 下的作用（spike 清单 3）—— ✅ 不需要

- `wails.json` 仅 wails **CLI** 读取（`internal/project/project.go:272` `Load()`）；运行库
  （`pkg/application`、`internal/app`、`internal/frontend`）从不读该文件。
- v2 CLI build 无 `-X` ldflags 版本注入（grep 无匹配）；`internal/goversion/min.go` 仅编译期
  最小 Go 版本校验（1.20），不依赖 wails.json。
- **结论**：manual build（`go build -tags desktop,production`）不需要 wails.json，缺失不影响
  `wails.Run`。GUI-3 若启用 wails CLI 双路径（`wails build`）再补 wails.json。
- **build tag 硬约束**：`internal/app/app_default_unix.go`
  （`!dev && !production && !bindings`）运行时报
  "Wails applications will not build without the correct build tags" —— manual build 必须带
  `production` tag（`app_production.go` 为 `//go:build production`）；`desktop` 是 CLI 惯例伴生
  tag（`pkg/commands/build/base.go:218`）。roadmap 的 `go build -tags desktop,production` 正确。
- 运行前提：`wails.Run` 需主 goroutine（macOS AppKit 主线程模型）；`guiFrontend.Run` 由
  `runPlayer` 在主 goroutine 调用，满足。`wails.Run` 阻塞调用方直到窗口退出。

---

## 4. 与既有前端共存（spike 清单 4）—— ✅ 零代码满足

- 聚合器：`internal/frontend/registration/registration.go` 空导入 `gui`；实机验证
  `frontend.Registered() = [wails headless tui webui]`（gui 仅 import frontend，无环）。
- 选择链：`resolveFrontendID()`（`internal/commands/netease.go:64-80`）`--frontend=wails`
  命中注册表 → `frontend.ByID` → `fe.Run`。
- `--once` fail-fast：`internal/commands/netease.go:45-47`
  `if GlobalOptions.Once != "" && id != "headless" { return errors.New("--once 仅支持 headless 前端") }`。
  实机验证：`musicfox --frontend=wails --once next netease` → `ERROR: --once 仅支持 headless 前端`，
  退出码 2。**GUI 天然满足，无需任何代码**。
- 占位 Run：`musicfox --frontend=wails netease` → `ERROR: wails frontend not wired yet`
  （GUI-2 实现前即达此分支，链路完整）。

---

## 5. spike 判定点总结（对照 §3.4）

| 判定点 | 结论 | 依据 |
|--------|------|------|
| Wails v2.15 + Go 1.26 可编译 | ✅ | §1；`go build -tags wails_spike` 通过 |
| `wails.Run` 在 `Assets:nil` 时可运行 | ❌ 不可行（启动期 log.Fatal） | §2.1；GUI-B 需 minimal AssetServer.Handler |
| 主窗口 Navigate 外部 http URL 的 API 形状 | 无公开 `Navigate`；用 `OnDomReady` + `runtime.WindowExecJS`（官方同款 `window.location.href`） | §2.2 |
| 窗口内加载 `http://127.0.0.1:<port>` | ✅ 三平台可行（macOS ATS 已放行/Windows 显式放行/Linux 无限制） | §2.3 |
| manual build 不依赖 wails CLI 生成物 | ✅ wails.json 非必需；必须带 `-tags desktop,production` | §3 |
| `--frontend=wails` 与 `--once` fail-fast | ✅ 天然满足（netease.go:45-47） | §4 |

**D-GUI-2 裁决：GUI-B 可行（保持优先路径）**；GUI-A 仍为回退方案。
**D-GUI-1 裁决：方式 B 成立**，`internal/frontend/gui/` 实现 Frontend 接口，manual build 可行。

## 6. 对 GUI-2 的落地影响清单

1. 删除 `internal/frontend/gui/wails_spike.go`（run.go 开始 import wails 后）。
2. `guiFrontend.Run` 装配：engine + wasm scope + `webui.NewServerWithOptions(localBackend,
   `ServerOptions{Auth:true})` → 起 server（`127.0.0.1:<port>`）→ `wails.Run`（minimal
   AssetServer.Handler + OnDomReady 导航 + OnShutdown 清理）。
3. OnDomReady 导航**仅执行一次**（外部页会二次触发 DomReady）。
4. Linux 无桌面环境回退：`desktopAvailable()`（DISPLAY/WAYLAND_DISPLAY）为 false 时转调
   `webui.RunWithOptions(ctx, LaunchOptions{Mode: ModeStandalone})`（复用系统浏览器路径）。
5. 窗口关闭与 server/engine 生命周期竞态：参考 webui standalone 的 `ShutdownCh` select 模式。
6. 构建：`go build -tags desktop,production`；GUI-3 落 Makefile/goreleaser 时补 wails.json
   （CLI 双路径）。
