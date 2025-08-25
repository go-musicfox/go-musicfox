package track

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/utils/app"
	"github.com/go-musicfox/netease-music/service"
)

var priority = map[service.SongQualityLevel]int{
	service.Standard: 1,
	service.Higher:   2,
	service.Exhigh:   3,
	service.Lossless: 4,
	service.Hires:    5,
}

type cacheFile struct {
	path    string
	size    int64
	modTime time.Time
}

// Cacher 负责管理歌曲的文件缓存。
type Cacher struct {
	musicDir string
	mu       sync.RWMutex
	maxSize  int64 // 最大缓存大小 (bytes)。0 禁用, -1 无限
}

// NewCacher 创建并初始化一个新的缓存管理器。
func NewCacher(maxSizeMB int64) *Cacher {
	var maxSize int64
	if maxSizeMB > 0 {
		maxSize = int64(maxSizeMB) * 1024 * 1024
	} else {
		maxSize = int64(maxSizeMB)
	}
	return &Cacher{
		musicDir: app.MusicCacheDir(),
		maxSize:  maxSize,
	}
}

func (m *Cacher) IsDisabled() bool {
	return m.maxSize == 0
}

func (m *Cacher) buildKey(songID int64, quality service.SongQualityLevel, fileType string) string {
	return fmt.Sprintf("%d-%d.%s", songID, priority[quality], fileType)
}

func (m *Cacher) ensureDirExists() error {
	return os.MkdirAll(m.musicDir, os.ModePerm)
}

// Put 将一首歌的数据流存入缓存。
func (m *Cacher) Put(song structs.Song, quality service.SongQualityLevel, fileType string, data io.Reader) error {
	if m.maxSize == 0 {
		_, err := io.Copy(io.Discard, data)
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if err := m.ensureDirExists(); err != nil {
		return fmt.Errorf("cannot put cache, directory check failed: %w", err)
	}

	key := m.buildKey(song.Id, quality, fileType)
	filePath := filepath.Join(m.musicDir, key)

	tempFile, err := os.CreateTemp(m.musicDir, "song-*.tmp")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	_, err = io.Copy(tempFile, data)
	if err != nil {
		return fmt.Errorf("failed to copy data to temp file: %w", err)
	}

	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	if err := os.Rename(tempFile.Name(), filePath); err != nil {
		return fmt.Errorf("failed to rename temp file to %s: %w", filePath, err)
	}

	slog.Debug("Song cached successfully.", "key", key)

	go func() {
		if err := m.prune(); err != nil {
			slog.Error("Failed to prune cache", "error", err)
		}
	}()

	return nil
}

// GetPath 尝试从缓存中获取一首歌的文件路径
// 它返回满足最低音质要求的最高音质的歌曲文件路径
func (m *Cacher) GetPath(songId int64, minQuality service.SongQualityLevel) (filePath string, fileType string, err error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if _, err := os.Stat(m.musicDir); os.IsNotExist(err) {
		return "", "", os.ErrNotExist
	}

	songIdStr := strconv.FormatInt(songId, 10)

	pattern := filepath.Join(m.musicDir, fmt.Sprintf("%s-*.*", songIdStr))
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return "", "", fmt.Errorf("failed to glob cache files with pattern %s: %w", pattern, err)
	}

	if len(matches) == 0 {
		return "", "", os.ErrNotExist
	}

	minPriority := priority[minQuality]
	bestKey := ""
	highestPriority := -1

	for _, matchPath := range matches {
		fileName := filepath.Base(matchPath)
		if strings.HasSuffix(fileName, ".tmp") {
			continue
		}

		ext := filepath.Ext(fileName)
		if ext == "" {
			continue
		}
		baseName := fileName[:len(fileName)-len(ext)]

		parts := strings.SplitN(baseName, "-", 2)
		if len(parts) != 2 {
			slog.Warn("Skipping malformed cache file.", "filename", fileName)
			continue
		}

		parsedSongId, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil || parsedSongId != songId {
			continue
		}

		p, err := strconv.Atoi(parts[1])
		if err != nil {
			slog.Warn("Skipping malformed cache file with invalid priority.", "filename", fileName)
			continue
		}

		if p >= minPriority && p > highestPriority {
			highestPriority = p
			bestKey = fileName
		}
	}

	if bestKey == "" {
		return "", "", os.ErrNotExist
	}

	slog.Debug("Cache hit.", "file", bestKey)
	filePath = filepath.Join(m.musicDir, bestKey)
	fileType = strings.TrimPrefix(filepath.Ext(bestKey), ".")

	return filePath, fileType, nil
}

// Get 尝试从缓存中获取一首歌的数据流
func (m *Cacher) Get(songId int64, minQuality service.SongQualityLevel) (io.ReadCloser, string, error) {
	filePath, fileType, err := m.GetPath(songId, minQuality)
	if err != nil {
		return nil, "", err
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to open cache file %s: %w", filePath, err)
	}

	return file, fileType, nil
}

// Clear 删除整个缓存目录并重建
func (m *Cacher) Clear() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err := os.RemoveAll(m.musicDir); err != nil {
		return err
	}
	return os.MkdirAll(m.musicDir, os.ModePerm)
}

// prune 检查当前缓存大小，如果超过上限，则删除最旧的文件直到满足要求
func (m *Cacher) prune() error {
	if m.maxSize <= 0 {
		return nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	entries, err := os.ReadDir(m.musicDir)
	if err != nil {
		return err
	}

	var files []cacheFile
	var totalSize int64

	for _, entry := range entries {
		if entry.IsDir() || strings.HasSuffix(entry.Name(), ".tmp") {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		totalSize += info.Size()
		files = append(files, cacheFile{
			path:    filepath.Join(m.musicDir, info.Name()),
			size:    info.Size(),
			modTime: info.ModTime(),
		})
	}

	if totalSize <= m.maxSize {
		return nil
	}

	slog.Debug("Cache size exceeds limit, pruning...", "currentSize", totalSize, "maxSize", m.maxSize)

	sort.Slice(files, func(i, j int) bool {
		return files[i].modTime.Before(files[j].modTime)
	})

	for _, file := range files {
		if err := os.Remove(file.path); err == nil {
			totalSize -= file.size
			slog.Debug("Pruned cache file", "path", file.path)
		}
		if totalSize <= m.maxSize {
			break
		}
	}

	return nil
}
