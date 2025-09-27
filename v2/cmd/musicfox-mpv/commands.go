package main

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/go-musicfox/go-musicfox/v2/pkg/event"
	"github.com/go-musicfox/go-musicfox/v2/pkg/model"
	"github.com/go-musicfox/go-musicfox/v2/plugins/audio"
	"github.com/go-musicfox/go-musicfox/v2/plugins/playlist"
)

// CommandHandler 命令处理器
type CommandHandler struct {
	audioPlugin    *audio.AudioPlugin
	playlistPlugin *playlist.PlaylistPluginImpl
	logger         *slog.Logger
}

// NewCommandHandler 创建命令处理器
func NewCommandHandler(audioPlugin *audio.AudioPlugin, playlistPlugin *playlist.PlaylistPluginImpl, logger *slog.Logger) *CommandHandler {
	return &CommandHandler{
		audioPlugin:    audioPlugin,
		playlistPlugin: playlistPlugin,
		logger:         logger,
	}
}

// HandlePlay 处理播放命令
func (h *CommandHandler) HandlePlay(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("play command requires a song URL or file path")
	}

	songURL := args[0]
	h.logger.Info("Playing song", "url", songURL, "backend", "mpv")

	// 创建歌曲对象
	song := &model.Song{
		ID:    fmt.Sprintf("mpv-%d", time.Now().UnixNano()),
		Title: extractTitleFromURL(songURL),
		URL:   songURL,
	}

	// 播放歌曲
	if err := h.audioPlugin.Play(song); err != nil {
		return fmt.Errorf("failed to play song: %w", err)
	}

	fmt.Printf("Playing: %s\n", song.Title)
	fmt.Printf("URL: %s\n", songURL)
	fmt.Println("Backend: MPV Player")
	return nil
}

// HandlePause 处理暂停命令
func (h *CommandHandler) HandlePause(ctx context.Context) error {
	h.logger.Info("Pausing playback")

	if err := h.audioPlugin.Pause(); err != nil {
		return fmt.Errorf("failed to pause: %w", err)
	}

	fmt.Println("Playback paused")
	return nil
}

// HandleResume 处理恢复命令
func (h *CommandHandler) HandleResume(ctx context.Context) error {
	h.logger.Info("Resuming playback")

	if err := h.audioPlugin.Resume(); err != nil {
		return fmt.Errorf("failed to resume: %w", err)
	}

	fmt.Println("Playback resumed")
	return nil
}

// HandleStop 处理停止命令
func (h *CommandHandler) HandleStop(ctx context.Context) error {
	h.logger.Info("Stopping playback")

	if err := h.audioPlugin.StopPlayback(); err != nil {
		return fmt.Errorf("failed to stop: %w", err)
	}

	fmt.Println("Playback stopped")
	return nil
}

// HandleNext 处理下一首命令
func (h *CommandHandler) HandleNext(ctx context.Context) error {
	h.logger.Info("Playing next song")

	// 从播放列表插件获取下一首歌曲
	currentSong := h.audioPlugin.GetPlayState().CurrentSong
	nextSong, err := h.playlistPlugin.GetNextSong(ctx, currentSong)
	if err != nil {
		return fmt.Errorf("failed to get next song: %w", err)
	}

	if nextSong == nil {
		fmt.Println("No next song available")
		return nil
	}

	if err := h.audioPlugin.Play(nextSong); err != nil {
		return fmt.Errorf("failed to play next song: %w", err)
	}

	fmt.Printf("Playing next song: %s\n", nextSong.Title)
	return nil
}

// HandlePrev 处理上一首命令
func (h *CommandHandler) HandlePrev(ctx context.Context) error {
	h.logger.Info("Playing previous song")

	// 从播放列表插件获取上一首歌曲
	currentSong := h.audioPlugin.GetPlayState().CurrentSong
	prevSong, err := h.playlistPlugin.GetPreviousSong(ctx, currentSong)
	if err != nil {
		return fmt.Errorf("failed to get previous song: %w", err)
	}

	if prevSong == nil {
		fmt.Println("No previous song available")
		return nil
	}

	if err := h.audioPlugin.Play(prevSong); err != nil {
		return fmt.Errorf("failed to play previous song: %w", err)
	}

	fmt.Printf("Playing previous song: %s\n", prevSong.Title)
	return nil
}

// HandleVolume 处理音量命令
func (h *CommandHandler) HandleVolume(ctx context.Context, args []string) error {
	if len(args) == 0 {
		// 显示当前音量
		volume := h.audioPlugin.GetVolume()
		fmt.Printf("Current volume: %d%%\n", int(volume*100))
		return nil
	}

	// 设置音量
	volumeStr := args[0]
	volumeInt, err := strconv.Atoi(volumeStr)
	if err != nil {
		return fmt.Errorf("invalid volume value: %s", volumeStr)
	}

	if volumeInt < 0 || volumeInt > 100 {
		return fmt.Errorf("volume must be between 0 and 100")
	}

	if err := h.audioPlugin.SetVolume(volumeInt); err != nil {
		return fmt.Errorf("failed to set volume: %w", err)
	}

	h.logger.Info("Volume changed", "volume", volumeInt)
	fmt.Printf("Volume set to: %d%%\n", volumeInt)
	return nil
}

// HandleStatus 处理状态命令
func (h *CommandHandler) HandleStatus(ctx context.Context) error {
	state := h.audioPlugin.GetPlayState()
	volume := h.audioPlugin.GetVolume()

	fmt.Println("=== MusicFox MPV Status ===")
	fmt.Printf("Backend: MPV Player\n")
	fmt.Printf("Status: %s\n", formatPlayStatus(state.Status))
	fmt.Printf("Volume: %d%%\n", volume)

	if state.CurrentSong != nil {
		fmt.Printf("Current Song: %s\n", state.CurrentSong.Title)
		fmt.Printf("URL: %s\n", state.CurrentSong.URL)
		if state.Duration > 0 {
			fmt.Printf("Duration: %s\n", formatDuration(state.Duration))
		}
		if state.Position > 0 {
			fmt.Printf("Position: %s\n", formatDuration(state.Position))
			progress := float64(state.Position) / float64(state.Duration) * 100
			fmt.Printf("Progress: %.1f%%\n", progress)
		}
	} else {
		fmt.Println("No song currently loaded")
	}

	if state.Playlist != nil {
		fmt.Printf("Playlist: %s (%d songs)\n", state.Playlist.Name, len(state.Playlist.Songs))
		fmt.Printf("Play Index: %d\n", state.PlayIndex)
	}

	fmt.Printf("Play Mode: %s\n", formatPlayMode(state.PlayMode))
	fmt.Printf("Shuffle: %t\n", state.Shuffle)
	fmt.Println("=========================")

	return nil
}

// HandlePlaylist 处理播放列表命令
func (h *CommandHandler) HandlePlaylist(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("playlist command requires a subcommand (create, list, show, add, remove)")
	}

	subcommand := args[0]
	subArgs := args[1:]

	switch subcommand {
	case "create":
		return h.handlePlaylistCreate(ctx, subArgs)
	case "list":
		return h.handlePlaylistList(ctx)
	case "show":
		return h.handlePlaylistShow(ctx, subArgs)
	case "add":
		return h.handlePlaylistAdd(ctx, subArgs)
	case "remove":
		return h.handlePlaylistRemove(ctx, subArgs)
	default:
		return fmt.Errorf("unknown playlist subcommand: %s", subcommand)
	}
}

// handlePlaylistCreate 创建播放列表
func (h *CommandHandler) handlePlaylistCreate(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("create command requires a playlist name")
	}

	name := strings.Join(args, " ")
	playlist, err := h.playlistPlugin.CreatePlaylist(ctx, name, "")
	if err != nil {
		return fmt.Errorf("failed to create playlist: %w", err)
	}

	h.logger.Info("Playlist created", "name", name, "id", playlist.ID)
	fmt.Printf("Created playlist: %s (ID: %s)\n", name, playlist.ID)
	return nil
}

// handlePlaylistList 列出播放列表
func (h *CommandHandler) handlePlaylistList(ctx context.Context) error {
	playlists, err := h.playlistPlugin.ListPlaylists(ctx)
	if err != nil {
		return fmt.Errorf("failed to get playlists: %w", err)
	}

	if len(playlists) == 0 {
		fmt.Println("No playlists found")
		return nil
	}

	fmt.Println("=== Playlists ===")
	for i, playlist := range playlists {
		fmt.Printf("%d. %s (ID: %s, %d songs)\n", i+1, playlist.Name, playlist.ID, len(playlist.Songs))
	}
	fmt.Println("=================")

	return nil
}

// handlePlaylistShow 显示播放列表详情
func (h *CommandHandler) handlePlaylistShow(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("show command requires a playlist ID")
	}

	playlistID := args[0]
	playlist, err := h.playlistPlugin.GetPlaylist(ctx, playlistID)
	if err != nil {
		return fmt.Errorf("failed to get playlist: %w", err)
	}

	fmt.Printf("=== Playlist: %s ===", playlist.Name)
	fmt.Printf("ID: %s\n", playlist.ID)
	fmt.Printf("Songs: %d\n", len(playlist.Songs))
	fmt.Println()

	if len(playlist.Songs) > 0 {
		fmt.Println("Songs:")
		for i, song := range playlist.Songs {
			fmt.Printf("%d. %s\n", i+1, song.Title)
			fmt.Printf("   URL: %s\n", song.URL)
		}
	} else {
		fmt.Println("No songs in this playlist")
	}

	fmt.Println("=========================")
	return nil
}

// handlePlaylistAdd 添加歌曲到播放列表
func (h *CommandHandler) handlePlaylistAdd(ctx context.Context, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("add command requires playlist ID and song URL")
	}

	playlistID := args[0]
	songURL := args[1]

	song := &model.Song{
		ID:     fmt.Sprintf("song-%d", time.Now().UnixNano()),
		Title:  extractTitleFromURL(songURL),
		URL:    songURL,
		Source: "mpv",
		Artist: "Unknown",
	}

	if err := h.playlistPlugin.AddSong(ctx, playlistID, song); err != nil {
		return fmt.Errorf("failed to add song to playlist: %w", err)
	}

	h.logger.Info("Song added to playlist", "playlist_id", playlistID, "song_url", songURL)
	fmt.Printf("Added song to playlist: %s\n", song.Title)
	return nil
}

// handlePlaylistRemove 从播放列表移除歌曲
func (h *CommandHandler) handlePlaylistRemove(ctx context.Context, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("remove command requires playlist ID and song ID")
	}

	playlistID := args[0]
	songID := args[1]

	if err := h.playlistPlugin.RemoveSong(ctx, playlistID, songID); err != nil {
		return fmt.Errorf("failed to remove song from playlist: %w", err)
	}

	h.logger.Info("Song removed from playlist", "playlist_id", playlistID, "song_id", songID)
	fmt.Printf("Removed song from playlist\n")
	return nil
}

// StatusMonitor 状态监控器
type StatusMonitor struct {
	audioPlugin    *audio.AudioPlugin
	playlistPlugin *playlist.PlaylistPluginImpl
	eventBus       event.EventBus
	logger         *slog.Logger
	running        bool
	stopChan       chan struct{}
}

// NewStatusMonitor 创建状态监控器
func NewStatusMonitor(audioPlugin *audio.AudioPlugin, playlistPlugin *playlist.PlaylistPluginImpl, eventBus event.EventBus, logger *slog.Logger) *StatusMonitor {
	return &StatusMonitor{
		audioPlugin:    audioPlugin,
		playlistPlugin: playlistPlugin,
		eventBus:       eventBus,
		logger:         logger,
		stopChan:       make(chan struct{}),
	}
}

// Start 启动状态监控
func (m *StatusMonitor) Start(ctx context.Context) error {
	if m.running {
		return fmt.Errorf("status monitor already running")
	}

	m.running = true
	m.logger.Info("Starting status monitor")

	// 启动监控协程
	go m.monitorLoop(ctx)

	return nil
}

// Stop 停止状态监控
func (m *StatusMonitor) Stop() {
	if !m.running {
		return
	}

	m.logger.Info("Stopping status monitor")
	m.running = false
	
	// 安全关闭channel，避免重复关闭
	select {
	case <-m.stopChan:
		// channel已经关闭
	default:
		close(m.stopChan)
	}
}

// monitorLoop 监控循环
func (m *StatusMonitor) monitorLoop(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-m.stopChan:
			return
		case <-ticker.C:
			m.checkStatus()
		}
	}
}

// checkStatus 检查播放状态
func (m *StatusMonitor) checkStatus() {
	state := m.audioPlugin.GetPlayState()
	m.logger.Debug("Status check", "status", state.Status, "backend", "mpv")

	// 发布状态事件
	event := &event.BaseEvent{
		ID:        fmt.Sprintf("status-check-%d", time.Now().UnixNano()),
		Type:      "audio.status.check",
		Source:    "mpv-monitor",
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"status":  state.Status,
			"backend": "mpv",
			"volume":  m.audioPlugin.GetVolume(),
		},
	}

	m.eventBus.PublishAsync(context.Background(), event)
}

// 辅助函数

// extractTitleFromURL 从URL提取标题
func extractTitleFromURL(url string) string {
	// 简单的标题提取逻辑
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		filename := parts[len(parts)-1]
		// 移除文件扩展名
		if dotIndex := strings.LastIndex(filename, "."); dotIndex > 0 {
			filename = filename[:dotIndex]
		}
		// 替换下划线和连字符为空格
		filename = strings.ReplaceAll(filename, "_", " ")
		filename = strings.ReplaceAll(filename, "-", " ")
		return filename
	}
	return "Unknown Song"
}

// formatPlayStatus 格式化播放状态
func formatPlayStatus(status model.PlayStatus) string {
	switch status {
	case model.PlayStatusPlaying:
		return "Playing"
	case model.PlayStatusPaused:
		return "Paused"
	case model.PlayStatusStopped:
		return "Stopped"
	case model.PlayStatusError:
		return "Error"
	default:
		return "Unknown"
	}
}

// formatPlayMode 格式化播放模式
func formatPlayMode(mode model.PlayMode) string {
	switch mode {
	case model.PlayModeSequential:
		return "Sequential"
	case model.PlayModeShuffle:
		return "Shuffle"
	case model.PlayModeRepeatOne:
		return "Repeat One"
	case model.PlayModeRepeatAll:
		return "Repeat All"
	default:
		return "Unknown"
	}
}

// formatDuration 格式化时长
func formatDuration(d time.Duration) string {
	totalSeconds := int(d.Seconds())
	minutes := totalSeconds / 60
	seconds := totalSeconds % 60
	return fmt.Sprintf("%02d:%02d", minutes, seconds)
}