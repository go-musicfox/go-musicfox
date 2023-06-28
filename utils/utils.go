package utils

import (
	"embed"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
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
func CheckUpdate() bool {
	response, err := http.Get(constants.AppCheckUpdateUrl)
	if err != nil {
		return false
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(response.Body)

	jsonBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return false
	}

	tag, err := jsonparser.GetString(jsonBytes, "tag_name")
	if err != nil {
		return false
	}

	return CompareVersion(tag, constants.AppVersion, false)
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

// DownloadMusic 下载音乐
func DownloadMusic(song structs.Song) {
	errHandler := func(errs ...error) {
		log.Printf("下载歌曲失败, err: %+v", errs)
	}

	url, musicType, err := GetSongUrl(song.Id)
	if err != nil {
		errHandler(err)
		return
	}

	go func(url string, musicType string) {
		downloadDir := configs.ConfigRegistry.MainDownloadDir
		if downloadDir == "" {
			downloadDir = path.Join(GetLocalDataDir(), "download")
		}
		if !FileOrDirExists(downloadDir) {
			_ = os.MkdirAll(downloadDir, os.ModePerm)
		}

		fileName := fmt.Sprintf("%s-%s.%s", song.Name, song.ArtistName(), musicType)
		targetFilename := path.Join(downloadDir, fileName)
		if _, err := os.Stat(targetFilename); err == nil {
			Notify(NotifyContent{
				Title:   "🙅🏻‍文件已存在",
				Text:    song.Name,
				Url:     FileUrl(downloadDir),
				GroupId: constants.GroupID,
			})
			return
		}

		resp, err := http.Get(url)
		if err != nil {
			errHandler(err)
			return
		}
		defer resp.Body.Close()

		f, err := os.CreateTemp("", fileName)
		if err != nil {
			errHandler(err)
			return
		}
		defer os.Remove(f.Name())

		Notify(NotifyContent{
			Title:   "👇🏻正在下载，请稍候...",
			Text:    song.Name,
			Url:     FileUrl(downloadDir),
			GroupId: constants.GroupID,
		})

		_, _ = io.Copy(f, resp.Body)

		version := songtag.CheckVersion(f)
		switch version {
		case songtag.VersionID3v22, songtag.VersionID3v23, songtag.VersionID3v24:
			tag, err := id3v2.ParseReader(f, id3v2.Options{Parse: true})
			if err != nil {
				_ = os.Rename(f.Name(), targetFilename)
				break
			}
			// defer tag.Close() //fix: "The process cannot access the file because it is being used by another process" Err on Windows
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
			tag.Close() //fix: "The process cannot access the file because it is being used by another process" Err on Windows
			err = os.Rename(f.Name(), targetFilename)
			if err != nil && runtime.GOOS == "windows" {
				//fix: Windows下载路径修改为其他盘符时报错：The system cannot move the file to a different disk drive.
				srcFile, _ := os.Open(f.Name())
				dstFile, _ := os.Create(targetFilename)
				defer dstFile.Close()
				_, _ = io.Copy(dstFile, srcFile)
				srcFile.Close()
			}
		default:
			metadata, err := songtag.Read(f)
			if err != nil {
				_ = os.Rename(f.Name(), targetFilename)
				break
			}
			defer metadata.Close()
			_ = metadata.SetAlbum(song.Album.Name)
			_ = metadata.SetArtist(song.ArtistName())
			_ = metadata.SetAlbumArtist(song.Album.ArtistName())
			_ = metadata.SetTitle(song.Name)
			if flac, ok := metadata.(*songtag.FLAC); ok && song.PicUrl != "" {
				if imgResp, err := http.Get(AddResizeParamForPicUrl(song.PicUrl, 1024)); err == nil {
					defer imgResp.Body.Close()
					if data, err := io.ReadAll(imgResp.Body); err == nil {
						img, _ := flacpicture.NewFromImageData(flacpicture.PictureTypeFrontCover, "cover", data, "image/jpeg")
						_ = flac.SetFlacPicture(img)
					}
				}
			}
			_ = metadata.SaveFile(targetFilename)
		}

		Notify(NotifyContent{
			Title:   "✅下载完成",
			Text:    song.Name,
			Url:     FileUrl(downloadDir),
			GroupId: constants.GroupID,
		})
	}(url, musicType)
}

var brMap = map[service.SongQualityLevel]string{
	service.Standard: "320000",
	service.Higher:   "320000",
	service.Exhigh:   "320000",
	service.Lossless: "999000",
	service.Hires:    "999000",
}

func GetSongUrl(songId int64) (url, musicType string, err error) {
	urlService := service.SongUrlV1Service{
		ID:      strconv.FormatInt(songId, 10),
		Level:   configs.ConfigRegistry.MainPlayerSongLevel,
		SkipUNM: true,
	}
	code, response := urlService.SongUrl()
	if code != 200 {
		return "", "", errors.New(string(response))
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
			ID: strconv.FormatInt(songId, 10),
			Br: br,
		}
		code, response = s.SongUrl()
		if code != 200 {
			return "", "", errors.New(string(response))
		}
	}

	url, _ = jsonparser.GetString(response, "data", "[0]", "url")
	musicType, _ = jsonparser.GetString(response, "data", "[0]", "type")
	if musicType = strings.ToLower(musicType); musicType == "" {
		musicType = "mp3"
	}

	return url, musicType, nil
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
