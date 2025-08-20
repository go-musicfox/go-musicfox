package netease

import (
	"errors"
	"log/slog"
	"strconv"
	"strings"

	"github.com/buger/jsonparser"
	"github.com/go-musicfox/go-musicfox/utils/mathx"
	"github.com/go-musicfox/netease-music/service"
)

// PlayableInfo 歌曲的 URL 等信息
type PlayableInfo struct {
	URL       string
	MusicType string
	Size      int64
	Quality   service.SongQualityLevel
}

var brMap = map[service.SongQualityLevel]string{
	service.Standard: "320000",
	service.Higher:   "320000",
	service.Exhigh:   "320000",
	service.Lossless: "999000",
	service.Hires:    "999000",
}

// FetchPlayableInfo 从网易云API获取一首歌的可播放信息
func FetchPlayableInfo(songID int64, quality service.SongQualityLevel) (PlayableInfo, error) {
	urlService := service.SongUrlV1Service{
		ID:      strconv.FormatInt(songID, 10),
		Level:   quality,
		SkipUNM: true,
	}
	code, response := urlService.SongUrl()
	if code != 200 {
		return PlayableInfo{}, errors.New(string(response))
	}

	var (
		err1, err2    error
		freeTrialInfo jsonparser.ValueType
	)
	url, err1 := jsonparser.GetString(response, "data", "[0]", "url")
	_, freeTrialInfo, _, err2 = jsonparser.Get(response, "data", "[0]", "freeTrialInfo")
	if err1 != nil || err2 != nil || url == "" || (freeTrialInfo != jsonparser.NotExist && freeTrialInfo != jsonparser.Null) {
		br, ok := brMap[quality]
		if !ok {
			br = "320000"
		}
		s := service.SongUrlService{
			ID: strconv.FormatInt(songID, 10),
			Br: br,
		}
		code, response = s.SongUrl()
		if code != 200 {
			return PlayableInfo{}, errors.New(string(response))
		}
	}

	url, _ = jsonparser.GetString(response, "data", "[0]", "url")

	size, _ := jsonparser.GetInt(response, "data", "[0]", "size")
	if size > 0 {
		slog.Info("music size", "size", mathx.FormatBytes(size))
	}

	musicType, _ := jsonparser.GetString(response, "data", "[0]", "type")
	if musicType = strings.ToLower(musicType); musicType == "" {
		musicType = "mp3"
	}

	return PlayableInfo{
		URL:       url,
		MusicType: musicType,
		Size:      size,
		Quality:   quality,
	}, nil
}
