package utils

import (
	"github.com/faiface/beep"
	"github.com/faiface/beep/effects"
	"github.com/faiface/beep/flac"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/vorbis"
	"github.com/faiface/beep/wav"
	"net/http"
	"time"
)

// State 播放器状态
type State uint8

const (
	Stopped State = iota
	Paused
	Playing
)

// SongType 歌曲类型
type SongType uint8

const (
	Mp3 SongType = iota
	Wav
	Ogg
	Flac
)

type UrlMusic struct {
	Url  string
	Type SongType
}

type Player struct {
	State    State
	ctrl     *beep.Ctrl
	volume   *effects.Volume
	timer    *Timer
	timeChan chan time.Duration
	done     chan struct{}

	musicChan chan UrlMusic
}

func NewPlayer() *Player {
	player := new(Player)
	player.timeChan = make(chan time.Duration)
	player.done = make(chan struct{})
	player.musicChan = make(chan UrlMusic)
	player.ctrl = &beep.Ctrl{
		Paused: false,
	}
	player.volume = &effects.Volume{
		Base:   2,
		Silent: false,
	}

	go func() {
		player.listen()
	}()

	return player
}

// listen 开始监听
func (p *Player) listen() {
	var sampleRate beep.SampleRate = 44100
	err := speaker.Init(sampleRate, sampleRate.N(time.Millisecond*200))
	if err != nil {
		panic(err)
	}
	done := make(chan bool)

	var (
		streamer, oldStreamer beep.StreamSeekCloser
	)

	for {
		select {
		case <-done:
			p.State = Stopped
			p.pushDone()
			break
		case music := <-p.musicChan:
			var (
				resp   *http.Response
				format beep.Format
			)

			speaker.Clear()

			// 关闭旧响应body
			if resp != nil {
				_ = resp.Body.Close()
			}

			// 关闭旧计时器
			if p.timer != nil {
				p.timer.Stop()
				p.timer = nil
			}

			resp, err = http.Get(music.Url)
			if err != nil {
				p.pushDone()
				break
			}

			oldStreamer = streamer
			switch music.Type {
			case Mp3:
				streamer, format, err = mp3.Decode(resp.Body)
			case Wav:
				streamer, format, err = wav.Decode(resp.Body)
			case Ogg:
				streamer, format, err = vorbis.Decode(resp.Body)
			case Flac:
				streamer, format, err = flac.Decode(resp.Body)
			default:
				p.pushDone()
				break
			}
			if err != nil {
				p.pushDone()
				break
			}

			p.State = Playing
			newStreamer := beep.Resample(3, format.SampleRate, sampleRate, streamer)
			p.ctrl.Streamer = beep.Seq(newStreamer, beep.Callback(func() {
				done <- true
			}))
			p.volume.Streamer = p.ctrl
			p.ctrl.Paused = false
			speaker.Play(p.volume)

			// 启动计时器
			p.timer = New(Options{
				Duration:       24*time.Hour,
				TickerInternal: 200*time.Millisecond,
				OnRun: func(started bool) {},
				OnPaused: func() {},
				OnDone: func(stopped bool) {},
				OnTick: func() {
					select {
					case p.timeChan <- p.timer.Passed():
					default:
					}
				},
			})
			go p.timer.Run()


			// 关闭旧Streamer，避免协程泄漏
			if oldStreamer != nil {
				_ = oldStreamer.Close()
			}
		}

	}

}

// Play 播放音乐
func (p *Player) Play(songType SongType, url string) {
	music := UrlMusic{
		url,
		songType,
	}
	p.musicChan <- music
}

func (p *Player) pushDone() {
	select {
	case p.done <- struct{}{}:
	default:
	}
}

// Timer 获取定时器
func (p *Player) Timer() <-chan time.Duration {
	return p.timeChan
}

// Done done chan, 如果播放完成往chan中写入struct{}
func (p *Player) Done() <-chan struct{} {
	return p.done
}

// UpVolume 调大音量
func (p *Player) UpVolume() {
	if p.volume.Volume > 0 {
		return
	}

	speaker.Lock()
	p.volume.Silent = false
	p.volume.Volume += 0.5
	speaker.Unlock()
}

// DownVolume 调小音量
func (p *Player) DownVolume() {
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
func (p *Player) Paused() {
	if p.State == Paused {
		return
	}

	speaker.Lock()
	p.ctrl.Paused = true
	speaker.Unlock()
	p.State = Paused
	p.timer.Pause()
}

// Resume 继续播放
func (p *Player) Resume() {
	if p.State == Playing {
		return
	}

	speaker.Lock()
	p.ctrl.Paused = false
	speaker.Unlock()
	p.State = Playing
	go p.timer.Run()
}

// Close 关闭
func (p *Player) Close() {
	p.timer.Stop()
	p.timer = nil
	speaker.Clear()
}
