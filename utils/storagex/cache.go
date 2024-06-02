package storagex

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/buger/jsonparser"
	"github.com/go-musicfox/netease-music/service"
	"github.com/pkg/errors"

	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/utils/app"
	"github.com/go-musicfox/go-musicfox/utils/filex"
)

func getCacheUri(songId int64) (uri string, ok bool) {
	cacheDir := app.CacheDir()
	if !filex.FileOrDirExists(cacheDir) {
		if configs.ConfigRegistry.Main.CacheLimit != 0 {
			_ = os.MkdirAll(cacheDir, os.ModePerm)
		}
		return
	}
	files, err := os.ReadDir(cacheDir)
	if err != nil || len(files) == 0 {
		return
	}
	for i := len(files) - 1; i >= 0; i-- {
		file := files[i]
		if strings.HasPrefix(file.Name(), strconv.FormatInt(songId, 10)) {
			uri = filepath.Join(cacheDir, file.Name())
			ok = true
			return
		}
	}
	return
}

func CopyCachedSong(song structs.Song) error {
	downloadDir := app.DownloadDir()
	cacheDir := app.CacheDir()
	if !filex.FileOrDirExists(downloadDir) {
		_ = os.MkdirAll(downloadDir, os.ModePerm)
	}
	if !filex.FileOrDirExists(cacheDir) {
		_ = os.MkdirAll(cacheDir, os.ModePerm)
	}
	oldFilename, ok := getCacheUri(song.Id)
	if !ok {
		return errors.New("cache file not exists")
	}
	split := strings.Split(path.Base(oldFilename), ".")
	musicType := split[len(split)-1]
	filename := fmt.Sprintf("%s-%s.%s", song.Name, song.ArtistName(), musicType)
	// Windows Linux 均不允许文件名中出现 / \ 替换为 _
	filename = strings.Replace(filename, "/", "_", -1)
	filename = strings.Replace(filename, "\\", "_", -1)
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

func PlayableUrlSong(song structs.Song) (url, musicType string, err error) {
	if configs.ConfigRegistry.Main.CacheLimit != 0 {
		var ok bool
		if url, musicType, ok = GetCacheUrl(song.Id); ok {
			return
		}
	}

	slog.Info("get play url of song",
		slog.Int64("song", song.Id),
		slog.String("level", string(configs.ConfigRegistry.Main.PlayerSongLevel)),
	)

	urlService := service.SongUrlV1Service{
		ID:      strconv.FormatInt(song.Id, 10),
		Level:   configs.ConfigRegistry.Main.PlayerSongLevel,
		SkipUNM: true,
	}
	code, response := urlService.SongUrl()
	if code != 200 {
		err = errors.New(string(response))
		return
	}

	var (
		err1, err2    error
		freeTrialInfo jsonparser.ValueType
	)
	url, err1 = jsonparser.GetString(response, "data", "[0]", "url")
	_, freeTrialInfo, _, err2 = jsonparser.Get(response, "data", "[0]", "freeTrialInfo")
	if err1 != nil || err2 != nil || url == "" || (freeTrialInfo != jsonparser.NotExist && freeTrialInfo != jsonparser.Null) {
		br, ok := brMap[urlService.Level]
		if !ok {
			br = "320000"
		}
		s := service.SongUrlService{
			ID: strconv.FormatInt(song.Id, 10),
			Br: br,
		}
		code, response = s.SongUrl()
		if code != 200 {
			err = errors.New(string(response))
			return
		}
	}

	if size, _ := jsonparser.GetInt(response, "data", "[0]", "size"); size != 0 {
		slog.Info("music size", slog.Int64("size", size))
	}

	url, _ = jsonparser.GetString(response, "data", "[0]", "url")
	musicType, _ = jsonparser.GetString(response, "data", "[0]", "type")
	if musicType = strings.ToLower(musicType); musicType == "" {
		musicType = "mp3"
	}
	err = nil
	if configs.ConfigRegistry.Main.CacheLimit != 0 {
		go CacheMusic(song, url, musicType, urlService.Level)
	}
	return
}
