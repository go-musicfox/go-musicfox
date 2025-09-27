package main

import (
	"fmt"
	"path/filepath"
	"time"
)

// loadAudioPlugin 加载音频插件
func (app *MVPApp) loadAudioPlugin() error {
	app.logger.Info("Loading real audio plugin from plugins/audio/...")

	// 创建真实的音频插件实例
	audio := &RealAudioPlugin{
		logger:    app.logger,
		pluginDir: filepath.Join("../../plugins/audio"),
		info: &PluginInfo{
			ID:          "audio-processor",
			Name:        "Real Audio Plugin",
			Version:     "1.0.0",
			Description: "Real audio processing plugin from plugins/audio/",
			Author:      "go-musicfox",
			Type:        "audio_processor",
			CreatedAt:   time.Now(),
		},
	}

	// 初始化插件
	if err := audio.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize real audio plugin: %w", err)
	}

	// 注册到插件管理器
	if err := app.pluginManager.RegisterPlugin(audio); err != nil {
		return fmt.Errorf("failed to register real audio plugin: %w", err)
	}
	app.logger.Info("Real audio plugin loaded successfully from plugins/audio/")
	return nil
}

// loadPlaylistPlugin 加载播放列表插件
func (app *MVPApp) loadPlaylistPlugin() error {
	app.logger.Info("Loading real playlist plugin from plugins/playlist/...")

	// 创建真实的播放列表插件实例
	playlist := &RealPlaylistPlugin{
		logger:    app.logger,
		pluginDir: filepath.Join("../../plugins/playlist"),
		info: &PluginInfo{
			ID:          "playlist-manager",
			Name:        "Real Playlist Plugin",
			Version:     "1.0.0",
			Description: "Real playlist management plugin from plugins/playlist/",
			Author:      "go-musicfox",
			Type:        "playlist",
			CreatedAt:   time.Now(),
		},
	}

	// 初始化插件
	if err := playlist.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize real playlist plugin: %w", err)
	}

	// 注册到插件管理器
	if err := app.pluginManager.RegisterPlugin(playlist); err != nil {
		return fmt.Errorf("failed to register real playlist plugin: %w", err)
	}
	app.logger.Info("Real playlist plugin loaded successfully from plugins/playlist/")
	return nil
}

// loadNeteasePlugin 加载网易云插件
func (app *MVPApp) loadNeteasePlugin() error {
	app.logger.Info("Loading real netease plugin from plugins/netease/...")

	// 创建真实的网易云插件实例
	netease := &RealNeteasePlugin{
		logger:    app.logger,
		pluginDir: filepath.Join("../../plugins/netease"),
		info: &PluginInfo{
			ID:          "netease-music",
			Name:        "Real Netease Music Plugin",
			Version:     "1.0.0",
			Description: "Real Netease Cloud Music plugin from plugins/netease/",
			Author:      "go-musicfox",
			Type:        "music_source",
			CreatedAt:   time.Now(),
		},
	}

	// 初始化插件
	if err := netease.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize real netease plugin: %w", err)
	}

	// 注册到插件管理器
	if err := app.pluginManager.RegisterPlugin(netease); err != nil {
		return fmt.Errorf("failed to register real netease plugin: %w", err)
	}
	app.logger.Info("Real netease plugin loaded successfully from plugins/netease/")
	return nil
}

// loadTUIPlugin 加载TUI插件
func (app *MVPApp) loadTUIPlugin() error {
	app.logger.Info("Loading real TUI plugin from plugins/tui/...")

	// 创建真实的TUI插件实例
	tuiPlugin := &RealTUIPluginWrapper{
		logger:    app.logger,
		pluginDir: filepath.Join("../../plugins/tui"),
		info: &PluginInfo{
			ID:          "tui",
			Name:        "Real TUI Plugin",
			Version:     "1.0.0",
			Description: "Real Terminal User Interface plugin from plugins/tui/",
			Author:      "go-musicfox",
			Type:        "ui",
			CreatedAt:   time.Now(),
		},
	}

	// 初始化插件
	if err := tuiPlugin.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize real TUI plugin: %w", err)
	}

	// 注册到插件管理器
	if err := app.pluginManager.RegisterPlugin(tuiPlugin); err != nil {
		return fmt.Errorf("failed to register real TUI plugin: %w", err)
	}

	app.tuiPlugin = tuiPlugin
	app.logger.Info("Real TUI plugin loaded successfully from plugins/tui/")
	return nil
}