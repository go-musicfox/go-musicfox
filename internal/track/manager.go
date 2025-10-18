package track

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/go-musicfox/netease-music/service"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/singleflight"

	"github.com/go-musicfox/go-musicfox/internal/composer"
	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/utils/app"
	"github.com/go-musicfox/go-musicfox/utils/netease"
)

var supportedFileExtensions = []string{"mp3", "flac"}

type persistJob struct {
	ctx           context.Context
	stream        io.ReadCloser
	finalFilePath string
	source        PlayableSource
	isFromCache   bool
}

// Manager 是 songmanager 包的统一入口和协调器。
// 它为上层应用提供了简单、健壮且并发安全的接口。
type Manager struct {
	cacher      *Cacher
	fetcher     Fetcher
	tagger      Tagger
	nameGen     *composer.FileNameGenerator
	downloadDir string
	lyricDir    string
	quality     service.SongQualityLevel
	sfGroup     singleflight.Group
}

// ManagerOption 是用于配置 Manager 的函数类型。
type ManagerOption func(*Manager)

// NewManager 创建一个新的 Manager 实例。
func NewManager(opts ...ManagerOption) *Manager {
	m := &Manager{
		downloadDir: app.DownloadDir(),
		lyricDir:    app.DownloadLyricDir(),
		quality:     service.Standard, // 默认音质
	}

	for _, opt := range opts {
		opt(m)
	}

	if m.cacher == nil {
		m.cacher = NewCacher(0) // 禁用
	}
	if m.fetcher == nil {
		m.fetcher = NewFetcher(WithFetcherSongQuality(m.quality))
	}
	if m.tagger == nil {
		m.tagger = NewTagger()
	}
	if m.nameGen == nil {
		m.nameGen = composer.NewFileNameGenerator()
	}

	return m
}

// WithCacheLimit 是一个配置选项，用于设置缓存大小(MB)。
func WithCacheLimit(sizeMB int64) ManagerOption {
	return func(m *Manager) {
		if m.cacher == nil {
			m.cacher = NewCacher(sizeMB)
		}
	}
}

// WithSongQuality 是一个配置选项，用于设置期望的歌曲音质。
func WithSongQuality(quality service.SongQualityLevel) ManagerOption {
	return func(m *Manager) {
		m.quality = quality
	}
}

// WithCacher 是一个配置选项，用于提供一个自定义的 Cacher 实例。
// 这在测试时非常有用，可以注入一个 mock cacher。
func WithCacher(cacher *Cacher) ManagerOption {
	return func(m *Manager) {
		m.cacher = cacher
	}
}

// WithFetcher 是一个配置选项，用于提供一个自定义的 Fetcher 实例。
func WithFetcher(fetcher Fetcher) ManagerOption {
	return func(m *Manager) {
		m.fetcher = fetcher
	}
}

// WithTagger 是一个配置选项，用于提供一个自定义的 Tagger 实例。
func WithTagger(tagger Tagger) ManagerOption {
	return func(m *Manager) {
		m.tagger = tagger
	}
}

// WithNameGenerator 是一个配置选项，用于提供一个自定义的文件名生成器。
func WithNameGenerator(nameGen *composer.FileNameGenerator) ManagerOption {
	return func(m *Manager) {
		m.nameGen = nameGen
	}
}

// WithDownloadDir 是一个配置选项，用于设置歌曲下载目录。
func WithDownloadDir(dir string) ManagerOption {
	return func(m *Manager) {
		m.downloadDir = dir
	}
}

// WithDownloadLyricDir 是一个配置选项，用于设置歌词下载目录。
func WithDownloadLyricDir(dir string) ManagerOption {
	return func(m *Manager) {
		m.lyricDir = dir
	}
}

// ResolvePlayableSource 是 Manager 最核心的公共方法。
// 它解析一首歌的最佳可播放源，查找顺序: 已下载文件 -> 缓存文件 -> 远程网络。
func (m *Manager) ResolvePlayableSource(ctx context.Context, song structs.Song) (PlayableSource, error) {
	source, err := m.resolveSongSource(ctx, song)
	if err != nil {
		return PlayableSource{}, err
	}

	if source.Type == SourceRemote && !m.cacher.IsDisabled() {
		go m.backgroundCache(ctx, source)
	}

	return source, nil
}

// DownloadSong 下载一首歌并返回其本地路径。
func (m *Manager) DownloadSong(ctx context.Context, song structs.Song) (string, error) {
	if song.Id == 0 {
		return "", fmt.Errorf("Song does not exist, id = 0")
	}
	key := fmt.Sprintf("song-download-%d", song.Id)
	result, err, _ := m.sfGroup.Do(key, func() (any, error) {
		source, err := m.resolveSongSource(ctx, song)
		if err != nil {
			return "", err
		}

		switch source.Type {
		case SourceDownloaded:
			return source.Path, os.ErrExist
		case SourceCached:
			slog.Debug("Persisting song from cache to downloads", "songId", song.Id)
			return m.persistCachedSource(ctx, source)
		case SourceRemote:
			slog.Debug("Persisting song from remote to downloads", "songId", song.Id)
			return m.persistRemoteSource(ctx, source)
		}
		return "", fmt.Errorf("unknown source type encountered for song %d", song.Id)
	})

	if err != nil && !errors.Is(err, os.ErrExist) {
		return "", err
	}
	return result.(string), err
}

// DownloadLyric 下载一首歌的歌词并返回其本地路径。
func (m *Manager) DownloadLyric(ctx context.Context, song structs.Song) (string, error) {
	if song.Id == 0 {
		return "", errors.New("Song does not exist, id = 0")
	}
	key := fmt.Sprintf("lyric-download-%d", song.Id)
	result, err, _ := m.sfGroup.Do(key, func() (any, error) {
		fileName, err := m.nameGen.Lyric(song, "lrc")
		if err != nil {
			return "", err
		}
		filePath := filepath.Join(m.lyricDir, fileName)

		if _, err := os.Stat(filePath); err == nil {
			return filePath, os.ErrExist
		}

		lrc, err := m.GetLyric(ctx, song.Id)
		if err != nil {
			return "", err
		}

		if err := m.ensureDirExists(m.lyricDir); err != nil {
			return "", err
		}
		if err = os.WriteFile(filePath, []byte(lrc.Original), 0644); err != nil {
			return "", err
		}
		return filePath, nil
	})

	if err != nil && !errors.Is(err, os.ErrExist) {
		return "", err
	}
	return result.(string), err
}

// GetLyric 获取一首歌的歌词。
func (m *Manager) GetLyric(ctx context.Context, songID int64) (structs.LRCData, error) {
	key := fmt.Sprintf("lyric-fetch-%d", songID)
	result, err, _ := m.sfGroup.Do(key, func() (any, error) {
		return m.fetcher.FetchLyric(ctx, songID)
	})

	if err != nil {
		return structs.LRCData{}, err
	}
	return result.(structs.LRCData), nil
}

func (m *Manager) ClearCache() error {
	slog.Info("Clearing all song cache...")
	return m.cacher.Clear()
}

// resolveSongSource 严格按 已下载 -> 缓存 -> 网络的顺序解析音源。
func (m *Manager) resolveSongSource(ctx context.Context, song structs.Song) (PlayableSource, error) {
	key := fmt.Sprintf("song-resolve-%d", song.Id)
	result, err, _ := m.sfGroup.Do(key, func() (any, error) {
		// 检查下载目录
		for _, ext := range supportedFileExtensions {
			fileName, err := m.nameGen.Song(song, ext)
			if err != nil {
				slog.Warn("Failed to generate potential filename", "songId", song.Id, "ext", ext, "error", err)
				continue
			}
			finalFilePath := filepath.Join(m.downloadDir, fileName)
			if _, err := os.Stat(finalFilePath); err == nil {
				slog.Debug("Resolved source: Downloaded", "songId", song.Id)
				return PlayableSource{
					Song: song,
					Type: SourceDownloaded,
					Path: finalFilePath,
					Info: &netease.PlayableInfo{
						URL:       "file://" + finalFilePath,
						MusicType: ext,
					},
				}, nil
			}
		}

		// 检查缓存
		if !m.cacher.IsDisabled() {
			cachePath, fileType, cacheErr := m.cacher.GetPath(song.Id, m.quality)
			if cacheErr == nil {
				slog.Debug("Resolved source: Cached", "songId", song.Id)
				return PlayableSource{
					Song: song,
					Type: SourceCached,
					Path: cachePath,
					Info: &netease.PlayableInfo{
						URL:       "file://" + cachePath,
						MusicType: fileType,
					},
				}, nil
			}
			if !errors.Is(cacheErr, os.ErrNotExist) {
				slog.Error("Cache system error during source resolution", "songId", song.Id, "error", cacheErr)
				return nil, cacheErr
			}
		}

		// 从网络获取
		slog.Debug("Local sources miss, resolving from network...", "songId", song.Id)
		info, err := m.fetcher.FetchPlayableInfo(ctx, song.Id)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch playable info: %w", err)
		}
		slog.Debug("Resolved source: Remote", "songId", song.Id, slog.Any("info", info))
		return PlayableSource{
			Song: song,
			Type: SourceRemote,
			Info: info,
		}, nil
	})

	if err != nil {
		return PlayableSource{}, err
	}
	return result.(PlayableSource), nil
}

func (m *Manager) backgroundCache(ctx context.Context, source PlayableSource) {
	cacheKey := fmt.Sprintf("song-cache-%d", source.Id)
	m.sfGroup.Do(cacheKey, func() (any, error) {
		slog.Debug("Starting background caching...", "songId", source.Id)
		stream, err := m.fetcher.FetchStream(ctx, source)
		if err != nil {
			slog.Error("Background cache: fetch stream failed", "songId", source.Id, "error", err)
			return nil, err
		}
		defer stream.Close()

		err = m.cacher.Put(source.Song, m.quality, source.Info.MusicType, stream)
		if err != nil {
			slog.Error("Background cache: put failed", "songId", source.Id, "error", err)
		} else {
			slog.Debug("Background caching succeeded", "songId", source.Id)
		}
		return nil, err
	})
}

func (m *Manager) persistCachedSource(ctx context.Context, source PlayableSource) (string, error) {
	if err := m.ensureDirExists(m.downloadDir); err != nil {
		return "", err
	}
	fileName, _ := m.nameGen.Song(source.Song, source.Info.MusicType)
	finalFilePath := filepath.Join(m.downloadDir, fileName)

	stream, _, err := m.cacher.Get(source.Id, m.quality)
	if err != nil {
		return "", err
	}

	job := persistJob{
		ctx:           ctx,
		stream:        stream,
		finalFilePath: finalFilePath,
		source:        source,
		isFromCache:   true,
	}
	if err := m.persistStream(job); err != nil {
		return "", err
	}
	if err := m.tagger.SetSongTag(finalFilePath, source.Song); err != nil {
		slog.Warn("Song persisted from cache, but failed to set metadata.", "file", finalFilePath, "error", err)
	}
	return finalFilePath, nil
}

func (m *Manager) persistRemoteSource(ctx context.Context, source PlayableSource) (string, error) {
	if err := m.ensureDirExists(m.downloadDir); err != nil {
		return "", err
	}
	stream, err := m.fetcher.FetchStream(ctx, source)
	if err != nil {
		return "", err
	}

	fileName, _ := m.nameGen.Song(source.Song, source.Info.MusicType)
	filePath := filepath.Join(m.downloadDir, fileName)

	job := persistJob{
		ctx:           ctx,
		stream:        stream,
		finalFilePath: filePath,
		source:        source,
		isFromCache:   false,
	}
	if err := m.persistStream(job); err != nil {
		return "", err
	}
	if err := m.tagger.SetSongTag(filePath, source.Song); err != nil {
		slog.Warn("Song downloaded, but failed to set metadata.", "file", filePath, "error", err)
	}
	return filePath, nil
}

func (m *Manager) persistStream(job persistJob) error {
	defer job.stream.Close()

	tempFile, err := os.CreateTemp(filepath.Dir(job.finalFilePath), "download-*.tmp")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tempFile.Name())

	var (
		writer     io.Writer = tempFile
		eg         errgroup.Group
		pipeWriter *io.PipeWriter
	)

	if !job.isFromCache && !m.cacher.IsDisabled() {
		var pipeReader *io.PipeReader
		pipeReader, pipeWriter = io.Pipe()

		eg.Go(func() error {
			defer pipeReader.Close()
			return m.cacher.Put(job.source.Song, m.quality, job.source.Info.MusicType, pipeReader)
		})

		writer = io.MultiWriter(tempFile, pipeWriter)
	}

	_, copyErr := io.Copy(writer, job.stream)

	if pipeWriter != nil {
		if copyErr != nil {
			pipeWriter.CloseWithError(copyErr)
		} else {
			pipeWriter.Close()
		}
	}

	waitErr := eg.Wait()
	closeErr := tempFile.Close()

	if finalErr := errors.Join(copyErr, waitErr, closeErr); finalErr != nil {
		return finalErr
	}

	return os.Rename(tempFile.Name(), job.finalFilePath)
}

func (m *Manager) ensureDirExists(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return os.MkdirAll(dir, os.ModePerm)
	}
	return nil
}
