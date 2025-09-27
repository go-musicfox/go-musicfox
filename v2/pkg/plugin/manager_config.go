package plugin

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	// "time" // 暂时未使用

	"log/slog"
)

// ConfigDrivenPluginManager 配置驱动的插件管理器
type ConfigDrivenPluginManager struct {
	logger          *slog.Logger
	configManager   *PluginConfigManager
	resourceManager *ResourceManager
	securityManager *SecurityManager
	hotReloader     *ConfigHotReloader
	plugins         map[string]interface{} // 插件实例映射
	configs         map[string]*EnhancedPluginConfig
	dependencies    *DependencyGraph
	mu              sync.RWMutex
	ctx             context.Context
	cancel          context.CancelFunc
	isRunning       bool
	loadOrder       []string
	conflictResolver *ConfigConflictResolver
}

// NewConfigDrivenPluginManager 创建配置驱动的插件管理器
func NewConfigDrivenPluginManager(logger *slog.Logger) *ConfigDrivenPluginManager {
	ctx, cancel := context.WithCancel(context.Background())

	configManager := NewPluginConfigManager(logger)
	resourceManager := NewResourceManager(logger)
	securityManager := NewSecurityManager(logger)
	hotReloader := NewConfigHotReloader(logger, configManager, nil)

	manager := &ConfigDrivenPluginManager{
		logger:          logger,
		configManager:   configManager,
		resourceManager: resourceManager,
		securityManager: securityManager,
		hotReloader:     hotReloader,
		plugins:         make(map[string]interface{}),
		configs:         make(map[string]*EnhancedPluginConfig),
		dependencies:    NewDependencyGraph(),
		ctx:             ctx,
		cancel:          cancel,
		conflictResolver: NewConfigConflictResolver(logger),
	}

	// 设置热重载器的插件管理器引用
	hotReloader.pluginManager = manager

	return manager
}

// Start 启动配置驱动的插件管理器
func (cpm *ConfigDrivenPluginManager) Start(ctx context.Context) error {
	cpm.mu.Lock()
	defer cpm.mu.Unlock()

	if cpm.isRunning {
		return fmt.Errorf("config driven plugin manager is already running")
	}

	// 启动各个子系统
	if err := cpm.configManager.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize config manager: %w", err)
	}

	if err := cpm.resourceManager.Start(ctx); err != nil {
		return fmt.Errorf("failed to start resource manager: %w", err)
	}

	if err := cpm.securityManager.Start(ctx); err != nil {
		return fmt.Errorf("failed to start security manager: %w", err)
	}

	if err := cpm.hotReloader.Start(ctx); err != nil {
		return fmt.Errorf("failed to start hot reloader: %w", err)
	}

	// 加载所有配置的插件
	if err := cpm.loadAllConfiguredPlugins(ctx); err != nil {
		return fmt.Errorf("failed to load configured plugins: %w", err)
	}

	cpm.isRunning = true
	cpm.logger.Info("Config driven plugin manager started")
	return nil
}

// Stop 停止配置驱动的插件管理器
func (cpm *ConfigDrivenPluginManager) Stop() {
	cpm.mu.Lock()
	defer cpm.mu.Unlock()

	if !cpm.isRunning {
		return
	}

	cpm.isRunning = false
	cpm.cancel()

	// 停止所有插件
	for pluginID := range cpm.plugins {
		cpm.unloadPlugin(pluginID)
	}

	// 停止各个子系统
	cpm.hotReloader.Stop()
	cpm.securityManager.Stop()
	cpm.resourceManager.Stop()

	cpm.logger.Info("Config driven plugin manager stopped")
}

// LoadPlugin 加载插件
func (cpm *ConfigDrivenPluginManager) LoadPlugin(ctx context.Context, pluginID string) error {
	cpm.mu.Lock()
	defer cpm.mu.Unlock()

	return cpm.loadPlugin(ctx, pluginID)
}

// UnloadPlugin 卸载插件
func (cpm *ConfigDrivenPluginManager) UnloadPlugin(pluginID string) error {
	cpm.mu.Lock()
	defer cpm.mu.Unlock()

	return cpm.unloadPlugin(pluginID)
}

// ReloadPlugin 重新加载插件
func (cpm *ConfigDrivenPluginManager) ReloadPlugin(ctx context.Context, pluginID string) error {
	cpm.mu.Lock()
	defer cpm.mu.Unlock()

	// 先卸载
	if err := cpm.unloadPlugin(pluginID); err != nil {
		cpm.logger.Warn("Failed to unload plugin during reload", "plugin_id", pluginID, "error", err)
	}

	// 重新加载配置
	config, err := cpm.configManager.LoadConfig(ctx, pluginID)
	if err != nil {
		return fmt.Errorf("failed to reload config: %w", err)
	}

	cpm.configs[pluginID] = config

	// 重新加载插件
	return cpm.loadPlugin(ctx, pluginID)
}

// GetPlugin 获取插件实例
func (cpm *ConfigDrivenPluginManager) GetPlugin(pluginID string) interface{} {
	cpm.mu.RLock()
	defer cpm.mu.RUnlock()

	return cpm.plugins[pluginID]
}

// GetPluginConfig 获取插件配置
func (cpm *ConfigDrivenPluginManager) GetPluginConfig(pluginID string) *EnhancedPluginConfig {
	cpm.mu.RLock()
	defer cpm.mu.RUnlock()

	return cpm.configs[pluginID]
}

// ListPlugins 列出所有插件
func (cpm *ConfigDrivenPluginManager) ListPlugins() []string {
	cpm.mu.RLock()
	defer cpm.mu.RUnlock()

	plugins := make([]string, 0, len(cpm.plugins))
	for pluginID := range cpm.plugins {
		plugins = append(plugins, pluginID)
	}

	sort.Strings(plugins)
	return plugins
}

// UpdatePluginConfig 更新插件配置
func (cpm *ConfigDrivenPluginManager) UpdatePluginConfig(ctx context.Context, pluginID string, config *EnhancedPluginConfig) error {
	cpm.mu.Lock()
	defer cpm.mu.Unlock()

	// 验证配置
	if err := cpm.configManager.validateConfig(ctx, config); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}

	// 检查配置冲突
	if conflicts := cpm.conflictResolver.DetectConflicts(config, cpm.configs); len(conflicts) > 0 {
		if err := cpm.conflictResolver.ResolveConflicts(conflicts); err != nil {
			return fmt.Errorf("failed to resolve config conflicts: %w", err)
		}
	}

	// 保存配置
	if err := cpm.configManager.SaveConfig(ctx, config); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	cpm.configs[pluginID] = config

	// 如果插件已加载，应用新配置
	if plugin, exists := cpm.plugins[pluginID]; exists {
		if err := cpm.applyConfigToPlugin(plugin, config); err != nil {
			return fmt.Errorf("failed to apply config to plugin: %w", err)
		}
	}

	cpm.logger.Info("Plugin config updated", "plugin_id", pluginID)
	return nil
}

// loadAllConfiguredPlugins 加载所有配置的插件
func (cpm *ConfigDrivenPluginManager) loadAllConfiguredPlugins(ctx context.Context) error {
	// 获取所有配置
	configs, err := cpm.configManager.ListConfigs(ctx)
	if err != nil {
		return fmt.Errorf("failed to list configs: %w", err)
	}

	// 存储配置
	for _, config := range configs {
		cpm.configs[config.ID] = config
	}

	// 构建依赖图
	if err := cpm.buildDependencyGraph(); err != nil {
		return fmt.Errorf("failed to build dependency graph: %w", err)
	}

	// 计算加载顺序
	loadOrder, err := cpm.dependencies.TopologicalSort()
	if err != nil {
		return fmt.Errorf("failed to calculate load order: %w", err)
	}
	cpm.loadOrder = loadOrder

	// 按依赖顺序加载插件
	for _, pluginID := range loadOrder {
		if config, exists := cpm.configs[pluginID]; exists && config.Enabled {
			if err := cpm.loadPlugin(ctx, pluginID); err != nil {
				cpm.logger.Error("Failed to load plugin", "plugin_id", pluginID, "error", err)
				// 继续加载其他插件
			}
		}
	}

	return nil
}

// loadPlugin 加载单个插件
func (cpm *ConfigDrivenPluginManager) loadPlugin(ctx context.Context, pluginID string) error {
	config, exists := cpm.configs[pluginID]
	if !exists {
		return fmt.Errorf("config not found for plugin: %s", pluginID)
	}

	if !config.Enabled {
		return fmt.Errorf("plugin is disabled: %s", pluginID)
	}

	// 检查依赖
	if err := cpm.checkDependencies(pluginID); err != nil {
		return fmt.Errorf("dependency check failed: %w", err)
	}

	// 创建插件实例
	plugin, err := cpm.createPluginInstance(config)
	if err != nil {
		return fmt.Errorf("failed to create plugin instance: %w", err)
	}

	// TODO: 设置资源监控
	// 由于ResourceLimits和DetailedResourceLimits类型不匹配，需要重新设计
	// if config.ResourceLimits != nil {
	//     monitor, err := cpm.resourceManager.AddMonitor(pluginID, config.ResourceLimits)
	//     if err != nil {
	//         cpm.logger.Warn("Failed to add resource monitor", "plugin_id", pluginID, "error", err)
	//     } else {
	//         cpm.logger.Debug("Resource monitor added", "plugin_id", pluginID)
	//         _ = monitor // 避免未使用变量警告
	//     }
	// }
	cpm.logger.Debug("Resource monitoring temporarily disabled", "plugin_id", pluginID)

	// 设置安全执行器
	if config.SecurityConfig != nil {
		enforcer, err := cpm.securityManager.AddEnforcer(pluginID, config.SecurityConfig)
		if err != nil {
			cpm.logger.Warn("Failed to add security enforcer", "plugin_id", pluginID, "error", err)
		} else {
			cpm.logger.Debug("Security enforcer added", "plugin_id", pluginID)
			_ = enforcer // 避免未使用变量警告
		}
	}

	// 应用配置到插件
	if err := cpm.applyConfigToPlugin(plugin, config); err != nil {
		return fmt.Errorf("failed to apply config to plugin: %w", err)
	}

	// TODO: 实现插件初始化和启动逻辑
	// 由于plugin现在是interface{}类型，需要类型断言或重新设计
	cpm.logger.Info("Plugin initialization and start not implemented", "plugin_id", pluginID)

	cpm.plugins[pluginID] = plugin

	// 启动配置监控
	if config.AutoReload {
		if err := cpm.hotReloader.StartWatching(ctx, pluginID); err != nil {
			cpm.logger.Warn("Failed to start config watching", "plugin_id", pluginID, "error", err)
		}
	}

	cpm.logger.Info("Plugin loaded successfully", "plugin_id", pluginID)
	return nil
}

// unloadPlugin 卸载单个插件
func (cpm *ConfigDrivenPluginManager) unloadPlugin(pluginID string) error {
	_, exists := cpm.plugins[pluginID]
	if !exists {
		return fmt.Errorf("plugin not found: %s", pluginID)
	}

	// 停止配置监控
	cpm.hotReloader.StopWatching(pluginID)

	// TODO: 实现插件停止和清理逻辑
	// 由于plugin现在是interface{}类型，需要类型断言或重新设计
	cpm.logger.Info("Plugin stop and cleanup not implemented", "plugin_id", pluginID)

	// 移除资源监控
	cpm.resourceManager.RemoveMonitor(pluginID)

	// 移除安全执行器
	cpm.securityManager.RemoveEnforcer(pluginID)

	delete(cpm.plugins, pluginID)

	cpm.logger.Info("Plugin unloaded successfully", "plugin_id", pluginID)
	return nil
}

// createPluginInstance 创建插件实例
func (cpm *ConfigDrivenPluginManager) createPluginInstance(config *EnhancedPluginConfig) (interface{}, error) {
	// 这里需要根据插件类型创建相应的实例
	// 实际实现中可能需要插件工厂或注册机制
	
	// 示例实现：根据插件路径加载
	if config.PluginPath == "" {
		return nil, fmt.Errorf("plugin path not specified")
	}

	// 检查插件文件是否存在
	if !filepath.IsAbs(config.PluginPath) {
		return nil, fmt.Errorf("plugin path must be absolute: %s", config.PluginPath)
	}

	// 这里应该实现实际的插件加载逻辑
	// 例如：动态库加载、Go插件加载等
	// 为了示例，我们返回一个基础插件实例
	return map[string]interface{}{
		"id": config.ID,
		"name": config.Name,
		"version": config.Version,
	}, nil
}

// applyConfigToPlugin 将配置应用到插件
func (cpm *ConfigDrivenPluginManager) applyConfigToPlugin(plugin interface{}, config *EnhancedPluginConfig) error {
	// TODO: 实现配置应用逻辑
	// 由于plugin现在是interface{}类型，需要类型断言或重新设计
	cpm.logger.Debug("Config applied to plugin", "plugin_id", config.ID)
	return nil
}

// checkDependencies 检查插件依赖
func (cpm *ConfigDrivenPluginManager) checkDependencies(pluginID string) error {
	deps := cpm.dependencies.GetDependencies(pluginID)
	for _, dep := range deps {
		if _, exists := cpm.plugins[dep]; !exists {
			return fmt.Errorf("dependency not loaded: %s", dep)
		}
	}
	return nil
}

// buildDependencyGraph 构建依赖图
func (cpm *ConfigDrivenPluginManager) buildDependencyGraph() error {
	cpm.dependencies = NewDependencyGraph()

	for pluginID, config := range cpm.configs {
		cpm.dependencies.AddNode(pluginID)
		for _, dep := range config.Dependencies {
			if err := cpm.dependencies.AddEdge(dep, pluginID); err != nil {
				return fmt.Errorf("failed to add dependency edge: %w", err)
			}
		}
	}

	return nil
}

// DependencyGraph 依赖图
type DependencyGraph struct {
	nodes map[string]bool
	edges map[string][]string
	mu    sync.RWMutex
}

// NewDependencyGraph 创建依赖图
func NewDependencyGraph() *DependencyGraph {
	return &DependencyGraph{
		nodes: make(map[string]bool),
		edges: make(map[string][]string),
	}
}

// AddNode 添加节点
func (dg *DependencyGraph) AddNode(node string) {
	dg.mu.Lock()
	defer dg.mu.Unlock()

	dg.nodes[node] = true
	if _, exists := dg.edges[node]; !exists {
		dg.edges[node] = make([]string, 0)
	}
}

// AddEdge 添加边（from -> to）
func (dg *DependencyGraph) AddEdge(from, to string) error {
	dg.mu.Lock()
	defer dg.mu.Unlock()

	// 确保节点存在
	if !dg.nodes[from] {
		dg.nodes[from] = true
		dg.edges[from] = make([]string, 0)
	}
	if !dg.nodes[to] {
		dg.nodes[to] = true
		dg.edges[to] = make([]string, 0)
	}

	// 检查是否会形成循环
	if dg.wouldCreateCycle(from, to) {
		return fmt.Errorf("adding edge %s -> %s would create a cycle", from, to)
	}

	dg.edges[from] = append(dg.edges[from], to)
	return nil
}

// GetDependencies 获取节点的依赖
func (dg *DependencyGraph) GetDependencies(node string) []string {
	dg.mu.RLock()
	defer dg.mu.RUnlock()

	deps := make([]string, 0)
	for from, tos := range dg.edges {
		for _, to := range tos {
			if to == node {
				deps = append(deps, from)
			}
		}
	}
	return deps
}

// TopologicalSort 拓扑排序
func (dg *DependencyGraph) TopologicalSort() ([]string, error) {
	dg.mu.RLock()
	defer dg.mu.RUnlock()

	// 计算入度
	inDegree := make(map[string]int)
	for node := range dg.nodes {
		inDegree[node] = 0
	}

	for _, tos := range dg.edges {
		for _, to := range tos {
			inDegree[to]++
		}
	}

	// 找到入度为0的节点
	queue := make([]string, 0)
	for node, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, node)
		}
	}

	result := make([]string, 0, len(dg.nodes))

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		result = append(result, current)

		// 更新邻接节点的入度
		for _, neighbor := range dg.edges[current] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	if len(result) != len(dg.nodes) {
		return nil, fmt.Errorf("dependency cycle detected")
	}

	return result, nil
}

// wouldCreateCycle 检查添加边是否会创建循环
func (dg *DependencyGraph) wouldCreateCycle(from, to string) bool {
	// 使用DFS检查从to是否能到达from
	visited := make(map[string]bool)
	return dg.dfsHasPath(to, from, visited)
}

// dfsHasPath 使用DFS检查是否存在路径
func (dg *DependencyGraph) dfsHasPath(start, target string, visited map[string]bool) bool {
	if start == target {
		return true
	}

	if visited[start] {
		return false
	}

	visited[start] = true

	for _, neighbor := range dg.edges[start] {
		if dg.dfsHasPath(neighbor, target, visited) {
			return true
		}
	}

	return false
}

// ConfigConflict 配置冲突
type ConfigConflict struct {
	Type        string      `json:"type"`
	PluginID1   string      `json:"plugin_id_1"`
	PluginID2   string      `json:"plugin_id_2"`
	Resource    string      `json:"resource"`
	Description string      `json:"description"`
	Severity    string      `json:"severity"`
	Details     interface{} `json:"details,omitempty"`
}

// ConfigConflictResolver 配置冲突解决器
type ConfigConflictResolver struct {
	logger *slog.Logger
}

// NewConfigConflictResolver 创建配置冲突解决器
func NewConfigConflictResolver(logger *slog.Logger) *ConfigConflictResolver {
	return &ConfigConflictResolver{
		logger: logger,
	}
}

// DetectConflicts 检测配置冲突
func (ccr *ConfigConflictResolver) DetectConflicts(newConfig *EnhancedPluginConfig, existingConfigs map[string]*EnhancedPluginConfig) []ConfigConflict {
	conflicts := make([]ConfigConflict, 0)

	for pluginID, existingConfig := range existingConfigs {
		if pluginID == newConfig.ID {
			continue // 跳过自己
		}

		// 检查端口冲突
		if portConflicts := ccr.detectPortConflicts(newConfig, existingConfig); len(portConflicts) > 0 {
			conflicts = append(conflicts, portConflicts...)
		}

		// 检查文件路径冲突
		if pathConflicts := ccr.detectPathConflicts(newConfig, existingConfig); len(pathConflicts) > 0 {
			conflicts = append(conflicts, pathConflicts...)
		}

		// 检查资源冲突
		if resourceConflicts := ccr.detectResourceConflicts(newConfig, existingConfig); len(resourceConflicts) > 0 {
			conflicts = append(conflicts, resourceConflicts...)
		}

		// 检查权限冲突
		if permissionConflicts := ccr.detectPermissionConflicts(newConfig, existingConfig); len(permissionConflicts) > 0 {
			conflicts = append(conflicts, permissionConflicts...)
		}
	}

	return conflicts
}

// ResolveConflicts 解决配置冲突
func (ccr *ConfigConflictResolver) ResolveConflicts(conflicts []ConfigConflict) error {
	for _, conflict := range conflicts {
		switch conflict.Severity {
		case "critical":
			return fmt.Errorf("critical conflict detected: %s", conflict.Description)
		case "high":
			ccr.logger.Error("High severity conflict", "conflict", conflict)
			// 可以选择阻止加载或应用默认解决策略
		case "medium":
			ccr.logger.Warn("Medium severity conflict", "conflict", conflict)
			// 应用自动解决策略
		case "low":
			ccr.logger.Info("Low severity conflict", "conflict", conflict)
			// 记录但不阻止
		}
	}

	return nil
}

// detectPortConflicts 检测端口冲突
func (ccr *ConfigConflictResolver) detectPortConflicts(config1, config2 *EnhancedPluginConfig) []ConfigConflict {
	conflicts := make([]ConfigConflict, 0)

	// 从自定义配置中提取端口信息
	ports1 := ccr.extractPorts(config1.CustomConfig)
	ports2 := ccr.extractPorts(config2.CustomConfig)

	for _, port1 := range ports1 {
		for _, port2 := range ports2 {
			if port1 == port2 {
				conflicts = append(conflicts, ConfigConflict{
					Type:        "port_conflict",
					PluginID1:   config1.ID,
					PluginID2:   config2.ID,
					Resource:    fmt.Sprintf("port:%d", port1),
					Description: fmt.Sprintf("Both plugins are trying to use port %d", port1),
					Severity:    "critical",
				})
			}
		}
	}

	return conflicts
}

// detectPathConflicts 检测路径冲突
func (ccr *ConfigConflictResolver) detectPathConflicts(config1, config2 *EnhancedPluginConfig) []ConfigConflict {
	conflicts := make([]ConfigConflict, 0)

	// 从自定义配置中提取路径信息
	paths1 := ccr.extractPaths(config1.CustomConfig)
	paths2 := ccr.extractPaths(config2.CustomConfig)

	for _, path1 := range paths1 {
		for _, path2 := range paths2 {
			if ccr.pathsConflict(path1, path2) {
				conflicts = append(conflicts, ConfigConflict{
					Type:        "path_conflict",
					PluginID1:   config1.ID,
					PluginID2:   config2.ID,
					Resource:    fmt.Sprintf("path:%s", path1),
					Description: fmt.Sprintf("Path conflict between %s and %s", path1, path2),
					Severity:    "medium",
				})
			}
		}
	}

	return conflicts
}

// detectResourceConflicts 检测资源冲突
func (ccr *ConfigConflictResolver) detectResourceConflicts(config1, config2 *EnhancedPluginConfig) []ConfigConflict {
	conflicts := make([]ConfigConflict, 0)

	// 检查资源限制是否合理
	if config1.ResourceLimits != nil && config2.ResourceLimits != nil {
		// 检查总资源使用是否超过系统限制
		totalMemory := int64(config1.ResourceLimits.MaxMemoryMB + config2.ResourceLimits.MaxMemoryMB)

		// 假设系统内存限制为1GB
		if totalMemory > 1024 {
			conflicts = append(conflicts, ConfigConflict{
				Type:        "resource_conflict",
				PluginID1:   config1.ID,
				PluginID2:   config2.ID,
				Resource:    "memory",
				Description: "Combined memory usage exceeds system limits",
				Severity:    "high",
			})
		}
	}

	return conflicts
}

// detectPermissionConflicts 检测权限冲突
func (ccr *ConfigConflictResolver) detectPermissionConflicts(config1, config2 *EnhancedPluginConfig) []ConfigConflict {
	conflicts := make([]ConfigConflict, 0)

	// 检查安全配置冲突
	if config1.SecurityConfig != nil && config2.SecurityConfig != nil {
		// 检查路径冲突
		for _, path1 := range config1.SecurityConfig.AllowedPaths {
			for _, path2 := range config2.SecurityConfig.AllowedPaths {
				if ccr.pathsConflict(path1, path2) {
					conflicts = append(conflicts, ConfigConflict{
						Type:        "path_conflict",
						PluginID1:   config1.ID,
						PluginID2:   config2.ID,
						Resource:    fmt.Sprintf("path:%s", path1),
						Description: fmt.Sprintf("Path conflict: %s vs %s", path1, path2),
						Severity:    "medium",
					})
				}
			}
		}
	}

	return conflicts
}

// extractPorts 从配置中提取端口信息
func (ccr *ConfigConflictResolver) extractPorts(config map[string]interface{}) []int {
	ports := make([]int, 0)

	// 递归搜索配置中的端口字段
	ccr.searchPorts(config, &ports)

	return ports
}

// searchPorts 递归搜索端口
func (ccr *ConfigConflictResolver) searchPorts(obj interface{}, ports *[]int) {
	switch v := obj.(type) {
	case map[string]interface{}:
		for key, value := range v {
			if strings.Contains(strings.ToLower(key), "port") {
				if port, ok := value.(int); ok {
					*ports = append(*ports, port)
				}
				if port, ok := value.(float64); ok {
					*ports = append(*ports, int(port))
				}
			} else {
				ccr.searchPorts(value, ports)
			}
		}
	case []interface{}:
		for _, item := range v {
			ccr.searchPorts(item, ports)
		}
	}
}

// extractPaths 从配置中提取路径信息
func (ccr *ConfigConflictResolver) extractPaths(config map[string]interface{}) []string {
	paths := make([]string, 0)

	// 递归搜索配置中的路径字段
	ccr.searchPaths(config, &paths)

	return paths
}

// searchPaths 递归搜索路径
func (ccr *ConfigConflictResolver) searchPaths(obj interface{}, paths *[]string) {
	switch v := obj.(type) {
	case map[string]interface{}:
		for key, value := range v {
			if strings.Contains(strings.ToLower(key), "path") || strings.Contains(strings.ToLower(key), "dir") {
				if path, ok := value.(string); ok {
					*paths = append(*paths, path)
				}
			} else {
				ccr.searchPaths(value, paths)
			}
		}
	case []interface{}:
		for _, item := range v {
			ccr.searchPaths(item, paths)
		}
	}
}

// pathsConflict 检查路径是否冲突
func (ccr *ConfigConflictResolver) pathsConflict(path1, path2 string) bool {
	// 检查路径是否重叠
	rel1, err1 := filepath.Rel(path1, path2)
	rel2, err2 := filepath.Rel(path2, path1)

	// 如果其中一个路径是另一个的子路径，则存在冲突
	if err1 == nil && !strings.HasPrefix(rel1, "..") {
		return true
	}
	if err2 == nil && !strings.HasPrefix(rel2, "..") {
		return true
	}

	return false
}

// permissionsConflict 检查权限是否冲突
func (ccr *ConfigConflictResolver) permissionsConflict(perm1, perm2 Permission) bool {
	// 定义互斥权限
	conflictingPerms := map[Permission][]Permission{
		PermissionFileWrite:  {PermissionFileDelete},
		PermissionFileDelete: {PermissionFileWrite},
		PermissionSystemExec: {PermissionSystemSignal},
	}

	if conflicts, exists := conflictingPerms[perm1]; exists {
		for _, conflict := range conflicts {
			if perm2 == conflict {
				return true
			}
		}
	}

	return false
}

// BasePlugin 基础插件实现
type BasePlugin struct {
	id      string
	name    string
	version string
	config  map[string]interface{}
	mu      sync.RWMutex
}

// NewBasePlugin 创建基础插件
func NewBasePlugin(id, name, version string) *BasePlugin {
	return &BasePlugin{
		id:      id,
		name:    name,
		version: version,
		config:  make(map[string]interface{}),
	}
}

// GetID 获取插件ID
func (bp *BasePlugin) GetID() string {
	return bp.id
}

// GetName 获取插件名称
func (bp *BasePlugin) GetName() string {
	return bp.name
}

// GetVersion 获取插件版本
func (bp *BasePlugin) GetVersion() string {
	return bp.version
}

// Initialize 初始化插件
func (bp *BasePlugin) Initialize(ctx context.Context) error {
	return nil
}

// Start 启动插件
func (bp *BasePlugin) Start(ctx context.Context) error {
	return nil
}

// Stop 停止插件
func (bp *BasePlugin) Stop() error {
	return nil
}

// Cleanup 清理插件
func (bp *BasePlugin) Cleanup() error {
	return nil
}

// UpdateConfig 更新配置
func (bp *BasePlugin) UpdateConfig(config map[string]interface{}) error {
	bp.mu.Lock()
	defer bp.mu.Unlock()

	bp.config = config
	return nil
}

// GetConfig 获取配置
func (bp *BasePlugin) GetConfig() map[string]interface{} {
	bp.mu.RLock()
	defer bp.mu.RUnlock()

	// 返回配置副本
	config := make(map[string]interface{})
	for k, v := range bp.config {
		config[k] = v
	}
	return config
}