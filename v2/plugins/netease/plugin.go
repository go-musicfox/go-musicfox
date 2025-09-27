package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/go-musicfox/netease-music/service"
	"github.com/go-musicfox/netease-music/util"
	"github.com/telanflow/cookiejar"

	core "github.com/go-musicfox/go-musicfox/v2/pkg/plugin/core"
	plugin "github.com/go-musicfox/go-musicfox/v2/pkg/plugin/music"
)

// NeteasePlugin 网易云音乐插件
type NeteasePlugin struct {
	*plugin.BaseMusicSourcePlugin
	cookieJar *cookiejar.Jar
	user      *UserInfo
	qrUniKey  string
	mu        sync.RWMutex
}

// UserInfo 用户信息
type UserInfo struct {
	UserID   int64  `json:"user_id"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

// NewNeteasePlugin 创建网易云音乐插件实例
func NewNeteasePlugin() *NeteasePlugin {
	info := &plugin.PluginInfo{
		ID:          "netease-music",
		Name:        "Netease Music",
		Version:     "1.0.0",
		Description: "Netease Cloud Music source plugin",
		Author:      "go-musicfox team",
		License:     "MIT",
		Type:        plugin.PluginTypeMusicSource,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	p := &NeteasePlugin{
		BaseMusicSourcePlugin: plugin.NewBaseMusicSourcePlugin(info),
	}

	// 添加支持的功能
	p.AddFeature(plugin.MusicSourceFeatureSearch)
	p.AddFeature(plugin.MusicSourceFeaturePlaylist)
	p.AddFeature(plugin.MusicSourceFeatureUser)
	p.AddFeature(plugin.MusicSourceFeatureLyrics)
	p.AddFeature(plugin.MusicSourceFeatureRecommendation)
	p.AddFeature(plugin.MusicSourceFeatureChart)

	return p
}

// Initialize 初始化插件
func (p *NeteasePlugin) Initialize(pluginCtx core.PluginContext) error {
	if err := p.BaseMusicSourcePlugin.Initialize(pluginCtx); err != nil {
		return err
	}

	// 初始化cookie jar
	dataDir := "/tmp/musicfox" // TODO: 从配置获取
	// 确保目录存在
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}
	
	cookieFile := filepath.Join(dataDir, "netease_cookies")
	// 如果cookie文件不存在，创建一个空文件
	if _, err := os.Stat(cookieFile); os.IsNotExist(err) {
		if file, err := os.Create(cookieFile); err != nil {
			return fmt.Errorf("failed to create cookie file: %w", err)
		} else {
			file.Close()
		}
	}
	
	jar, err := cookiejar.NewFileJar(cookieFile, nil)
	if err != nil {
		return fmt.Errorf("failed to create cookie jar: %w", err)
	}
	if cookieJar, ok := jar.(*cookiejar.Jar); ok {
		p.cookieJar = cookieJar
	} else {
		return fmt.Errorf("failed to cast cookie jar to *cookiejar.Jar")
	}
	util.SetGlobalCookieJar(jar)

	p.UpdateServiceStatus(plugin.ServiceStatusRunning)
	p.UpdateHealthStatus(plugin.HealthStatusHealthy())

	return nil
}

// Start 启动插件
func (p *NeteasePlugin) Start(ctx context.Context) error {
	if err := p.BaseMusicSourcePlugin.Start(); err != nil {
		return err
	}

	// 尝试从cookie恢复登录状态
	if err := p.tryRestoreLogin(); err != nil {
		slog.Warn("Failed to restore login from cookies", "error", err)
	}

	return nil
}

// Stop 停止插件
func (p *NeteasePlugin) Stop(ctx context.Context) error {
	p.UpdateServiceStatus(plugin.ServiceStatusStopped)
	return p.BaseMusicSourcePlugin.Stop()
}

// Cleanup 清理插件资源
func (p *NeteasePlugin) Cleanup(ctx context.Context) error {
	return p.BaseMusicSourcePlugin.Cleanup()
}

// tryRestoreLogin 尝试从cookie恢复登录状态
func (p *NeteasePlugin) tryRestoreLogin() error {
	accountService := &service.UserAccountService{}
	code, resp := accountService.AccountInfo()
	if code != 200 {
		return fmt.Errorf("failed to get account info: code %f", code)
	}

	// TODO: 解析用户信息
	_ = resp
	return nil
}

// GetSupportedFeatures 获取支持的功能
func (p *NeteasePlugin) GetSupportedFeatures() []plugin.MusicSourceFeature {
	return p.BaseMusicSourcePlugin.GetSupportedFeatures()
}

// GetServiceInfo 获取服务信息
func (p *NeteasePlugin) GetServiceInfo() *plugin.ServiceInfo {
	return p.BaseMusicSourcePlugin.GetServiceInfo()
}

// 插件入口点
func main() {
	plugin := NewNeteasePlugin()
	// TODO: 启动插件服务器或注册到插件管理器
	_ = plugin
}