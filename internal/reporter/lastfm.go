package reporter

import (
	"log/slog"
	"time"

	"github.com/go-musicfox/go-musicfox/internal/lastfm"
	"github.com/go-musicfox/go-musicfox/internal/storage"
	"github.com/go-musicfox/go-musicfox/internal/structs"
)

type lastFMReporter struct {
	tracker     *lastfm.Tracker
	skipDjRadio bool
}

func newLastFMReporter(tracker *lastfm.Tracker, skipDjRadio bool) reporter {
	return &lastFMReporter{
		tracker:     tracker,
		skipDjRadio: skipDjRadio,
	}
}

func (l *lastFMReporter) reportStart(song structs.Song) {
	if l.tracker == nil {
		return
	}

	if l.skipDjRadio {
		if song.DjRadio.Id != 0 {
			slog.Debug("skip report playing djRadio", "name", song.Name, "id", song.Id)
			return
		}
	}

	l.tracker.Playing(*storage.NewScrobble(song, 0))
}

func (l *lastFMReporter) reportEnd(song structs.Song, passedTime time.Duration) {
	if l.tracker == nil {
		return
	}

	if l.skipDjRadio {
		if song.DjRadio.Id != 0 {
			slog.Debug("skip report played djRadio", "name", song.Name, "id", song.Id)
			return
		}
	}

	if l.tracker.IsScrobbleable(song.Duration.Seconds(), passedTime.Seconds()) {
		l.tracker.Scrobble(*storage.NewScrobble(song, passedTime))
	}
}

func (l *lastFMReporter) close() {}
