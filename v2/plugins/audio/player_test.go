package audio

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestPlayerFactory_NewPlayerFactory 测试播放器工厂创建
func TestPlayerFactory_NewPlayerFactory(t *testing.T) {
	factory := NewPlayerFactory()

	assert.NotNil(t, factory)
	assert.NotNil(t, factory.backends)

	// 检查内置后端是否注册
	backends := factory.GetAvailableBackends()
	assert.NotEmpty(t, backends)
	assert.Contains(t, backends, "beep") // beep应该总是可用
}

// TestPlayerFactory_RegisterBackend 测试注册播放器后端
func TestPlayerFactory_RegisterBackend(t *testing.T) {
	factory := NewPlayerFactory()

	// 注册自定义后端
	err := factory.RegisterBackend("test", func(config map[string]interface{}) (PlayerBackend, error) {
		mockPlayer := NewMockPlayerBackend()
		mockPlayer.On("IsAvailable").Return(true)
		return mockPlayer, nil
	})

	assert.NoError(t, err)

	// 测试空名称
	err = factory.RegisterBackend("", func(config map[string]interface{}) (PlayerBackend, error) {
		return nil, nil
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "backend name cannot be empty")

	// 测试nil创建函数
	err = factory.RegisterBackend("test2", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "backend creator cannot be nil")
}

// TestPlayerFactory_CreatePlayer 测试创建播放器
func TestPlayerFactory_CreatePlayer(t *testing.T) {
	factory := NewPlayerFactory()

	// 注册测试后端
	mockPlayer := NewMockPlayerBackend()
	mockPlayer.On("IsAvailable").Return(true)

	factory.RegisterBackend("test", func(config map[string]interface{}) (PlayerBackend, error) {
		return mockPlayer, nil
	})

	// 创建播放器
	player, err := factory.CreatePlayer("test", nil)
	assert.NoError(t, err)
	assert.NotNil(t, player)
	assert.Equal(t, mockPlayer, player)

	// 测试不存在的后端
	_, err = factory.CreatePlayer("nonexistent", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "player backend 'nonexistent' not found")
}

// TestPlayerFactory_CreatePlayerUnavailable 测试创建不可用的播放器
func TestPlayerFactory_CreatePlayerUnavailable(t *testing.T) {
	factory := NewPlayerFactory()

	// 注册不可用的后端
	mockPlayer := NewMockPlayerBackend()
	mockPlayer.On("IsAvailable").Return(false)

	factory.RegisterBackend("unavailable", func(config map[string]interface{}) (PlayerBackend, error) {
		return mockPlayer, nil
	})

	// 尝试创建不可用的播放器
	_, err := factory.CreatePlayer("unavailable", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "is not available on this system")
}

// TestPlayerFactory_GetBackendInfo 测试获取后端信息
func TestPlayerFactory_GetBackendInfo(t *testing.T) {
	factory := NewPlayerFactory()

	// 注册测试后端
	mockPlayer := NewMockPlayerBackend()
	mockPlayer.name = "Test Player"
	mockPlayer.version = "2.0.0"
	mockPlayer.supportedFormats = []string{"mp3", "wav", "flac"}
	mockPlayer.On("IsAvailable").Return(true)

	factory.RegisterBackend("test", func(config map[string]interface{}) (PlayerBackend, error) {
		return mockPlayer, nil
	})

	// 获取后端信息
	info, err := factory.GetBackendInfo("test")
	assert.NoError(t, err)
	assert.NotNil(t, info)
	assert.Equal(t, "Test Player", info["name"])
	assert.Equal(t, "2.0.0", info["version"])
	assert.Equal(t, []string{"mp3", "wav", "flac"}, info["supported_formats"])
	assert.Equal(t, true, info["available"])

	// 测试不存在的后端
	_, err = factory.GetBackendInfo("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "backend 'nonexistent' not found")
}

// TestBasePlayer_NewBasePlayer 测试基础播放器创建
func TestBasePlayer_NewBasePlayer(t *testing.T) {
	formats := []string{"mp3", "wav"}
	player := NewBasePlayerWithInfo("Test Player", "1.0.0", formats)

	assert.NotNil(t, player)
	assert.Equal(t, "Test Player", player.GetName())
	assert.Equal(t, "1.0.0", player.GetVersion())
	assert.Equal(t, formats, player.GetSupportedFormats())
	assert.Equal(t, 0.8, player.volume) // 默认音量80%
	assert.False(t, player.IsPlaying())
}

// TestBasePlayer_SetVolume 测试基础播放器音量设置
func TestBasePlayer_SetVolume(t *testing.T) {
	player := NewBasePlayerWithInfo("Test", "1.0.0", []string{"mp3"})

	// 设置有效音量
	err := player.SetVolume(0.5)
	assert.NoError(t, err)
	volume, err := player.GetVolume()
	assert.NoError(t, err)
	assert.Equal(t, 0.5, volume)

	// 测试无效音量范围
	err = player.SetVolume(-0.1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "volume must be between 0 and 1")

	err = player.SetVolume(1.1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "volume must be between 0 and 1")
}

// TestBasePlayer_Position 测试基础播放器位置管理
func TestBasePlayer_Position(t *testing.T) {
	player := NewBasePlayerWithInfo("Test", "1.0.0", []string{"mp3"})

	// 初始位置应该为0
	pos, err := player.GetPosition()
	assert.NoError(t, err)
	assert.Equal(t, time.Duration(0), pos)

	// 设置位置
	newPos := 30 * time.Second
	player.setPosition(newPos)
	pos, err = player.GetPosition()
	assert.NoError(t, err)
	assert.Equal(t, newPos, pos)
}

// TestBasePlayer_Duration 测试基础播放器时长管理
func TestBasePlayer_Duration(t *testing.T) {
	player := NewBasePlayerWithInfo("Test", "1.0.0", []string{"mp3"})

	// 初始时长应该为0
	dur, err := player.GetDuration()
	assert.NoError(t, err)
	assert.Equal(t, time.Duration(0), dur)

	// 设置时长
	newDur := 3 * time.Minute
	player.setDuration(newDur)
	dur, err = player.GetDuration()
	assert.NoError(t, err)
	assert.Equal(t, newDur, dur)
}

// TestBasePlayer_PlayingState 测试基础播放器播放状态
func TestBasePlayer_PlayingState(t *testing.T) {
	player := NewBasePlayerWithInfo("Test", "1.0.0", []string{"mp3"})

	// 初始状态应该是未播放
	assert.False(t, player.IsPlaying())

	// 设置为播放状态
	player.setPlaying(true)
	assert.True(t, player.IsPlaying())

	// 设置为停止状态
	player.setPlaying(false)
	assert.False(t, player.IsPlaying())
}

// TestBeepPlayer_NewBeepPlayer 测试Beep播放器创建
func TestBeepPlayer_NewBeepPlayer(t *testing.T) {
	config := map[string]interface{}{
		"sample_rate": 48000,
	}

	player := NewBeepPlayer(config)
	assert.NotNil(t, player)
	assert.Equal(t, "Beep Player", player.GetName())
	assert.Equal(t, "1.0.0", player.GetVersion())
	assert.True(t, player.IsAvailable()) // Beep应该总是可用

	formats := player.GetSupportedFormats()
	assert.Contains(t, formats, "mp3")
	assert.Contains(t, formats, "wav")
	assert.Contains(t, formats, "flac")
	assert.Contains(t, formats, "ogg")
}

// TestBeepPlayer_Initialize 测试Beep播放器初始化
func TestBeepPlayer_Initialize(t *testing.T) {
	player := NewBeepPlayer(nil)

	// 注意：实际的初始化可能会失败，因为需要音频设备
	// 这里主要测试接口调用不会panic
	err := player.Initialize(context.Background())
	// 不检查错误，因为在CI环境中可能没有音频设备
	_ = err

	// 清理
	player.Cleanup()
}

// TestMPVPlayer_NewMPVPlayer 测试MPV播放器创建
func TestMPVPlayer_NewMPVPlayer(t *testing.T) {
	config := map[string]interface{}{
		"extra_args": []string{"--volume=50"},
	}

	player := NewMPVPlayer(config)
	assert.NotNil(t, player)
	assert.Equal(t, "MPV Player", player.GetName())
	assert.Equal(t, "1.0.0", player.GetVersion())

	formats := player.GetSupportedFormats()
	assert.Contains(t, formats, "mp3")
	assert.Contains(t, formats, "wav")
	assert.Contains(t, formats, "flac")
	assert.Contains(t, formats, "ogg")
	assert.Contains(t, formats, "m4a")
	assert.Contains(t, formats, "aac")
	assert.Contains(t, formats, "wma")
	assert.Contains(t, formats, "ape")
}

// TestMPVPlayer_IsAvailable 测试MPV播放器可用性检查
func TestMPVPlayer_IsAvailable(t *testing.T) {
	player := NewMPVPlayer(nil)

	// MPV的可用性取决于系统是否安装了mpv
	available := player.IsAvailable()
	// 不做断言，因为CI环境可能没有安装mpv
	_ = available
}

// BenchmarkPlayerFactory_CreatePlayer 基准测试播放器创建性能
func BenchmarkPlayerFactory_CreatePlayer(b *testing.B) {
	factory := NewPlayerFactory()

	// 注册快速创建的测试后端
	factory.RegisterBackend("benchmark", func(config map[string]interface{}) (PlayerBackend, error) {
		mockPlayer := NewMockPlayerBackend()
		mockPlayer.On("IsAvailable").Return(true)
		return mockPlayer, nil
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		player, err := factory.CreatePlayer("benchmark", nil)
		if err != nil {
			b.Fatal(err)
		}
		_ = player
	}
}

// BenchmarkBasePlayer_SetVolume 基准测试音量设置性能
func BenchmarkBasePlayer_SetVolume(b *testing.B) {
	player := NewBasePlayerWithInfo("Benchmark", "1.0.0", []string{"mp3"})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		volume := float64(i%100) / 100.0
		err := player.SetVolume(volume)
		if err != nil {
			b.Fatal(err)
		}
	}
}
