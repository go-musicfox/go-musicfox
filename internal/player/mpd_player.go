package player

import (
	"fmt"
	"log/slog"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fhs/gompd/v2/mpd"

	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/types"
	"github.com/go-musicfox/go-musicfox/utils/errorx"
	"github.com/go-musicfox/go-musicfox/utils/slogx"
	"github.com/go-musicfox/go-musicfox/utils/timex"
)

var stateMapping = map[string]types.State{
	"play":  types.Playing,
	"pause": types.Paused,
	"stop":  types.Stopped,
}

func mpdErrorHandler(err error, ignore bool) {
	if err == nil {
		return
	}

	slog.Error("mpdPlayer caught err", slogx.Error(err))
	if !ignore {
		panic(err)
	}
}

type mpdPlayer struct {
	conf *MpdConfig

	watcher *mpd.Watcher
	l       sync.Mutex

	curMusic       URLMusic
	curSongId      int
	timer          *timex.Timer
	latestPlayTime time.Time // 避免切歌时产生的stop信号造成影响

	volume    int
	state     types.State
	timeChan  chan time.Duration
	stateChan chan types.State
	musicChan chan URLMusic

	close chan struct{}
}

type MpdConfig struct {
	Bin        string
	ConfigFile string
	Network    string
	Address    string
	AutoStart  bool
}

func NewMpdPlayer(conf *MpdConfig) *mpdPlayer {
	cmd := exec.Command(conf.Bin)
	if conf.ConfigFile != "" {
		cmd.Args = append(cmd.Args, conf.ConfigFile)
	}

	if conf.AutoStart {
		// 启动前kill
		{
			killCmd := *cmd
			killCmd.Args = append(killCmd.Args, "--kill")
			output, err := killCmd.CombinedOutput()
			if err != nil {
				slog.Warn("MPD kill失败", slogx.Error(err), slogx.Bytes("detail", output))
			}
		}

		output, err := cmd.CombinedOutput()
		if err != nil {
			panic(fmt.Sprintf("[ERROR] MPD启动失败:%s, 详情:\n%s", err, output))
		}
	}

	client, err := mpd.Dial(conf.Network, conf.Address)
	mpdErrorHandler(err, false)

	err = client.Clear()
	mpdErrorHandler(err, false)

	err = client.Single(true)
	mpdErrorHandler(err, true)

	err = client.Repeat(false)
	mpdErrorHandler(err, false)

	watcher, err := mpd.NewWatcher(conf.Network, conf.Address, "", "player", "mixer")
	mpdErrorHandler(err, false)

	p := &mpdPlayer{
		conf:      conf,
		watcher:   watcher,
		state:     types.Stopped,
		timeChan:  make(chan time.Duration, 1),
		stateChan: make(chan types.State, 10),
		musicChan: make(chan URLMusic, 1),
		close:     make(chan struct{}),
	}

	errorx.WaitGoStart(p.listen)
	errorx.WaitGoStart(p.watch)

	p.syncMpdStatus("")
	return p
}

var _client *mpd.Client

func (p *mpdPlayer) client() *mpd.Client {
	var err error
	if _client != nil {
		if err = _client.Ping(); err == nil {
			return _client
		}
	}
	_client, err = mpd.Dial(p.conf.Network, p.conf.Address)
	mpdErrorHandler(err, false)
	return _client
}

func (p *mpdPlayer) syncMpdStatus(subsystem string) {
	status, err := p.client().Status()
	mpdErrorHandler(err, true)

	state := stateMapping[status["state"]]
	if subsystem == "player" && (state != types.Stopped || time.Since(p.latestPlayTime) >= time.Second*2) {
		switch state {
		case types.Playing:
			if p.timer != nil {
				go p.timer.Run()
			}
			p.setState(types.Playing)
		case types.Paused:
			if p.timer != nil {
				p.timer.Pause()
			}
			p.setState(types.Paused)
		case types.Stopped:
			if p.timer != nil {
				p.timer.Stop()
			}
			p.setState(types.Stopped)
		default:
		}
	}
	if vol := status["volume"]; vol != "" {
		p.volume, _ = strconv.Atoi(vol)
	}
	if elapsed := status["elapsed"]; elapsed != "" {
		duration, _ := time.ParseDuration(elapsed + "s")
		if p.timer != nil {
			p.timer.SetPassed(duration)
			select {
			case p.timeChan <- p.timer.Passed():
			default:
			}
		}
	}
}

// listen 开始监听
func (p *mpdPlayer) listen() {
	var err error

	for {
		select {
		case <-p.close:
			return
		case p.curMusic = <-p.musicChan:
			p.Pause()
			if p.timer != nil {
				p.timer.SetPassed(0)
			}
			p.latestPlayTime = time.Now()
			// 重置
			{
				if p.timer != nil {
					p.timer.Stop()
				}
				if p.curSongId != 0 {
					err = p.client().DeleteID(p.curSongId)
					mpdErrorHandler(err, true)
				}
			}

			var (
				url     = p.curMusic.URL
				isLocal = strings.HasPrefix(p.curMusic.URL, "file://")
			)
			if isLocal {
				url = path.Base(p.curMusic.URL)
			}

			if isLocal {
				_, err = p.client().Rescan(url)
				mpdErrorHandler(err, false)
				for {
					var attr map[string]string
					if attr, err = p.client().Status(); err != nil {
						mpdErrorHandler(err, true)
						break
					}
					if _, ok := attr["updating_db"]; ok {
						continue
					}
					// 确保更新完成
					break
				}
			}

			p.curSongId, err = p.client().AddID(url, 0)
			mpdErrorHandler(err, false)

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

			err = p.client().PlayID(p.curSongId)
			mpdErrorHandler(err, false)
			if !isLocal {
				// Doing this because github.com/fhs/gompd/v2/mpd hasn't implement "addtagid" yet
				command := "addtagid %d %s %s"
				err = p.client().Command(command, p.curSongId, "artist", p.curMusic.ArtistName()).OK()
				mpdErrorHandler(err, true)
				err = p.client().Command(command, p.curSongId, "album", p.curMusic.Album.Name).OK()
				mpdErrorHandler(err, true)
				err = p.client().Command(command, p.curSongId, "title", p.curMusic.Name).OK()
				mpdErrorHandler(err, true)
			}
			p.Resume()
		}
	}
}

func (p *mpdPlayer) watch() {
	for {
		select {
		case <-p.close:
			return
		case subsystem := <-p.watcher.Event:
			if subsystem == "player" || subsystem == "mixer" {
				p.syncMpdStatus(subsystem)
			}
		}
	}
}

func (p *mpdPlayer) setState(state types.State) {
	p.state = state
	select {
	case p.stateChan <- state:
	case <-time.After(time.Second * 2):
	}
}

func (p *mpdPlayer) Play(music URLMusic) {
	timer := time.NewTimer(time.Second)
	defer timer.Stop()
	select {
	case p.musicChan <- music:
	case <-timer.C:
	}
}

func (p *mpdPlayer) CurMusic() URLMusic {
	return p.curMusic
}

func (p *mpdPlayer) Pause() {
	p.l.Lock()
	defer p.l.Unlock()
	err := p.client().Pause(true)
	mpdErrorHandler(err, false)
}

func (p *mpdPlayer) Resume() {
	p.l.Lock()
	defer p.l.Unlock()
	err := p.client().Pause(false)
	mpdErrorHandler(err, false)
}

func (p *mpdPlayer) Stop() {
	p.l.Lock()
	defer p.l.Unlock()
	err := p.client().Pause(true)
	mpdErrorHandler(err, false)
}

func (p *mpdPlayer) Toggle() {
	switch p.State() {
	case types.Paused, types.Stopped:
		p.Resume()
	case types.Playing:
		p.Pause()
	default:
		p.Resume()
	}
}

func (p *mpdPlayer) Seek(duration time.Duration) {
	p.l.Lock()
	defer p.l.Unlock()
	err := p.client().SeekCur(duration, false)
	mpdErrorHandler(err, true)
	if err == nil {
		p.timer.SetPassed(duration)
	}
}

func (p *mpdPlayer) PassedTime() time.Duration {
	if p.timer == nil {
		return 0
	}
	return p.timer.Passed()
}

func (p *mpdPlayer) PlayedTime() time.Duration {
	if p.timer == nil {
		return 0
	}
	return p.timer.ActualRuntime()
}

func (p *mpdPlayer) TimeChan() <-chan time.Duration {
	return p.timeChan
}

func (p *mpdPlayer) State() types.State {
	return p.state
}

func (p *mpdPlayer) StateChan() <-chan types.State {
	return p.stateChan
}

func (p *mpdPlayer) UpVolume() {
	p.l.Lock()
	defer p.l.Unlock()
	if p.volume+5 >= 100 {
		p.volume = 100
	} else {
		p.volume += 5
	}
	mpdErrorHandler(p.client().SetVolume(p.volume), true)
}

func (p *mpdPlayer) DownVolume() {
	p.l.Lock()
	defer p.l.Unlock()
	if p.volume-5 <= 0 {
		p.volume = 0
	} else {
		p.volume -= 5
	}
	mpdErrorHandler(p.client().SetVolume(p.volume), true)
}

func (p *mpdPlayer) Volume() int {
	return p.volume
}

func (p *mpdPlayer) SetVolume(volume int) {
	if volume > 100 {
		volume = 100
	}
	if volume < 0 {
		volume = 0
	}
	p.l.Lock()
	defer p.l.Unlock()

	p.volume = volume
	mpdErrorHandler(p.client().SetVolume(volume), true)
}

func (p *mpdPlayer) Close() {
	p.l.Lock()
	defer p.l.Unlock()

	if p.timer != nil {
		p.timer.Stop()
	}

	err := p.watcher.Close()
	mpdErrorHandler(err, true)

	if p.close != nil {
		close(p.close)
		p.close = nil
	}

	err = p.client().Stop()
	mpdErrorHandler(err, true)

	err = p.client().Close()
	mpdErrorHandler(err, true)

	if p.conf.AutoStart {
		cmd := exec.Command(p.conf.Bin)
		if p.conf.ConfigFile != "" {
			cmd.Args = append(cmd.Args, p.conf.ConfigFile)
		}
		cmd.Args = append(cmd.Args, "--kill")
		_ = cmd.Run()
	}
}
