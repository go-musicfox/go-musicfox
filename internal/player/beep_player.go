package player

import (
	"context"
	"io"
	"log/slog"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gopxl/beep"
	"github.com/gopxl/beep/effects"
	"github.com/gopxl/beep/speaker"

	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/types"
	"github.com/go-musicfox/go-musicfox/utils/app"
	"github.com/go-musicfox/go-musicfox/utils/errorx"
	"github.com/go-musicfox/go-musicfox/utils/filex"
	"github.com/go-musicfox/go-musicfox/utils/iox"
	"github.com/go-musicfox/go-musicfox/utils/slogx"
	"github.com/go-musicfox/go-musicfox/utils/timex"
)

const (
	sampleRate       = beep.SampleRate(44100)
	resampleQuiality = 4
)

type beepPlayer struct {
	l sync.Mutex

	curMusic URLMusic
	timer    *timex.Timer

	cacheReader     *os.File
	cacheWriter     *os.File
	cacheDownloaded bool

	curStreamer beep.StreamSeekCloser
	curFormat   beep.Format

	state      types.State
	ctrl       *beep.Ctrl
	volume     *effects.Volume
	timeChan   chan time.Duration
	stateChan  chan types.State
	musicChan  chan URLMusic
	httpClient *http.Client

	close chan struct{}
}

func NewBeepPlayer() *beepPlayer {
	p := &beepPlayer{
		state: types.Stopped,

		timeChan:  make(chan time.Duration, 1),
		stateChan: make(chan types.State, 10),
		musicChan: make(chan URLMusic, 1),
		ctrl: &beep.Ctrl{
			Paused: false,
		},
		volume: &effects.Volume{
			Base:   2,
			Silent: false,
		},
		httpClient: &http.Client{},
		close:      make(chan struct{}),
	}

	errorx.WaitGoStart(p.listen)

	return p
}

// listen 开始监听
func (p *beepPlayer) listen() {
	var (
		done       = make(chan struct{})
		resp       *http.Response
		reader     io.ReadCloser
		err        error
		ctx        context.Context
		cancel     context.CancelFunc
		prevSongId int64
		doneHandle = func() {
			select {
			case done <- struct{}{}:
			case <-p.close:
			}
		}
	)

	if err = speaker.Init(sampleRate, sampleRate.N(time.Millisecond*200)); err != nil {
		panic(err)
	}

	cacheFile := filepath.Join(app.RuntimeDir(), "beep_playing")
	for {
		select {
		case <-p.close:
			if cancel != nil {
				cancel()
			}
			return
		case <-done:
			p.Stop()
		case p.curMusic = <-p.musicChan:
			p.l.Lock()
			p.pausedNoLock()
			if p.timer != nil {
				p.timer.SetPassed(0)
			}
			// 清理上一轮
			if cancel != nil {
				cancel()
			}
			p.reset()
			ctx, cancel = context.WithCancel(context.Background())

			if prevSongId != p.curMusic.Id || !filex.FileOrDirExists(cacheFile) {
				// FIXME: 先这样处理，暂时没想到更好的办法
				_ = os.Remove(cacheFile)
				if p.cacheReader, err = os.OpenFile(cacheFile, os.O_CREATE|os.O_TRUNC|os.O_RDONLY, 0666); err != nil {
					panic(err)
				}
				if p.cacheWriter, err = os.OpenFile(cacheFile, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666); err != nil {
					panic(err)
				}

				if strings.HasPrefix(p.curMusic.URL, "file://") {
					reader, err = os.Open(strings.TrimPrefix(p.curMusic.URL, "file://"))
					if err != nil {
						panic(err)
					}
				} else if resp, err = p.httpClient.Get(p.curMusic.URL); err != nil {
					p.stopNoLock()
					goto nextLoop
				} else {
					reader = resp.Body
				}

				// 边下载边播放
				go func(ctx context.Context, cacheWFile *os.File, read io.ReadCloser) {
					_, _ = iox.CopyClose(ctx, cacheWFile, read)
					p.l.Lock()
					defer p.l.Unlock()
					if p.curStreamer == nil {
						// nil说明外层解析还没开始或解析失败，这里直接退出
						return
					}
					// 除了MP3格式，其他格式无需重载
					if p.curMusic.Type == Mp3 && configs.AppConfig.Player.Beep.Mp3Decoder != types.BeepMiniMp3Decoder {
						// 需再开一次文件，保证其指针变化，否则将概率导致 p.ctrl.Streamer = beep.Seq(……) 直接停止播放
						cacheReader, _ := os.OpenFile(cacheFile, os.O_RDONLY, 0666)
						// 使用新的文件后需手动Seek到上次播放处
						lastStreamer := p.curStreamer
						defer func() { _ = lastStreamer.Close() }()
						pos := lastStreamer.Position()
						if p.curStreamer, p.curFormat, err = DecodeSong(p.curMusic.Type, cacheReader); err != nil {
							p.stopNoLock()
							return
						}
						if pos >= p.curStreamer.Len() {
							pos = p.curStreamer.Len() - 1
						}
						if pos < 0 {
							pos = 1
						}
						_ = p.curStreamer.Seek(pos)
						p.ctrl.Streamer = beep.Seq(p.resampleStreamer(p.curFormat.SampleRate), beep.Callback(doneHandle))
					}
					p.cacheDownloaded = true
				}(ctx, p.cacheWriter, reader)

				N := 512
				if p.curMusic.Type == Flac {
					N *= 4
				}
				if err = iox.WaitForNBytes(p.cacheReader, N, time.Millisecond*100, 50); err != nil {
					slog.Error("WaitForNBytes err", slogx.Error(err))
					p.stopNoLock()
					goto nextLoop
				}
			} else {
				// 单曲循环以及歌单只有一首歌时不再请求网络
				p.cacheDownloaded = true
				if p.cacheReader, err = os.OpenFile(cacheFile, os.O_RDONLY, 0666); err != nil {
					panic(err)
				}
			}

			if p.curStreamer, p.curFormat, err = DecodeSong(p.curMusic.Type, p.cacheReader); err != nil {
				p.stopNoLock()
				goto nextLoop
			}

			slog.Info("current song sample rate", slog.Int("sample_rate", int(p.curFormat.SampleRate)))

			p.ctrl.Streamer = beep.Seq(p.resampleStreamer(p.curFormat.SampleRate), beep.Callback(doneHandle))
			p.volume.Streamer = p.ctrl
			speaker.Play(p.volume)

			// 计时器
			p.timer = timex.NewTimer(timex.Options{
				Duration:       8760 * time.Hour,
				TickerInternal: configs.AppConfig.Main.FrameRate.Interval(),
				OnRun:          func(started bool) {},
				OnPause:        func() {},
				OnDone:         func(stopped bool) {},
				OnTick: func() {
					select {
					case p.timeChan <- p.timer.Passed():
					default:
					}
				},
			})
			p.resumeNoLock()
			prevSongId = p.curMusic.Id

		nextLoop:
			p.l.Unlock()
		}
	}
}

// Play 播放音乐
func (p *beepPlayer) Play(music URLMusic) {
	timer := time.NewTimer(time.Second)
	defer timer.Stop()
	select {
	case p.musicChan <- music:
	case <-timer.C:
	}
}

func (p *beepPlayer) CurMusic() URLMusic {
	return p.curMusic
}

func (p *beepPlayer) setState(state types.State) {
	p.state = state
	select {
	case p.stateChan <- state:
	case <-time.After(time.Second * 2):
	}
}

// State 当前状态
func (p *beepPlayer) State() types.State {
	return p.state
}

// StateChan 状态发生变更
func (p *beepPlayer) StateChan() <-chan types.State {
	return p.stateChan
}

func (p *beepPlayer) PassedTime() time.Duration {
	if p.timer == nil {
		return 0
	}
	return p.timer.Passed()
}

func (p *beepPlayer) PlayedTime() time.Duration {
	if p.timer == nil {
		return 0
	}
	return p.timer.ActualRuntime()
}

// TimeChan 获取定时器
func (p *beepPlayer) TimeChan() <-chan time.Duration {
	return p.timeChan
}

func (p *beepPlayer) Seek(duration time.Duration) {
	if duration < 0 || !p.cacheDownloaded {
		return
	}
	// FIXME: 暂时仅对MP3格式提供跳转功能
	// FLAC格式(其他未测)跳转会占用大量CPU资源，比特率越高占用越高
	// 导致Seek方法卡住20-40秒的时间，之后方可随意跳转
	// minimp3未实现Seek
	if p.curStreamer == nil || p.curMusic.Type != Mp3 || configs.AppConfig.Player.Beep.Mp3Decoder == types.BeepMiniMp3Decoder {
		return
	}
	if p.state == types.Playing || p.state == types.Paused {
		speaker.Lock()
		newPos := p.curFormat.SampleRate.N(duration)

		if newPos < 0 {
			newPos = 0
		}
		if newPos >= p.curStreamer.Len() {
			newPos = p.curStreamer.Len() - 1
		}
		if p.curStreamer != nil {
			err := p.curStreamer.Seek(newPos)
			if err != nil {
				slog.Error("seek error", slogx.Error(err))
			}
		}
		if p.timer != nil {
			p.timer.SetPassed(duration)
		}
		speaker.Unlock()
	}
}

// UpVolume 调大音量
func (p *beepPlayer) UpVolume() {
	if p.volume.Volume >= 0 {
		return
	}
	p.l.Lock()
	defer p.l.Unlock()

	p.volume.Silent = false
	p.volume.Volume += 0.25
}

// DownVolume 调小音量
func (p *beepPlayer) DownVolume() {
	if p.volume.Volume <= -5 {
		return
	}

	p.l.Lock()
	defer p.l.Unlock()

	p.volume.Volume -= 0.25
	if p.volume.Volume <= -5 {
		p.volume.Silent = true
	}
}

func (p *beepPlayer) Volume() int {
	p.l.Lock()
	defer p.l.Unlock()
	floatVolume := (p.volume.Volume + 5) * 100 / 5
	return int(math.Floor(floatVolume + 0.5 + 1e-9)) // 转为0~100存储
}

func (p *beepPlayer) SetVolume(volume int) {
	if volume > 100 {
		volume = 100
	}
	if volume < 0 {
		volume = 0
	}

	p.l.Lock()
	defer p.l.Unlock()
	p.volume.Volume = float64(volume)*5/100 - 5
}

func (p *beepPlayer) pausedNoLock() {
	if p.state != types.Playing {
		return
	}
	p.ctrl.Paused = true
	p.timer.Pause()
	p.setState(types.Paused)
}

// Pause 暂停播放
func (p *beepPlayer) Pause() {
	p.l.Lock()
	defer p.l.Unlock()
	p.pausedNoLock()
}

func (p *beepPlayer) resumeNoLock() {
	if p.state == types.Playing {
		return
	}
	p.ctrl.Paused = false
	go p.timer.Run()
	p.setState(types.Playing)
}

// Resume 继续播放
func (p *beepPlayer) Resume() {
	p.l.Lock()
	defer p.l.Unlock()
	p.resumeNoLock()
}

func (p *beepPlayer) stopNoLock() {
	if p.state == types.Stopped {
		return
	}
	p.ctrl.Paused = true
	p.timer.Pause()
	p.setState(types.Stopped)
}

// Stop 停止
func (p *beepPlayer) Stop() {
	p.l.Lock()
	defer p.l.Unlock()
	p.stopNoLock()
}

// Toggle 切换状态
func (p *beepPlayer) Toggle() {
	switch p.State() {
	case types.Paused, types.Stopped:
		p.Resume()
	case types.Playing:
		p.Pause()
	default:
		p.Resume()
	}
}

// Close 关闭
func (p *beepPlayer) Close() {
	p.l.Lock()
	defer p.l.Unlock()

	if p.timer != nil {
		p.timer.Stop()
	}
	close(p.close)
	speaker.Clear()
	speaker.Close()
}

func (p *beepPlayer) reset() {
	// 关闭旧计时器
	if p.timer != nil {
		p.timer.Stop()
	}
	if p.cacheReader != nil {
		_ = p.cacheReader.Close()
	}
	if p.cacheWriter != nil {
		_ = p.cacheWriter.Close()
	}
	if p.curStreamer != nil {
		_ = p.curStreamer.Close()
		p.curStreamer = nil
	}
	p.cacheDownloaded = false
	speaker.Clear()
}

func (p *beepPlayer) streamer(samples [][2]float64) (n int, ok bool) {
	p.l.Lock()
	defer p.l.Unlock()

	pos := p.curStreamer.Position()
	n, ok = p.curStreamer.Stream(samples)
	err := p.curStreamer.Err()
	if err == nil && (ok || p.cacheDownloaded) {
		return
	}
	p.pausedNoLock()

	retry := 4
	for !ok && retry > 0 {
		if p.curMusic.Type == Flac {
			if err = p.curStreamer.Seek(pos); err != nil {
				return
			}
		}
		errorx.ResetError(p.curStreamer)

		select {
		case <-time.After(time.Second * 5):
			n, ok = p.curStreamer.Stream(samples)
		case <-p.close:
			return
		}
		retry--
	}
	p.resumeNoLock()
	return
}

func (p *beepPlayer) resampleStreamer(old beep.SampleRate) beep.Streamer {
	if old == sampleRate {
		return beep.StreamerFunc(p.streamer)
	}
	return beep.Resample(resampleQuiality, old, sampleRate, beep.StreamerFunc(p.streamer))
}
