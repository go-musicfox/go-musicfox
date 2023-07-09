package player

import (
	"fmt"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-musicfox/go-musicfox/utils"

	"github.com/fhs/gompd/v2/mpd"
)

var stateMapping = map[string]State{
	"play":  Playing,
	"pause": Paused,
	"stop":  Stopped,
}

func mpdErrorHandler(err error, ignore bool) {
	if err == nil {
		return
	}

	utils.Logger().Printf("[ERROR] mpdPlayer, err: %+v", err)
	if !ignore {
		panic(err)
	}
}

type mpdPlayer struct {
	bin        string
	configFile string
	network    string
	address    string

	watcher *mpd.Watcher
	l       sync.Mutex

	curMusic       UrlMusic
	curSongId      int
	timer          *utils.Timer
	latestPlayTime time.Time //避免切歌时产生的stop信号造成影响

	volume    int
	state     State
	timeChan  chan time.Duration
	stateChan chan State
	musicChan chan UrlMusic

	close chan struct{}
}

func NewMpdPlayer(bin, configFile, network, address string) Player {
	cmd := exec.Command(bin)
	if configFile != "" {
		cmd.Args = append(cmd.Args, configFile)
	}

	// 启动前kill
	{
		var killCmd = *cmd
		killCmd.Args = append(killCmd.Args, "--kill")
		output, err := killCmd.CombinedOutput()
		if err != nil {
			utils.Logger().Printf("[WARNIG] MPD kill失败:%s, 详情:\n%s", err, output)
		}
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		panic(fmt.Sprintf("[ERROR] MPD启动失败:%s, 详情:\n%s", err, output))
	}

	client, err := mpd.Dial(network, address)
	mpdErrorHandler(err, false)

	err = client.Clear()
	mpdErrorHandler(err, true)

	err = client.Single(true)
	mpdErrorHandler(err, true)

	watcher, err := mpd.NewWatcher(network, address, "", "player", "mixer")
	mpdErrorHandler(err, false)

	p := &mpdPlayer{
		bin:        bin,
		configFile: configFile,
		network:    network,
		address:    address,
		watcher:    watcher,
		state:      Stopped,
		timeChan:   make(chan time.Duration),
		stateChan:  make(chan State),
		musicChan:  make(chan UrlMusic),
		close:      make(chan struct{}),
	}

	go func() {
		defer utils.Recover(false)
		p.listen()
	}()

	go func() {
		defer utils.Recover(false)
		p.watch()
	}()

	p.SyncMpdStatus("")
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
	_client, err = mpd.Dial(p.network, p.address)
	mpdErrorHandler(err, false)
	return _client
}

func (p *mpdPlayer) SyncMpdStatus(subsystem string) {
	status, err := p.client().Status()
	mpdErrorHandler(err, true)

	state := stateMapping[status["state"]]
	if subsystem == "player" && (state != Stopped || time.Now().Sub(p.latestPlayTime) >= time.Second*2) {
		switch state {
		case Playing:
			if p.timer != nil {
				go p.timer.Run()
			}
			p.setState(Playing)
		case Paused:
			if p.timer != nil {
				p.timer.Pause()
			}
			p.setState(Paused)
		case Stopped:
			if p.timer != nil {
				p.timer.Stop()
			}
			p.setState(Stopped)
		}
	}
	p.volume, _ = strconv.Atoi(status["volume"])
	duration, _ := time.ParseDuration(status["elapsed"] + "s")

	if p.timer != nil {
		p.timer.SetPassed(duration)
		select {
		case p.timeChan <- p.timer.Passed():
		default:
		}
	}
}

// listen 开始监听
func (p *mpdPlayer) listen() {
	var (
		err error
	)

	for {
		select {
		case <-p.close:
			return
		case p.curMusic = <-p.musicChan:
			p.Paused()
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
				url     string
				isCache bool
			)
			if strings.HasPrefix(p.curMusic.Url, "http") {
				url = p.curMusic.Url
				isCache = false
			} else {
				url = path.Base(p.curMusic.Url)
				isCache = true
			}

			if isCache {
				_, err := p.client().Rescan(url)
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

			err = p.client().PlayID(p.curSongId)
			mpdErrorHandler(err, false)
			if !isCache {
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
				p.SyncMpdStatus(subsystem)
			}
		}
	}
}

func (p *mpdPlayer) setState(state State) {
	p.state = state
	select {
	case p.stateChan <- state:
	case <-time.After(time.Second * 2):
	}
}

func (p *mpdPlayer) Play(music UrlMusic) {
	select {
	case p.musicChan <- music:
	default:
	}
}

func (p *mpdPlayer) CurMusic() UrlMusic {
	return p.curMusic
}

func (p *mpdPlayer) Paused() {
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
	case Paused, Stopped:
		p.Resume()
	case Playing:
		p.Paused()
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

func (p *mpdPlayer) TimeChan() <-chan time.Duration {
	return p.timeChan
}

func (p *mpdPlayer) State() State {
	return p.state
}

func (p *mpdPlayer) StateChan() <-chan State {
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
	_ = p.client().SetVolume(p.volume)
}

func (p *mpdPlayer) DownVolume() {
	p.l.Lock()
	defer p.l.Unlock()
	if p.volume-5 <= 0 {
		p.volume = 0
	} else {
		p.volume -= 5
	}
	_ = p.client().SetVolume(p.volume)
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
	_ = p.client().SetVolume(volume)
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

	cmd := exec.Command(p.bin)
	if p.configFile != "" {
		cmd.Args = append(cmd.Args, p.configFile)
	}
	cmd.Args = append(cmd.Args, "--kill")
	_ = cmd.Run()
}
