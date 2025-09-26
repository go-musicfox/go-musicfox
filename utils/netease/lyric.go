package netease

import (
	"fmt"
	"strconv"

	"github.com/buger/jsonparser"
	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/netease-music/service"
)

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

	if data.Original == "" {
		data.Original = "[00:00.00] 暂无歌词~"
	}
	if data.Translated == "" {
		data.Translated = "[00:00.00]"
	}

	return data, nil
}
