package netease

import (
	"fmt"
	"strconv"

	"github.com/buger/jsonparser"
	"github.com/go-musicfox/netease-music/service"
)

// FetchLyric 从网易云API获取一首歌的歌词。
func FetchLyric(songID int64) (string, error) {
	lrcService := service.LyricService{
		ID: strconv.FormatInt(songID, 10),
	}
	code, response := lrcService.Lyric()
	if code != 200 {
		return "", fmt.Errorf("API returned invalid status code: %v", code)
	}

	lrc, err := jsonparser.GetString(response, "lrc", "lyric")
	if err != nil {
		return "", fmt.Errorf("failed to parse lyric: %w", err)
	}

	return lrc, nil
}
