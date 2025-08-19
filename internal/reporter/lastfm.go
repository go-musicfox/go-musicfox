package reporter

import (
	"time"

	"github.com/go-musicfox/go-musicfox/internal/lastfm"
	"github.com/go-musicfox/go-musicfox/internal/storage"
	"github.com/go-musicfox/go-musicfox/internal/structs"
)

type lastFMReporter struct {
	tracker *lastfm.Tracker
}

func newLastFMReporter(tracker *lastfm.Tracker) reporter {
	return &lastFMReporter{
		tracker: tracker,
	}
}

func (l *lastFMReporter) reportStart(song structs.Song) {
	if l.tracker == nil {
		return
	}
	l.tracker.Playing(*storage.NewScrobble(song, 0))
}

func (l *lastFMReporter) reportEnd(song structs.Song, passedTime time.Duration) {
	if l.tracker == nil {
		return
	}
	if l.tracker.IsScrobbleable(song.Duration.Seconds(), passedTime.Seconds()) {
		l.tracker.Scrobble(*storage.NewScrobble(song, passedTime))
	}
}

func (l *lastFMReporter) close() {}
