package utils

import (
	"errors"
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

type Player struct {
	State  State
	ctrl   *beep.Ctrl
	volume *effects.Volume
	timer  chan time.Duration
	done   chan struct{}
	isInit bool // 是否初始化
}

func NewPlayer() *Player {
	player := new(Player)
	player.timer = make(chan time.Duration)
	player.done = make(chan struct{})
	player.ctrl = &beep.Ctrl{
		Paused: false,
	}
	player.volume = &effects.Volume{
		Base:   2,
		Silent: false,
	}

	return player
}

// Play 播放音乐
func (p *Player) Play(songType SongType, url string) error {
	// 清理旧streamer
	speaker.Clear()

	resp, err := http.Get(url)
	if err != nil {
		p.done <- struct{}{}
		return err
	}
	defer resp.Body.Close()

	var (
		streamer beep.StreamSeekCloser
		format beep.Format
	)

	switch songType {
	case Mp3:
		streamer, format, err = mp3.Decode(resp.Body)
	case Wav:
		streamer, format, err = wav.Decode(resp.Body)
	case Ogg:
		streamer, format, err = vorbis.Decode(resp.Body)
	case Flac:
		streamer, format, err = flac.Decode(resp.Body)
	default:
		p.done <- struct{}{}
		return errors.New("类型错误")
	}

	if err != nil {
		p.done <- struct{}{}
		return err
	}
	defer streamer.Close()

	if !p.isInit {
		err = speaker.Init(format.SampleRate, format.SampleRate.N(time.Millisecond*200))
		if err != nil {
			p.done <- struct{}{}
			return err
		}
		p.isInit = true
	}

	p.State = Playing
	done := make(chan bool)
	p.ctrl.Streamer = beep.Seq(streamer, beep.Callback(func() {
		done <- true
	}))
	p.volume.Streamer = p.ctrl
	p.ctrl.Paused = false
	speaker.Play(p.volume)

	for {
		select {
		case <-done:
			p.State = Stopped
			select {
			case p.done <- struct{}{}:
			default:
			}
			return nil
		case <-time.After(time.Second):
			speaker.Lock()
			select {
			case p.timer<-format.SampleRate.D(streamer.Position()).Round(time.Second):
			default:
			}
			speaker.Unlock()
		}
	}
}

// Timer 获取定时器
func (p *Player) Timer() <-chan time.Duration {
	return p.timer
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
}

// Close 关闭
func (p *Player) Close() {
	speaker.Clear()
}