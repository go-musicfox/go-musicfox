package netease

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/buger/jsonparser"
	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/netease-music/service"
	"github.com/go-musicfox/netease-music/util"
)

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// FetchLyric 从网易云API获取一首歌的歌词。
func FetchLyric(songID int64) (structs.LRCData, error) {
	lrcService := service.LyricService{
		ID: strconv.FormatInt(songID, 10),
	}
	code, response := lrcService.Lyric()

	var data structs.LRCData
	if code != 200 {
		return data, fmt.Errorf("netease api returned status code: %v", code)
	}

	if lrc, err := jsonparser.GetString(response, "lrc", "lyric"); err == nil && lrc != "" {
		data.Original = lrc
	}

	if tlyric, err := jsonparser.GetString(response, "tlyric", "lyric"); err == nil && tlyric != "" {
		data.Translated = tlyric
	}

	// 进一步尝试通过新版接口获取逐字歌词(YRC)及其翻译/音译
	options := &util.Options{Crypto: "eapi", Url: "/api/song/lyric/v1", Cookies: []*http.Cookie{{Name: "os", Value: "pc"}}}
	newParams := map[string]string{
		"id":  strconv.FormatInt(songID, 10),
		"cp":  "false",
		"tv":  "0",
		"lv":  "0",
		"rv":  "0",
		"kv":  "0",
		"yv":  "0",
		"ytv": "0",
		"yrv": "0",
	}
	code2, resp2, _ := util.CreateRequest("POST", "https://interface.music.163.com/eapi/song/lyric/v1", newParams, options)
	slog.Debug("[YRC API] response", "code", code2, "respLen", len(resp2))
	if code2 == 200 && len(resp2) > 0 {
		preview := string(resp2[:min(200, len(resp2))])
		slog.Debug("[YRC API] response preview", "preview", preview)
		if yrc, err := jsonparser.GetString(resp2, "yrc", "lyric"); err == nil && yrc != "" {
			data.Yrc = yrc
			slog.Debug("[YRC API] yrc payload", "len", len(yrc))
		} else {
			slog.Debug("[YRC API] missing yrc", "error", err)
		}
		if ytlrc, err := jsonparser.GetString(resp2, "ytlrc", "lyric"); err == nil && ytlrc != "" {
			data.Ytlrc = ytlrc
		}
		if yromalrc, err := jsonparser.GetString(resp2, "yromalrc", "lyric"); err == nil && yromalrc != "" {
			data.Yromalrc = yromalrc
		}
		// 若新版接口也返回了 lrc/tlyric，则以新版为准进行覆盖
		if lrc2, err := jsonparser.GetString(resp2, "lrc", "lyric"); err == nil && lrc2 != "" {
			data.Original = lrc2
		}
		if tlyric2, err := jsonparser.GetString(resp2, "tlyric", "lyric"); err == nil && tlyric2 != "" {
			data.Translated = tlyric2
		}
	}

	if data.Original == "" {
		data.Original = "[00:00.00] 暂无歌词~"
	}
	if data.Translated == "" {
		data.Translated = "[00:00.00]"
	}

	return data, nil
}
