---
status: root_cause_found
trigger: 当前项目当配置中的frameRate设置比较高时(>=30)，cpu占用很高，检查定位下看有什么方案可以优化的，先不要修改代码，仅做计划
created: 2026-07-18
updated: 2026-07-18
---

# Debug Session: framerate-high-cpu

## Symptoms

- **Expected behavior**: frameRate 设置较高时（>=30），CPU 占用应保持在合理水平
- **Actual behavior**: frameRate >= 30 时 CPU 占用飙升，frameRate=10 时正常
- **Error messages**: 无报错，纯性能问题
- **Timeline**: 播放歌曲时出现，与频谱（visualizer）无关，纯 UI 渲染问题
- **Reproduction**: 将 config 中 frameRate 设置为 >=30，播放任意歌曲，观察 CPU 占用
- **Goal**: diagnose only — 定位根因并输出优化方案计划，不修改代码

## Current Focus

- **hypothesis**: 双重渲染 + 全量重绘 = 高 CPU
- **test**: 代码分析确认 — 每次 TimeChan tick 触发 2 次 View() 调用
- **expecting**: 根因定位完成
- **next_action**: 撰写完整根因报告与优化方案

## Root Cause Analysis

### 根因总览

frameRate>=30 时高 CPU 占用由**三个叠加因素**导致：

1. **双重渲染 (Double Render)**：每次 TimeChan tick 触发 **2 次**完整 View() 渲染
2. **全量重绘 (No Dirty Checking)**：每帧重建所有 UI 组件，无缓存/记忆化
3. **frameRate 线性放大**：渲染频率随 frameRate 线性增长，30fps 时实际渲染 ~60 次/秒

### 1. 双重渲染问题（ROOT CAUSE #1）

**文件：`internal/ui/player.go` 第 149-175 行**

```go
// 时间监听 goroutine
case duration := <-p.TimeChan():
    p.lyricService.UpdatePosition(duration)

    // 触发点 1：向 renderTicker 发信号（非阻塞）
    if p.renderTicker != nil {
        select {
        case p.renderTicker.c <- time.Now():
        default:
        }
    }

    // 触发点 2：直接调用 Rerender
    p.netease.Rerender(false)   // <--- 这一行是冗余的！
```

**文件：`vendor/github.com/anhoder/foxful-cli/model/app.go` 第 54-58 行**

```go
if a.options.Ticker != nil {
    go func() {
        for range a.options.Ticker.Ticker() {  // 监听 renderTicker.c
            a.Rerender(false)                   // 触发完整 View() 周期
        }
    }()
}
```

**流程追踪**：
```
TimeChan tick (每 33ms @30fps)
  │
  ├─► renderTicker.c <- time.Now()
  │     └─► App goroutine: a.Rerender(false)
  │           └─► program.Send(RerenderCmd)
  │                 └─► bubbletea: View()  ← 第 1 次渲染
  │
  └─► p.netease.Rerender(false)  直接调用
        └─► program.Send(RerenderCmd)
              └─► bubbletea: View()  ← 第 2 次渲染（冗余！）
```

**影响**：每次 TimeChan tick 产生 **~2x** 的渲染。以 frameRate=30 为例，实际渲染约 60 次/秒。

### 2. 全量重绘——每帧无条件渲染所有组件（ROOT CAUSE #2）

**文件：`vendor/github.com/anhoder/foxful-cli/model/main.go` 第 242-295 行**

`Main.View()` 方法每帧执行以下操作，无任何 dirty checking：

```
Main.View() 每帧执行：
├─ TitleView()            — 标题重建（轻量）
├─ MenuTitleView()        — 菜单标题重建（轻量）
├─ menuListView()         — 20+ 菜单项：居中 + 截断 + 填充 + ANSI 样式
├─ searchInputView()      — 搜索框（通常空字符串）
└─ for each component:    — 无条件调用所有组件 View()
    ├─ SpectrumRenderer.View()     — 频谱渲染（通常 disabled）
    ├─ LyricRenderer.View()        — 歌词处理 + ANSI 颜色 + 滚动 tick
    ├─ SongInfoRenderer.View()     — 歌曲信息构建 + like list 查询
    └─ ProgressRenderer.View()     — 进度条 + 颜色渐变计算
```

**各组件每帧开销分析**：

| 组件 | 每帧操作 | 开销 |
|------|---------|------|
| **menuListView** | 20+ 菜单项：runewidth.Truncate/FillRight + ANSI 样式 + 居中计算 | **中** |
| **LyricRenderer** | 获取歌词状态 → 构建 YRC/LRC → ANSI 颜色处理 → 截断/填充 → ScrollBar.Tick | **高** |
| **SongInfoRenderer** | 构建分段 → likelist 查询 → truncate → ANSI 样式 → 宽度计算 | **中** |
| **ProgressRenderer** | 进度计算 → 颜色渐变生成（条件缓存）→ 时间格式化 | **中** |
| **CoverRenderer** | 大多数帧 short-circuit 返回空（已有良好缓存） | **低** |
| **SpectrumRenderer** | 默认 disabled，直接返回 | **低** |

### 3. frameRate 与渲染频率的线性关系

| frameRate | Player tick 间隔 | 实际渲染/秒 (双重) | CPU 表现 |
|-----------|-----------------|-------------------|---------|
| 5 (默认)  | 200ms           | ~10               | 极低    |
| 10        | 100ms           | ~20               | 正常    |
| 30        | 33ms            | ~60               | **高**  |
| 60        | 16ms            | ~120              | **极高** |

### 4. 额外的渲染触发源

除了 TimeChan 循环，还有 `StateChan` 监听器也会触发渲染：

**文件：`internal/ui/player.go` 第 131-146 行**

```go
case s := <-p.StateChan():
    p.stateHandler.SetPlayingInfo(p.PlayingInfo())
    p.updateDesktopLyrics()
    if s != types.Stopped {
        p.netease.Rerender(false)   // 状态变更时额外触发
        break
    }
    p.NextSong(false)
```

## Evidence

- **timestamp**: 2026-07-18 — Code analysis
- **finding**: `internal/ui/player.go:166-172` — 双重触发：一次通过 renderTicker 通道，一次直接调用 Rerender
- **finding**: `vendor/.../foxful-cli/model/app.go:54-58` — App goroutine 监听 renderTicker 并触发 Rerender
- **finding**: `vendor/.../foxful-cli/model/main.go:242-295` — Main.View() 无条件重建所有 UI 组件
- **finding**: `internal/configs/framerate.go:22` — `Interval()` 返回 `1000/frameRate ms`，例如 30fps = 33ms
- **finding**: `internal/player/beep_player.go:212` — beep player 使用 `FrameRate.Interval()` 作为 ticker 间隔
- **finding**: `internal/player/mpv_player.go:297` — mpv player 同样使用 frameRate 驱动 ticker
- **finding**: `internal/ui/song_info_renderer.go:79` — 每帧检查 `likelist.IsLikeSong()`，可能涉及磁盘/网络 I/O
- **finding**: `internal/ui/progress_renderer.go:92-94` — 窗口宽度变化时重建整个颜色渐变数组
- **finding**: `internal/ui/lyric_renderer.go:224-246` — prepareLyricLines() 每帧重新构建 lyrics 数组

## Eliminated

- **频谱 (visualizer)**：已确认用户场景中 visualizer 未启用，排除
- **桌面歌词**：仅在有桌面歌词时才触发更新，排除
- **CoverRenderer**：已有良好缓存机制，大多数帧 short-circuit 返回，排除

## Resolution

### root_cause
frameRate>=30 时 CPU 高的三个叠加根因：
1. **双重渲染（P0）**：`player.go:172` 的直接 `Rerender(false)` 调用与 App goroutine 的 ticker 监听形成双重渲染，使实际渲染次数翻倍
2. **全量重绘无缓存（P1）**：`Main.View()` 每帧无条件重建所有 UI 组件（菜单列表、歌词、歌曲信息、进度条），缺乏 dirty checking 和输出缓存
3. **frameRate 线性放大（P1）**：渲染工作量随 frameRate 线性增长，而大部分帧之间没有足够的视觉差异值得重新渲染

### fix
（仅计划，未实施）

### optimization_plan

#### 阶段一：立即修复（高收益、低风险）

**1.1 消除双重渲染**
- **位置**：`internal/ui/player.go` 第 172 行
- **动作**：删除 `p.netease.Rerender(false)` 直接调用
- **原因**：App goroutine (`app.go:54-58`) 已通过 renderTicker 机制触发渲染，直接调用是冗余的
- **预期收益**：渲染次数减半，CPU 降低 ~40-50%

**1.2 条件渲染——时间/进度跳过检查**
- **位置**：`internal/ui/player.go` 第 149-175 行，TimeChan 监听器
- **动作**：跟踪上次渲染时的播放时间，仅当进度变化超过阈值时才发送 ticker 信号
- **伪代码**：
```go
var lastRenderedMs int64
case duration := <-p.TimeChan():
    currentMs := duration.Milliseconds()
    if abs(currentMs - lastRenderedMs) < 50 { // 进度未明显变化
        return  // 跳过本次渲染
    }
    lastRenderedMs = currentMs
    // ... 发送 ticker 信号
```
- **预期收益**：进一步减少无效渲染 ~10-20%

#### 阶段二：短期优化（中等收益）

**2.1 歌词渲染缓存**
- **位置**：`internal/ui/lyric_renderer.go` `View()` 方法
- **动作**：缓存上次渲染的歌词输出（字符串），仅当 `currentTimeMs`、歌词状态、窗口尺寸任一变化时才重新计算
- **预期收益**：减少每帧字符串分配和 ANSI 样式处理

**2.2 歌曲信息渲染缓存**
- **位置**：`internal/ui/song_info_renderer.go` `View()` 方法
- **动作**：缓存上次构建的输出，仅当 songId 或播放状态变化时重建
- **注意**：`likelist.IsLikeSong()` 查询可移出高频渲染路径

**2.3 进度条渲染缓存**
- **位置**：`internal/ui/progress_renderer.go` `View()` 方法
- **动作**：缓存颜色渐变数组，仅当窗口宽度变化时重建（已部分实现）

**2.4 菜单列表条件渲染**
- **位置**：`vendor/.../foxful-cli/model/main.go` `menuListView()`
- **动作**：添加 dirty flag，仅当选中项、菜单数据或窗口尺寸变化时重建菜单视图
- **注意**：需要 fork foxful-cli 或提交上游 PR

#### 阶段三：架构级优化（长期）

**3.1 自适应 frameRate**
- **位置**：`internal/configs/framerate.go`
- **动作**：在播放暂停、无动画时自动降低 frameRate 到 1-2 FPS；有歌词动画或频谱时恢复配置值
- **预期收益**：暂停期间 CPU 降至几乎为 0

**3.2 增量渲染**
- **动作**：仅重绘发生变化的行，而非整个终端视图
- **复杂度**：需要框架级改造，但收益巨大

**3.3 内存池化字符串构建**
- **动作**：使用 `sync.Pool` 复用 `strings.Builder` 实例，减少 GC 压力
- **收益**：减少 GC 暂停和内存分配

### specialist_hint
general

## Specialist Review

（待执行 specialist dispatch）

---

*This is a diagnose-only session. No code modifications were made.*
