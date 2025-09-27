package main

import (
	"testing"
)

// TestNeteasePlugin 测试网易云插件基本功能
func TestNeteasePlugin(t *testing.T) {
	// 创建插件实例
	plugin := NewNeteasePlugin()
	if plugin == nil {
		t.Fatal("Failed to create Netease plugin")
	}

	// 测试插件信息
	info := plugin.GetInfo()
	if info.ID != "netease-music" {
		t.Errorf("Expected plugin ID 'netease-music', got '%s'", info.ID)
	}

	if info.Name != "Netease Music" {
		t.Errorf("Expected plugin name 'Netease Music', got '%s'", info.Name)
	}

	// 测试支持的功能
	features := plugin.GetSupportedFeatures()
	t.Logf("Plugin supports %d features", len(features))

	// 测试服务信息
	serviceInfo := plugin.GetServiceInfo()
	if serviceInfo == nil {
		t.Error("Service info should not be nil")
	}
}

// TestNeteasePluginLoginMethods 测试登录方法
func TestNeteasePluginLoginMethods(t *testing.T) {
	plugin := NewNeteasePlugin()

	// 测试获取登录方法
	methods := plugin.GetLoginMethods()
	expectedMethods := []string{"phone", "email", "cookie", "qr"}

	if len(methods) != len(expectedMethods) {
		t.Errorf("Expected %d login methods, got %d", len(expectedMethods), len(methods))
	}

	for _, expected := range expectedMethods {
		found := false
		for _, method := range methods {
			if method == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected login method '%s' not found", expected)
		}
	}

	// 测试基本凭据验证
	err := plugin.ValidateCredentials(map[string]string{
		"type":     "phone",
		"phone":    "13800138000",
		"password": "password123",
	})
	if err != nil {
		t.Errorf("Valid phone credentials should not error: %v", err)
	}

	err = plugin.ValidateCredentials(map[string]string{
		"type":     "email",
		"email":    "test@example.com",
		"password": "password123",
	})
	if err != nil {
		t.Errorf("Valid email credentials should not error: %v", err)
	}

	err = plugin.ValidateCredentials(map[string]string{
		"type": "qr",
	})
	if err != nil {
		t.Errorf("QR credentials should not error: %v", err)
	}
}