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
    State  State
    ctrl   *beep.Ctrl
    volume *effects.Volume
    timer  chan time.Duration
    done   chan struct{}

    musicChan chan UrlMusic
}

func NewPlayer() *Player {
    player := new(Player)
    player.timer = make(chan time.Duration)
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
        case <-time.After(time.Second):
            if streamer == nil {
                break
            }
            select {
            case p.timer <- sampleRate.D(streamer.Position()).Round(time.Second):
            default:
            }
        case music := <-p.musicChan:
            var resp *http.Response

            speaker.Clear()

            // 关闭旧响应body
            if resp != nil {
                _ = resp.Body.Close()
            }

            resp, err = http.Get(music.Url)
            if err != nil {
                p.pushDone()
                break
            }

            oldStreamer = streamer
            switch music.Type {
            case Mp3:
                streamer, _, err = mp3.Decode(resp.Body)
            case Wav:
                streamer, _, err = wav.Decode(resp.Body)
            case Ogg:
                streamer, _, err = vorbis.Decode(resp.Body)
            case Flac:
                streamer, _, err = flac.Decode(resp.Body)
            default:
                p.pushDone()
                break
            }
            if err != nil {
                p.pushDone()
                break
            }

            p.State = Playing
            p.ctrl.Streamer = beep.Seq(streamer, beep.Callback(func() {
                done <- true
            }))
            p.volume.Streamer = p.ctrl
            p.ctrl.Paused = false
            speaker.Play(p.volume)

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
    p.musicChan<-music
}

func (p *Player) pushDone() {
    select {
    case p.done <- struct{}{}:
    default:
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
