# tmux 封面显示与兼容设置

封面渲染是可选功能。普通终端仍使用原有 Kitty 绝对放置；tmux 内使用 Unicode 占位符，让封面随窗格文字一起移动。不会轮询窗格位置，也不修改播放器、歌曲下载或歌词解析。

## 配置

以下为相关设置的默认关闭状态：

```toml
[main.lyric.cover]
show = false
tmuxPassthrough = false
tmuxSyncOutput = false
```

还需要支持 Kitty 图形协议的外层终端，以及 tmux ≥3.3 的 `set -g allow-passthrough on`。能力检查失败时不启用图像透传。`tmuxPassthrough` 默认关闭；开启仍属实验性功能，tmux 内强制使用静态封面。

`tmuxPassthrough` 与 `tmuxSyncOutput` 均默认关闭，旧配置缺少这些键时也保持关闭。需要 tmux 封面时，手动将 `show` 与 `tmuxPassthrough` 设为 `true`；若右侧窗格出现多个格子、重复顶部条带，或切歌提示期间失真，再将 `tmuxSyncOutput` 设为 `true` 并重启 musicfox。它只在上述封面功能开启时生效，给含封面占位符的最终 TUI 写入添加标准 `CSI ?2026h/l` 同步刷新。通知、弹窗和行差分之后才包装，因此切歌提示期间也能保持图片行列坐标。普通文字写入原样通过。

tmux 3.7c 在右侧窗格增量输出组合字符时遗漏 pane 的横向偏移，造成封面每行重复显示图像顶部；同步刷新避开这条增量路径。上游修复偏移计算后仍可使用同步刷新，但可能增加重绘。确认所用 tmux 已修复后可将 `tmuxSyncOutput` 设为 `false` 并重启 musicfox。不依据版本号自动关闭，避免下游版本与回移补丁不一致。

要完全关闭 tmux 图像传输，设 `tmuxPassthrough = false`；要在所有终端关闭封面，设 `show = false`。

## 隔离与限流

- `internal/ui/kitty/` 负责协议生成、终端检测与最终输出包装。
- `internal/ui/cover_renderer_tmux.go` 负责静态图像传输、虚拟放置和占位符接口。
- `internal/ui/cover_renderer_limiter.go` 负责单包、速率预算和拥塞退避。
- 歌词渲染器只嵌入指定位置的占位符，并用封面标识与尺寸使缓存失效。

tmux 静态源图缩放至 320px；图像 DCS 单包上限 768 KiB，持续预算 1 MiB/s，突发预算 1 MiB。失败或超过 200ms 的写入触发 5–30 秒冷却。这些措施由 #660 引入，用于降低图像传输负载。写耗时用于判断传输拥塞，不代表真实 GPU 指标。调试日志仅按事件记录，无周期性 RSS 查询；macOS 的 RSS 子进程查询有超时。

## 回归验证

普通单测不需要启动 tmux。Unix 上安装 tmux 与 Python 3 后运行：

```sh
MUSICFOX_TMUX_INTEGRATION=1 go test ./internal/ui/kitty -run TestUnicodePlaceholderTmuxOutput -count=1 -v
```

测试启动独立 tmux 服务，读取实际发往终端的 PTY 字节，逐帧验证左右窗格、通知合成、单行刷新和交换窗格。测试不向终端传输图像，不替代 Ghostty/Kitty 的最终视觉与 GPU 检验。

测试已修复的 tmux 时，可通过 `PATH` 选择另一份二进制，并添加 `MUSICFOX_TMUX_TEST_NO_SYNC=1` 验证关闭兼容处理；该变量只影响测试，不是运行时配置。原版 tmux 3.7c 在此模式下应复现失败。
