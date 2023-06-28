package utils

import (
	"strconv"
	"time"

	"github.com/go-musicfox/netease-music/service"
)

// 播放结束上报网易云
func ReportSongEnd(songID int64, sourceID int64, passedTime time.Duration) {
	playendService := service.ScrobbleService{
		ID:       strconv.FormatInt(songID, 10),
		Sourceid: strconv.FormatInt(sourceID, 10),
		Time:     int64(passedTime),
	}
	code, response := playendService.Scrobble()
	if code != 200 {
		Logger().Printf("网易云上报播放结束失败: %v", string(response))
		return
	}
}
