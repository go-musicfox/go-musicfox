package lastfm

import (
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/storage"
)

type Tracker struct {
	client *Client
	l      sync.Mutex

	enable          bool // 是否启用，功能不可用时为本地模式
	scrobblePoint   int  // 可 Scrobble 百分比
	onlyFirstArtist bool
	pending         *storage.ScrobbleList // 待上报项
	nowPlaying      storage.Scrobble      // 当前播放

	// processing 保证同一时间至多一个上报 goroutine；consecutiveFailures
	// 记录队头连续可重试失败次数，超过阈值后把队头挪到队尾，避免阻塞
	// 其后所有条目
	processing          bool
	consecutiveFailures int
}

const (
	// maxHeadRetries 是队头连续可重试失败的最大次数，超过后挪到队尾
	maxHeadRetries = 3
	// scrobbleRetryInterval 是可重试失败后的重试间隔
	scrobbleRetryInterval = 5 * time.Second
)

func NewTracker(client *Client) *Tracker {
	t := &Tracker{
		client:          client,
		enable:          configs.AppConfig.Reporter.Lastfm.Enable,
		onlyFirstArtist: configs.AppConfig.Reporter.Lastfm.OnlyFirstArtist,
		pending:         &storage.ScrobbleList{},
	}
	t.setScrobblePoint(configs.AppConfig.Reporter.Lastfm.ScrobblePoint)
	t.pending.InitFromStorage()

	return t
}

func (t *Tracker) processPendingScrobbles() {
	for {
		t.l.Lock()
		if !t.Status() || len(t.pending.Scrobbles) == 0 {
			t.processing = false
			t.l.Unlock()
			return
		}

		head := t.pending.Scrobbles[0]
		retry, err := t.scrobble(head)
		switch {
		case err == nil:
			t.pending.Scrobbles = t.pending.Scrobbles[1:]
			t.consecutiveFailures = 0
		case !retry:
			// 永久失败（已过期/鉴权失效）：丢弃队头
			slog.Error("上报失败，已丢弃", slog.Any("error", err.Error()))
			t.pending.Scrobbles = t.pending.Scrobbles[1:]
			t.consecutiveFailures = 0
		case t.consecutiveFailures >= maxHeadRetries:
			// 队头持续可重试失败：挪到队尾（每条记录自带时间戳，顺序不影响
			// 正确性；14 天过期策略最终兜底），避免阻塞其后所有条目无限堆积
			slog.Warn("队头连续失败，挪到队尾", slog.Any("error", err.Error()))
			t.pending.Scrobbles = append(t.pending.Scrobbles[1:], head)
			t.consecutiveFailures = 0
		default:
			t.consecutiveFailures++
		}

		if err != nil {
			t.pending.Store() // 持久化进度
			t.l.Unlock()
			if retry {
				time.Sleep(scrobbleRetryInterval)
			}
			continue
		}
		t.l.Unlock()
	}
}

func (t *Tracker) updateNowPlaying(scrobble storage.Scrobble) error {
	maxRetries := 3
	retries := 0

	var attempt func() error
	attempt = func() error {
		_, err := t.client.api.Track.UpdateNowPlaying(map[string]any{
			"artist":   scrobble.Artist,
			"track":    scrobble.Track,
			"album":    scrobble.Album,
			"duration": scrobble.Duration,
		})

		retry, err := t.client.errorHandle(err)
		if t.client.NeedAuth() {
			return err
		}
		if retry && retries < maxRetries {
			retries++
			return attempt()
		}
		return err
	}

	return attempt()
}

func (t *Tracker) scrobble(scrobble storage.Scrobble) (retry bool, err error) {
	if t.client.api == nil {
		return false, errors.New("lastfm api not initialized")
	}
	if t.IsScrobbleExpired(scrobble) {
		return false, errors.New("scrobble 已过期")
	}
	_, err = t.client.api.Track.Scrobble(map[string]any{
		"artist":    scrobble.Artist,
		"track":     scrobble.Track,
		"album":     scrobble.Album,
		"timestamp": scrobble.Timestamp,
		"duration":  scrobble.Duration,
	})

	retry, err = t.client.errorHandle(err)
	return
}

func (t *Tracker) Scrobble(scrobble storage.Scrobble) {
	if !t.Status() {
		return
	}
	if t.onlyFirstArtist {
		scrobble.FilterArtist()
	}
	t.l.Lock()
	t.pending.Add(scrobble)
	processing := t.processing
	t.processing = true
	t.l.Unlock()
	// 有界唤醒：同一时间至多一个上报 goroutine，避免每个 Scrobble spawn 一个
	if !processing {
		go t.processPendingScrobbles()
	}
}

func (t *Tracker) Playing(scrobble storage.Scrobble) {
	if t.onlyFirstArtist {
		scrobble.FilterArtist()
	}
	t.nowPlaying = scrobble
	if t.client.NeedAuth() || !t.Status() {
		return
	}
	if err := t.updateNowPlaying(scrobble); err != nil {
		slog.Error("上报当前播放出错: ", slog.Any("error", err.Error()))
	}
}

func (t *Tracker) setScrobblePoint(point int) {
	if point < 50 || point > 100 {
		slog.Error("ScrobblePoint 大小须为 50~100，使用默认值 50")
		point = 50
	}
	t.scrobblePoint = point
}

func (t *Tracker) Status() bool {
	return t.enable
}

func (t *Tracker) Toggle() {
	t.enable = !t.Status()
	if !t.client.NeedAuth() && t.Status() {
		if t.nowPlaying.Track != "" {
			go t.Playing(t.nowPlaying)
		}
		t.l.Lock()
		pending := len(t.pending.Scrobbles)
		processing := t.processing
		t.processing = true
		t.l.Unlock()
		if pending > 0 && !processing {
			go t.processPendingScrobbles()
		}
	}
}

func (t *Tracker) close() {
	t.l.Lock()
	defer t.l.Unlock()
	t.pending.Store()
}

// 清除所有本地及当前 scrobble 待上报记录
func (t *Tracker) Clear() {
	t.pending.Clear()
	t.l.Lock()
	defer t.l.Unlock()
	t.pending = &storage.ScrobbleList{}
}

// IsScrobbleable 对比实际播放时间与音乐总时长（秒）
// 检测是否符合上报条件，见 https://www.last.fm/api/scrobbling
func (t *Tracker) IsScrobbleable(duration, played float64) bool {
	if played <= 30 { // 必须大于 30s
		return false
	}
	if played >= 4*60 { // 大于 4min 直接上报
		return true
	}
	if t.scrobblePoint == 100 && duration-played < 3 { // 近似完整播放
		return true
	}
	return played >= duration*(float64(t.scrobblePoint)/100) // 自定义上报起始比例
}

const scrobbleExpiryDays = 14

// IsScrobbleExpired 校验 Scrobble 是否已过期
func (t *Tracker) IsScrobbleExpired(scrobble storage.Scrobble) bool {
	fourteenDaysAgo := time.Now().AddDate(0, 0, -int(scrobbleExpiryDays)).Unix()
	return scrobble.Timestamp < fourteenDaysAgo
}

func (m *Tracker) Count() int {
	m.l.Lock()
	defer m.l.Unlock()
	return len(m.pending.Scrobbles)
}
