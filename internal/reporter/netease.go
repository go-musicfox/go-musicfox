package reporter

import (
	"math"
	"strconv"
	"time"

	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/netease-music/service"
)

type neteaseReporter struct {
	svc  *service.ReportService
}

func newNeteaseReporter() reporter {
	return &neteaseReporter{}
}

func (n *neteaseReporter) reportStart(song structs.Song) {
	n.buildNeteaseReportService(song, 0).Playstart()
}

func (n *neteaseReporter) reportEnd(song structs.Song, passedTime time.Duration) {
	svc := n.buildNeteaseReportService(song, passedTime) 

	switch {
	case math.Abs(song.Duration.Seconds()-passedTime.Seconds()) <= 10:
		svc.EndType = "playend"
	default:
		svc.EndType = "ui"
	}
	svc.Playend()
}

func (n *neteaseReporter) close() {}

func (n *neteaseReporter) buildNeteaseReportService(song structs.Song, passedTime time.Duration) *service.ReportService {
	svc := &service.ReportService{
		ID:      song.Id,
		Alg:     song.Alg,
		Type:    "song",
		Time:    int64(passedTime.Seconds()),
		EndType: "playend",
	}

	// 提供基础支持
	if svc.SourceType == "" && song.Album.Id != 0 {
		svc.SourceType = "album"
		svc.SourceId = strconv.FormatInt(song.Album.Id, 10)
	}

	return svc
}
