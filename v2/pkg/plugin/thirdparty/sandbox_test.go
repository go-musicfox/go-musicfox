// Package thirdparty 沙箱测试
package thirdparty

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewSandbox 测试创建沙箱
func TestNewSandbox(t *testing.T) {
	config := &SandboxConfig{
		Enabled:          true,
		AllowedPaths:     []string{"/tmp"},
		AllowedNetworks:  []string{"localhost"},
		AllowedSyscalls:  []string{"read", "write"},
		TrustedSources:   []string{},
		IsolationLevel:   IsolationLevelBasic,
		NetworkAccess:    false,
		FileSystemAccess: false,
	}

	sandbox, err := NewSandbox(config)
	assert.NoError(t, err)
	assert.NotNil(t, sandbox)
	assert.Equal(t, config, sandbox.config)
	assert.False(t, sandbox.active)

	// 测试nil配置
	_, err = NewSandbox(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be nil")
}

// TestSandboxActivation 测试沙箱激活和停用
func TestSandboxActivation(t *testing.T) {
	config := &SandboxConfig{
		Enabled:          true,
		IsolationLevel:   IsolationLevelBasic,
		NetworkAccess:    false,
		FileSystemAccess: false,
	}

	sandbox, err := NewSandbox(config)
	require.NoError(t, err)

	// 测试激活
	err = sandbox.Activate()
	assert.NoError(t, err)
	assert.True(t, sandbox.active)

	// 测试重复激活
	err = sandbox.Activate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already active")

	// 测试停用
	err = sandbox.Deactivate()
	assert.NoError(t, err)
	assert.False(t, sandbox.active)

	// 测试重复停用
	err = sandbox.Deactivate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not active")
}

// TestSandboxExecution 测试沙箱执行
func TestSandboxExecution(t *testing.T) {
	config := &SandboxConfig{
		Enabled:        true,
		IsolationLevel: IsolationLevelBasic,
	}

	sandbox, err := NewSandbox(config)
	require.NoError(t, err)

	// 激活沙箱
	err = sandbox.Activate()
	require.NoError(t, err)

	// 测试成功执行
	ctx := context.Background()
	result, err := sandbox.Execute(ctx, func() (interface{}, error) {
		return "success", nil
	})
	assert.NoError(t, err)
	assert.Equal(t, "success", result)

	// 测试执行错误
	_, err = sandbox.Execute(ctx, func() (interface{}, error) {
		return nil, assert.AnError
	})
	assert.Error(t, err)

	// 测试在未激活的沙箱中执行
	sandbox.Deactivate()
	_, err = sandbox.Execute(ctx, func() (interface{}, error) {
		return "test", nil
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not active")
}

// TestSandboxTimeout 测试沙箱超时
func TestSandboxTimeout(t *testing.T) {
	config := &SandboxConfig{
		Enabled:        true,
		IsolationLevel: IsolationLevelBasic,
	}

	sandbox, err := NewSandbox(config)
	require.NoError(t, err)

	err = sandbox.Activate()
	require.NoError(t, err)

	// 测试超时执行
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err = sandbox.Execute(ctx, func() (interface{}, error) {
		// 模拟长时间运行的函数
		time.Sleep(200 * time.Millisecond)
		return "should timeout", nil
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "timeout")
}

// TestFileAccessValidation 测试文件访问验证
func TestFileAccessValidation(t *testing.T) {
	config := &SandboxConfig{
		Enabled:          true,
		FileSystemAccess: true,
		AllowedPaths:     []string{"/tmp", "/var/tmp"},
		IsolationLevel:   IsolationLevelBasic,
	}

	sandbox, err := NewSandbox(config)
	require.NoError(t, err)

	// 测试允许的路径
	err = sandbox.ValidateFileAccess("/tmp/test.txt")
	assert.NoError(t, err)

	err = sandbox.ValidateFileAccess("/var/tmp/data.log")
	assert.NoError(t, err)

	// 测试不允许的路径
	err = sandbox.ValidateFileAccess("/etc/passwd")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "denied")

	err = sandbox.ValidateFileAccess("/home/user/secret.txt")
	assert.Error(t, err)

	// 测试禁用文件系统访问
	config.FileSystemAccess = false
	sandbox.UpdateConfig(config)

	err = sandbox.ValidateFileAccess("/tmp/test.txt")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "access denied")
}

// TestNetworkAccessValidation 测试网络访问验证
func TestNetworkAccessValidation(t *testing.T) {
	config := &SandboxConfig{
		Enabled:         true,
		NetworkAccess:   true,
		AllowedNetworks: []string{"localhost", "127.0.0.1", "example.com"},
		IsolationLevel:  IsolationLevelBasic,
	}

	sandbox, err := NewSandbox(config)
	require.NoError(t, err)

	// 测试允许的网络
	err = sandbox.ValidateNetworkAccess("localhost", 8080)
	assert.NoError(t, err)

	err = sandbox.ValidateNetworkAccess("127.0.0.1", 3000)
	assert.NoError(t, err)

	err = sandbox.ValidateNetworkAccess("example.com", 443)
	assert.NoError(t, err)

	// 测试不允许的网络
	err = sandbox.ValidateNetworkAccess("malicious.com", 80)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "denied")

	// 测试禁用网络访问
	config.NetworkAccess = false
	sandbox.UpdateConfig(config)

	err = sandbox.ValidateNetworkAccess("localhost", 8080)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "access denied")
}

// TestSecurityViolations 测试安全违规记录
func TestSecurityViolations(t *testing.T) {
	config := &SandboxConfig{
		Enabled:          true,
		FileSystemAccess: false,
		NetworkAccess:    false,
		IsolationLevel:   IsolationLevelStrict,
	}

	sandbox, err := NewSandbox(config)
	require.NoError(t, err)

	// 初始状态应该没有违规记录
	violations := sandbox.GetViolations()
	assert.Empty(t, violations)

	// 触发文件访问违规
	err = sandbox.ValidateFileAccess("/etc/passwd")
	assert.Error(t, err)

	// 触发网络访问违规
	err = sandbox.ValidateNetworkAccess("example.com", 80)
	assert.Error(t, err)

	// 检查违规记录
	violations = sandbox.GetViolations()
	assert.Len(t, violations, 2)

	// 验证违规记录内容
	fileViolation := violations[0]
	assert.Equal(t, ViolationTypeFileAccess, fileViolation.Type)
	assert.True(t, fileViolation.Blocked)
	assert.NotZero(t, fileViolation.Timestamp)

	networkViolation := violations[1]
	assert.Equal(t, ViolationTypeNetworkAccess, networkViolation.Type)
	assert.True(t, networkViolation.Blocked)

	// 测试清除违规记录
	sandbox.ClearViolations()
	violations = sandbox.GetViolations()
	assert.Empty(t, violations)
}

// TestIsolationLevels 测试隔离级别
func TestIsolationLevels(t *testing.T) {
	testCases := []struct {
		name  string
		level IsolationLevel
	}{
		{"None", IsolationLevelNone},
		{"Basic", IsolationLevelBasic},
		{"Strict", IsolationLevelStrict},
		{"Complete", IsolationLevelComplete},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := &SandboxConfig{
				Enabled:        true,
				IsolationLevel: tc.level,
			}

			sandbox, err := NewSandbox(config)
			assert.NoError(t, err)

			err = sandbox.Activate()
			assert.NoError(t, err)

			// 验证隔离级别设置正确
			assert.Equal(t, tc.level, sandbox.config.IsolationLevel)

			sandbox.Deactivate()
		})
	}
}

// TestSandboxStats 测试沙箱统计信息
func TestSandboxStats(t *testing.T) {
	config := &SandboxConfig{
		Enabled:          true,
		IsolationLevel:   IsolationLevelBasic,
		NetworkAccess:    true,
		FileSystemAccess: true,
	}

	sandbox, err := NewSandbox(config)
	require.NoError(t, err)

	// 激活沙箱
	err = sandbox.Activate()
	require.NoError(t, err)

	// 触发一些违规
	sandbox.ValidateFileAccess("/forbidden/path")
	sandbox.ValidateNetworkAccess("blocked.com", 80)

	// 获取统计信息
	stats := sandbox.GetStats()
	assert.NotNil(t, stats)

	// 验证统计信息
	assert.Equal(t, true, stats["active"])
	assert.Equal(t, "basic", stats["isolation_level"])
	assert.Equal(t, true, stats["network_access"])
	assert.Equal(t, true, stats["filesystem_access"])
	assert.Greater(t, stats["violation_count"], 0)

	// 验证违规类型统计
	violationTypes, ok := stats["violation_types"].(map[string]int)
	assert.True(t, ok)
	assert.Greater(t, violationTypes["file_access"], 0)
	assert.Greater(t, violationTypes["network_access"], 0)
}

// TestSandboxCleanup 测试沙箱清理
func TestSandboxCleanup(t *testing.T) {
	config := &SandboxConfig{
		Enabled:        true,
		IsolationLevel: IsolationLevelBasic,
	}

	sandbox, err := NewSandbox(config)
	require.NoError(t, err)

	// 激活沙箱并产生一些违规记录
	err = sandbox.Activate()
	require.NoError(t, err)

	sandbox.ValidateFileAccess("/forbidden")
	assert.NotEmpty(t, sandbox.GetViolations())

	// 测试清理
	err = sandbox.Cleanup()
	assert.NoError(t, err)
	assert.False(t, sandbox.active)
	assert.Nil(t, sandbox.violations)

	// 测试清理未激活的沙箱
	err = sandbox.Cleanup()
	assert.NoError(t, err)
}

// TestViolationTypeString 测试违规类型字符串表示
func TestViolationTypeString(t *testing.T) {
	testCases := []struct {
		violationType ViolationType
		expected      string
	}{
		{ViolationTypeFileAccess, "file_access"},
		{ViolationTypeNetworkAccess, "network_access"},
		{ViolationTypeSyscall, "syscall"},
		{ViolationTypeResourceLimit, "resource_limit"},
		{ViolationTypePermission, "permission"},
		{ViolationTypeUnsafeOperation, "unsafe_operation"},
		{ViolationType(999), "unknown"},
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.violationType.String())
		})
	}
}

// TestSeverityString 测试严重程度字符串表示
func TestSeverityString(t *testing.T) {
	testCases := []struct {
		severity Severity
		expected string
	}{
		{SeverityLow, "low"},
		{SeverityMedium, "medium"},
		{SeverityHigh, "high"},
		{SeverityCritical, "critical"},
		{Severity(999), "unknown"},
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.severity.String())
		})
	}
}

// TestIsolationLevelString 测试隔离级别字符串表示
func TestIsolationLevelString(t *testing.T) {
	testCases := []struct {
		level    IsolationLevel
		expected string
	}{
		{IsolationLevelNone, "none"},
		{IsolationLevelBasic, "basic"},
		{IsolationLevelStrict, "strict"},
		{IsolationLevelComplete, "complete"},
		{IsolationLevel(999), "unknown"},
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.level.String())
		})
	}
}