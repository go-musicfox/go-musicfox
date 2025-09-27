package main

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/go-musicfox/go-musicfox/v2/pkg/model"
	"github.com/go-musicfox/go-musicfox/v2/plugins/audio"
	"github.com/go-musicfox/go-musicfox/v2/plugins/playlist"
	"github.com/google/uuid"
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
func (ch *CommandHandler) HandlePlay(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("play command requires a song URL or file path")
	}

	songURL := args[0]
	ch.logger.Info("Playing song", "url", songURL)

	// 创建歌曲对象
	song := &model.Song{
		ID:       uuid.New().String(),
		Title:    extractSongName(songURL),
		Artist:   "Unknown Artist",
		Album:    "Unknown Album",
		URL:      songURL,
		Source:   "local",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 播放歌曲
	if err := ch.audioPlugin.Play(song); err != nil {
		return fmt.Errorf("failed to play song: %w", err)
	}

	// 添加到播放历史
	if err := ch.playlistPlugin.AddToHistory(ctx, song); err != nil {
		ch.logger.Warn("Failed to add song to history", "error", err)
	}

	fmt.Printf("Now playing: %s\n", song.Title)
	return nil
}

// HandlePause 处理暂停命令
func (ch *CommandHandler) HandlePause(ctx context.Context) error {
	ch.logger.Info("Pausing playback")

	if err := ch.audioPlugin.Pause(); err != nil {
		return fmt.Errorf("failed to pause: %w", err)
	}

	fmt.Println("Playback paused")
	return nil
}

// HandleResume 处理恢复命令
func (ch *CommandHandler) HandleResume(ctx context.Context) error {
	ch.logger.Info("Resuming playback")

	if err := ch.audioPlugin.Resume(); err != nil {
		return fmt.Errorf("failed to resume: %w", err)
	}

	fmt.Println("Playback resumed")
	return nil
}

// HandleStop 处理停止命令
func (ch *CommandHandler) HandleStop(ctx context.Context) error {
	ch.logger.Info("Stopping playback")

	if err := ch.audioPlugin.Stop(); err != nil {
		return fmt.Errorf("failed to stop: %w", err)
	}

	fmt.Println("Playback stopped")
	return nil
}

// HandleNext 处理下一首命令
func (ch *CommandHandler) HandleNext(ctx context.Context) error {
	ch.logger.Info("Playing next song")

	// 获取当前播放状态
	state := ch.audioPlugin.GetState()
	if state == nil || state.CurrentSong == nil {
		return fmt.Errorf("no current song playing")
	}

	// 获取下一首歌曲
	nextSong, err := ch.playlistPlugin.GetNextSong(ctx, state.CurrentSong)
	if err != nil {
		return fmt.Errorf("failed to get next song: %w", err)
	}

	if nextSong == nil {
		fmt.Println("No next song available")
		return nil
	}

	// 播放下一首
	if err := ch.audioPlugin.Play(nextSong); err != nil {
		return fmt.Errorf("failed to play next song: %w", err)
	}

	fmt.Printf("Now playing: %s\n", nextSong.Title)
	return nil
}

// HandlePrev 处理上一首命令
func (ch *CommandHandler) HandlePrev(ctx context.Context) error {
	ch.logger.Info("Playing previous song")

	// 获取当前播放状态
	state := ch.audioPlugin.GetState()
	if state == nil || state.CurrentSong == nil {
		return fmt.Errorf("no current song playing")
	}

	// 获取上一首歌曲
	prevSong, err := ch.playlistPlugin.GetPreviousSong(ctx, state.CurrentSong)
	if err != nil {
		return fmt.Errorf("failed to get previous song: %w", err)
	}

	if prevSong == nil {
		fmt.Println("No previous song available")
		return nil
	}

	// 播放上一首
	if err := ch.audioPlugin.Play(prevSong); err != nil {
		return fmt.Errorf("failed to play previous song: %w", err)
	}

	fmt.Printf("Now playing: %s\n", prevSong.Title)
	return nil
}

// HandleVolume 处理音量命令
func (ch *CommandHandler) HandleVolume(ctx context.Context, args []string) error {
	if len(args) == 0 {
		// 显示当前音量
		volume := ch.audioPlugin.GetVolume()
		fmt.Printf("Current volume: %d%%\n", volume)
		return nil
	}

	// 设置音量
	volumeStr := args[0]
	volume, err := strconv.Atoi(volumeStr)
	if err != nil {
		return fmt.Errorf("invalid volume level: %s", volumeStr)
	}

	if volume < 0 || volume > 100 {
		return fmt.Errorf("volume must be between 0 and 100")
	}

	ch.logger.Info("Setting volume", "level", volume)

	if err := ch.audioPlugin.SetVolume(volume); err != nil {
		return fmt.Errorf("failed to set volume: %w", err)
	}

	fmt.Printf("Volume set to %d%%\n", volume)
	return nil
}

// HandleStatus 处理状态命令
func (ch *CommandHandler) HandleStatus(ctx context.Context) error {
	ch.logger.Debug("Getting playback status")

	// 获取播放状态
	state := ch.audioPlugin.GetState()
	if state == nil {
		fmt.Println("Status: No playback information available")
		return nil
	}

	fmt.Println("=== Playback Status ===")
	fmt.Printf("Status: %s\n", formatPlayStatus(state.Status))

	if state.CurrentSong != nil {
		fmt.Printf("Song: %s\n", state.CurrentSong.Title)
		fmt.Printf("Artist: %s\n", state.CurrentSong.Artist)
		fmt.Printf("Album: %s\n", state.CurrentSong.Album)
		fmt.Printf("Duration: %s\n", formatDuration(state.Duration))
		fmt.Printf("Position: %s\n", formatDuration(state.Position))
		fmt.Printf("Progress: %s\n", formatProgress(state.Position, state.Duration))
	}

	fmt.Printf("Volume: %d%%\n", ch.audioPlugin.GetVolume())
	fmt.Printf("Play Mode: %s\n", formatPlayMode(ch.playlistPlugin.GetPlayMode(ctx)))

	// 显示播放队列信息
	queue, err := ch.playlistPlugin.GetCurrentQueue(ctx)
	if err == nil && len(queue) > 0 {
		fmt.Printf("Queue: %d songs\n", len(queue))
	} else {
		fmt.Println("Queue: Empty")
	}

	return nil
}

// HandlePlaylist 处理播放列表命令
func (ch *CommandHandler) HandlePlaylist(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("playlist command requires a subcommand (create, list, show, add, remove, delete)")
	}

	subcommand := args[0]
	subArgs := args[1:]

	switch subcommand {
	case "create":
		return ch.handlePlaylistCreate(ctx, subArgs)
	case "list":
		return ch.handlePlaylistList(ctx)
	case "show":
		return ch.handlePlaylistShow(ctx, subArgs)
	case "add":
		return ch.handlePlaylistAdd(ctx, subArgs)
	case "remove":
		return ch.handlePlaylistRemove(ctx, subArgs)
	case "delete":
		return ch.handlePlaylistDelete(ctx, subArgs)
	case "queue":
		return ch.handleQueueCommand(ctx, subArgs)
	default:
		return fmt.Errorf("unknown playlist subcommand: %s", subcommand)
	}
}

// handlePlaylistCreate 处理创建播放列表
func (ch *CommandHandler) handlePlaylistCreate(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("create command requires a playlist name")
	}

	name := strings.Join(args, " ")
	playlist, err := ch.playlistPlugin.CreatePlaylist(ctx, name, "")
	if err != nil {
		return fmt.Errorf("failed to create playlist: %w", err)
	}

	fmt.Printf("Created playlist: %s (ID: %s)\n", playlist.Name, playlist.ID)
	return nil
}

// handlePlaylistList 处理列出播放列表
func (ch *CommandHandler) handlePlaylistList(ctx context.Context) error {
	playlists, err := ch.playlistPlugin.ListPlaylists(ctx)
	if err != nil {
		return fmt.Errorf("failed to list playlists: %w", err)
	}

	if len(playlists) == 0 {
		fmt.Println("No playlists found")
		return nil
	}

	fmt.Println("=== Playlists ===")
	for _, playlist := range playlists {
		fmt.Printf("ID: %s, Name: %s, Songs: %d\n", playlist.ID, playlist.Name, len(playlist.Songs))
	}

	return nil
}

// handlePlaylistShow 处理显示播放列表详情
func (ch *CommandHandler) handlePlaylistShow(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("show command requires a playlist ID")
	}

	playlistID := args[0]
	playlist, err := ch.playlistPlugin.GetPlaylist(ctx, playlistID)
	if err != nil {
		return fmt.Errorf("failed to get playlist: %w", err)
	}

	fmt.Printf("=== Playlist: %s ===\n", playlist.Name)
	fmt.Printf("ID: %s\n", playlist.ID)
	fmt.Printf("Description: %s\n", playlist.Description)
	fmt.Printf("Songs: %d\n", len(playlist.Songs))
	fmt.Println()

	for i, song := range playlist.Songs {
		fmt.Printf("%d. %s - %s\n", i+1, song.Title, song.Artist)
	}

	return nil
}

// handlePlaylistAdd 处理添加歌曲到播放列表
func (ch *CommandHandler) handlePlaylistAdd(ctx context.Context, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("add command requires playlist ID and song URL")
	}

	playlistID := args[0]
	songURL := args[1]

	// 创建歌曲对象
	song := &model.Song{
		ID:       uuid.New().String(),
		Title:    extractSongName(songURL),
		Artist:   "Unknown Artist",
		Album:    "Unknown Album",
		URL:      songURL,
		Source:   "local",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := ch.playlistPlugin.AddSong(ctx, playlistID, song); err != nil {
		return fmt.Errorf("failed to add song to playlist: %w", err)
	}

	fmt.Printf("Added song '%s' to playlist\n", song.Title)
	return nil
}

// handlePlaylistRemove 处理从播放列表移除歌曲
func (ch *CommandHandler) handlePlaylistRemove(ctx context.Context, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("remove command requires playlist ID and song ID")
	}

	playlistID := args[0]
	songID := args[1]

	if err := ch.playlistPlugin.RemoveSong(ctx, playlistID, songID); err != nil {
		return fmt.Errorf("failed to remove song from playlist: %w", err)
	}

	fmt.Println("Song removed from playlist")
	return nil
}

// handlePlaylistDelete 处理删除播放列表
func (ch *CommandHandler) handlePlaylistDelete(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("delete command requires a playlist ID")
	}

	playlistID := args[0]

	if err := ch.playlistPlugin.DeletePlaylist(ctx, playlistID); err != nil {
		return fmt.Errorf("failed to delete playlist: %w", err)
	}

	fmt.Println("Playlist deleted")
	return nil
}

// handleQueueCommand 处理队列命令
func (ch *CommandHandler) handleQueueCommand(ctx context.Context, args []string) error {
	if len(args) == 0 {
		// 显示当前队列
		queue, err := ch.playlistPlugin.GetCurrentQueue(ctx)
		if err != nil {
			return fmt.Errorf("failed to get queue: %w", err)
		}

		if len(queue) == 0 {
			fmt.Println("Queue is empty")
			return nil
		}

		fmt.Println("=== Current Queue ===")
		for i, song := range queue {
			fmt.Printf("%d. %s - %s\n", i+1, song.Title, song.Artist)
		}
		return nil
	}

	subcommand := args[0]
	subArgs := args[1:]

	switch subcommand {
	case "add":
		if len(subArgs) == 0 {
			return fmt.Errorf("queue add requires a song URL")
		}
		songURL := subArgs[0]
		song := &model.Song{
			ID:       uuid.New().String(),
			Title:    extractSongName(songURL),
			Artist:   "Unknown Artist",
			URL:      songURL,
			Source:   "local",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := ch.playlistPlugin.AddToQueue(ctx, song); err != nil {
			return fmt.Errorf("failed to add song to queue: %w", err)
		}
		fmt.Printf("Added '%s' to queue\n", song.Title)
	case "clear":
		if err := ch.playlistPlugin.ClearQueue(ctx); err != nil {
			return fmt.Errorf("failed to clear queue: %w", err)
		}
		fmt.Println("Queue cleared")
	case "shuffle":
		if err := ch.playlistPlugin.ShuffleQueue(ctx); err != nil {
			return fmt.Errorf("failed to shuffle queue: %w", err)
		}
		fmt.Println("Queue shuffled")
	default:
		return fmt.Errorf("unknown queue subcommand: %s", subcommand)
	}

	return nil
}

// 辅助函数

// extractSongName 从URL或文件路径提取歌曲名称
func extractSongName(url string) string {
	// 简单实现：从路径中提取文件名
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		filename := parts[len(parts)-1]
		// 移除文件扩展名
		if dotIndex := strings.LastIndex(filename, "."); dotIndex > 0 {
			return filename[:dotIndex]
		}
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
	case model.PlayStatusBuffering:
		return "Buffering"
	case model.PlayStatusError:
		return "Error"
	default:
		return "Unknown"
	}
}

// formatDuration 格式化时长
func formatDuration(d time.Duration) string {
	if d == 0 {
		return "--:--"
	}
	minutes := int(d.Minutes())
	seconds := int(d.Seconds()) % 60
	return fmt.Sprintf("%02d:%02d", minutes, seconds)
}

// formatProgress 格式化播放进度
func formatProgress(position, duration time.Duration) string {
	if duration == 0 {
		return "---%"
	}
	progress := float64(position) / float64(duration) * 100
	return fmt.Sprintf("%.1f%%", progress)
}

// formatPlayMode 格式化播放模式
func formatPlayMode(mode model.PlayMode) string {
	switch mode {
	case model.PlayModeSequential:
		return "Sequential"
	case 2: // Random mode
		return "Random"
	case 1: // Single mode
		return "Single"
	default:
		return "Unknown"
	}
}