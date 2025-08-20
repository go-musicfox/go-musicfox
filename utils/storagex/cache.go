package storagex

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/pkg/errors"

	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/utils/app"
	"github.com/go-musicfox/go-musicfox/utils/filex"
)

func tryFindCache(songId int64) (fpath string) {
	cacheDir := app.MusicCacheDir()
	if !filex.FileOrDirExists(cacheDir) {
		if configs.ConfigRegistry.Main.CacheLimit != 0 {
			_ = os.MkdirAll(cacheDir, os.ModePerm)
		}
		return
	}
	// TODO: refactor this, read subdir to to improve performance
	files, err := os.ReadDir(cacheDir)
	if err != nil || len(files) == 0 {
		return
	}
	prior := priority[configs.ConfigRegistry.Main.PlayerSongLevel]
	for i := len(files) - 1; i >= 0; i-- {
		file := files[i]
		if strings.HasPrefix(file.Name(), strconv.FormatInt(songId, 10)) && !strings.HasSuffix(file.Name(), "-tmp") && file.Name() >= fmt.Sprintf("%d-%d", songId, prior) {
			fpath = filepath.Join(cacheDir, file.Name())
			return
		}
	}
	return
}

func CopyCachedSong(song structs.Song) error {
	initNameGen()
	downloadDir := app.DownloadDir()
	cacheDir := app.MusicCacheDir()
	if !filex.FileOrDirExists(downloadDir) {
		_ = os.MkdirAll(downloadDir, os.ModePerm)
	}
	if !filex.FileOrDirExists(cacheDir) {
		_ = os.MkdirAll(cacheDir, os.ModePerm)
	}
	oldFilename := tryFindCache(song.Id)
	if oldFilename == "" {
		return errors.New("cache file not exists")
	}
	split := strings.Split(path.Base(oldFilename), ".")
	musicType := split[len(split)-1]
	filename, _ := nameGen.Song(song, musicType)
	targetFilename := filepath.Join(downloadDir, filename)

	if _, err := os.Stat(targetFilename); err == nil {
		return errors.Wrap(os.ErrExist, "file exists: "+targetFilename)
	}
	err := os.Rename(oldFilename, targetFilename)
	if err != nil && (runtime.GOOS == "windows" || strings.HasSuffix(err.Error(), "invalid cross-device link")) {
		// fix: 当临时文件系统和目标下载位置不在同一磁盘时无法下载文件
		src, _ := os.Open(oldFilename)
		defer func() { src.Close() }()
		dst, _ := os.Create(targetFilename)
		defer func() { dst.Close() }()
		_, _ = io.Copy(dst, src)
	}
	return nil
}

func GetCacheURL(songID int64) (fpath, musicType string) {
	fpath = tryFindCache(songID)
	if fpath == "" {
		return
	}
	split := strings.Split(path.Base(fpath), ".")
	musicType = split[len(split)-1]
	fpath = "file://" + fpath
	return
}
