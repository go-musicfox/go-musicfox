package config

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPluginConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *PluginConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: &PluginConfig{
				Name:      "test-plugin",
				Type:      "rpc",
				Path:      "/path/to/plugin",
				Enabled:   true,
				AutoStart: true,
				Priority:  50,
				Config: map[string]string{
					"key1": "value1",
					"key2": "value2",
				},
				Resources: ResourceLimits{
					MaxMemory:     512 * 1024 * 1024,
					MaxCPU:       0.5,
					MaxGoroutines: 100,
					Timeout:       30 * time.Second,
				},
						Security: PluginSecurityConfig{
				Sandbox:         true,
				AllowedPaths:    []string{"/tmp"},
				AllowedNetworks: []string{"127.0.0.1"},
				TrustedSources:  []string{"official"},
			},
			},
			wantErr: false,
		},
		{
			name: "empty name",
			config: &PluginConfig{
				Name: "",
				Type: "rpc",
				Path: "/path/to/plugin",
			},
			wantErr: true,
		},
		{
			name: "empty type",
			config: &PluginConfig{
				Name: "test-plugin",
				Type: "",
				Path: "/path/to/plugin",
			},
			wantErr: true,
		},
		{
			name: "empty path",
			config: &PluginConfig{
				Name: "test-plugin",
				Type: "rpc",
				Path: "",
			},
			wantErr: true,
		},
		{
			name: "invalid priority",
			config: &PluginConfig{
				Name:     "test-plugin",
				Type:     "rpc",
				Path:     "/path/to/plugin",
				Priority: -1,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
		}
		})
	}
}

func TestPluginConfig_SetDefaults(t *testing.T) {
	config := &PluginConfig{
		Name: "test-plugin",
		Type: "rpc",
		Path: "/path/to/plugin",
	}

	config.SetDefaults()

	assert.Equal(t, 50, config.Priority)
	assert.NotNil(t, config.Config)
	assert.NotZero(t, config.Resources.MaxMemory)
	assert.NotZero(t, config.Resources.MaxCPU)
	assert.NotNil(t, config.Security.AllowedPaths)
}

func TestPluginConfig_JSONSerialization(t *testing.T) {
	original := &PluginConfig{
		Name:      "test-plugin",
		Type:      "rpc",
		Path:      "/path/to/plugin",
		Enabled:   true,
		AutoStart: false,
		Priority:  75,
		Config: map[string]string{
			"timeout": "30",
			"retries": "3",
		},
		Resources: ResourceLimits{
			MaxMemory:     256 * 1024 * 1024,
			MaxCPU:       0.25,
			MaxGoroutines: 50,
			Timeout:       15 * time.Second,
		},
		Security: PluginSecurityConfig{
			Sandbox:         false,
			AllowedPaths:    []string{"/tmp"},
			AllowedNetworks: []string{"127.0.0.1"},
			TrustedSources:  []string{"official"},
		},
	}

	// 序列化
	data, err := json.Marshal(original)
	require.NoError(t, err)

	// 反序列化
	var deserialized PluginConfig
	err = json.Unmarshal(data, &deserialized)
	require.NoError(t, err)

	// 验证
	assert.Equal(t, original.Name, deserialized.Name)
	assert.Equal(t, original.Type, deserialized.Type)
	assert.Equal(t, original.Path, deserialized.Path)
	assert.Equal(t, original.Enabled, deserialized.Enabled)
	assert.Equal(t, original.AutoStart, deserialized.AutoStart)
	assert.Equal(t, original.Priority, deserialized.Priority)
	assert.Equal(t, original.Config, deserialized.Config)

	// 验证资源限制
	assert.Equal(t, original.Resources.MaxMemory, deserialized.Resources.MaxMemory)
	assert.Equal(t, original.Resources.MaxCPU, deserialized.Resources.MaxCPU)
	assert.Equal(t, original.Resources.MaxGoroutines, deserialized.Resources.MaxGoroutines)
	assert.Equal(t, original.Resources.Timeout, deserialized.Resources.Timeout)

	// 验证安全配置
	assert.Equal(t, original.Security.Sandbox, deserialized.Security.Sandbox)
	assert.Equal(t, original.Security.AllowedPaths, deserialized.Security.AllowedPaths)
	assert.Equal(t, original.Security.AllowedNetworks, deserialized.Security.AllowedNetworks)
	assert.Equal(t, original.Security.TrustedSources, deserialized.Security.TrustedSources)
}

func TestResourceLimits_Validate(t *testing.T) {
	tests := []struct {
		name      string
		limits    *ResourceLimits
		wantErr   bool
	}{
		{
			name: "valid limits",
			limits: &ResourceLimits{
				MaxMemory:     100 * 1024 * 1024,
				MaxCPU:       0.5,
				MaxGoroutines: 50,
				Timeout:       30 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "negative memory",
			limits: &ResourceLimits{
				MaxMemory: -1,
			},
			wantErr: true,
		},
		{
			name: "invalid CPU percent",
			limits: &ResourceLimits{
				MaxCPU: 1.5,
			},
			wantErr: true,
		},
		{
			name: "zero goroutines",
			limits: &ResourceLimits{
				MaxGoroutines: 0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.limits.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestResourceLimits_SetDefaults(t *testing.T) {
	limits := &ResourceLimits{}
	limits.SetDefaults()

	assert.Equal(t, int64(100*1024*1024), limits.MaxMemory)
	assert.Equal(t, 0.5, limits.MaxCPU)
	assert.Equal(t, 100, limits.MaxGoroutines)
	assert.Equal(t, 30*time.Second, limits.Timeout)
}

func TestPluginSecurityConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *PluginSecurityConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: &PluginSecurityConfig{
				Sandbox:         true,
				AllowedPaths:    []string{"/tmp"},
				AllowedNetworks: []string{"127.0.0.1"},
				TrustedSources:  []string{"official"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.config.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("PluginSecurityConfig.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPluginSecurityConfig_SetDefaults(t *testing.T) {
	config := &PluginSecurityConfig{}
	config.SetDefaults()

	assert.NotNil(t, config.AllowedPaths)
	assert.NotNil(t, config.AllowedNetworks)
	assert.NotNil(t, config.TrustedSources)
}