package utils

import (
	"bytes"
	"embed"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"path"
	"runtime"
	"strconv"
	"strings"

	"go-musicfox/pkg/configs"
	"go-musicfox/pkg/constants"
	"go-musicfox/pkg/structs"

	"github.com/anhoder/netease-music/service"
	"github.com/bogem/id3v2/v2"
	"github.com/buger/jsonparser"
	songtag "github.com/frolovo22/tag"
	"github.com/go-flac/flacpicture"
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
		// Home目录
		homeDir, err := Home()
		if nil != err {
			panic("未获取到用户Home目录: " + err.Error())
		}
		projectDir = path.Join(homeDir, constants.AppLocalDataDir)
	}

	if _, err := os.Stat(projectDir); os.IsNotExist(err) {
		_ = os.MkdirAll(projectDir, os.ModePerm)
	}
	return projectDir
}

// Home 获取当前用户的Home目录
func Home() (string, error) {
	curUser, err := user.Current()
	if nil == err {
		return curUser.HomeDir, nil
	}

	// cross compile support
	if "windows" == runtime.GOOS {
		return homeWindows()
	}

	// Unix-like system, so just assume Unix
	return homeUnix()
}

func homeUnix() (string, error) {
	// First prefer the HOME environmental variable
	if home := os.Getenv("HOME"); home != "" {
		return home, nil
	}

	// If that fails, try the shell
	var stdout bytes.Buffer
	cmd := exec.Command("sh", "-c", "eval echo ~$USER")
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return "", err
	}

	result := strings.TrimSpace(stdout.String())
	if result == "" {
		return "", errors.New("blank output when reading home directory")
	}

	return result, nil
}

func homeWindows() (string, error) {
	drive := os.Getenv("HOMEDRIVE")
	path := os.Getenv("HOMEPATH")
	home := drive + path
	if drive == "" || path == "" {
		home = os.Getenv("USERPROFILE")
	}
	if home == "" {
		return "", errors.New("HOMEDRIVE, HOMEPATH, and USERPROFILE are blank")
	}

	return home, nil
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
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
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
		if _, err = os.Stat(downloadDir); os.IsNotExist(err) {
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
			defer tag.Close()
			tag.SetDefaultEncoding(id3v2.EncodingUTF8)
			if imgResp, err := http.Get(song.PicUrl + "?param=1024y1024"); err == nil {
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
			_ = os.Rename(f.Name(), targetFilename)
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
				if imgResp, err := http.Get(song.PicUrl + "?param=1024y1024"); err == nil {
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
	service.Standard: "128000",
	service.Higher:   "192000",
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

func FileUrl(filepath string) string {
	return "file://" + filepath
}
