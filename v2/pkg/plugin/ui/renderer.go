package ui

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
)

// MultiUIRenderer 多UI类型渲染器
type MultiUIRenderer struct {
	renderers map[UIType]UIRenderer
	mutex     sync.RWMutex
	logger    *slog.Logger
}

// NewMultiUIRenderer 创建多UI类型渲染器
func NewMultiUIRenderer(logger *slog.Logger) *MultiUIRenderer {
	renderer := &MultiUIRenderer{
		renderers: make(map[UIType]UIRenderer),
		logger:    logger,
	}

	// 注册默认渲染器
	renderer.registerDefaultRenderers()

	return renderer
}

// registerDefaultRenderers 注册默认渲染器
func (r *MultiUIRenderer) registerDefaultRenderers() {
	r.renderers[UITypeDesktop] = NewDesktopRenderer(r.logger)
	r.renderers[UITypeWeb] = NewWebRenderer(r.logger)
	r.renderers[UITypeMobile] = NewMobileRenderer(r.logger)
	r.renderers[UITypeTerminal] = NewTerminalRenderer(r.logger)
	r.renderers[UITypeEmbedded] = NewEmbeddedRenderer(r.logger)
}

// Render 渲染组件
func (r *MultiUIRenderer) Render(ctx context.Context, component *UIComponent, state *AppState) ([]byte, error) {
	if component == nil {
		return nil, fmt.Errorf("component cannot be nil")
	}

	if state == nil {
		return nil, fmt.Errorf("state cannot be nil")
	}

	// 根据当前UI类型选择渲染器
	uiType := UITypeDesktop // 默认桌面UI
	if uiTypeStr, ok := state.Config["ui_type"]; ok {
		if parsedType, err := parseUIType(uiTypeStr); err == nil {
			uiType = parsedType
		}
	}

	r.mutex.RLock()
	renderer, exists := r.renderers[uiType]
	r.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("no renderer found for UI type: %s", uiType.String())
	}

	return renderer.Render(ctx, component, state)
}

// SupportsType 检查是否支持指定UI类型
func (r *MultiUIRenderer) SupportsType(uiType UIType) bool {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	_, exists := r.renderers[uiType]
	return exists
}

// GetSupportedTypes 获取支持的UI类型
func (r *MultiUIRenderer) GetSupportedTypes() []UIType {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	types := make([]UIType, 0, len(r.renderers))
	for uiType := range r.renderers {
		types = append(types, uiType)
	}

	return types
}

// RegisterRenderer 注册渲染器
func (r *MultiUIRenderer) RegisterRenderer(uiType UIType, renderer UIRenderer) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.renderers[uiType] = renderer
	r.logger.Info("Renderer registered", "type", uiType.String())
}

// DesktopRenderer 桌面UI渲染器
type DesktopRenderer struct {
	logger *slog.Logger
}

// NewDesktopRenderer 创建桌面UI渲染器
func NewDesktopRenderer(logger *slog.Logger) *DesktopRenderer {
	return &DesktopRenderer{
		logger: logger,
	}
}

// Render 渲染桌面UI组件
func (r *DesktopRenderer) Render(ctx context.Context, component *UIComponent, state *AppState) ([]byte, error) {
	// 根据组件类型生成桌面UI代码
	switch component.Type {
	case ComponentTypeButton:
		return r.renderButton(component, state)
	case ComponentTypeInput:
		return r.renderInput(component, state)
	case ComponentTypeLabel:
		return r.renderLabel(component, state)
	case ComponentTypePlayer:
		return r.renderPlayer(component, state)
	case ComponentTypePlaylist:
		return r.renderPlaylist(component, state)
	default:
		return r.renderCustom(component, state)
	}
}

// renderButton 渲染按钮
func (r *DesktopRenderer) renderButton(component *UIComponent, state *AppState) ([]byte, error) {
	html := fmt.Sprintf(`<button id="%s" class="ui-button" style="%s">%s</button>`,
		component.ID,
		component.Styles,
		component.Name)
	return []byte(html), nil
}

// renderInput 渲染输入框
func (r *DesktopRenderer) renderInput(component *UIComponent, state *AppState) ([]byte, error) {
	placeholder := ""
	if val, ok := component.Props["placeholder"]; ok {
		if str, ok := val.(string); ok {
			placeholder = str
		}
	}

	html := fmt.Sprintf(`<input id="%s" class="ui-input" placeholder="%s" style="%s" />`,
		component.ID,
		placeholder,
		component.Styles)
	return []byte(html), nil
}

// renderLabel 渲染标签
func (r *DesktopRenderer) renderLabel(component *UIComponent, state *AppState) ([]byte, error) {
	text := component.Name
	if val, ok := component.Props["text"]; ok {
		if str, ok := val.(string); ok {
			text = str
		}
	}

	html := fmt.Sprintf(`<label id="%s" class="ui-label" style="%s">%s</label>`,
		component.ID,
		component.Styles,
		text)
	return []byte(html), nil
}

// renderPlayer 渲染播放器
func (r *DesktopRenderer) renderPlayer(component *UIComponent, state *AppState) ([]byte, error) {
	playerHTML := `<div id="%s" class="ui-player" style="%s">
`
	playerHTML += `  <div class="player-info">
`

	if state.Player != nil && state.Player.CurrentSong != nil {
		playerHTML += fmt.Sprintf(`    <div class="song-title">%s</div>\n`, state.Player.CurrentSong.Title)
		playerHTML += fmt.Sprintf(`    <div class="song-artist">%s</div>\n`, state.Player.CurrentSong.Artist)
	}

	playerHTML += `  </div>\n`
	playerHTML += `  <div class="player-controls">\n`
	playerHTML += `    <button class="btn-prev">⏮</button>\n`
	playerHTML += `    <button class="btn-play">⏯</button>\n`
	playerHTML += `    <button class="btn-next">⏭</button>\n`
	playerHTML += `  </div>\n`
	playerHTML += `</div>`

	html := fmt.Sprintf(playerHTML, component.ID, component.Styles)
	return []byte(html), nil
}

// renderPlaylist 渲染播放列表
func (r *DesktopRenderer) renderPlaylist(component *UIComponent, state *AppState) ([]byte, error) {
	playlistHTML := `<div id="%s" class="ui-playlist" style="%s">\n`
	playlistHTML += `  <div class="playlist-header">Playlist</div>\n`
	playlistHTML += `  <div class="playlist-items">\n`

	if state.Player != nil && len(state.Player.Queue) > 0 {
		for i, song := range state.Player.Queue {
			playlistHTML += fmt.Sprintf(`    <div class="playlist-item" data-index="%d">\n`, i)
			playlistHTML += fmt.Sprintf(`      <span class="song-title">%s</span>\n`, song.Title)
			playlistHTML += fmt.Sprintf(`      <span class="song-artist">%s</span>\n`, song.Artist)
			playlistHTML += `    </div>\n`
		}
	}

	playlistHTML += `  </div>\n`
	playlistHTML += `</div>`

	html := fmt.Sprintf(playlistHTML, component.ID, component.Styles)
	return []byte(html), nil
}

// renderCustom 渲染自定义组件
func (r *DesktopRenderer) renderCustom(component *UIComponent, state *AppState) ([]byte, error) {
	if component.Template != "" {
		// 使用模板渲染
		return []byte(component.Template), nil
	}

	// 默认渲染
	html := fmt.Sprintf(`<div id="%s" class="ui-custom" style="%s">%s</div>`,
		component.ID,
		component.Styles,
		component.Name)
	return []byte(html), nil
}

// SupportsType 检查是否支持指定UI类型
func (r *DesktopRenderer) SupportsType(uiType UIType) bool {
	return uiType == UITypeDesktop
}

// GetSupportedTypes 获取支持的UI类型
func (r *DesktopRenderer) GetSupportedTypes() []UIType {
	return []UIType{UITypeDesktop}
}

// WebRenderer Web UI渲染器
type WebRenderer struct {
	logger *slog.Logger
}

// NewWebRenderer 创建Web UI渲染器
func NewWebRenderer(logger *slog.Logger) *WebRenderer {
	return &WebRenderer{
		logger: logger,
	}
}

// Render 渲染Web UI组件
func (r *WebRenderer) Render(ctx context.Context, component *UIComponent, state *AppState) ([]byte, error) {
	// Web UI渲染逻辑
	html := fmt.Sprintf(`<div id="%s" class="web-component %s" style="%s">`,
		component.ID,
		strings.ToLower(component.Type.String()),
		component.Styles)

	if component.Template != "" {
		html += component.Template
	} else {
		html += component.Name
	}

	html += `</div>`

	return []byte(html), nil
}

// SupportsType 检查是否支持指定UI类型
func (r *WebRenderer) SupportsType(uiType UIType) bool {
	return uiType == UITypeWeb
}

// GetSupportedTypes 获取支持的UI类型
func (r *WebRenderer) GetSupportedTypes() []UIType {
	return []UIType{UITypeWeb}
}

// MobileRenderer 移动端UI渲染器
type MobileRenderer struct {
	logger *slog.Logger
}

// NewMobileRenderer 创建移动端UI渲染器
func NewMobileRenderer(logger *slog.Logger) *MobileRenderer {
	return &MobileRenderer{
		logger: logger,
	}
}

// Render 渲染移动端UI组件
func (r *MobileRenderer) Render(ctx context.Context, component *UIComponent, state *AppState) ([]byte, error) {
	// 移动端UI渲染逻辑
	html := fmt.Sprintf(`<div id="%s" class="mobile-component %s" style="%s; touch-action: manipulation;">`,
		component.ID,
		strings.ToLower(component.Type.String()),
		component.Styles)

	if component.Template != "" {
		html += component.Template
	} else {
		html += component.Name
	}

	html += `</div>`

	return []byte(html), nil
}

// SupportsType 检查是否支持指定UI类型
func (r *MobileRenderer) SupportsType(uiType UIType) bool {
	return uiType == UITypeMobile
}

// GetSupportedTypes 获取支持的UI类型
func (r *MobileRenderer) GetSupportedTypes() []UIType {
	return []UIType{UITypeMobile}
}

// TerminalRenderer 终端UI渲染器
type TerminalRenderer struct {
	logger *slog.Logger
}

// NewTerminalRenderer 创建终端UI渲染器
func NewTerminalRenderer(logger *slog.Logger) *TerminalRenderer {
	return &TerminalRenderer{
		logger: logger,
	}
}

// Render 渲染终端UI组件
func (r *TerminalRenderer) Render(ctx context.Context, component *UIComponent, state *AppState) ([]byte, error) {
	// 终端UI渲染逻辑
	switch component.Type {
	case ComponentTypeButton:
		return []byte(fmt.Sprintf("[%s]", component.Name)), nil
	case ComponentTypeLabel:
		return []byte(component.Name), nil
	case ComponentTypePlayer:
		if state.Player != nil && state.Player.CurrentSong != nil {
			return []byte(fmt.Sprintf("♪ %s - %s", state.Player.CurrentSong.Title, state.Player.CurrentSong.Artist)), nil
		}
		return []byte("♪ No song playing"), nil
	default:
		return []byte(fmt.Sprintf("%s: %s", component.Type.String(), component.Name)), nil
	}
}

// SupportsType 检查是否支持指定UI类型
func (r *TerminalRenderer) SupportsType(uiType UIType) bool {
	return uiType == UITypeTerminal
}

// GetSupportedTypes 获取支持的UI类型
func (r *TerminalRenderer) GetSupportedTypes() []UIType {
	return []UIType{UITypeTerminal}
}

// EmbeddedRenderer 嵌入式UI渲染器
type EmbeddedRenderer struct {
	logger *slog.Logger
}

// NewEmbeddedRenderer 创建嵌入式UI渲染器
func NewEmbeddedRenderer(logger *slog.Logger) *EmbeddedRenderer {
	return &EmbeddedRenderer{
		logger: logger,
	}
}

// Render 渲染嵌入式UI组件
func (r *EmbeddedRenderer) Render(ctx context.Context, component *UIComponent, state *AppState) ([]byte, error) {
	// 嵌入式UI渲染逻辑（简化版）
	data := fmt.Sprintf("%s:%s", component.ID, component.Name)
	return []byte(data), nil
}

// SupportsType 检查是否支持指定UI类型
func (r *EmbeddedRenderer) SupportsType(uiType UIType) bool {
	return uiType == UITypeEmbedded
}

// GetSupportedTypes 获取支持的UI类型
func (r *EmbeddedRenderer) GetSupportedTypes() []UIType {
	return []UIType{UITypeEmbedded}
}

// parseUIType 解析UI类型字符串
func parseUIType(typeStr string) (UIType, error) {
	switch strings.ToLower(typeStr) {
	case "desktop":
		return UITypeDesktop, nil
	case "web":
		return UITypeWeb, nil
	case "mobile":
		return UITypeMobile, nil
	case "terminal":
		return UITypeTerminal, nil
	case "embedded":
		return UITypeEmbedded, nil
	default:
		return UITypeDesktop, fmt.Errorf("unknown UI type: %s", typeStr)
	}
}

// ComponentType.String 返回组件类型的字符串表示
func (c ComponentType) String() string {
	switch c {
	case ComponentTypeButton:
		return "button"
	case ComponentTypeInput:
		return "input"
	case ComponentTypeLabel:
		return "label"
	case ComponentTypeImage:
		return "image"
	case ComponentTypeList:
		return "list"
	case ComponentTypeTable:
		return "table"
	case ComponentTypeTree:
		return "tree"
	case ComponentTypeChart:
		return "chart"
	case ComponentTypePlayer:
		return "player"
	case ComponentTypePlaylist:
		return "playlist"
	case ComponentTypeLyrics:
		return "lyrics"
	case ComponentTypeVisualizer:
		return "visualizer"
	case ComponentTypeEqualizer:
		return "equalizer"
	case ComponentTypeCustom:
		return "custom"
	default:
		return "unknown"
	}
}