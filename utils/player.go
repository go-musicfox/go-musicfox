package utils

import (
    "fmt"
    "github.com/faiface/beep"
    "github.com/faiface/beep/effects"
    "github.com/faiface/beep/flac"
    "github.com/faiface/beep/mp3"
    "github.com/faiface/beep/speaker"
    "github.com/faiface/beep/vorbis"
    "github.com/faiface/beep/wav"
    "io"
    "net/http"
    "os"
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
    Url      string
    Type     SongType
    Duration time.Duration
}

type Player struct {
    State    State
    Progress int
    CurMusic UrlMusic
    ctrl     *beep.Ctrl
    volume   *effects.Volume
    Timer    *Timer
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
        streamer, oldStreamer  beep.StreamSeekCloser
        cacheRFile, cacheWFile *os.File
    )

    cacheFile := fmt.Sprintf("%s/music_cache", GetLocalDataDir())

    for {
        select {
        case <-done:
            p.State = Stopped
            p.pushDone()
            break
        case p.CurMusic = <-p.musicChan:
            var (
                resp   *http.Response
                format beep.Format
            )

            // 打开缓存文件
            if cacheRFile != nil {
                _ = cacheRFile.Close()
            }
            if cacheWFile != nil {
                _ = cacheWFile.Close()
            }
            cacheRFile, err = os.OpenFile(cacheFile, os.O_CREATE|os.O_TRUNC|os.O_RDONLY, 0666)
            if err != nil {
                panic(err)
            }
            cacheWFile, err = os.OpenFile(cacheFile, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
            if err != nil {
                panic(err)
            }

            speaker.Clear()

            // 关闭旧响应body
            if resp != nil {
                _ = resp.Body.Close()
            }

            // 关闭旧计时器
            if p.Timer != nil {
                p.Timer.Stop()
            }
            p.Progress = 0

            resp, err = http.Get(p.CurMusic.Url)
            if err != nil {
                p.pushDone()
                break
            }

            go func(cacheWFile *os.File, read io.ReadCloser) {
                _, _ = io.Copy(cacheWFile, read)
            }(cacheWFile, resp.Body)

            for {
                t := make([]byte, 5)
                _, err = io.ReadFull(cacheRFile, t)
                if err != io.EOF {
                    _, _ = cacheRFile.Seek(0, 0)
                    break
                }
            }

            oldStreamer = streamer
            switch p.CurMusic.Type {
            case Mp3:
                streamer, format, err = mp3.Decode(cacheRFile)
            case Wav:
                streamer, format, err = wav.Decode(cacheRFile)
            case Ogg:
                streamer, format, err = vorbis.Decode(cacheRFile)
            case Flac:
                streamer, format, err = flac.Decode(cacheRFile)
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
            p.Timer = New(Options{
                Duration:       24 * time.Hour,
                TickerInternal: 200 * time.Millisecond,
                OnRun:          func(started bool) {},
                OnPaused:       func() {},
                OnDone:         func(stopped bool) {},
                OnTick: func() {
                    if p.CurMusic.Duration > 0 {
                        p.Progress = int(p.Timer.Passed() * 100 / p.CurMusic.Duration)
                    }
                    select {
                    case p.timeChan <- p.Timer.Passed():
                    default:
                    }
                },
            })
            go p.Timer.Run()

            // 关闭旧Streamer，避免协程泄漏
            if oldStreamer != nil {
                _ = oldStreamer.Close()
            }
        }

    }

}

// Play 播放音乐
func (p *Player) Play(songType SongType, url string, duration time.Duration) {
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

func (p *Player) pushDone() {
    select {
    case p.done <- struct{}{}:
    default:
    }
}

// TimeChan 获取定时器
func (p *Player) TimeChan() <-chan time.Duration {
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
    p.Timer.Pause()
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
    go p.Timer.Run()
}

// Close 关闭
func (p *Player) Close() {
    p.Timer.Stop()
    speaker.Clear()
}
