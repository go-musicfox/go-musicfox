package utils

import (
    "bytes"
    "encoding/binary"
    "errors"
    "fmt"
    "github.com/buger/jsonparser"
    "go-musicfox/constants"
    "go-musicfox/ds"
    "os"
    "os/exec"
    "os/user"
    "runtime"
    "strings"
)

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
func CheckUserInfo(user *ds.User) ResCode {
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
func GetDailySongs(data []byte) (list []ds.Song) {
    _, _ = jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
       if song, err := ds.NewSongFromDailySongsJson(value); err == nil {
           list = append(list, song)
       }

    }, "data", "dailySongs")

    return
}

// GetDailyPlaylists 获取播放列表
func GetDailyPlaylists(data []byte) (list []ds.Playlist) {

    _, _ = jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
        if playlist, err := ds.NewPlaylistFromJson(value); err == nil {
            list = append(list, playlist)
        }
    }, "recommend")

    return
}

// GetSongsOfPlaylist 获取播放列表的歌曲
func GetSongsOfPlaylist(data []byte) (list []ds.Song) {
    _, _ = jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
        if song, err := ds.NewSongFromPlaylistSongsJson(value); err == nil {
            list = append(list, song)
        }

    }, "playlist", "tracks")

    return
}

// GetPlaylists 获取播放列表
func GetPlaylists(data []byte) (list []ds.Playlist) {

    _, _ = jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
        if playlist, err := ds.NewPlaylistFromJson(value); err == nil {
            list = append(list, playlist)
        }
    }, "playlist")

    return
}

// GetFmSongs 获取每日歌曲列表
func GetFmSongs(data []byte) (list []ds.Song) {
    _, _ = jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
        if song, err := ds.NewSongFromFmJson(value); err == nil {
            list = append(list, song)
        }

    }, "data")

    return
}

// GetIntelligenceSongs 获取心动模式歌曲列表
func GetIntelligenceSongs(data []byte) (list []ds.Song) {
    _, _ = jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
        if song, err := ds.NewSongFromIntelligenceJson(value); err == nil {
            list = append(list, song)
        }

    }, "data")

    return
}