package storagex

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/go-musicfox/netease-music/service"
	foldersize "github.com/markthree/go-get-folder-size/src"
	"github.com/pkg/errors"

	"github.com/go-musicfox/go-musicfox/internal/composer"
	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/internal/track"
	"github.com/go-musicfox/go-musicfox/internal/types"
	"github.com/go-musicfox/go-musicfox/utils/app"
	"github.com/go-musicfox/go-musicfox/utils/filex"
	"github.com/go-musicfox/go-musicfox/utils/netease"
	"github.com/go-musicfox/go-musicfox/utils/notify"
	"github.com/go-musicfox/go-musicfox/utils/slogx"
)

func DownloadFile(url, filename, dirname string) error {
	targetFilename := filepath.Join(dirname, filename)
	if !filex.FileOrDirExists(dirname) {
		_ = os.MkdirAll(dirname, os.ModePerm)
	}
	if _, err := os.Stat(targetFilename); err == nil {
		return errors.Wrap(os.ErrExist, "file exists: "+targetFilename)
	}
	client := &http.Client{
		Timeout: 60 * time.Second,
	}
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	f, err := os.CreateTemp("", filename)
	if err != nil {
		return err
	}
	defer func() { _ = os.Remove(f.Name()) }()

	_, err = io.Copy(f, resp.Body)
	if err != nil {
		slog.Error("下载歌曲失败", slogx.Error(err))
		return err
	}
	err = os.Rename(f.Name(), targetFilename)
	if err != nil && (runtime.GOOS == "windows" || strings.HasSuffix(err.Error(), "invalid cross-device link")) {
		// fix: 当临时文件系统和目标下载位置不在同一磁盘时无法下载文件
		srcFile, _ := os.Open(f.Name())
		dstFile, _ := os.Create(targetFilename)
		defer dstFile.Close()
		_, _ = io.Copy(dstFile, srcFile)
		_ = srcFile.Close()
	}
	return nil
}

var (
	tplInitd sync.Once
	nameGen  *composer.FileNameGenerator
	tagger   = track.NewTagger()
)

func initNameGen() {
	tplInitd.Do(func() {
		nameGen = composer.NewFileNameGenerator()
		tplDownloadStr := configs.ConfigRegistry.Main.DownloadFileNameTpl
		if tplDownloadStr != "" {
			if err := nameGen.RegisterSongTemplate(tplDownloadStr); err != nil {
				panic(fmt.Sprintf("加载自定义下载模板失败, %v", err))
			}
			nameGen.RegisterLyricTemplate(tplDownloadStr)
		}
	})
}

func downloadMusic(url, musicType string, song structs.Song, downloadDir string) error {
	initNameGen()
	filename, _ := nameGen.Song(song, musicType)
	err := DownloadFile(url, filename, downloadDir)
	if err != nil {
		return err
	}
	SetSongTag(filepath.Join(downloadDir, filename), song)
	slog.Info("下载歌曲成功", slog.String("file", filename))
	return nil
}

// DownloadMusic 下载音乐
func DownloadMusic(song structs.Song) {
	var (
		errHandler = func(errs ...error) {
			slog.Error("下载歌曲失败", slog.Any("error", errs))
		}
		err error
	)

	info, err := netease.FetchPlayableInfo(song.Id, configs.ConfigRegistry.Main.PlayerSongLevel)
	if err != nil {
		errHandler(err)
		return
	}

	downloadDir := app.DownloadDir()
	notify.Notify(notify.NotifyContent{
		Title: "👇🏻正在下载，请稍候...",
		Text:  song.Name,
		// Url:     FileUrl(downloadDir),
		GroupId: types.GroupID,
	})

	if fpath := tryFindCache(song.Id); fpath != "" {
		err = CopyCachedSong(song)
	} else {
		err = downloadMusic(info.URL, info.MusicType, song, downloadDir)
	}
	switch true {
	case err == nil:
		notify.Notify(notify.NotifyContent{
			Title: "✅下载完成",
			Text:  song.Name,
			// Url:     FileUrl(ownloadDir),
			GroupId: types.GroupID,
		})
	case errors.Is(err, os.ErrExist):
		notify.Notify(notify.NotifyContent{
			Title: "🙅🏻‍文件已存在",
			Text:  song.Name,
			// Url:     FileUrl(downloadDir),
			GroupId: types.GroupID,
		})
	default:
		notify.Notify(notify.NotifyContent{
			Title: "❌下载失败",
			Text:  err.Error(),
			// Url:     FileUrl(downloadDir),
			GroupId: types.GroupID,
		})
		errHandler(err)
	}
}

func DownLoadLrc(song structs.Song) {
	lrc, err := netease.FetchLyric(song.Id)
	if err != nil {
		return
	}

	filename, _ := nameGen.Lyric(song, "lrc")
	savepath := filepath.Join(app.DownloadLyricDir(), filename)

	err = os.WriteFile(savepath, []byte(lrc), 0644)
	if err != nil {
		notify.Notify(notify.NotifyContent{
			Title:   "下载歌词失败",
			Text:    err.Error(),
			Url:     types.AppGithubUrl,
			GroupId: types.GroupID,
		})
	} else {
		notify.Notify(notify.NotifyContent{
			Title:   "下载歌词成功",
			Text:    song.Name + ".lrc 已保存到指定目录",
			Url:     types.AppGithubUrl,
			GroupId: types.GroupID,
		})
	}
}

var priority = map[service.SongQualityLevel]int{
	service.Standard: 1,
	service.Higher:   2,
	service.Exhigh:   3,
	service.Lossless: 4,
	service.Hires:    5,
}

func CacheMusic(song structs.Song, url string, musicType string, quality service.SongQualityLevel) {
	errHandler := func(errs ...error) {
		slog.Error("缓存歌曲失败", slog.Any("error", errs))
	}
	var err error
	cacheDir := app.MusicCacheDir()
	size, err := foldersize.Invoke(cacheDir)
	if err != nil {
		errHandler(err)
		return
	}
	if configs.ConfigRegistry.Main.CacheLimit != -1 && size > configs.ConfigRegistry.Main.CacheLimit*1024*1024 {
		return
	}
	filename := fmt.Sprintf("%d-%d.%s", song.Id, priority[quality], musicType)
	err = DownloadFile(url, filename, cacheDir)
	if err != nil {
		errHandler(err)
		return
	}
	SetSongTag(filepath.Join(cacheDir, filename), song)
	slog.Info("缓存歌曲成功", slog.String("file", filename))
}

func ClearMusicCache() error {
	cacheDir := app.MusicCacheDir()
	return ClearDir(cacheDir)
}

func ClearDir(dir string) error {
	if filex.FileOrDirExists(dir) {
		if err := os.RemoveAll(dir); err != nil {
			return err
		}
		return os.MkdirAll(dir, os.ModePerm)
	}
	return nil
}

func SetSongTag(filePath string, song structs.Song) {
	if err := tagger.SetSongTag(filePath, song); err != nil {
		slog.Error("元数据写入失败", "error", err)
	}
}
