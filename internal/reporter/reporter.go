package reporter

import (
	"sync"
	"time"

	"github.com/go-musicfox/go-musicfox/internal/lastfm"
	"github.com/go-musicfox/go-musicfox/internal/structs"
)

type Service interface {
	// ReportEnd 上报一首歌的结束
	ReportEnd(passedTime time.Duration)

	// ReportStart 上报一首歌的开始
	ReportStart(song structs.Song)

	Shutdown()
}

// MasterReporter 上报服务核心，在内部维护一个当前播放信息
type MasterReporter struct {
	mu          sync.Mutex
	currentSong structs.Song
	reporters   []reporter
}

type Option func(*MasterReporter)

func NewService(options ...Option) Service {
	master := &MasterReporter{}
	for _, option := range options {
		option(master)
	}
	return master
}

func WithLastFM(tracker *lastfm.Tracker) Option {
	return func(m *MasterReporter) {
		if tracker == nil {
			return
		}
		m.reporters = append(m.reporters, newLastFMReporter(tracker))
	}
}

func WithNetease() Option {
	return func(m *MasterReporter) {
		m.reporters = append(m.reporters, newNeteaseReporter())
	}
}

func (m *MasterReporter) ReportStart(song structs.Song) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if song.Id == 0 {
		return
	}

	m.currentSong = song
	for _, r := range m.reporters {
		go func(rp reporter) {
			rp.reportStart(song)
		}(r)
	}
}

func (m *MasterReporter) ReportEnd(passedTime time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.currentSong.Id == 0 {
		return
	}

	if passedTime.Seconds() < 20 {
		return
	}

	song := m.currentSong
	for _, r := range m.reporters {
		go func(rp reporter) {
			rp.reportEnd(song, passedTime)
		}(r)
	}

	m.currentSong = structs.Song{}
}

func (m *MasterReporter) Shutdown() {
	for _, r := range m.reporters {
		r.close()
	}
}

type reporter interface {
	reportStart(song structs.Song)
	reportEnd(song structs.Song, passedTime time.Duration)
	close()
}
