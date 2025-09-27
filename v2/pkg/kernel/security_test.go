package kernel

import (
	"testing"
	"time"
	"context"
	"log/slog"
	"os"

	"github.com/knadh/koanf/v2"
)

// MockSandbox 模拟沙箱实现
type MockSandbox struct {
	id       string
	pluginID string
	status   SandboxStatus
	resourceUsage *ResourceUsage
}

func (m *MockSandbox) GetID() string {
	return m.id
}

func (m *MockSandbox) GetPluginID() string {
	return m.pluginID
}

func (m *MockSandbox) GetStatus() SandboxStatus {
	return m.status
}

func (m *MockSandbox) Start(ctx context.Context) error {
	m.status = SandboxStatusRunning
	return nil
}

func (m *MockSandbox) Stop() error {
	m.status = SandboxStatusStopped
	return nil
}

func (m *MockSandbox) Execute(command string, args []string) ([]byte, error) {
	return []byte("mock output"), nil
}

func (m *MockSandbox) GetResourceUsage() (*ResourceUsage, error) {
	if m.resourceUsage == nil {
		m.resourceUsage = &ResourceUsage{
			MemoryUsed:    1024 * 1024, // 1MB
			CPUUsage:      10.5,        // 10.5%
			NetworkIORate: 1024,        // 1KB/s
			LastUpdated:   time.Now(),
		}
	}
	return m.resourceUsage, nil
}

func (m *MockSandbox) SetResourceLimits(limits *ResourceLimits) error {
	return nil
}

func (m *MockSandbox) Cleanup() error {
	m.status = SandboxStatusStopped
	return nil
}

// TestNewSecurityManager 测试安全管理器创建
func TestNewSecurityManager(t *testing.T) {
	ctx := context.Background()
	config := koanf.New(".")
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	sm := NewSecurityManager(config, logger)
	if sm == nil {
		t.Fatal("NewSecurityManager returned nil")
	}

	// 初始化安全管理器
	if err := sm.Initialize(ctx); err != nil {
		t.Errorf("Initialize failed: %v", err)
	}
}

// TestVerifyPluginSignature 测试插件签名验证
func TestVerifyPluginSignature(t *testing.T) {
	ctx := context.Background()
	config := koanf.New(".")
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	sm := NewSecurityManager(config, logger)

	// 初始化安全管理器
	if err := sm.Initialize(ctx); err != nil {
		t.Errorf("Initialize failed: %v", err)
		return
	}

	// 测试空路径
	err := sm.VerifySignature("", []byte(""))
	if err == nil {
		t.Error("Expected error for empty plugin path")
	}

	// 测试不存在的文件
	err = sm.VerifySignature("/nonexistent/plugin.so", []byte("signature"))
	if err == nil {
		t.Error("Expected error for nonexistent plugin")
	}
}

// TestCreateSandbox 测试沙箱创建
func TestCreateSandbox(t *testing.T) {
	ctx := context.Background()
	config := koanf.New(".")
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	sm := NewSecurityManager(config, logger)

	// 初始化安全管理器
	if err := sm.Initialize(ctx); err != nil {
		t.Errorf("Initialize failed: %v", err)
		return
	}

	limits := ResourceLimits{
		MaxMemory:     1024 * 1024 * 100, // 100MB
		MaxCPUPercent: 50.0,
		MaxNetworkIO:  1024 * 1024, // 1MB/s
	}

	sandbox, err := sm.CreateSandbox("test-plugin", limits)
	if err != nil {
		t.Errorf("CreateSandbox failed: %v", err)
		return
	}

	if sandbox == nil {
		t.Error("CreateSandbox returned nil sandbox")
		return
	}

	if sandbox.GetID() == "" {
		t.Error("CreateSandbox returned sandbox with empty ID")
	}

	// 验证沙箱是否存在
	retrievedSandbox, err := sm.GetSandbox(sandbox.GetID())
	if err != nil {
		t.Errorf("GetSandbox failed: %v", err)
	}
	if retrievedSandbox == nil {
		t.Error("Created sandbox not found")
	}
}

// TestPermissionManagement 测试权限管理
func TestPermissionManagement(t *testing.T) {
	ctx := context.Background()
	config := koanf.New(".")
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	sm := NewSecurityManager(config, logger)

	// 初始化安全管理器
	if err := sm.Initialize(ctx); err != nil {
		t.Errorf("Initialize failed: %v", err)
		return
	}

	pluginID := "test-plugin"
	resource := "file"
	action := "read"

	// 测试添加权限
	err := sm.GrantPermission(pluginID, resource, []string{action})
	if err != nil {
		t.Errorf("GrantPermission failed: %v", err)
	}

	// 测试检查权限
	hasPermission, err := sm.CheckPermission(pluginID, resource, action)
	if err != nil {
		t.Errorf("CheckPermission failed: %v", err)
	}
	if !hasPermission {
		t.Error("Plugin should have the granted permission")
	}

	// 测试撤销权限
	err = sm.RevokePermission(pluginID, resource, []string{action})
	if err != nil {
		t.Errorf("RevokePermission failed: %v", err)
	}

	// 验证权限已撤销
	hasPermission, err = sm.CheckPermission(pluginID, resource, action)
	if err != nil {
		t.Errorf("CheckPermission failed: %v", err)
	}
	if hasPermission {
		t.Error("Plugin should not have the revoked permission")
	}
}

// TestACLManagement 测试访问控制列表管理
func TestACLManagement(t *testing.T) {
	ctx := context.Background()
	config := koanf.New(".")
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	sm := NewSecurityManager(config, logger)

	// 初始化安全管理器
	if err := sm.Initialize(ctx); err != nil {
		t.Errorf("Initialize failed: %v", err)
		return
	}

	pluginID := "test-plugin"
	resource := "file:/tmp/test.txt"

	// 测试添加ACL规则
	err := sm.AddACLRule(pluginID, resource, "read", true)
	if err != nil {
		t.Errorf("AddACLRule failed: %v", err)
	}

	// 测试检查ACL
	allowed, err := sm.CheckACL(pluginID, resource, "read")
	if err != nil {
		t.Errorf("CheckACL failed: %v", err)
	}
	if !allowed {
		t.Error("ACL check should allow read access")
	}

	allowed, err = sm.CheckACL(pluginID, resource, "execute")
	if err != nil {
		t.Errorf("CheckACL failed: %v", err)
	}
	if allowed {
		t.Error("ACL check should deny execute access")
	}

	// 测试移除ACL规则
	err = sm.RemoveACLRule(pluginID, resource, "read")
	if err != nil {
		t.Errorf("RemoveACLRule failed: %v", err)
	}

	// 验证规则已移除
	allowed, err = sm.CheckACL(pluginID, resource, "read")
	if err != nil {
		t.Errorf("CheckACL failed: %v", err)
	}
	if allowed {
		t.Error("ACL check should deny access after rule removal")
	}
}

// TestResourceLimits 测试资源限制
func TestResourceLimits(t *testing.T) {
	ctx := context.Background()
	config := koanf.New(".")
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	sm := NewSecurityManager(config, logger)

	// 初始化安全管理器
	if err := sm.Initialize(ctx); err != nil {
		t.Errorf("Initialize failed: %v", err)
		return
	}

	pluginID := "test-plugin"

	limits := ResourceLimits{
		MaxMemory:     1024 * 1024 * 50, // 50MB
		MaxCPUPercent: 30.0,
		MaxNetworkIO:  512 * 1024, // 512KB/s
	}

	// 设置资源限制
	err := sm.SetResourceLimits(pluginID, &limits)
	if err != nil {
		t.Errorf("SetResourceLimits failed: %v", err)
	}

	// 获取资源限制
	retrievedLimits, err := sm.GetResourceLimits(pluginID)
	if err != nil {
		t.Errorf("GetResourceLimits failed: %v", err)
		return
	}

	if retrievedLimits.MaxMemory != limits.MaxMemory {
		t.Errorf("MaxMemory mismatch: expected %d, got %d", limits.MaxMemory, retrievedLimits.MaxMemory)
	}

	if retrievedLimits.MaxCPUPercent != limits.MaxCPUPercent {
		t.Errorf("MaxCPUPercent mismatch: expected %.2f, got %.2f", limits.MaxCPUPercent, retrievedLimits.MaxCPUPercent)
	}
}

// TestSecurityPolicy 测试安全策略
func TestSecurityPolicy(t *testing.T) {
	ctx := context.Background()
	config := koanf.New(".")
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	sm := NewSecurityManager(config, logger)

	// 初始化安全管理器
	if err := sm.Initialize(ctx); err != nil {
		t.Errorf("Initialize failed: %v", err)
		return
	}

	policy := &SecurityPolicy{
		TrustedSources: []string{"trusted.example.com"},
		BlockedPlugins: []string{"malicious-plugin"},
		EnforceSignatureVerification: true,
	}

	// 设置安全策略
	err := sm.UpdateSecurityPolicy(policy)
	if err != nil {
		t.Errorf("UpdateSecurityPolicy failed: %v", err)
	}

	// 获取安全策略
	retrievedPolicy := sm.GetSecurityPolicy()
	if retrievedPolicy == nil {
		t.Error("GetSecurityPolicy returned nil")
		return
	}

	if len(retrievedPolicy.TrustedSources) != 1 {
		t.Errorf("Expected 1 trusted source, got %d", len(retrievedPolicy.TrustedSources))
	}

	if retrievedPolicy.TrustedSources[0] != "trusted.example.com" {
		t.Errorf("Trusted source mismatch: expected %s, got %s", "trusted.example.com", retrievedPolicy.TrustedSources[0])
	}

	// 测试插件阻止检查
	if !sm.IsPluginBlocked("malicious-plugin") {
		t.Error("IsPluginBlocked should return true for blocked plugin")
	}

	if sm.IsPluginBlocked("safe-plugin") {
		t.Error("IsPluginBlocked should return false for non-blocked plugin")
	}
}

// TestSecurityEvents 测试安全事件
func TestSecurityEvents(t *testing.T) {
	// 添加超时机制
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	config := koanf.New(".")
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	sm := NewSecurityManager(config, logger)

	// 初始化安全管理器
	if err := sm.Initialize(ctx); err != nil {
		t.Errorf("Initialize failed: %v", err)
		return
	}

	// 记录安全事件
	event := &SecurityEvent{
		Type:      SecurityEventTypePermissionDenied,
		Level:     SecurityEventLevelWarning,
		PluginID:  "test-plugin",
		Resource:  "file",
		Message:   "Permission denied",
		Timestamp: time.Now(),
	}
	sm.LogSecurityEvent(event)

	// 获取安全事件
	filter := &SecurityEventFilter{
		Limit:  100,
		Offset: 0,
	}
	events, err := sm.GetSecurityEvents(filter)
	if err != nil {
		t.Errorf("GetSecurityEvents failed: %v", err)
		return
	}
	if len(events) == 0 {
		t.Error("GetSecurityEvents returned empty events")
	}

	// 按插件获取事件
	pluginFilter := &SecurityEventFilter{
		PluginIDs: []string{"test-plugin"},
		Limit:     100,
		Offset:    0,
	}
	pluginEvents, err := sm.GetSecurityEvents(pluginFilter)
	if err != nil {
		t.Errorf("GetSecurityEvents for plugin failed: %v", err)
		return
	}
	if len(pluginEvents) == 0 {
		t.Error("GetSecurityEvents for plugin returned empty events")
	}

	// 按级别获取事件
	levelFilter := &SecurityEventFilter{
		Levels: []SecurityEventLevel{SecurityEventLevelWarning},
		Limit:  100,
		Offset: 0,
	}
	levelEvents, err := sm.GetSecurityEvents(levelFilter)
	if err != nil {
		t.Errorf("GetSecurityEvents for level failed: %v", err)
		return
	}
	if len(levelEvents) == 0 {
		t.Error("GetSecurityEvents for level returned empty events")
	}

	// 清除安全事件 - 使用未来时间来清除所有事件
	err = sm.ClearSecurityEvents(time.Now().Add(time.Hour))
	if err != nil {
		t.Errorf("ClearSecurityEvents failed: %v", err)
		return
	}
	events, err = sm.GetSecurityEvents(filter)
	if err != nil {
		t.Errorf("GetSecurityEvents failed: %v", err)
		return
	}
	// ClearSecurityEvents会记录一个新的审计事件，所以应该只剩下1个事件
	if len(events) != 1 {
		t.Errorf("Expected 1 event after clearing (audit event), got %d", len(events))
	}
	// 验证剩余的事件是审计事件
	if len(events) > 0 && events[0].Type != "audit" {
		t.Errorf("Expected remaining event to be audit type, got %s", events[0].Type)
	}
}

// TestEnforceSandbox 测试沙箱强制执行
func TestEnforceSandbox(t *testing.T) {
	// 添加超时机制
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	config := koanf.New(".")
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	sm := NewSecurityManager(config, logger)

	// 初始化安全管理器
	if err := sm.Initialize(ctx); err != nil {
		t.Errorf("Initialize failed: %v", err)
		return
	}

	// 设置资源限制
	limits := ResourceLimits{
		MaxMemory:     1024 * 1024 * 10, // 10MB
		MaxCPUPercent: 20.0,
		MaxNetworkIO:  1024 * 512, // 512KB/s
	}
	sm.SetResourceLimits("test-plugin", &limits)

	// 测试沙箱创建
	sandbox, err := sm.CreateSandbox("test-plugin", limits)
	if err != nil {
		t.Errorf("CreateSandbox failed: %v", err)
		return
	}

	// 测试获取沙箱 - 使用sandbox ID而不是plugin ID
	retrievedSandbox, err := sm.GetSandbox(sandbox.GetID())
	if err != nil {
		t.Errorf("GetSandbox failed: %v", err)
		return
	}

	if retrievedSandbox.GetID() != sandbox.GetID() {
		t.Error("Retrieved sandbox ID does not match created sandbox ID")
	}

	// 测试销毁沙箱 - 使用sandbox ID而不是plugin ID
	err = sm.DestroySandbox(sandbox.GetID())
	if err != nil {
		t.Errorf("DestroySandbox failed: %v", err)
	}
}

// TestResourceMonitoring 测试资源监控
func TestResourceMonitoring(t *testing.T) {
	// 添加超时机制
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	config := koanf.New(".")
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	sm := NewSecurityManager(config, logger)
	pluginID := "test-plugin"

	// 使用ctx避免未使用变量警告
	_ = ctx

	// 测试资源监控
	usage, err := sm.MonitorResources(pluginID)
	if err != nil {
		t.Errorf("MonitorResources failed: %v", err)
		return
	}

	if usage == nil {
		t.Error("MonitorResources returned nil usage")
	}
}

// TestSecurityStatistics 测试安全统计
func TestSecurityStatistics(t *testing.T) {
	// 添加超时机制
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	config := koanf.New(".")
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	sm := NewSecurityManager(config, logger)

	// 使用ctx避免未使用变量警告
	_ = ctx

	// 记录一些安全事件
	event1 := &SecurityEvent{
		Type:      SecurityEventTypePermissionDenied,
		Level:     SecurityEventLevelWarning,
		PluginID:  "plugin1",
		Resource:  "file",
		Message:   "Permission denied",
		Timestamp: time.Now(),
	}
	sm.LogSecurityEvent(event1)

	event2 := &SecurityEvent{
		Type:      SecurityEventTypeResourceLimit,
		Level:     SecurityEventLevelError,
		PluginID:  "plugin2",
		Resource:  "memory",
		Message:   "Memory limit exceeded",
		Timestamp: time.Now(),
	}
	sm.LogSecurityEvent(event2)

	// 获取安全事件来验证记录成功
	filter := &SecurityEventFilter{
		Limit:  100,
		Offset: 0,
	}
	events, err := sm.GetSecurityEvents(filter)
	if err != nil {
		t.Errorf("GetSecurityEvents failed: %v", err)
		return
	}

	if len(events) == 0 {
		t.Error("Expected non-zero events")
	}
}

// TestSafeExecute 测试安全执行
func TestSafeExecute(t *testing.T) {
	// 添加超时机制
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	config := koanf.New(".")
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	sm := NewSecurityManager(config, logger)

	// 初始化安全管理器
	if err := sm.Initialize(ctx); err != nil {
		t.Errorf("Initialize failed: %v", err)
		return
	}

	// 先授予权限
	err := sm.GrantPermission("test-plugin", "test-resource", []string{"read"})
	if err != nil {
		t.Errorf("GrantPermission failed: %v", err)
		return
	}

	// 测试权限检查
	allowed, err := sm.CheckPermission("test-plugin", "test-resource", "read")
	if err != nil {
		t.Errorf("CheckPermission failed: %v", err)
		return
	}
	if !allowed {
		t.Error("Expected permission to be allowed")
	}

	// 添加ACL规则
	err = sm.AddACLRule("test-plugin", "test-resource", "read", true)
	if err != nil {
		t.Errorf("AddACLRule failed: %v", err)
		return
	}

	// 测试ACL检查
	allowed, err = sm.CheckACL("test-plugin", "test-resource", "read")
	if err != nil {
		t.Errorf("CheckACL failed: %v", err)
		return
	}
	if !allowed {
		t.Error("Expected ACL to allow access")
	}
}

// BenchmarkCheckPermission 权限检查性能测试
func BenchmarkCheckPermission(b *testing.B) {
	config := koanf.New(".")
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	sm := NewSecurityManager(config, logger)
	pluginID := "bench-plugin"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sm.CheckPermission(pluginID, "file", "read")
	}
}

// BenchmarkCheckACL ACL检查性能测试
func BenchmarkCheckACL(b *testing.B) {
	config := koanf.New(".")
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	sm := NewSecurityManager(config, logger)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sm.CheckACL("plugin-50", "database", "read")
	}
}