package app

import (
	"encoding/binary"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"

	"github.com/adrg/xdg"
	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/types"
)

// pathManager 统一管理应用的所有路径，遵循XDG规范
type pathManager struct {
	isPortable bool   // 是否为便携模式
	rootDir    string // 根目录，仅在便携模式下有意义

	configDir   string // 配置文件目录 (XDG_CONFIG_HOME)
	dataDir     string // 数据文件目录 (XDG_DATA_HOME)
	stateDir    string // 状态文件目录 (XDG_STATE_HOME)
	cacheDir    string // 缓存目录 (XDG_CACHE_HOME)
	downloadDir string // 下载目录 (XDG_DOWNLOAD_DIR)

	dbDir         string // 数据库目录
	logDir        string // 日志文件目录
	musicCacheDir string // 音乐缓存子目录的完整路径
}

var (
	paths         pathManager
	bootstrapOnce sync.Once
	initPathsOnce sync.Once
)

func initPaths() {
	bootstrapOnce.Do(func() {
		portableRoot := os.Getenv("MUSICFOX_ROOT")
		if portableRoot != "" {
			absRoot, err := filepath.Abs(portableRoot)
			if err != nil {
				panic(fmt.Sprintf("无法解析便携模式根目录: %v", err))
			}
			paths.isPortable = true
			paths.rootDir = absRoot
			paths.configDir = absRoot
			paths.stateDir = absRoot
			paths.dataDir = filepath.Join(absRoot, "data")
			paths.cacheDir = filepath.Join(absRoot, "cache")
			paths.downloadDir = filepath.Join(absRoot, "download")
			mustCreateDirectory(absRoot)
		} else {
			paths.dataDir = filepath.Join(xdg.DataHome, types.AppLocalDataDir)
			paths.stateDir = filepath.Join(xdg.StateHome, types.AppLocalDataDir)
			paths.cacheDir = filepath.Join(xdg.CacheHome, types.AppLocalDataDir)
			paths.downloadDir = filepath.Join(xdg.UserDirs.Download, types.AppLocalDataDir)
			path, err := xdg.ConfigFile(types.AppLocalDataDir)
			if err != nil {
				panic(fmt.Sprintf("无法获取配置目录: %v", err))
			}
			paths.configDir = path
		}
		paths.logDir = filepath.Join(paths.stateDir, "log")
		paths.dbDir = filepath.Join(paths.dataDir, "db")

		mustCreateDirectory(paths.configDir, paths.dataDir, paths.logDir)
	})
}

func initPathsWithConfig() {
	initPathsOnce.Do(func() {
		initPaths()
		if userCacheDir := configs.AppConfig.Storage.Cache.Dir; userCacheDir != "" {
			if paths.isPortable {
				paths.cacheDir = filepath.Join(paths.rootDir, userCacheDir)
			} else {
				paths.cacheDir, _ = filepath.Abs(userCacheDir)
			}
		}
		musicCacheDir := configs.AppConfig.Storage.Cache.MusicDir
		if filepath.IsAbs(musicCacheDir) {
			paths.musicCacheDir, _ = filepath.Abs(musicCacheDir)
		} else {
			paths.musicCacheDir = filepath.Join(paths.cacheDir, musicCacheDir)
		}

		if userDownloadDir := configs.AppConfig.Storage.DownloadDir; userDownloadDir != "" {
			if paths.isPortable {
				paths.downloadDir = filepath.Join(paths.rootDir, userDownloadDir)
			} else {
				paths.downloadDir, _ = filepath.Abs(userDownloadDir)
			}
		}

		mustCreateDirectory(paths.cacheDir)
	})
}

func mustCreateDirectory(dirs ...string) {
	for _, dir := range dirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			if err := os.MkdirAll(dir, 0755); err != nil {
				slog.Error("创建目录失败", "dir", dir, "error", err)
			}
		}
	}
}

func ConfigDir() string {
	initPaths()
	return paths.configDir
}

func ConfigFilePath() string {
	resolved := configs.ResolveConfigFile(ConfigDir())
	return resolved.Path
}

// DataDir 获取数据文件目录 (db, cookie, logo.png)
func DataDir() string {
	initPaths()
	return paths.dataDir
}

// RuntimeDir 用于 beep 临时文件、二维码图片路径等
func RuntimeDir() string {
	initPaths()
	dir := filepath.Join(xdg.RuntimeDir, types.AppLocalDataDir)

	// 尝试创建目录
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			// 创建失败（如只读系统），回退到 CacheDir
			slog.Warn("无法创建 RuntimeDir，回退到 CacheDir", "dir", dir, "error", err)
			return CacheDir()
		}
	}

	return dir
}

func DBDir() string {
	return filepath.Join(DataDir(), "db")
}

func StateDir() string {
	initPaths()
	return paths.stateDir
}

func LogDir() string {
	initPaths()
	return paths.logDir
}

func CacheDir() string {
	initPathsWithConfig()
	return paths.cacheDir
}

// MusicCacheDir 获取存放音乐文件的特定缓存子目录
func MusicCacheDir() string {
	initPathsWithConfig()
	return paths.musicCacheDir
}

// DownloadDir 下载目录
func DownloadDir() string {
	initPathsWithConfig()
	return paths.downloadDir
}

// DownloadLyricDir 歌词下载目录，同 DownloadDir
func DownloadLyricDir() string {
	customDir := configs.AppConfig.Storage.LyricDir
	if customDir == "" {
		return DownloadDir()
	}

	if paths.isPortable {
		return filepath.Join(paths.rootDir, customDir)
	} else {
		dir, _ := filepath.Abs(customDir)
		return dir
	}
}

// IDToBin convert autoincrement ID to []byte
func IDToBin(ID uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, ID)
	return b
}

// BinToID convert []byte to autoincrement ID
func BinToID(bin []byte) uint64 {
	ID := binary.BigEndian.Uint64(bin)

	return ID
}

func AddResizeParamForPicUrl(picurl string, size int64) string {
	if picurl == "" {
		return ""
	}
	return fmt.Sprintf("%s?param=%dy%d", picurl, size, size)
}
