package storagex

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/bogem/id3v2/v2"
	songtag "github.com/frolovo22/tag"
	"github.com/go-flac/flacpicture"
	"github.com/go-musicfox/netease-music/service"
	foldersize "github.com/markthree/go-get-folder-size/src"
	"github.com/pkg/errors"

	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/internal/types"
	"github.com/go-musicfox/go-musicfox/utils/app"
	"github.com/go-musicfox/go-musicfox/utils/filex"
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
	songNameTpl *template.Template
	tplInitd    sync.Once
)

func downloadMusic(url, musicType string, song structs.Song, downloadDir string) error {
	tplInitd.Do(func() {
		tpl := template.New("songName")
		if configs.ConfigRegistry.Main.DownloadFileNameTpl != "" {
			songNameTpl = template.Must(tpl.Parse(configs.ConfigRegistry.Main.DownloadFileNameTpl))
		} else {
			songNameTpl = template.Must(tpl.Parse("{{.SongName}}-{{.ArtistName}}.{{.SongType}}"))
		}
	})
	var filenameBuilder strings.Builder
	_ = songNameTpl.Execute(&filenameBuilder, map[string]string{
		"SongName":   song.Name,
		"ArtistName": song.ArtistName(),
		"SongType":   musicType,
	})
	filename := filenameBuilder.String()

	// Windows Linux 均不允许文件名中出现 / \ 替换为 _
	filename = strings.ReplaceAll(filename, "/", "_")
	filename = strings.ReplaceAll(filename, "\\", "_")
	err := DownloadFile(url, filename, downloadDir)
	if err != nil {
		return err
	}
	file, _ := os.OpenFile(filepath.Join(downloadDir, filename), os.O_RDWR, os.ModePerm)
	SetSongTag(file, song)
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

	url, musicType, err := PlayableUrlSong(song)
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

	if _, ok := getCacheUri(song.Id); ok {
		err = CopyCachedSong(song)
	} else {
		err = downloadMusic(url, musicType, song, downloadDir)
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
	cacheDir := app.CacheDir()
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
	file, err := os.OpenFile(filepath.Join(cacheDir, filename), os.O_RDWR, os.ModePerm)
	if err != nil {
		return
	}
	SetSongTag(file, song)
	slog.Info("缓存歌曲成功", slog.String("file", filename))
}

func GetCacheUrl(songId int64) (url, musicType string, ok bool) {
	url, ok = getCacheUri(songId)
	if !ok || path.Base(url) < fmt.Sprintf("%d-%d", songId, priority[configs.ConfigRegistry.Main.PlayerSongLevel]) {
		ok = false
		return
	}
	split := strings.Split(path.Base(url), ".")
	musicType = split[len(split)-1]
	ok = true
	return
}

func ClearMusicCache() error {
	cacheDir := app.CacheDir()
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

var brMap = map[service.SongQualityLevel]string{
	service.Standard: "320000",
	service.Higher:   "320000",
	service.Exhigh:   "320000",
	service.Lossless: "999000",
	service.Hires:    "999000",
}

func SetSongTag(file *os.File, song structs.Song) {
	defer file.Close()
	version := songtag.CheckVersion(file)
	switch version {
	case songtag.VersionID3v22, songtag.VersionID3v23, songtag.VersionID3v24:
		tag, err := id3v2.ParseReader(file, id3v2.Options{Parse: true})
		if err != nil {
			return
		}
		tag.SetDefaultEncoding(id3v2.EncodingUTF8)
		if imgResp, err := http.Get(app.AddResizeParamForPicUrl(song.PicUrl, 1024)); err == nil {
			defer imgResp.Body.Close()
			if data, err := io.ReadAll(imgResp.Body); err == nil {
				tag.AddAttachedPicture(id3v2.PictureFrame{
					Encoding:    id3v2.EncodingUTF8,
					MimeType:    "image/jpg",
					PictureType: id3v2.PTOther,
					Picture:     data,
				})
			}
		}
		tag.SetTitle(song.Name)
		tag.SetAlbum(song.Album.Name)
		tag.SetArtist(song.ArtistName())
		_ = tag.Save()
		_ = tag.Close()
	default:
		metadata, err := songtag.Read(file)
		if err != nil {
			return
		}
		defer metadata.Close()
		_ = metadata.SetAlbum(song.Album.Name)
		_ = metadata.SetArtist(song.ArtistName())
		_ = metadata.SetAlbumArtist(song.Album.ArtistName())
		_ = metadata.SetTitle(song.Name)
		if _, ok := metadata.(*songtag.FLAC); !ok {
			return
		}
		if imgResp, err := http.Get(app.AddResizeParamForPicUrl(song.PicUrl, 1024)); err == nil {
			defer imgResp.Body.Close()
			if data, err := io.ReadAll(imgResp.Body); err == nil {
				img, _ := flacpicture.NewFromImageData(flacpicture.PictureTypeFrontCover, "cover", data, "image/jpeg")
				_ = metadata.(*songtag.FLAC).SetFlacPicture(img)
			}
		}
		_ = metadata.SaveFile(file.Name() + "-tmp")
		_ = os.Rename(file.Name()+"-tmp", file.Name())
	}
}
