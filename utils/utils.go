package utils

import (
	"bytes"
	"embed"
	"encoding/binary"
	"errors"
	"fmt"
	"go-musicfox/pkg/constants"
	"io"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"strings"

	"github.com/buger/jsonparser"
	"go-musicfox/configs"
	ds2 "go-musicfox/pkg/structs"
)

//go:embed embed
var embedDir embed.FS

// GetLocalDataDir 获取本地数据存储目录
func GetLocalDataDir() string {
	// Home目录
	homeDir, err := Home()
	if nil != err {
		panic("未获取到用户Home目录: " + err.Error())
	}

	projectDir := fmt.Sprintf("%s/%s", homeDir, constants.AppLocalDataDir)

	if _, err := os.Stat(projectDir); os.IsNotExist(err) {
		_ = os.Mkdir(projectDir, os.ModePerm)
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

type ResCode uint8

const (
	Success ResCode = iota
	UnknownError
	NetworkError
	NeedLogin
	PasswordError
)

// CheckCode 验证响应码
func CheckCode(code float64) ResCode {
	switch code {
	case 301, 302:
		return NeedLogin
	case 520:
		return NetworkError
	case 200:
		return Success
	}

	return PasswordError
}

// CheckUserInfo 验证用户信息
func CheckUserInfo(user *ds2.User) ResCode {
	if user == nil || user.UserId == 0 {
		return NeedLogin
	}

	return Success
}

// ReplaceSpecialStr 替换特殊字符
func ReplaceSpecialStr(str string) string {
	replaceStr := map[string]string{
		"“": "\"",
		"”": "\"",
		"·": ".",
	}
	for oldStr, newStr := range replaceStr {
		str = strings.ReplaceAll(str, oldStr, newStr)
	}

	return str
}

// GetDailySongs 获取每日歌曲列表
func GetDailySongs(data []byte) (list []ds2.Song) {
	_, _ = jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if song, err := ds2.NewSongFromDailySongsJson(value); err == nil {
			list = append(list, song)
		}

	}, "data", "dailySongs")

	return
}

// GetDailyPlaylists 获取播放列表
func GetDailyPlaylists(data []byte) (list []ds2.Playlist) {

	_, _ = jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if playlist, err := ds2.NewPlaylistFromJson(value); err == nil {
			list = append(list, playlist)
		}
	}, "recommend")

	return
}

// GetSongsOfPlaylist 获取播放列表的歌曲
func GetSongsOfPlaylist(data []byte) (list []ds2.Song) {
	_, _ = jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if song, err := ds2.NewSongFromPlaylistSongsJson(value); err == nil {
			list = append(list, song)
		}

	}, "playlist", "tracks")

	return
}

// GetSongsOfAlbum 获取专辑的歌曲
func GetSongsOfAlbum(data []byte) (list []ds2.Song) {
	_, _ = jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if song, err := ds2.NewSongFromAlbumSongsJson(value); err == nil {
			list = append(list, song)
		}

	}, "songs")

	return
}

// GetPlaylists 获取播放列表
func GetPlaylists(data []byte) (list []ds2.Playlist) {

	_, _ = jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if playlist, err := ds2.NewPlaylistFromJson(value); err == nil {
			list = append(list, playlist)
		}
	}, "playlist")

	return
}

// GetPlaylistsFromHighQuality 获取精品歌单
func GetPlaylistsFromHighQuality(data []byte) (list []ds2.Playlist) {

	_, _ = jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if playlist, err := ds2.NewPlaylistFromJson(value); err == nil {
			list = append(list, playlist)
		}
	}, "playlists")

	return
}

// GetFmSongs 获取每日歌曲列表
func GetFmSongs(data []byte) (list []ds2.Song) {
	_, _ = jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if song, err := ds2.NewSongFromFmJson(value); err == nil {
			list = append(list, song)
		}

	}, "data")

	return
}

// GetIntelligenceSongs 获取心动模式歌曲列表
func GetIntelligenceSongs(data []byte) (list []ds2.Song) {
	_, _ = jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if song, err := ds2.NewSongFromIntelligenceJson(value); err == nil {
			list = append(list, song)
		}

	}, "data")

	return
}

// GetNewAlbums 获取最新专辑列表
func GetNewAlbums(data []byte) (albums []ds2.Album) {
	_, _ = jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {

		if album, err := ds2.NewAlbumFromAlbumJson(value); err == nil {
			albums = append(albums, album)
		}

	}, "albums")

	return
}

// GetTopAlbums 获取专辑列表
func GetTopAlbums(data []byte) (albums []ds2.Album) {
	_, _ = jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {

		if album, err := ds2.NewAlbumFromAlbumJson(value); err == nil {
			albums = append(albums, album)
		}

	}, "monthData")

	return
}

// GetArtistHotAlbums 获取歌手热门专辑列表
func GetArtistHotAlbums(data []byte) (albums []ds2.Album) {
	_, _ = jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {

		if album, err := ds2.NewAlbumFromAlbumJson(value); err == nil {
			albums = append(albums, album)
		}

	}, "hotAlbums")

	return
}

// GetSongsOfSearchResult 获取搜索结果的歌曲
func GetSongsOfSearchResult(data []byte) (list []ds2.Song) {
	_, _ = jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if song, err := ds2.NewSongFromSearchResultJson(value); err == nil {
			list = append(list, song)
		}

	}, "result", "songs")

	return
}

// GetAlbumsOfSearchResult 获取搜索结果的专辑
func GetAlbumsOfSearchResult(data []byte) (list []ds2.Album) {
	_, _ = jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if album, err := ds2.NewAlbumFromAlbumJson(value); err == nil {
			list = append(list, album)
		}

	}, "result", "albums")

	return
}

// GetPlaylistsOfSearchResult 获取搜索结果的歌单
func GetPlaylistsOfSearchResult(data []byte) (list []ds2.Playlist) {
	_, _ = jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if playlist, err := ds2.NewPlaylistFromJson(value); err == nil {
			list = append(list, playlist)
		}

	}, "result", "playlists")

	return
}

// GetArtistsOfSearchResult 获取搜索结果的歌手
func GetArtistsOfSearchResult(data []byte) (list []ds2.Artist) {
	_, _ = jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if artist, err := ds2.NewArtist(value); err == nil {
			list = append(list, artist)
		}

	}, "result", "artists")

	return
}

// GetArtistsOfTopArtists 获取热门歌手
func GetArtistsOfTopArtists(data []byte) (list []ds2.Artist) {
	_, _ = jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if artist, err := ds2.NewArtist(value); err == nil {
			list = append(list, artist)
		}

	}, "artists")

	return
}

// GetSongsOfArtist 获取歌手的歌曲
func GetSongsOfArtist(data []byte) (list []ds2.Song) {
	_, _ = jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if song, err := ds2.NewSongFromArtistSongsJson(value); err == nil {
			list = append(list, song)
		}

	}, "songs")

	return
}

// GetUsersOfSearchResult 从搜索结果中获取用户列表
func GetUsersOfSearchResult(data []byte) (list []ds2.User) {
	_, _ = jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if song, err := ds2.NewUserFromSearchResultJson(value); err == nil {
			list = append(list, song)
		}

	}, "result", "userprofiles")

	return
}

// GetDjRadiosOfSearchResult 从搜索结果中获取电台列表
func GetDjRadiosOfSearchResult(data []byte) (list []ds2.DjRadio) {
	_, _ = jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if radio, err := ds2.NewDjRadioFromJson(value); err == nil {
			list = append(list, radio)
		}

	}, "result", "djRadios")

	return
}

// GetDjRadios 获取电台列表
func GetDjRadios(data []byte) (list []ds2.DjRadio) {
	_, _ = jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if radio, err := ds2.NewDjRadioFromJson(value); err == nil {
			list = append(list, radio)
		}

	}, "djRadios")

	return
}

// GetDjRadiosOfToday 获取今日优选电台列表
func GetDjRadiosOfToday(data []byte) (list []ds2.DjRadio) {
	_, _ = jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if radio, err := ds2.NewDjRadioFromJson(value); err == nil {
			list = append(list, radio)
		}

	}, "data")

	return
}

// GetDjRadiosOfTopDj 获取热门电台列表
func GetDjRadiosOfTopDj(data []byte) (list []ds2.DjRadio) {
	_, _ = jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if radio, err := ds2.NewDjRadioFromJson(value); err == nil {
			list = append(list, radio)
		}

	}, "toplist")

	return
}

// GetSongsOfDjRadio 获取电台节目列表的歌曲
func GetSongsOfDjRadio(data []byte) (list []ds2.Song) {
	_, _ = jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if song, err := ds2.NewSongFromDjRadioProgramJson(value); err == nil {
			list = append(list, song)
		}
	}, "programs")

	return
}

// GetSongsOfDjRank 获取电台节目排行榜列表的歌曲
func GetSongsOfDjRank(data []byte) (list []ds2.Song) {
	_, _ = jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if song, err := ds2.NewSongFromDjRankProgramJson(value); err == nil {
			list = append(list, song)
		}
	}, "toplist")

	return
}

// GetSongsOfDjHoursRank 获取电台节目24小时排行榜列表的歌曲
func GetSongsOfDjHoursRank(data []byte) (list []ds2.Song) {
	_, _ = jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if song, err := ds2.NewSongFromDjRankProgramJson(value); err == nil {
			list = append(list, song)
		}
	}, "data", "list")

	return
}

// GetRanks 获取排行榜
func GetRanks(data []byte) (list []ds2.Rank) {

	_, _ = jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if rank, err := ds2.NewRankFromJson(value); err == nil {
			list = append(list, rank)
		}
	}, "list")

	return
}

// GetSongsOfCloud 获取云盘的歌曲
func GetSongsOfCloud(data []byte) (list []ds2.Song) {
	_, _ = jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if song, err := ds2.NewSongFromCloudJson(value); err == nil {
			list = append(list, song)
		}
	}, "data")

	return
}

// GetDjCategory 获取电台分类
func GetDjCategory(data []byte) (list []ds2.DjCategory) {
	_, _ = jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if cate, err := ds2.NewDjCategoryFromJson(value); err == nil {
			list = append(list, cate)
		}

	}, "categories")

	return
}

// OpenUrl 打开链接
func OpenUrl(url string) error {
	commands := map[string]string{
		"windows": "start",
		"darwin":  "open",
		"linux":   "xdg-open",
	}

	run, ok := commands[runtime.GOOS]
	if !ok {
		return errors.New(fmt.Sprintf("don't know how to open things on %s platform", runtime.GOOS))
	}

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", run, url)
	} else {
		cmd = exec.Command(run, url)
	}
	return cmd.Start()
}

// LoadIniConfig 加载ini配置信息
func LoadIniConfig() *configs.Registry {
	projectDir := GetLocalDataDir()
	return configs.NewRegistryFromIniFile(fmt.Sprintf("%s/%s", projectDir, constants.AppIniFile))
}

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

	curTagArr := strings.Split(strings.Trim(constants.AppVersion, "v"), ".")
	tagArr := strings.Split(strings.Trim(tag, "v"), ".")
	if len(tagArr) >= 1 && len(curTagArr) >= 1 {
		if tagArr[0] > curTagArr[0] {
			return true
		}

		if tagArr[0] < curTagArr[0] {
			return false
		}
	}

	if len(tagArr) >= 2 && len(curTagArr) >= 2 {
		if tagArr[1] > curTagArr[1] {
			return true
		}

		if tagArr[1] < curTagArr[1] {
			return false
		}
	}

	if len(tagArr) >= 3 && len(curTagArr) >= 3 {
		if tagArr[2] > curTagArr[2] {
			return true
		}

		if tagArr[2] < curTagArr[2] {
			return false
		}
	}

	return false
}
