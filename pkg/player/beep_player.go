package player

import (
	"context"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/effects"
	"github.com/faiface/beep/flac"
	"github.com/faiface/beep/minimp3"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/vorbis"
	"github.com/faiface/beep/wav"
	minimp3pkg "github.com/tosone/minimp3"
	"go-musicfox/utils"
)

type beepPlayer struct {
	l sync.Mutex

	curMusic UrlMusic
	timer    *utils.Timer

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
	done := make(chan bool)

	var (
		cacheRFile *os.File
		cacheWFile *os.File
		resp       *http.Response
		err        error
		ctx        context.Context
		cancel     context.CancelFunc
	)

	cacheFile := utils.GetLocalDataDir() + "/music_cache"
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
			// 重置
			{
				speaker.Clear()
				speaker.Close()
				if cancel != nil {
					cancel()
				}
				// 关闭旧响应body
				if resp != nil {
					_ = resp.Body.Close()
				}
				// 关闭旧计时器
				if p.timer != nil {
					p.timer.Stop()
				}
				if cacheRFile != nil {
					_ = cacheRFile.Close()
				}
				if cacheWFile != nil {
					_ = cacheWFile.Close()
				}
				if p.curStreamer != nil {
					_ = p.curStreamer.Close()
				}
			}

			ctx, cancel = context.WithCancel(context.Background())

			// FIXME 先这样处理，暂时没想到更好的办法
			cacheRFile, err = os.OpenFile(cacheFile, os.O_CREATE|os.O_TRUNC|os.O_RDONLY, 0666)
			if err != nil {
				panic(err)
			}
			cacheWFile, err = os.OpenFile(cacheFile, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
			if err != nil {
				panic(err)
			}

			resp, err = p.httpClient.Get(p.curMusic.Url)
			if err != nil {
				p.Stop()
				break
			}

			go func(ctx context.Context, cacheWFile *os.File, read io.ReadCloser) {
				defer utils.Recover(false)
				_, _ = utils.Copy(ctx, cacheWFile, read)
			}(ctx, cacheWFile, resp.Body)

			for i := 0; i < 50; i++ {
				t := make([]byte, 256)
				_, err = io.ReadFull(cacheRFile, t)
				_, _ = cacheRFile.Seek(0, 0)
				if err != io.EOF {
					break
				}
				<-time.After(time.Millisecond * 100)
			}

			switch p.curMusic.Type {
			case Mp3:
				minimp3pkg.BufferSize = 1024 * 60
				p.curStreamer, p.curFormat, err = minimp3.Decode(cacheRFile)
			case Wav:
				p.curStreamer, p.curFormat, err = wav.Decode(cacheRFile)
			case Ogg:
				p.curStreamer, p.curFormat, err = vorbis.Decode(cacheRFile)
			case Flac:
				p.curStreamer, p.curFormat, err = flac.Decode(cacheRFile)
			default:
				p.Stop()
				break
			}
			if err != nil {
				p.Stop()
				break
			}

			if err = speaker.Init(p.curFormat.SampleRate, p.curFormat.SampleRate.N(time.Millisecond*200)); err != nil {
				panic(err)
			}

			p.ctrl.Streamer = beep.Seq(p.curStreamer, beep.Callback(func() {
				done <- true
			}))
			p.volume.Streamer = p.ctrl
			speaker.Play(p.volume)

			// 计时器
			p.timer = utils.NewTimer(utils.Options{
				Duration:       8760 * time.Hour,
				TickerInternal: 200 * time.Millisecond,
				OnRun:          func(started bool) {},
				OnPaused:       func() {},
				OnDone:         func(stopped bool) {},
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
	// 还有问题，暂时不实现
	//if p.curStreamer != nil {
	//	err := p.curStreamer.Seek(p.curStreamer.Position())
	//	fmt.Println(err)
	//	if err != nil {
	//		utils.Logger().Printf("seek error: %+v", err)
	//	}
	//}
	//if p.timer != nil {
	//	p.timer.SetPassed(duration)
	//}
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
	p.l.Lock()
	defer p.l.Unlock()
	if p.volume.Volume <= -5 {
		p.volume.Silent = true
		return
	}

	p.volume.Volume -= 0.25
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
	if p.timer != nil {
		p.timer.Stop()
	}
	speaker.Clear()
	p.close <- struct{}{}
}
