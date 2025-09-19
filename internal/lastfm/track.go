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
}

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
	t.l.Lock()
	defer t.l.Unlock()
	for len(t.pending.Scrobbles) > 0 {
		if !t.Status() { // 随时终止
			break
		}
		if retry, err := t.scrobble(t.pending.Scrobbles[0]); err != nil {
			slog.Error("上报出错，已暂停", slog.Any("error", err.Error()))
			if !retry {
				t.pending.Scrobbles = t.pending.Scrobbles[1:]
			}
			break
		}
		t.pending.Scrobbles = t.pending.Scrobbles[1:]
	}
	if len(t.pending.Scrobbles) > 0 {
		t.pending.Store() // 更新本地存储队列
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
	t.l.Lock()
	defer t.l.Unlock()
	if t.onlyFirstArtist {
		scrobble.FilterArtist()
	}
	t.pending.Add(scrobble)
	go t.processPendingScrobbles()
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
		if len(t.pending.Scrobbles) > 0 {
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
	return len(m.pending.Scrobbles)
}
