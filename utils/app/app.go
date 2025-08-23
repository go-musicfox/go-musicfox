package app

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	_ "github.com/adrg/xdg"
	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/types"
)

var (
	projectDir     string
	projectDirOnce sync.Once
)

// DataRootDir 获取本地数据存储目录
func DataRootDir() string {
	projectDirOnce.Do(func() {
		if root := os.Getenv("MUSICFOX_ROOT"); root != "" {
			projectDir = root
		} else {
			configDir, err := os.UserConfigDir()
			if nil != err {
				panic("未获取到本地数据目录：" + err.Error())
			}
			projectDir = filepath.Join(configDir, types.AppLocalDataDir)
		}
		if _, err := os.Stat(projectDir); os.IsNotExist(err) {
			_ = os.MkdirAll(projectDir, os.ModePerm)
		}
	})
	return projectDir
}

func CacheDir() string {
	cacheDir := configs.ConfigRegistry.Main.CacheDir
	if cacheDir == "" {
		cache, err := os.UserCacheDir()
		if nil != err {
			panic("未获取到本地缓存目录：" + err.Error())
		}
		cacheDir = filepath.Join(cache, types.AppLocalDataDir)
	}
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		_ = os.MkdirAll(cacheDir, os.ModePerm)
	}
	return cacheDir
}

func DownloadDir() string {
	downloadDir := configs.ConfigRegistry.Main.DownloadDir
	return resolvePath(downloadDir, DataRootDir(), "download")
}

func DownloadLyricDir() string {
	downloadLyricDir := configs.ConfigRegistry.Main.DownloadLyricDir
	return resolvePath(downloadLyricDir, DataRootDir(), "download")
}

func resolvePath(pathA, basePathB string, additional ...string) string {
	var resolvedBase string
	if filepath.IsAbs(pathA) {
		resolvedBase = pathA
	} else {
		resolvedBase = filepath.Join(basePathB, pathA)
	}
	allSegments := append([]string{resolvedBase}, additional...)
	fullPath := filepath.Join(allSegments...)
	absPath, _ := filepath.Abs(fullPath)
	return absPath
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
