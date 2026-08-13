package reporter

import (
	"log/slog"
	"sync"
	"time"

	"github.com/go-musicfox/go-musicfox/internal/lastfm"
	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/utils/errorx"
)

type Service interface {
	// ReportEnd 上报一首歌的结束
	ReportEnd(passedTime time.Duration)

	// ReportStart 上报一首歌的开始
	ReportStart(song structs.Song)

	Shutdown()
}

// reporterTask 是发给单个 reporter 串行队列的任务
type reporterTask struct {
	start      bool
	song       structs.Song
	passedTime time.Duration
}

// MasterReporter 上报服务核心，在内部维护一个当前播放信息
type MasterReporter struct {
	mu          sync.Mutex
	currentSong structs.Song
	reporters   []reporter
	// queues 与 reporters 一一对应：每个 reporter 一个串行 worker，任务严格
	// 按发起顺序执行（playend(A) 先于 playstart(B)），服务端不会收到乱序事件；
	// 队列缓冲防止上报阻塞切歌路径
	queues []chan reporterTask
	wg     sync.WaitGroup
}

type Option func(*MasterReporter)

func NewService(options ...Option) Service {
	master := &MasterReporter{}
	for _, option := range options {
		option(master)
	}
	master.queues = make([]chan reporterTask, len(master.reporters))
	for i, r := range master.reporters {
		master.queues[i] = make(chan reporterTask, 8)
		master.wg.Add(1)
		errorx.Go(func() {
			defer master.wg.Done()
			for task := range master.queues[i] {
				if task.start {
					r.reportStart(task.song)
				} else {
					r.reportEnd(task.song, task.passedTime)
				}
			}
		})
	}
	return master
}

func WithLastFM(tracker *lastfm.Tracker, skipDjRadio bool) Option {
	return func(m *MasterReporter) {
		if tracker == nil {
			return
		}
		m.reporters = append(m.reporters, newLastFMReporter(tracker, skipDjRadio))
	}
}

func WithNetease() Option {
	return func(m *MasterReporter) {
		m.reporters = append(m.reporters, newNeteaseReporter())
	}
}

// enqueue 把任务投递给单个 reporter 的串行队列；队列满（reporter 卡死）时
// 丢弃并告警，绝不阻塞调用方（切歌路径）
func (m *MasterReporter) enqueue(i int, task reporterTask) {
	select {
	case m.queues[i] <- task:
	default:
		slog.Warn("上报队列已满，丢弃本次上报", slog.Any("task", task))
	}
}

func (m *MasterReporter) ReportStart(song structs.Song) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if song.Id == 0 {
		return
	}

	m.currentSong = song
	task := reporterTask{start: true, song: song}
	for i := range m.reporters {
		m.enqueue(i, task)
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
	task := reporterTask{song: song, passedTime: passedTime}
	for i := range m.reporters {
		m.enqueue(i, task)
	}

	m.currentSong = structs.Song{}
}

func (m *MasterReporter) Shutdown() {
	// 关闭队列并等待 worker 排空：退出前正在进行的 netease/lastfm 上报
	// 不会被丢弃
	for _, q := range m.queues {
		close(q)
	}
	m.wg.Wait()
	for _, r := range m.reporters {
		r.close()
	}
}

type reporter interface {
	reportStart(song structs.Song)
	reportEnd(song structs.Song, passedTime time.Duration)
	close()
}
