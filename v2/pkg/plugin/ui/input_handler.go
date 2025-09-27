package ui

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"
)

// MultiInputHandler 多输入类型处理器
type MultiInputHandler struct {
	handlers map[InputType]InputHandler
	mutex    sync.RWMutex
	logger   *slog.Logger

	// 输入过滤器
	filters []InputFilter

	// 快捷键映射
	shortcuts map[string]ShortcutAction

	// 输入历史
	history []InputEvent
	maxHistory int
}

// InputFilter 输入过滤器接口
type InputFilter interface {
	Filter(event *InputEvent) bool
	GetPriority() int
}

// ShortcutAction 快捷键动作接口
type ShortcutAction interface {
	Execute(ctx context.Context, event *InputEvent) error
	GetDescription() string
}

// NewMultiInputHandler 创建多输入类型处理器
func NewMultiInputHandler(logger *slog.Logger) *MultiInputHandler {
	handler := &MultiInputHandler{
		handlers:   make(map[InputType]InputHandler),
		logger:     logger,
		filters:    make([]InputFilter, 0),
		shortcuts:  make(map[string]ShortcutAction),
		history:    make([]InputEvent, 0),
		maxHistory: 100,
	}

	// 注册默认处理器
	handler.registerDefaultHandlers()

	// 注册默认快捷键
	handler.registerDefaultShortcuts()

	return handler
}

// registerDefaultHandlers 注册默认处理器
func (h *MultiInputHandler) registerDefaultHandlers() {
	h.handlers[InputTypeKeyboard] = NewKeyboardHandler(h.logger)
	h.handlers[InputTypeMouse] = NewMouseHandler(h.logger)
	h.handlers[InputTypeTouch] = NewTouchHandler(h.logger)
	h.handlers[InputTypeGamepad] = NewGamepadHandler(h.logger)
	h.handlers[InputTypeVoice] = NewVoiceHandler(h.logger)
}

// registerDefaultShortcuts 注册默认快捷键
func (h *MultiInputHandler) registerDefaultShortcuts() {
	h.shortcuts["ctrl+p"] = &PlayPauseAction{}
	h.shortcuts["ctrl+n"] = &NextTrackAction{}
	h.shortcuts["ctrl+b"] = &PrevTrackAction{}
	h.shortcuts["ctrl+m"] = &MuteAction{}
	h.shortcuts["ctrl+q"] = &QuitAction{}
	h.shortcuts["space"] = &PlayPauseAction{}
	h.shortcuts["left"] = &SeekBackwardAction{}
	h.shortcuts["right"] = &SeekForwardAction{}
	h.shortcuts["up"] = &VolumeUpAction{}
	h.shortcuts["down"] = &VolumeDownAction{}
}

// Handle 处理输入事件
func (h *MultiInputHandler) Handle(ctx context.Context, event *InputEvent) error {
	if event == nil {
		return fmt.Errorf("event cannot be nil")
	}

	// 应用输入过滤器
	if !h.applyFilters(event) {
		h.logger.Debug("Input event filtered out", "type", event.Type, "key", event.Key)
		return nil
	}

	// 记录输入历史
	h.addToHistory(*event)

	// 检查快捷键
	if h.handleShortcut(ctx, event) {
		return nil
	}

	// 根据输入类型选择处理器
	h.mutex.RLock()
	handler, exists := h.handlers[event.Type]
	h.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("no handler found for input type: %d", event.Type)
	}

	return handler.Handle(ctx, event)
}

// applyFilters 应用输入过滤器
func (h *MultiInputHandler) applyFilters(event *InputEvent) bool {
	for _, filter := range h.filters {
		if !filter.Filter(event) {
			return false
		}
	}
	return true
}

// handleShortcut 处理快捷键
func (h *MultiInputHandler) handleShortcut(ctx context.Context, event *InputEvent) bool {
	if event.Type != InputTypeKeyboard {
		return false
	}

	// 构建快捷键字符串
	shortcutKey := h.buildShortcutKey(event)
	if shortcutKey == "" {
		return false
	}

	// 查找快捷键动作
	action, exists := h.shortcuts[shortcutKey]
	if !exists {
		return false
	}

	// 执行快捷键动作
	if err := action.Execute(ctx, event); err != nil {
		h.logger.Error("Failed to execute shortcut action", "shortcut", shortcutKey, "error", err)
		return false
	}

	h.logger.Debug("Shortcut executed", "shortcut", shortcutKey)
	return true
}

// buildShortcutKey 构建快捷键字符串
func (h *MultiInputHandler) buildShortcutKey(event *InputEvent) string {
	if event.Key == "" {
		return ""
	}

	parts := make([]string, 0)

	// 添加修饰键
	for _, modifier := range event.Modifiers {
		parts = append(parts, strings.ToLower(modifier))
	}

	// 添加主键
	parts = append(parts, strings.ToLower(event.Key))

	return strings.Join(parts, "+")
}

// addToHistory 添加到输入历史
func (h *MultiInputHandler) addToHistory(event InputEvent) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	// 添加时间戳
	event.Timestamp = time.Now()

	// 添加到历史
	h.history = append(h.history, event)

	// 限制历史长度
	if len(h.history) > h.maxHistory {
		h.history = h.history[1:]
	}
}

// SupportsType 检查是否支持指定输入类型
func (h *MultiInputHandler) SupportsType(inputType InputType) bool {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	_, exists := h.handlers[inputType]
	return exists
}

// GetSupportedTypes 获取支持的输入类型
func (h *MultiInputHandler) GetSupportedTypes() []InputType {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	types := make([]InputType, 0, len(h.handlers))
	for inputType := range h.handlers {
		types = append(types, inputType)
	}

	return types
}

// RegisterHandler 注册输入处理器
func (h *MultiInputHandler) RegisterHandler(inputType InputType, handler InputHandler) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.handlers[inputType] = handler
	h.logger.Info("Input handler registered", "type", inputType)
}

// RegisterFilter 注册输入过滤器
func (h *MultiInputHandler) RegisterFilter(filter InputFilter) {
	h.filters = append(h.filters, filter)
	h.logger.Info("Input filter registered")
}

// RegisterShortcut 注册快捷键
func (h *MultiInputHandler) RegisterShortcut(key string, action ShortcutAction) {
	h.shortcuts[strings.ToLower(key)] = action
	h.logger.Info("Shortcut registered", "key", key, "description", action.GetDescription())
}

// GetInputHistory 获取输入历史
func (h *MultiInputHandler) GetInputHistory() []InputEvent {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	// 返回历史副本
	history := make([]InputEvent, len(h.history))
	copy(history, h.history)
	return history
}

// KeyboardHandler 键盘输入处理器
type KeyboardHandler struct {
	logger *slog.Logger
}

// NewKeyboardHandler 创建键盘输入处理器
func NewKeyboardHandler(logger *slog.Logger) *KeyboardHandler {
	return &KeyboardHandler{
		logger: logger,
	}
}

// Handle 处理键盘输入
func (h *KeyboardHandler) Handle(ctx context.Context, event *InputEvent) error {
	h.logger.Debug("Keyboard input", "key", event.Key, "modifiers", event.Modifiers)

	// 处理特殊键
	switch strings.ToLower(event.Key) {
	case "enter":
		return h.handleEnter(ctx, event)
	case "escape":
		return h.handleEscape(ctx, event)
	case "tab":
		return h.handleTab(ctx, event)
	case "backspace":
		return h.handleBackspace(ctx, event)
	default:
		return h.handleRegularKey(ctx, event)
	}
}

// handleEnter 处理回车键
func (h *KeyboardHandler) handleEnter(ctx context.Context, event *InputEvent) error {
	h.logger.Debug("Enter key pressed")
	// 触发确认事件
	return nil
}

// handleEscape 处理ESC键
func (h *KeyboardHandler) handleEscape(ctx context.Context, event *InputEvent) error {
	h.logger.Debug("Escape key pressed")
	// 触发取消事件
	return nil
}

// handleTab 处理Tab键
func (h *KeyboardHandler) handleTab(ctx context.Context, event *InputEvent) error {
	h.logger.Debug("Tab key pressed")
	// 触发焦点切换事件
	return nil
}

// handleBackspace 处理退格键
func (h *KeyboardHandler) handleBackspace(ctx context.Context, event *InputEvent) error {
	h.logger.Debug("Backspace key pressed")
	// 触发删除事件
	return nil
}

// handleRegularKey 处理普通按键
func (h *KeyboardHandler) handleRegularKey(ctx context.Context, event *InputEvent) error {
	h.logger.Debug("Regular key pressed", "key", event.Key)
	// 触发字符输入事件
	return nil
}

// SupportsType 检查是否支持指定输入类型
func (h *KeyboardHandler) SupportsType(inputType InputType) bool {
	return inputType == InputTypeKeyboard
}

// GetSupportedTypes 获取支持的输入类型
func (h *KeyboardHandler) GetSupportedTypes() []InputType {
	return []InputType{InputTypeKeyboard}
}

// MouseHandler 鼠标输入处理器
type MouseHandler struct {
	logger *slog.Logger
}

// NewMouseHandler 创建鼠标输入处理器
func NewMouseHandler(logger *slog.Logger) *MouseHandler {
	return &MouseHandler{
		logger: logger,
	}
}

// Handle 处理鼠标输入
func (h *MouseHandler) Handle(ctx context.Context, event *InputEvent) error {
	h.logger.Debug("Mouse input", "key", event.Key, "position", event.Position)

	switch strings.ToLower(event.Key) {
	case "left":
		return h.handleLeftClick(ctx, event)
	case "right":
		return h.handleRightClick(ctx, event)
	case "middle":
		return h.handleMiddleClick(ctx, event)
	case "wheel_up":
		return h.handleWheelUp(ctx, event)
	case "wheel_down":
		return h.handleWheelDown(ctx, event)
	default:
		return h.handleMouseMove(ctx, event)
	}
}

// handleLeftClick 处理左键点击
func (h *MouseHandler) handleLeftClick(ctx context.Context, event *InputEvent) error {
	h.logger.Debug("Left mouse button clicked", "position", event.Position)
	return nil
}

// handleRightClick 处理右键点击
func (h *MouseHandler) handleRightClick(ctx context.Context, event *InputEvent) error {
	h.logger.Debug("Right mouse button clicked", "position", event.Position)
	return nil
}

// handleMiddleClick 处理中键点击
func (h *MouseHandler) handleMiddleClick(ctx context.Context, event *InputEvent) error {
	h.logger.Debug("Middle mouse button clicked", "position", event.Position)
	return nil
}

// handleWheelUp 处理滚轮向上
func (h *MouseHandler) handleWheelUp(ctx context.Context, event *InputEvent) error {
	h.logger.Debug("Mouse wheel up", "position", event.Position)
	return nil
}

// handleWheelDown 处理滚轮向下
func (h *MouseHandler) handleWheelDown(ctx context.Context, event *InputEvent) error {
	h.logger.Debug("Mouse wheel down", "position", event.Position)
	return nil
}

// handleMouseMove 处理鼠标移动
func (h *MouseHandler) handleMouseMove(ctx context.Context, event *InputEvent) error {
	h.logger.Debug("Mouse moved", "position", event.Position)
	return nil
}

// SupportsType 检查是否支持指定输入类型
func (h *MouseHandler) SupportsType(inputType InputType) bool {
	return inputType == InputTypeMouse
}

// GetSupportedTypes 获取支持的输入类型
func (h *MouseHandler) GetSupportedTypes() []InputType {
	return []InputType{InputTypeMouse}
}

// TouchHandler 触摸输入处理器
type TouchHandler struct {
	logger *slog.Logger
}

// NewTouchHandler 创建触摸输入处理器
func NewTouchHandler(logger *slog.Logger) *TouchHandler {
	return &TouchHandler{
		logger: logger,
	}
}

// Handle 处理触摸输入
func (h *TouchHandler) Handle(ctx context.Context, event *InputEvent) error {
	h.logger.Debug("Touch input", "key", event.Key, "position", event.Position)
	return nil
}

// SupportsType 检查是否支持指定输入类型
func (h *TouchHandler) SupportsType(inputType InputType) bool {
	return inputType == InputTypeTouch
}

// GetSupportedTypes 获取支持的输入类型
func (h *TouchHandler) GetSupportedTypes() []InputType {
	return []InputType{InputTypeTouch}
}

// GamepadHandler 手柄输入处理器
type GamepadHandler struct {
	logger *slog.Logger
}

// NewGamepadHandler 创建手柄输入处理器
func NewGamepadHandler(logger *slog.Logger) *GamepadHandler {
	return &GamepadHandler{
		logger: logger,
	}
}

// Handle 处理手柄输入
func (h *GamepadHandler) Handle(ctx context.Context, event *InputEvent) error {
	h.logger.Debug("Gamepad input", "key", event.Key)
	return nil
}

// SupportsType 检查是否支持指定输入类型
func (h *GamepadHandler) SupportsType(inputType InputType) bool {
	return inputType == InputTypeGamepad
}

// GetSupportedTypes 获取支持的输入类型
func (h *GamepadHandler) GetSupportedTypes() []InputType {
	return []InputType{InputTypeGamepad}
}

// VoiceHandler 语音输入处理器
type VoiceHandler struct {
	logger *slog.Logger
}

// NewVoiceHandler 创建语音输入处理器
func NewVoiceHandler(logger *slog.Logger) *VoiceHandler {
	return &VoiceHandler{
		logger: logger,
	}
}

// Handle 处理语音输入
func (h *VoiceHandler) Handle(ctx context.Context, event *InputEvent) error {
	h.logger.Debug("Voice input", "key", event.Key)
	return nil
}

// SupportsType 检查是否支持指定输入类型
func (h *VoiceHandler) SupportsType(inputType InputType) bool {
	return inputType == InputTypeVoice
}

// GetSupportedTypes 获取支持的输入类型
func (h *VoiceHandler) GetSupportedTypes() []InputType {
	return []InputType{InputTypeVoice}
}

// 快捷键动作实现

// PlayPauseAction 播放/暂停动作
type PlayPauseAction struct{}

func (a *PlayPauseAction) Execute(ctx context.Context, event *InputEvent) error {
	// 实现播放/暂停逻辑
	return nil
}

func (a *PlayPauseAction) GetDescription() string {
	return "Play/Pause music"
}

// NextTrackAction 下一曲动作
type NextTrackAction struct{}

func (a *NextTrackAction) Execute(ctx context.Context, event *InputEvent) error {
	// 实现下一曲逻辑
	return nil
}

func (a *NextTrackAction) GetDescription() string {
	return "Next track"
}

// PrevTrackAction 上一曲动作
type PrevTrackAction struct{}

func (a *PrevTrackAction) Execute(ctx context.Context, event *InputEvent) error {
	// 实现上一曲逻辑
	return nil
}

func (a *PrevTrackAction) GetDescription() string {
	return "Previous track"
}

// MuteAction 静音动作
type MuteAction struct{}

func (a *MuteAction) Execute(ctx context.Context, event *InputEvent) error {
	// 实现静音逻辑
	return nil
}

func (a *MuteAction) GetDescription() string {
	return "Mute/Unmute"
}

// QuitAction 退出动作
type QuitAction struct{}

func (a *QuitAction) Execute(ctx context.Context, event *InputEvent) error {
	// 实现退出逻辑
	return nil
}

func (a *QuitAction) GetDescription() string {
	return "Quit application"
}

// SeekBackwardAction 快退动作
type SeekBackwardAction struct{}

func (a *SeekBackwardAction) Execute(ctx context.Context, event *InputEvent) error {
	// 实现快退逻辑
	return nil
}

func (a *SeekBackwardAction) GetDescription() string {
	return "Seek backward"
}

// SeekForwardAction 快进动作
type SeekForwardAction struct{}

func (a *SeekForwardAction) Execute(ctx context.Context, event *InputEvent) error {
	// 实现快进逻辑
	return nil
}

func (a *SeekForwardAction) GetDescription() string {
	return "Seek forward"
}

// VolumeUpAction 音量增加动作
type VolumeUpAction struct{}

func (a *VolumeUpAction) Execute(ctx context.Context, event *InputEvent) error {
	// 实现音量增加逻辑
	return nil
}

func (a *VolumeUpAction) GetDescription() string {
	return "Volume up"
}

// VolumeDownAction 音量减少动作
type VolumeDownAction struct{}

func (a *VolumeDownAction) Execute(ctx context.Context, event *InputEvent) error {
	// 实现音量减少逻辑
	return nil
}

func (a *VolumeDownAction) GetDescription() string {
	return "Volume down"
}