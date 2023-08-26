package utils

import (
	"embed"
	"encoding/binary"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/go-musicfox/go-musicfox/pkg/configs"
	"github.com/go-musicfox/go-musicfox/pkg/constants"
	"github.com/go-musicfox/go-musicfox/pkg/structs"

	"github.com/bogem/id3v2/v2"
	"github.com/buger/jsonparser"
	songtag "github.com/frolovo22/tag"
	"github.com/go-flac/flacpicture"
	"github.com/go-musicfox/netease-music/service"
	foldersize "github.com/markthree/go-get-folder-size/src"
	"github.com/pkg/errors"
	"github.com/skip2/go-qrcode"
)

//go:embed embed
var embedDir embed.FS

// GetLocalDataDir 获取本地数据存储目录
func GetLocalDataDir() string {
	var projectDir string
	if root := os.Getenv("MUSICFOX_ROOT"); root != "" {
		projectDir = root
	} else {
		configDir, err := os.UserConfigDir()
		if nil != err {
			panic("未获取到本地数据目录：" + err.Error())
		}
		projectDir = path.Join(configDir, constants.AppLocalDataDir)
	}

	// 如果 projectDir 不存在且未设置 MUSICFOX_ROOT 环境变量
	// 则尝试从默认路径迁移配置
	if !FileOrDirExists(projectDir) {
		home, _ := os.UserHomeDir()
		oldPath := path.Join(home, "."+constants.AppLocalDataDir)
		if !FileOrDirExists(oldPath) {
			_ = os.MkdirAll(projectDir, os.ModePerm)
			return projectDir
		}
		if os.Getenv("MUSICFOX_ROOT") == "" {
			_ = autoMigrateConfigDir(oldPath, projectDir)
		}
	}

	return projectDir
}

// 检查默认路径和 os.UserHomeDir 是否已存在配置文件
// 如果存在则将它们移动到 newPath
func autoMigrateConfigDir(oldPath, newPath string) error {
	if FileOrDirExists(oldPath) {
		return moveDir(oldPath, newPath)
	}

	return errors.New(oldPath + " not exists")
}

func moveDir(oldPath, newPath string) error {
	if oldPath == newPath {
		return errors.New(oldPath + " is the same path as " + newPath)
	}

	_ = os.MkdirAll(path.Dir(newPath), os.ModePerm)
	return os.Rename(oldPath, newPath)
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

// LoadIniConfig 加载ini配置信息
func LoadIniConfig() {
	projectDir := GetLocalDataDir()
	configFile := path.Join(projectDir, constants.AppIniFile)
	if !FileOrDirExists(configFile) {
		_ = CopyFileFromEmbed("embed/go-musicfox.ini", configFile)
	}
	configs.ConfigRegistry = configs.NewRegistryFromIniFile(configFile)
}

// CheckUpdate 检查更新
func CheckUpdate() (bool, string) {
	response, err := http.Get(constants.AppCheckUpdateUrl)
	if err != nil {
		return false, ""
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(response.Body)

	jsonBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return false, ""
	}

	tag, err := jsonparser.GetString(jsonBytes, "tag_name")
	if err != nil {
		return false, ""
	}

	return CompareVersion(tag, constants.AppVersion, false), tag
}

func CompareVersion(v1, v2 string, equal bool) bool {
	var (
		v1IsDev = strings.HasSuffix(v1, "-dev")
		v2IsDev = strings.HasSuffix(v2, "-dev")
	)
	if v1IsDev && !v2IsDev {
		return true
	}
	if !v1IsDev && v2IsDev {
		return false
	}

	v1 = strings.Trim(v1, "v")
	v2 = strings.Trim(v2, "v")
	if equal && v1 == v2 {
		return true
	}
	if v1 != "" && v2 == "" {
		return true
	}

	v1Arr := strings.Split(v1, ".")
	v2Arr := strings.Split(v2, ".")
	if len(v1Arr) >= 1 && len(v2Arr) >= 1 {
		if v1Arr[0] > v2Arr[0] {
			return true
		}
		if v1Arr[0] < v2Arr[0] {
			return false
		}
	}

	if len(v1Arr) >= 2 && len(v2Arr) >= 2 {
		if v1Arr[1] > v2Arr[1] {
			return true
		}
		if v1Arr[1] < v2Arr[1] {
			return false
		}
	}

	if len(v1Arr) >= 3 && len(v2Arr) >= 3 {
		if v1Arr[2] > v2Arr[2] {
			return true
		}
		if v1Arr[2] < v2Arr[2] {
			return false
		}
	}
	return false
}

type FileExistsError struct {
	path string
}

func (e FileExistsError) Error() string {
	return fmt.Sprintf("file %s already exists", e.path)
}

func GetCacheDir() string {
	cacheDir := configs.ConfigRegistry.MainCacheDir
	if cacheDir == "" {
		cacheDir = path.Join(GetLocalDataDir(), "cache")
	}
	return cacheDir
}

func GetDownloadDir() string {
	downloadDir := configs.ConfigRegistry.MainDownloadDir
	if downloadDir == "" {
		downloadDir = path.Join(GetLocalDataDir(), "download")
	}
	return downloadDir
}

func DownloadFile(url, filename, dirname string) error {
	targetFilename := path.Join(dirname, filename)
	if !FileOrDirExists(dirname) {
		_ = os.MkdirAll(dirname, os.ModePerm)
	}
	if _, err := os.Stat(targetFilename); err == nil {
		return FileExistsError{path: targetFilename}
	}
	client := &http.Client{
		Timeout: 60 * time.Second,
	}
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	f, err := os.CreateTemp("", filename)
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())

	_, err = io.Copy(f, resp.Body)
	if err != nil {
		Logger().Printf("[ERROR] 下载歌曲失败, err: %+v", err)
		return err
	}
	err = os.Rename(f.Name(), targetFilename)
	if err != nil && (runtime.GOOS == "windows" || strings.HasSuffix(err.Error(), "invalid cross-device link")) {
		// fix: 当临时文件系统和目标下载位置不在同一磁盘时无法下载文件
		srcFile, _ := os.Open(f.Name())
		dstFile, _ := os.Create(targetFilename)
		defer dstFile.Close()
		_, _ = io.Copy(dstFile, srcFile)
		srcFile.Close()
	}
	return nil
}

func getCacheUri(songId int64) (uri string, ok bool) {
	cacheDir := GetCacheDir()
	if !FileOrDirExists(cacheDir) {
		if configs.ConfigRegistry.MainCacheLimit != 0 {
			_ = os.MkdirAll(cacheDir, os.ModePerm)
		}
		return
	}
	files, err := os.ReadDir(cacheDir)
	if err != nil || len(files) == 0 {
		return
	}
	for _, file := range files {
		if strings.HasPrefix(file.Name(), strconv.FormatInt(songId, 10)) {
			uri = path.Join(cacheDir, file.Name())
			ok = true
			return
		}
	}
	return
}

func CopyCachedSong(song structs.Song) error {
	downloadDir := GetDownloadDir()
	cacheDir := GetCacheDir()
	if !FileOrDirExists(downloadDir) {
		_ = os.MkdirAll(downloadDir, os.ModePerm)
	}
	if !FileOrDirExists(cacheDir) {
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
	targetFilename := path.Join(downloadDir, filename)

	if _, err := os.Stat(targetFilename); err == nil {
		return FileExistsError{path: targetFilename}
	}
	err := os.Rename(oldFilename, targetFilename)
	if err != nil && (runtime.GOOS == "windows" || strings.HasSuffix(err.Error(), "invalid cross-device link")) {
		// fix: 当临时文件系统和目标下载位置不在同一磁盘时无法下载文件
		src, _ := os.Open(oldFilename)
		defer src.Close()
		dst, _ := os.Create(targetFilename)
		defer dst.Close()
		_, _ = io.Copy(dst, src)
	}
	return nil
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
		if imgResp, err := http.Get(AddResizeParamForPicUrl(song.PicUrl, 1024)); err == nil {
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
		tag.Close()
	default:
		metadata, err := songtag.Read(file)
		defer metadata.Close()
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
		if imgResp, err := http.Get(AddResizeParamForPicUrl(song.PicUrl, 1024)); err == nil {
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

func downloadMusic(url, musicType string, song structs.Song, downloadDir string) error {
	filename := fmt.Sprintf("%s-%s.%s", song.Name, song.ArtistName(), musicType)
	// Windows Linux 均不允许文件名中出现 / \ 替换为 _
	filename = strings.Replace(filename, "/", "_", -1)
	filename = strings.Replace(filename, "\\", "_", -1)
	err := DownloadFile(url, filename, downloadDir)
	if err != nil {
		return err
	}
	file, _ := os.OpenFile(path.Join(downloadDir, filename), os.O_RDWR, os.ModePerm)
	SetSongTag(file, song)
	return nil
}

// DownloadMusic 下载音乐
func DownloadMusic(song structs.Song) {
	var (
		errHandler = func(errs ...error) {
			Logger().Printf("[ERROR] 下载歌曲失败, err: %+v", errs)
		}
		err error
	)

	url, musicType, err := GetSongUrl(song)
	if err != nil {
		errHandler(err)
		return
	}

	downloadDir := GetDownloadDir()
	Notify(NotifyContent{
		Title:   "👇🏻正在下载，请稍候...",
		Text:    song.Name,
		Url:     FileUrl(downloadDir),
		GroupId: constants.GroupID,
	})

	if _, ok := getCacheUri(song.Id); ok {
		err = CopyCachedSong(song)
	} else {
		err = downloadMusic(url, musicType, song, downloadDir)
	}
	switch err.(type) {
	case nil:
		Notify(NotifyContent{
			Title:   "✅下载完成",
			Text:    song.Name,
			Url:     FileUrl(downloadDir),
			GroupId: constants.GroupID,
		})
	case FileExistsError:
		Notify(NotifyContent{
			Title:   "🙅🏻‍文件已存在",
			Text:    song.Name,
			Url:     FileUrl(downloadDir),
			GroupId: constants.GroupID,
		})
	default:
		Notify(NotifyContent{
			Title:   "❌下载失败",
			Text:    err.Error(),
			Url:     FileUrl(downloadDir),
			GroupId: constants.GroupID,
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
		Logger().Printf("[ERROR] 缓存歌曲失败, err: %+v", errs)
	}
	var err error
	cacheDir := GetCacheDir()
	size, err := foldersize.Invoke(cacheDir)
	if err != nil {
		errHandler(err)
		return
	}
	if configs.ConfigRegistry.MainCacheLimit != -1 && size > configs.ConfigRegistry.MainCacheLimit*1024*1024 {
		return
	}
	filename := fmt.Sprintf("%d-%d.%s", song.Id, priority[quality], musicType)
	err = DownloadFile(url, filename, cacheDir)
	if err != nil {
		errHandler(err)
		return
	}
	file, err := os.OpenFile(path.Join(cacheDir, filename), os.O_RDWR, os.ModePerm)
	if err != nil {
		return
	}
	SetSongTag(file, song)
	Logger().Printf("[INFO] 缓存歌曲成功: %s", filename)
}

func GetCacheUrl(songId int64) (url, musicType string, ok bool) {
	url, ok = getCacheUri(songId)
	if !ok || path.Base(url) < fmt.Sprintf("%d-%d", songId, priority[configs.ConfigRegistry.MainPlayerSongLevel]) {
		return
	}
	split := strings.Split(path.Base(url), ".")
	musicType = split[len(split)-1]
	ok = true
	return
}

func ClearMusicCache() error {
	cacheDir := GetCacheDir()
	return ClearDir(cacheDir)
}

func ClearDir(dir string) error {
	if FileOrDirExists(dir) {
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

func GetSongUrl(song structs.Song) (url, musicType string, err error) {
	if configs.ConfigRegistry.MainCacheLimit != 0 {
		var ok bool
		if url, musicType, ok = GetCacheUrl(song.Id); ok {
			return
		}
	}

	urlService := service.SongUrlV1Service{
		ID:      strconv.FormatInt(song.Id, 10),
		Level:   configs.ConfigRegistry.MainPlayerSongLevel,
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

	url, _ = jsonparser.GetString(response, "data", "[0]", "url")
	musicType, _ = jsonparser.GetString(response, "data", "[0]", "type")
	if musicType = strings.ToLower(musicType); musicType == "" {
		musicType = "mp3"
	}
	err = nil
	if configs.ConfigRegistry.MainCacheLimit != 0 {
		go CacheMusic(song, url, musicType, urlService.Level)
	}
	return
}

func CopyFileFromEmbed(src, dst string) error {
	var (
		err   error
		srcfd fs.File
		dstfd *os.File
	)

	if srcfd, err = embedDir.Open(src); err != nil {
		return err
	}
	defer srcfd.Close()

	if dstfd, err = os.OpenFile(dst, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0766); err != nil {
		return err
	}
	defer dstfd.Close()

	if _, err = io.Copy(dstfd, srcfd); err != nil {
		return err
	}
	return nil
}

func CopyDirFromEmbed(src, dst string) error {
	var (
		err error
		fds []fs.DirEntry
	)

	if err = os.MkdirAll(dst, 0766); err != nil {
		return err
	}
	if fds, err = embedDir.ReadDir(src); err != nil {
		return err
	}
	for _, fd := range fds {
		srcfp := path.Join(src, fd.Name())
		dstfp := path.Join(dst, fd.Name())

		if fd.IsDir() {
			if err = CopyDirFromEmbed(srcfp, dstfp); err != nil {
				return err
			}
		} else {
			if err = CopyFileFromEmbed(srcfp, dstfp); err != nil {
				return err
			}
		}
	}
	return nil
}

func GenQRCode(filename, content string) (string, error) {
	localDir := GetLocalDataDir()
	filepath := path.Join(localDir, filename)
	if err := qrcode.WriteFile(content, qrcode.Medium, 256, filepath); err != nil {
		return "", err
	}
	return filepath, nil
}

func WebUrlOfPlaylist(playlistId int64) string {
	return "https://music.163.com/#/my/m/music/playlist?id=" + strconv.FormatInt(playlistId, 10)
}

func WebUrlOfSong(songId int64) string {
	return "https://music.163.com/#/song?id=" + strconv.FormatInt(songId, 10)
}

func WebUrlOfArtist(artistId int64) string {
	return "https://music.163.com/#/artist?id=" + strconv.FormatInt(artistId, 10)
}

func WebUrlOfAlbum(artistId int64) string {
	return "https://music.163.com/#/album?id=" + strconv.FormatInt(artistId, 10)
}

func FileUrl(filepath string) string {
	return "file://" + filepath
}

func IsSameDate(t1, t2 time.Time) bool {
	y1, m1, d1 := t1.Date()
	y2, m2, d2 := t2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

func AddResizeParamForPicUrl(picurl string, size int64) string {
	if picurl == "" {
		return ""
	}
	return fmt.Sprintf("%s?param=%dy%d", picurl, size, size)
}

func FileOrDirExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}
