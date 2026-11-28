# herdr 封面兼容

musicfox 通过 `HERDR_PANE_ID` 识别 herdr pane。herdr 0.9.0 内置的 Kitty 图片执行器尚未实现原生动画，因此 `spin = true` 也会自动降级为 320px 静态封面。

静态图片使用 768 KiB 单包、1 MiB/s 持续速率和 1 MiB 突发预算；写入失败或耗时超过 200ms 后进入 5–30 秒冷却。冷却在同步写入返回后才生效，不能中断已经阻塞的终端写入。

未嵌套 tmux 时，herdr 路径输出裸 Kitty 序列，不需要开启 `tmuxPassthrough`。同时检测到 tmux 时，仍遵循 tmux 的透传开关与 DCS 包装规则。

这些措施处理 #671 中确认的动画兼容性与图片负载问题；macOS 内核锁循环的触发根因仍未确定，不能据此保证整机崩溃已修复。可设置 `[main.lyric.cover]` 下的 `show = false` 关闭封面。
