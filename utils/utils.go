package utils

import (
    "bytes"
    "errors"
    "github.com/buger/jsonparser"
    "go-musicfox/ds"
    "os"
    "os/exec"
    "os/user"
    "runtime"
    "strings"
    "time"
)

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

type ResCode uint8
const (
    Success ResCode = iota
    UnknownError
    NetworkError
    NeedLogin
    PasswordError
)

// CheckCode check response code
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

func ReplaceSpecialStr(str string) string {
    replaceStr := map[string]string{
        "“": "\"",
        "”": "\"",
    }
    for oldStr, newStr := range replaceStr {
        str = strings.ReplaceAll(str, oldStr, newStr)
    }

    return str
}

// GetDailySongs 获取每日歌曲列表
func GetDailySongs(data string) (list []ds.Song) {
    _, _ = jsonparser.ArrayEach([]byte(data), func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
       song := ds.Song{}
       if id, err := jsonparser.GetInt(value, "id"); err == nil {
           song.Id = id
       }
       if name, err := jsonparser.GetString(value, "name"); err == nil {
           song.Name = name
       }
       if duration, err := jsonparser.GetInt(value, "dt"); err == nil {
           song.Duration = time.Millisecond * time.Duration(duration)
       }
       if alId, err := jsonparser.GetInt(value, "al", "id"); err == nil {
           song.Album.Id = alId
       }
       if alName, err := jsonparser.GetString(value, "al", "name"); err == nil {
           song.Album.Name = alName
       }
       if alPic, err := jsonparser.GetString(value, "al", "picUrl"); err == nil {
           song.Album.PicUrl = alPic
       }

       _, _ = jsonparser.ArrayEach(value, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
           artist := ds.Artist{}
           if arId, err := jsonparser.GetInt(value, "id"); err == nil {
               artist.Id = arId
           }
           if arName, err := jsonparser.GetString(value, "name"); err == nil {
               artist.Name = arName
           }
           song.Artists = append(song.Artists, artist)
       }, "ar")

        list = append(list, song)
    }, "data", "dailySongs")

    return
}

// GetPlaylists 获取播放列表
func GetPlaylists(data string) (list []ds.Playlist) {

    _, _ = jsonparser.ArrayEach([]byte(data), func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
        playlist := ds.Playlist{}
        if id, err := jsonparser.GetInt(value, "id"); err == nil {
            playlist.Id = id
        }
        if name, err := jsonparser.GetString(value, "name"); err == nil {
            playlist.Name = name
        }
        if playlistType, err := jsonparser.GetInt(value, "type"); err == nil {
            playlist.Type = int(playlistType)
        }
        if copywriter, err := jsonparser.GetString(value, "copywriter"); err == nil {
            playlist.Copywriter = copywriter
        }

        list = append(list, playlist)
    }, "recommend")

    return
}

// GetSongsOfPlaylist 获取播放列表的歌曲
func GetSongsOfPlaylist(data string) (list []ds.Song) {
    _, _ = jsonparser.ArrayEach([]byte(data), func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
        song := ds.Song{}
        if id, err := jsonparser.GetInt(value, "id"); err == nil {
            song.Id = id
        }
        if name, err := jsonparser.GetString(value, "name"); err == nil {
            song.Name = name
        }
        if duration, err := jsonparser.GetInt(value, "dt"); err == nil {
            song.Duration = time.Millisecond * time.Duration(duration)
        }
        if alId, err := jsonparser.GetInt(value, "al", "id"); err == nil {
            song.Album.Id = alId
        }
        if alName, err := jsonparser.GetString(value, "al", "name"); err == nil {
            song.Album.Name = alName
        }
        if alPic, err := jsonparser.GetString(value, "al", "picUrl"); err == nil {
            song.Album.PicUrl = alPic
        }

        _, _ = jsonparser.ArrayEach(value, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
            artist := ds.Artist{}
            if arId, err := jsonparser.GetInt(value, "id"); err == nil {
                artist.Id = arId
            }
            if arName, err := jsonparser.GetString(value, "name"); err == nil {
                artist.Name = arName
            }
            song.Artists = append(song.Artists, artist)
        }, "ar")

        list = append(list, song)
    }, "playlist", "tracks")

    return
}