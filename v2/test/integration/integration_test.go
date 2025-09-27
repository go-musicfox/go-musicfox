package integration

import (
	"testing"
)

// TestIntegrationSuite 运行所有集成测试
func TestIntegrationSuite(t *testing.T) {
	t.Log("Starting go-musicfox microkernel plugin architecture integration tests...")

	// 1. 微内核集成测试
	t.Run("KernelIntegration", TestKernelIntegration)

	// 2. 插件系统集成测试
	t.Run("PluginSystemIntegration", TestPluginSystemIntegration)

	// 3. 阶段2集成测试 - 音频和播放列表插件
	t.Run("Stage2Integration", TestStage2Integration)

	// 4. 性能和稳定性测试
	t.Run("PerformanceStability", TestPerformanceStability)

	t.Log("All integration tests completed successfully!")
}