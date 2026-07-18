---
status: root_cause_found
trigger: "当前桌面歌词无法进行拖拽，请遵循第一性原则进行排查定位"
created: 2026-07-18T00:00:00Z
updated: 2026-07-18T00:00:00Z
---

# Symptoms

- **Expected behavior**: 按住鼠标拖拽移动桌面歌词窗口位置，松手后窗口停留在新位置
- **Actual behavior**: 完全不能拖拽，鼠标按住桌面歌词窗口后无论如何拖拽都无法移动
- **Error messages**: 无任何错误日志
- **Timeline**: 不确定什么时候开始的
- **Reproduction**: 启动应用，显示桌面歌词，尝试用鼠标拖拽窗口
- **Platform**: macOS

# Current Focus

- hypothesis: "NSTextField labels inside bgView intercept mouse events — NSControl subclass consumes mouseDown without forwarding to nextResponder"
- test: "Verified via code analysis: labels have no custom mouse override, bgView and containerView have mouse handlers but never receive events due to AppKit hit-test routing to deepest subview"
- expecting: "mouseDown reaches NSTextField label first, event consumed, never reaches LyricsBackgroundView → drag never starts"
- next_action: "root_cause_found"
- reasoning_checkpoint: "done"
- tdd_checkpoint: "n/a"

# Architecture Context

The desktop lyrics drag system on macOS uses a custom NSView subclass approach:

**Files involved:**
- `internal/desktop_lyrics/helper_darwin.go` (220 lines) — ObjC class registration, mouse event dispatch
- `internal/desktop_lyrics/desktop_lyrics_darwin.go` (1997 lines) — Main controller, window creation, drag logic
- `internal/configs/main.go` (line 197) — `Draggable bool` config field

**Drag flow:**
1. Two NSView subclasses registered: `LyricsDragView` (containerView) and `LyricsBackgroundView` (bgView)
2. Both override `mouseDown:`, `mouseDragged:`, `mouseUp:` → Go handlers via purego/objc
3. `handleDragViewMouseDown` → `ctrl.handleDragStart()` — records screen mouse position, sets dragActive
4. `handleDragViewMouseDragged` → `ctrl.handleDragMove()` — computes delta, updates xFactor/yFactor [0-1], calls `layoutContent(false)`
5. `handleDragViewMouseUp` → `ctrl.handleDragEnd()` → `persistPositionFactors()`

**View hierarchy:**
```
Window.contentView (borderless, full-size-content)
  └─ containerView (LyricsDragView) — fills entire window
       └─ bgView (LyricsBackgroundView) — fills entire container
            ├─ labels[0] (NSTextField) — positioned top
            ├─ labels[1] (NSTextField) — positioned bottom  
            └─ spectrum CALayer sublayers
```

**Critical gates:**
1. `c.window.SetIgnoresMouseEvents(!c.cfg.Draggable)` at line 297 — if Draggable=false, window receives NO mouse events at all
2. `handleDragStart` checks `if !c.cfg.Draggable || c.closed { return }` at line 835
3. Default config: `draggable = true` in `utils/filex/embed/config.toml:165`
4. No logging in drag handlers — silent failures

**Hit testing concern:** NSTextField labels are subviews of bgView. AppKit hit testing delivers mouseDown to the deepest subview containing the point. Labels have NO mouseDown override → they receive mouseDown but do nothing → event is consumed by label, NOT forwarded to bgView. This means clicking ON a label area cannot start a drag. Clicking on background area SHOULD work.

# Evidence

- timestamp: 2026-07-18T00:00:00Z
  type: code-analysis
  summary: |
    
    **Git history trace of drag system evolution:**
    
    1. `ec98d115` (original): Used `SetMovableByWindowBackground(true)` — window-level drag, worked correctly.
    2. `4111f4f0` ("feat: optimize"): Introduced `LyricsDragView` custom subclass with mouseDown/mouseDragged/mouseUp handlers. Replaced `SetMovableByWindowBackground(true)` with `c.window.SetIgnoresMouseEvents(!c.cfg.Draggable)`. This was the REGRESSION COMMIT — custom view-level drag replaced proven window-level drag.
    3. `1c27c6a6` ("feat: add visualizer"): Added spectrum bars. Drag system unchanged from step 2.
    4. `3d6620a4` ("fix: render error"): Added `sel_hitTest` selector registration (unused!), made `handleWindowWillMove`/`handleWindowDidMove` no-ops.
    5. `72d0b0e0` ("fix: panic and render err", HEAD, with UNCOMMITTED changes): Added `LyricsBackgroundView` (bgViewClass) to prevent bgView from intercepting events. BUT the NSTextField labels inside bgView still intercept.

  files:
    - internal/desktop_lyrics/helper_darwin.go
    - internal/desktop_lyrics/desktop_lyrics_darwin.go

- timestamp: 2026-07-18T00:00:00Z
  type: code-analysis
  summary: |
    
    **Event flow verification (AppKit hit testing):**
    
    - Window.contentView → containerView (LyricsDragView, has mouse* handlers) → bgView (LyricsBackgroundView, has mouse* handlers) → labels[0..1] (NSTextField, NO mouse* handlers)
    - AppKit hit testing sends mouseDown to DEEPEST subview: NSTextField label
    - NSTextField is NSControl subclass → default `mouseDown:` processes event internally without forwarding to nextResponder
    - Even when set to `Editable(false)` and `Selectable(false)`, NSTextField still receives and processes mouse events
    - Result: mouseDown/mouseDragged never reach LyricsBackgroundView or LyricsDragView → drag never starts
    
    **View coverage analysis:**
    - `needLargeHeight=true` initially: label height = 64px, gap = 8px, padding = 16px
    - Labels cover ~84% of content area height in two-line mode
    - Only the ~8px gap and 16px padding edges are clickable for drag
    - From user perspective: "completely cannot drag"

  files:
    - internal/desktop_lyrics/desktop_lyrics_darwin.go:426-443

- timestamp: 2026-07-18T00:00:00Z
  type: code-analysis
  summary: |
    
    **sel_hitTest registered but unused — planned fix never implemented:**
    
    In `helper_darwin.go:37`: `sel_hitTest = objc.RegisterName("hitTest:")`
    This selector is registered but NEVER used as a method override in any objc.RegisterClass call.
    
    The developer clearly recognized the need for a `hitTest:` override (to return nil for labels and pass-through events to bgView), registered the selector, but never implemented the method. This was likely the intended fix path.

  files:
    - internal/desktop_lyrics/helper_darwin.go:37

# Eliminated

- Checked: `Draggable` config defaults to `true` (config.toml:189) — NOT the cause
- Checked: `SetIgnoresMouseEvents(!c.cfg.Draggable)` at line 297 — when Draggable=true, window accepts mouse events — NOT the cause
- Checked: `handleDragStart` gate `if !c.cfg.Draggable || c.closed` — when Draggable=true, gate passes — NOT the cause

# Resolution

- root_cause: "NSTextField labels (NSControl subclass) intercept mouse events inside bgView. AppKit hit-testing sends mouseDown to the deepest subview (label), which consumes the event via default NSControl::mouseDown: without forwarding to the nextResponder. The LyricsBackgroundView/LyricsDragView drag handlers never receive mouseDown/mouseDragged → drag never starts. This was introduced in commit 4111f4f0 when SetMovableByWindowBackground(true) was replaced with custom view-level drag handling. A hitTest: override was planned (sel_hitTest registered) but never implemented."
- fix: "See specialist review for fix direction recommendation."
- verification: "After fix: mouseDown on any area of the desktop lyrics window (including over text) should start drag via LyricsBackgroundView/LyricsDragView mouse handlers"
- files_changed: []
