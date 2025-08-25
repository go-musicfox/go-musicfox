package netease

import (
	"log/slog"
	"strconv"
	"time"

	"github.com/go-musicfox/netease-music/service"
)

func ReportSongEnd(songID int64, sourceID int64, passedTime time.Duration) {
	playendService := service.ScrobbleService{
		ID:       strconv.FormatInt(songID, 10),
		Sourceid: strconv.FormatInt(sourceID, 10),
		Time:     int64(passedTime.Seconds()),
	}
	code, response, err := playendService.Scrobble()
	if err != nil {
		slog.Error("网易云上报播放结束失败", slog.String("error", err.Error()))
		return
	}
	if code != 200 {
		slog.Error("网易云上报播放结束失败", slog.String("response", string(response)))
		return
	}
}
