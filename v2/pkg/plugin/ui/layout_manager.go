package ui

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"sort"
	"sync"
	"time"
)

// DefaultLayoutManager 默认布局管理器
type DefaultLayoutManager struct {
	currentLayout *Layout
	layouts       map[string]*Layout
	mutex         sync.RWMutex
	logger        *slog.Logger

	// 布局计算器
	calculators map[LayoutType]LayoutCalculator

	// 响应式管理
	responsiveManager *ResponsiveManager

	// 布局历史
	layoutHistory []LayoutHistoryEntry
	maxHistory    int
}

// LayoutCalculator 布局计算器接口
type LayoutCalculator interface {
	Calculate(layout *Layout, components []*UIComponent, viewport *Viewport) (*LayoutResult, error)
	Validate(layout *Layout) error
	GetType() LayoutType
}

// LayoutResult 布局计算结果
type LayoutResult struct {
	Components []*ComponentLayout `json:"components"`
	Viewport   *Viewport          `json:"viewport"`
	Metrics    *LayoutMetrics     `json:"metrics"`
}

// ComponentLayout 组件布局信息
type ComponentLayout struct {
	ComponentID string    `json:"component_id"`
	Position    *Position `json:"position"`
	Size        *Size     `json:"size"`
	ZIndex      int       `json:"z_index"`
	Visible     bool      `json:"visible"`
	Clipped     bool      `json:"clipped"`
}

// Viewport 视口信息
type Viewport struct {
	Width  int `json:"width"`
	Height int `json:"height"`
	DPI    int `json:"dpi"`
}

// LayoutMetrics 布局指标
type LayoutMetrics struct {
	TotalComponents int           `json:"total_components"`
	VisibleComponents int         `json:"visible_components"`
	CalculationTime time.Duration `json:"calculation_time"`
	MemoryUsage     int64         `json:"memory_usage"`
}

// LayoutHistoryEntry 布局历史条目
type LayoutHistoryEntry struct {
	Layout    *Layout   `json:"layout"`
	Timestamp time.Time `json:"timestamp"`
	Reason    string    `json:"reason"`
}

// ResponsiveManager 响应式管理器
type ResponsiveManager struct {
	breakpoints map[string]int
	currentBreakpoint string
	logger      *slog.Logger
}

// NewDefaultLayoutManager 创建默认布局管理器
func NewDefaultLayoutManager(logger *slog.Logger) *DefaultLayoutManager {
	manager := &DefaultLayoutManager{
		layouts:       make(map[string]*Layout),
		logger:        logger,
		calculators:   make(map[LayoutType]LayoutCalculator),
		layoutHistory: make([]LayoutHistoryEntry, 0),
		maxHistory:    50,
	}

	// 初始化布局计算器
	manager.initializeCalculators()

	// 初始化响应式管理器
	manager.responsiveManager = NewResponsiveManager(logger)

	// 注册默认布局
	manager.registerDefaultLayouts()

	return manager
}

// initializeCalculators 初始化布局计算器
func (m *DefaultLayoutManager) initializeCalculators() {
	m.calculators[LayoutTypeGrid] = NewGridCalculator(m.logger)
	m.calculators[LayoutTypeFlex] = NewFlexCalculator(m.logger)
	m.calculators[LayoutTypeAbsolute] = NewAbsoluteCalculator(m.logger)
	m.calculators[LayoutTypeFloat] = NewFloatCalculator(m.logger)
	m.calculators[LayoutTypeTable] = NewTableCalculator(m.logger)
}

// registerDefaultLayouts 注册默认布局
func (m *DefaultLayoutManager) registerDefaultLayouts() {
	// 网格布局
	gridLayout := &Layout{
		ID:          "default-grid",
		Name:        "Default Grid",
		Description: "Default grid layout",
		Type:        LayoutTypeGrid,
		Grid: &GridConfig{
			Rows:    3,
			Columns: 3,
			Gap:     "10px",
		},
		IsDefault: true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	m.layouts[gridLayout.ID] = gridLayout

	// 弹性布局
	flexLayout := &Layout{
		ID:          "default-flex",
		Name:        "Default Flex",
		Description: "Default flex layout",
		Type:        LayoutTypeFlex,
		Flex: &FlexConfig{
			Direction: "column",
			Wrap:      "nowrap",
			Justify:   "flex-start",
			Align:     "stretch",
			Gap:       "10px",
		},
		IsDefault: false,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	m.layouts[flexLayout.ID] = flexLayout

	// 设置默认布局
	m.currentLayout = gridLayout
}

// ApplyLayout 应用布局
func (m *DefaultLayoutManager) ApplyLayout(ctx context.Context, layout *Layout) error {
	if layout == nil {
		return fmt.Errorf("layout cannot be nil")
	}

	// 验证布局
	if err := m.ValidateLayout(layout); err != nil {
		return fmt.Errorf("layout validation failed: %w", err)
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 记录布局历史
	if m.currentLayout != nil {
		m.addToHistory(m.currentLayout, "layout_change")
	}

	// 应用新布局
	m.currentLayout = layout
	m.layouts[layout.ID] = layout

	m.logger.Info("Layout applied", "layout", layout.Name, "type", layout.Type)
	return nil
}

// GetCurrentLayout 获取当前布局
func (m *DefaultLayoutManager) GetCurrentLayout() *Layout {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return m.currentLayout
}

// ValidateLayout 验证布局
func (m *DefaultLayoutManager) ValidateLayout(layout *Layout) error {
	if layout == nil {
		return fmt.Errorf("layout cannot be nil")
	}

	if layout.ID == "" {
		return fmt.Errorf("layout ID cannot be empty")
	}

	if layout.Name == "" {
		return fmt.Errorf("layout name cannot be empty")
	}

	// 使用对应的计算器验证
	calculator, exists := m.calculators[layout.Type]
	if !exists {
		return fmt.Errorf("unsupported layout type: %d", layout.Type)
	}

	return calculator.Validate(layout)
}

// GetSupportedTypes 获取支持的布局类型
func (m *DefaultLayoutManager) GetSupportedTypes() []LayoutType {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	types := make([]LayoutType, 0, len(m.calculators))
	for layoutType := range m.calculators {
		types = append(types, layoutType)
	}

	return types
}

// CalculateLayout 计算布局
func (m *DefaultLayoutManager) CalculateLayout(components []*UIComponent, viewport *Viewport) (*LayoutResult, error) {
	m.mutex.RLock()
	currentLayout := m.currentLayout
	m.mutex.RUnlock()

	if currentLayout == nil {
		return nil, fmt.Errorf("no current layout set")
	}

	calculator, exists := m.calculators[currentLayout.Type]
	if !exists {
		return nil, fmt.Errorf("no calculator found for layout type: %d", currentLayout.Type)
	}

	return calculator.Calculate(currentLayout, components, viewport)
}

// RegisterLayout 注册布局
func (m *DefaultLayoutManager) RegisterLayout(layout *Layout) error {
	if err := m.ValidateLayout(layout); err != nil {
		return err
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.layouts[layout.ID] = layout
	m.logger.Info("Layout registered", "id", layout.ID, "name", layout.Name)
	return nil
}

// GetLayout 获取布局
func (m *DefaultLayoutManager) GetLayout(layoutID string) (*Layout, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	layout, exists := m.layouts[layoutID]
	if !exists {
		return nil, fmt.Errorf("layout not found: %s", layoutID)
	}

	return layout, nil
}

// ListLayouts 列出所有布局
func (m *DefaultLayoutManager) ListLayouts() []*Layout {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	layouts := make([]*Layout, 0, len(m.layouts))
	for _, layout := range m.layouts {
		layouts = append(layouts, layout)
	}

	// 按名称排序
	sort.Slice(layouts, func(i, j int) bool {
		return layouts[i].Name < layouts[j].Name
	})

	return layouts
}

// addToHistory 添加到布局历史
func (m *DefaultLayoutManager) addToHistory(layout *Layout, reason string) {
	entry := LayoutHistoryEntry{
		Layout:    layout,
		Timestamp: time.Now(),
		Reason:    reason,
	}

	m.layoutHistory = append(m.layoutHistory, entry)

	// 限制历史长度
	if len(m.layoutHistory) > m.maxHistory {
		m.layoutHistory = m.layoutHistory[1:]
	}
}

// GetLayoutHistory 获取布局历史
func (m *DefaultLayoutManager) GetLayoutHistory() []LayoutHistoryEntry {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	history := make([]LayoutHistoryEntry, len(m.layoutHistory))
	copy(history, m.layoutHistory)
	return history
}

// GridCalculator 网格布局计算器
type GridCalculator struct {
	logger *slog.Logger
}

// NewGridCalculator 创建网格布局计算器
func NewGridCalculator(logger *slog.Logger) *GridCalculator {
	return &GridCalculator{
		logger: logger,
	}
}

// Calculate 计算网格布局
func (c *GridCalculator) Calculate(layout *Layout, components []*UIComponent, viewport *Viewport) (*LayoutResult, error) {
	start := time.Now()

	if layout.Grid == nil {
		return nil, fmt.Errorf("grid config is required for grid layout")
	}

	grid := layout.Grid
	cellWidth := viewport.Width / grid.Columns
	cellHeight := viewport.Height / grid.Rows

	result := &LayoutResult{
		Components: make([]*ComponentLayout, 0, len(components)),
		Viewport:   viewport,
		Metrics: &LayoutMetrics{
			TotalComponents: len(components),
		},
	}

	// 计算每个组件的位置
	for i, component := range components {
		if !component.Visible {
			continue
		}

		row := i / grid.Columns
		col := i % grid.Columns

		// 检查是否超出网格范围
		if row >= grid.Rows {
			break
		}

		componentLayout := &ComponentLayout{
			ComponentID: component.ID,
			Position: &Position{
				X: col * cellWidth,
				Y: row * cellHeight,
				Z: component.Position.Z,
			},
			Size: &Size{
				Width:  cellWidth,
				Height: cellHeight,
			},
			ZIndex:  component.Position.Z,
			Visible: true,
			Clipped: false,
		}

		result.Components = append(result.Components, componentLayout)
		result.Metrics.VisibleComponents++
	}

	result.Metrics.CalculationTime = time.Since(start)
	return result, nil
}

// Validate 验证网格布局
func (c *GridCalculator) Validate(layout *Layout) error {
	if layout.Grid == nil {
		return fmt.Errorf("grid config is required")
	}

	if layout.Grid.Rows <= 0 {
		return fmt.Errorf("grid rows must be positive")
	}

	if layout.Grid.Columns <= 0 {
		return fmt.Errorf("grid columns must be positive")
	}

	return nil
}

// GetType 获取布局类型
func (c *GridCalculator) GetType() LayoutType {
	return LayoutTypeGrid
}

// FlexCalculator 弹性布局计算器
type FlexCalculator struct {
	logger *slog.Logger
}

// NewFlexCalculator 创建弹性布局计算器
func NewFlexCalculator(logger *slog.Logger) *FlexCalculator {
	return &FlexCalculator{
		logger: logger,
	}
}

// Calculate 计算弹性布局
func (c *FlexCalculator) Calculate(layout *Layout, components []*UIComponent, viewport *Viewport) (*LayoutResult, error) {
	start := time.Now()

	if layout.Flex == nil {
		return nil, fmt.Errorf("flex config is required for flex layout")
	}

	flex := layout.Flex
	result := &LayoutResult{
		Components: make([]*ComponentLayout, 0, len(components)),
		Viewport:   viewport,
		Metrics: &LayoutMetrics{
			TotalComponents: len(components),
		},
	}

	// 计算主轴和交叉轴
	isColumn := flex.Direction == "column" || flex.Direction == "column-reverse"
	mainSize := viewport.Width
	crossSize := viewport.Height
	if isColumn {
		mainSize = viewport.Height
		crossSize = viewport.Width
	}

	// 计算组件大小
	visibleComponents := make([]*UIComponent, 0)
	for _, component := range components {
		if component.Visible {
			visibleComponents = append(visibleComponents, component)
		}
	}

	if len(visibleComponents) == 0 {
		result.Metrics.CalculationTime = time.Since(start)
		return result, nil
	}

	itemSize := mainSize / len(visibleComponents)
	currentPos := 0

	// 布局组件
	for _, component := range visibleComponents {
		var position *Position
		var size *Size

		if isColumn {
			position = &Position{
				X: 0,
				Y: currentPos,
				Z: component.Position.Z,
			}
			size = &Size{
				Width:  crossSize,
				Height: itemSize,
			}
		} else {
			position = &Position{
				X: currentPos,
				Y: 0,
				Z: component.Position.Z,
			}
			size = &Size{
				Width:  itemSize,
				Height: crossSize,
			}
		}

		componentLayout := &ComponentLayout{
			ComponentID: component.ID,
			Position:    position,
			Size:        size,
			ZIndex:      component.Position.Z,
			Visible:     true,
			Clipped:     false,
		}

		result.Components = append(result.Components, componentLayout)
		result.Metrics.VisibleComponents++
		currentPos += itemSize
	}

	result.Metrics.CalculationTime = time.Since(start)
	return result, nil
}

// Validate 验证弹性布局
func (c *FlexCalculator) Validate(layout *Layout) error {
	if layout.Flex == nil {
		return fmt.Errorf("flex config is required")
	}

	validDirections := []string{"row", "row-reverse", "column", "column-reverse"}
	validDirection := false
	for _, dir := range validDirections {
		if layout.Flex.Direction == dir {
			validDirection = true
			break
		}
	}

	if !validDirection {
		return fmt.Errorf("invalid flex direction: %s", layout.Flex.Direction)
	}

	return nil
}

// GetType 获取布局类型
func (c *FlexCalculator) GetType() LayoutType {
	return LayoutTypeFlex
}

// AbsoluteCalculator 绝对定位布局计算器
type AbsoluteCalculator struct {
	logger *slog.Logger
}

// NewAbsoluteCalculator 创建绝对定位布局计算器
func NewAbsoluteCalculator(logger *slog.Logger) *AbsoluteCalculator {
	return &AbsoluteCalculator{
		logger: logger,
	}
}

// Calculate 计算绝对定位布局
func (c *AbsoluteCalculator) Calculate(layout *Layout, components []*UIComponent, viewport *Viewport) (*LayoutResult, error) {
	start := time.Now()

	result := &LayoutResult{
		Components: make([]*ComponentLayout, 0, len(components)),
		Viewport:   viewport,
		Metrics: &LayoutMetrics{
			TotalComponents: len(components),
		},
	}

	// 使用组件自身的位置和大小
	for _, component := range components {
		if !component.Visible {
			continue
		}

		position := component.Position
		size := component.Size

		// 使用默认值如果未设置
		if position == nil {
			position = &Position{X: 0, Y: 0, Z: 0}
		}
		if size == nil {
			size = &Size{Width: 100, Height: 100}
		}

		// 检查是否超出视口
		clipped := position.X+size.Width > viewport.Width || position.Y+size.Height > viewport.Height

		componentLayout := &ComponentLayout{
			ComponentID: component.ID,
			Position:    position,
			Size:        size,
			ZIndex:      position.Z,
			Visible:     true,
			Clipped:     clipped,
		}

		result.Components = append(result.Components, componentLayout)
		result.Metrics.VisibleComponents++
	}

	result.Metrics.CalculationTime = time.Since(start)
	return result, nil
}

// Validate 验证绝对定位布局
func (c *AbsoluteCalculator) Validate(layout *Layout) error {
	// 绝对定位布局不需要特殊配置
	return nil
}

// GetType 获取布局类型
func (c *AbsoluteCalculator) GetType() LayoutType {
	return LayoutTypeAbsolute
}

// FloatCalculator 浮动布局计算器
type FloatCalculator struct {
	logger *slog.Logger
}

// NewFloatCalculator 创建浮动布局计算器
func NewFloatCalculator(logger *slog.Logger) *FloatCalculator {
	return &FloatCalculator{
		logger: logger,
	}
}

// Calculate 计算浮动布局
func (c *FloatCalculator) Calculate(layout *Layout, components []*UIComponent, viewport *Viewport) (*LayoutResult, error) {
	start := time.Now()

	result := &LayoutResult{
		Components: make([]*ComponentLayout, 0, len(components)),
		Viewport:   viewport,
		Metrics: &LayoutMetrics{
			TotalComponents: len(components),
		},
	}

	// 简单的浮动布局实现
	currentX := 0
	currentY := 0
	rowHeight := 0

	for _, component := range components {
		if !component.Visible {
			continue
		}

		size := component.Size
		if size == nil {
			size = &Size{Width: 100, Height: 100}
		}

		// 检查是否需要换行
		if currentX+size.Width > viewport.Width {
			currentX = 0
			currentY += rowHeight
			rowHeight = 0
		}

		position := &Position{
			X: currentX,
			Y: currentY,
			Z: component.Position.Z,
		}

		componentLayout := &ComponentLayout{
			ComponentID: component.ID,
			Position:    position,
			Size:        size,
			ZIndex:      component.Position.Z,
			Visible:     true,
			Clipped:     false,
		}

		result.Components = append(result.Components, componentLayout)
		result.Metrics.VisibleComponents++

		currentX += size.Width
		rowHeight = int(math.Max(float64(rowHeight), float64(size.Height)))
	}

	result.Metrics.CalculationTime = time.Since(start)
	return result, nil
}

// Validate 验证浮动布局
func (c *FloatCalculator) Validate(layout *Layout) error {
	// 浮动布局不需要特殊配置
	return nil
}

// GetType 获取布局类型
func (c *FloatCalculator) GetType() LayoutType {
	return LayoutTypeFloat
}

// TableCalculator 表格布局计算器
type TableCalculator struct {
	logger *slog.Logger
}

// NewTableCalculator 创建表格布局计算器
func NewTableCalculator(logger *slog.Logger) *TableCalculator {
	return &TableCalculator{
		logger: logger,
	}
}

// Calculate 计算表格布局
func (c *TableCalculator) Calculate(layout *Layout, components []*UIComponent, viewport *Viewport) (*LayoutResult, error) {
	// 表格布局类似网格布局，但更严格
	gridCalc := NewGridCalculator(c.logger)
	return gridCalc.Calculate(layout, components, viewport)
}

// Validate 验证表格布局
func (c *TableCalculator) Validate(layout *Layout) error {
	gridCalc := NewGridCalculator(c.logger)
	return gridCalc.Validate(layout)
}

// GetType 获取布局类型
func (c *TableCalculator) GetType() LayoutType {
	return LayoutTypeTable
}

// NewResponsiveManager 创建响应式管理器
func NewResponsiveManager(logger *slog.Logger) *ResponsiveManager {
	return &ResponsiveManager{
		breakpoints: map[string]int{
			"xs": 480,
			"sm": 768,
			"md": 1024,
			"lg": 1280,
			"xl": 1920,
		},
		currentBreakpoint: "md",
		logger:            logger,
	}
}

// GetCurrentBreakpoint 获取当前断点
func (r *ResponsiveManager) GetCurrentBreakpoint(width int) string {
	for name, breakpoint := range r.breakpoints {
		if width <= breakpoint {
			return name
		}
	}
	return "xl"
}

// UpdateBreakpoint 更新断点
func (r *ResponsiveManager) UpdateBreakpoint(width int) bool {
	newBreakpoint := r.GetCurrentBreakpoint(width)
	if newBreakpoint != r.currentBreakpoint {
		r.currentBreakpoint = newBreakpoint
		r.logger.Info("Breakpoint changed", "breakpoint", newBreakpoint, "width", width)
		return true
	}
	return false
}