package player

import (
	"context"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/effects"
	"github.com/faiface/beep/flac"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/vorbis"
	"github.com/faiface/beep/wav"
	"go-musicfox/utils"
)

type player struct {
	curMusic UrlMusic
	timer    *utils.Timer

	state     State
	ctrl      *beep.Ctrl
	volume    *effects.Volume
	timeChan  chan time.Duration
	stateChan chan State

	musicChan chan UrlMusic
}

func NewPlayer() Player {
	p := &player{
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
	}

	go func() {
		defer utils.Recover(false)
		p.listen()
	}()

	return p
}

// listen 开始监听
func (p *player) listen() {
	done := make(chan bool)

	var (
		streamer   beep.StreamSeekCloser
		cacheRFile *os.File
		cacheWFile *os.File
		resp       *http.Response
		format     beep.Format
		err        error
		ctx        context.Context
		cancel     context.CancelFunc
	)

	cacheFile := utils.GetLocalDataDir() + "/music_cache"

	for {
		select {
		case <-done:
			p.setState(Stopped)
			break
		case p.curMusic = <-p.musicChan:
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
				if streamer != nil {
					_ = streamer.Close()
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

			resp, err = http.Get(p.curMusic.Url)
			if err != nil {
				p.setState(Stopped)
				break
			}

			go func(ctx context.Context, cacheWFile *os.File, read io.ReadCloser) {
				defer utils.Recover(false)
				_, _ = utils.Copy(ctx, cacheWFile, read)
			}(ctx, cacheWFile, resp.Body)

			for {
				t := make([]byte, 5)
				_, err = io.ReadFull(cacheRFile, t)
				if err != io.EOF {
					_, _ = cacheRFile.Seek(0, 0)
					break
				}
			}

			switch p.curMusic.Type {
			case Mp3:
				streamer, format, err = mp3.Decode(cacheRFile)
			case Wav:
				streamer, format, err = wav.Decode(cacheRFile)
			case Ogg:
				streamer, format, err = vorbis.Decode(cacheRFile)
			case Flac:
				streamer, format, err = flac.Decode(cacheRFile)
			default:
				p.setState(Stopped)
				break
			}
			if err != nil {
				p.setState(Stopped)
				break
			}

			sampleRate := format.SampleRate
			if err = speaker.Init(sampleRate, sampleRate.N(time.Millisecond*200)); err != nil {
				panic(err)
			}

			newStreamer := beep.Resample(3, format.SampleRate, sampleRate, streamer)
			p.ctrl.Streamer = beep.Seq(newStreamer, beep.Callback(func() {
				done <- true
			}))
			p.volume.Streamer = p.ctrl
			p.ctrl.Paused = false
			speaker.Play(p.volume)
			p.setState(Playing)

			// 启动计时器
			p.timer = utils.NewTimer(utils.Options{
				Duration:       24 * time.Hour,
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

			go p.timer.Run()
		}
	}
}

// Play 播放音乐
func (p *player) Play(songType SongType, url string, duration time.Duration) {
	music := UrlMusic{
		url,
		songType,
		duration,
	}
	select {
	case p.musicChan <- music:
	default:
	}
}

func (p *player) CurMusic() UrlMusic {
	return p.curMusic
}

func (p *player) setState(state State) {
	p.state = state
	select {
	case p.stateChan <- state:
	default:
	}
}

// State 当前状态
func (p *player) State() State {
	return p.state
}

// StateChan 状态发生变更
func (p *player) StateChan() <-chan State {
	return p.stateChan
}

func (p *player) PassedTime() time.Duration {
	if p.timer == nil {
		return 0
	}
	return p.timer.Passed()
}

// TimeChan 获取定时器
func (p *player) TimeChan() <-chan time.Duration {
	return p.timeChan
}

// UpVolume 调大音量
func (p *player) UpVolume() {
	if p.volume.Volume > 0 {
		return
	}

	speaker.Lock()
	p.volume.Silent = false
	p.volume.Volume += 0.5
	speaker.Unlock()
}

// DownVolume 调小音量
func (p *player) DownVolume() {
	if p.volume.Volume <= -5 {
		speaker.Lock()
		p.volume.Silent = true
		speaker.Unlock()
		return
	}

	speaker.Lock()
	p.volume.Volume -= 0.5
	speaker.Unlock()
}

// Paused 暂停播放
func (p *player) Paused() {
	speaker.Lock()
	p.ctrl.Paused = true
	p.timer.Pause()
	p.setState(Paused)
	speaker.Unlock()
}

// Resume 继续播放
func (p *player) Resume() {
	speaker.Lock()
	p.ctrl.Paused = false
	go p.timer.Run()
	p.setState(Playing)
	speaker.Unlock()
}

// Stop 停止
func (p *player) Stop() {
	speaker.Lock()
	p.ctrl.Paused = true
	p.timer.Pause()
	p.setState(Stopped)
	speaker.Unlock()
}

// Close 关闭
func (p *player) Close() {
	p.timer.Stop()
	speaker.Clear()
}
