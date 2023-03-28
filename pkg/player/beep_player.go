package player

import (
	"context"
	"io"
	"net/http"
	"os"
	"path"
	"sync"
	"time"

	"github.com/go-musicfox/go-musicfox/utils"

	"github.com/faiface/beep"
	"github.com/faiface/beep/effects"
	"github.com/faiface/beep/speaker"
)

type beepPlayer struct {
	l sync.Mutex

	curMusic UrlMusic
	timer    *utils.Timer

	cacheReader *os.File
	cacheWriter *os.File

	curStreamer beep.StreamSeekCloser
	curFormat   beep.Format

	state      State
	ctrl       *beep.Ctrl
	volume     *effects.Volume
	timeChan   chan time.Duration
	stateChan  chan State
	musicChan  chan UrlMusic
	httpClient *http.Client

	close chan struct{}
}

func NewBeepPlayer() Player {
	p := &beepPlayer{
		state: Stopped,

		timeChan:  make(chan time.Duration),
		stateChan: make(chan State),
		musicChan: make(chan UrlMusic),
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

	go func() {
		defer utils.Recover(false)
		p.listen()
	}()
	return p
}

// listen 开始监听
func (p *beepPlayer) listen() {

	var (
		done   = make(chan struct{})
		resp   *http.Response
		err    error
		ctx    context.Context
		cancel context.CancelFunc
	)

	cacheFile := path.Join(utils.GetLocalDataDir(), "music_cache")
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
			p.Paused()
			if p.timer != nil {
				p.timer.SetPassed(0)
			}
			// 清理上一轮
			if cancel != nil {
				cancel()
			}
			p.reset()

			ctx, cancel = context.WithCancel(context.Background())

			// FIXME 先这样处理，暂时没想到更好的办法
			if p.cacheReader, err = os.OpenFile(cacheFile, os.O_CREATE|os.O_TRUNC|os.O_RDONLY, 0666); err != nil {
				panic(err)
			}
			if p.cacheWriter, err = os.OpenFile(cacheFile, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666); err != nil {
				panic(err)
			}

			if resp, err = p.httpClient.Get(p.curMusic.Url); err != nil {
				p.Stop()
				break
			}

			go func(ctx context.Context, cacheWFile *os.File, read io.ReadCloser) {
				defer utils.Recover(false)
				_, _ = utils.CopyClose(ctx, cacheWFile, read)
				// 除了MP3格式，其他格式无需重载
				if p.curMusic.Type == Mp3 {
					if p.curStreamer, p.curFormat, err = DecodeSong(p.curMusic.Type, p.cacheReader); err != nil {
						p.Stop()
					}
				}

			}(ctx, p.cacheWriter, resp.Body)

			if err = utils.WaitForNBytes(p.cacheReader, 512, time.Millisecond*100, 50); err != nil {
				p.Stop()
				break
			}

			if p.curStreamer, p.curFormat, err = DecodeSong(p.curMusic.Type, p.cacheReader); err != nil {
				p.Stop()
				break
			}

			if err = speaker.Init(p.curFormat.SampleRate, p.curFormat.SampleRate.N(time.Millisecond*200)); err != nil {
				panic(err)
			}

			p.ctrl.Streamer = beep.Seq(p.curStreamer, beep.Callback(func() { done <- struct{}{} }))
			p.volume.Streamer = p.ctrl
			speaker.Play(p.volume)

			// 计时器
			p.timer = utils.NewTimer(utils.Options{
				Duration:       8760 * time.Hour,
				TickerInternal: 200 * time.Millisecond,
				OnRun:          func(started bool) {},
				OnPaused: func() {
					// 暂停播放仍然更新界面
					ticker := time.NewTicker(time.Millisecond * 200)
					for {
						select {
						case <-ticker.C:
							p.timeChan <- p.timer.Passed()
							if p.state != Paused {
								ticker.Stop()
								return
							}
						}
					}
				},
				OnDone: func(stopped bool) {},
				OnTick: func() {
					select {
					case p.timeChan <- p.timer.Passed():
					default:
					}
				},
			})
			p.Resume()
		}
	}
}

// Play 播放音乐
func (p *beepPlayer) Play(music UrlMusic) {
	select {
	case p.musicChan <- music:
	default:
	}
}

func (p *beepPlayer) CurMusic() UrlMusic {
	return p.curMusic
}

func (p *beepPlayer) setState(state State) {
	p.state = state
	select {
	case p.stateChan <- state:
	case <-time.After(time.Second * 2):
	}
}

// State 当前状态
func (p *beepPlayer) State() State {
	return p.state
}

// StateChan 状态发生变更
func (p *beepPlayer) StateChan() <-chan State {
	return p.stateChan
}

func (p *beepPlayer) PassedTime() time.Duration {
	if p.timer == nil {
		return 0
	}
	return p.timer.Passed()
}

// TimeChan 获取定时器
func (p *beepPlayer) TimeChan() <-chan time.Duration {
	return p.timeChan
}

func (p *beepPlayer) Seek(duration time.Duration) {
	// FIXME: 暂时仅对MP3格式提供跳转功能
	// FLAC格式(其他未测)跳转会占用大量CPU资源，比特率越高占用越高
	// 导致Seek方法卡住20-40秒的时间，之后方可随意跳转
	if p.curMusic.Type != Mp3 {
		return
	}
	if p.state == Playing || p.state == Paused {

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
				utils.Logger().Printf("seek error: %+v", err)
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
	return int((p.volume.Volume + 5) * 100 / 5) // 转为0~100存储
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

// Paused 暂停播放
func (p *beepPlayer) Paused() {
	if p.state != Playing {
		return
	}
	p.l.Lock()
	defer p.l.Unlock()
	p.ctrl.Paused = true
	p.timer.Pause()
	p.setState(Paused)
}

// Resume 继续播放
func (p *beepPlayer) Resume() {
	if p.state == Playing {
		return
	}
	p.l.Lock()
	defer p.l.Unlock()
	p.ctrl.Paused = false
	go p.timer.Run()
	p.setState(Playing)
}

// Stop 停止
func (p *beepPlayer) Stop() {
	if p.state == Stopped {
		return
	}
	p.l.Lock()
	defer p.l.Unlock()
	p.ctrl.Paused = true
	p.timer.Pause()
	p.setState(Stopped)
}

// Toggle 切换状态
func (p *beepPlayer) Toggle() {
	switch p.State() {
	case Paused, Stopped:
		p.Resume()
	case Playing:
		p.Paused()
	}
}

// Close 关闭
func (p *beepPlayer) Close() {
	p.l.Lock()
	defer p.l.Unlock()

	if p.timer != nil {
		p.timer.Stop()
	}
	speaker.Clear()
	if p.close != nil {
		close(p.close)
		p.close = nil
	}
}

func (p *beepPlayer) reset() {
	speaker.Clear()
	speaker.Close()
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
	}
}
