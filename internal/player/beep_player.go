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
	"sync/atomic"
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

	state      atomic.Uint32 // types.State, atomically read/written (setState runs under p.l; State() is called from the UI thread)
	ctrl       *beep.Ctrl
	volume     *effects.Volume
	timeChan   chan time.Duration
	stateChan  chan types.State
	musicChan  chan URLMusic
	httpClient *http.Client

	close chan struct{}

	spectrum         *PCMAnalyzer
	spectrumConsumer func(sampleRate float64, samplesL, samplesR []float32)

	// spectrumSamplesL/R 是 streamer 馈送频谱数据的复用缓冲（仅 streamer
	// 使用，持 p.l 访问，无并发）
	spectrumSamplesL []float32
	spectrumSamplesR []float32
}

func NewBeepPlayer() *beepPlayer {
	p := &beepPlayer{
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

	if configs.AppConfig.Main.Visualizer.Enable || (configs.AppConfig.Main.Lyric.DesktopLyrics.SpectrumEnabled && desktopLyricsAvailable) {
		p.spectrum = NewPCMAnalyzer(configs.AppConfig.Main.FrameRate.Interval())
	}

	p.state.Store(uint32(types.Stopped))

	errorx.WaitGoStart(p.listen)

	return p
}

// listen 开始监听
func (p *beepPlayer) listen() {
	var (
		// done 是当前歌曲的结束信号通道。每首歌创建独立的缓冲通道：
		// 旧歌残留的结束信号不会误停新歌；缓冲 1 + 非阻塞发送保证拉取路径
		// （持有 speaker 全局锁）绝不会阻塞在发送上。
		done       chan struct{}
		resp       *http.Response
		reader     io.ReadCloser
		err        error
		ctx        context.Context
		cancel     context.CancelFunc
		prevSongId int64
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
		case music := <-p.musicChan:
			started := false
			p.l.Lock()
			// curMusic is read by CurMusic() (UI thread) under p.l, so assign
			// it while holding the lock too.
			p.curMusic = music
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

			// 每首歌独立的结束信号通道（见 listen 顶部注释）
			curDone := make(chan struct{}, 1)
			done = curDone
			doneHandle := func() {
				select {
				case curDone <- struct{}{}:
				default:
				}
			}

			if prevSongId != p.curMusic.Id || !filex.FileOrDirExists(cacheFile) {
				// FIXME: 先这样处理，暂时没想到更好的办法
				_ = os.Remove(cacheFile)
				if p.cacheReader, err = os.OpenFile(cacheFile, os.O_CREATE|os.O_TRUNC|os.O_RDONLY, 0o666); err != nil {
					panic(err)
				}
				if p.cacheWriter, err = os.OpenFile(cacheFile, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o666); err != nil {
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
				songID := p.curMusic.Id
				go func(ctx context.Context, cacheWFile *os.File, read io.ReadCloser, songID int64) {
					_, _ = iox.CopyClose(ctx, cacheWFile, read)
					p.l.Lock()
					defer p.l.Unlock()
					// 下载期间可能已切歌（cancel 打断）：此时 p.curMusic/p.curStreamer
					// 已属于新歌，绝不能用自己（旧歌）的缓存数据替换新歌的 streamer
					//（否则新歌会从旧歌位置开始、被瞬间跳过，甚至被切走）。
					if songID != p.curMusic.Id || p.curStreamer == nil {
						// nil说明外层解析还没开始或解析失败，这里直接退出
						return
					}
					// 除了MP3格式，其他格式无需重载
					if p.curMusic.Type == Mp3 && configs.AppConfig.Player.Beep.Mp3Decoder != types.BeepMiniMp3Decoder {
						// 需再开一次文件，保证其指针变化，否则将概率导致 p.ctrl.Streamer = beep.Seq(……) 直接停止播放
						cacheReader, _ := os.OpenFile(cacheFile, os.O_RDONLY, 0o666)
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
				}(ctx, p.cacheWriter, reader, songID)

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
				if p.cacheReader, err = os.OpenFile(cacheFile, os.O_RDONLY, 0o666); err != nil {
					panic(err)
				}
			}

			if p.curStreamer, p.curFormat, err = DecodeSong(p.curMusic.Type, p.cacheReader); err != nil {
				p.stopNoLock()
				goto nextLoop
			}

			slog.Info("current song sample rate", slog.Int("sample_rate", int(p.curFormat.SampleRate)))

			if p.spectrum != nil {
				p.spectrumConsumer = p.spectrum.NewConsumer()
			}

			p.ctrl.Streamer = beep.Seq(p.resampleStreamer(p.curFormat.SampleRate), beep.Callback(doneHandle))
			p.volume.Streamer = p.ctrl

			// 计时器
			p.timer = timex.NewTimer(timex.Options{
				Duration:       8760 * time.Hour,
				TickerInternal: configs.AppConfig.Main.FrameRate.Interval(),
				OnRun:          func(started bool) {},
				OnPause:        func() {},
				OnDone:         func(stopped bool) {},
				OnTick: func() {
					// Guard with p.l: curStreamer/curFormat are replaced (and
					// curStreamer set to nil) under p.l by reset() during song
					// switches. Reading them lock-free from the timer goroutine
					// races with that and can panic on a nil interface.
					p.l.Lock()
					defer p.l.Unlock()
					if p.curStreamer == nil || p.curFormat.SampleRate == 0 {
						return
					}
					select {
					case p.timeChan <- time.Duration(p.curStreamer.Position()) * time.Second / time.Duration(p.curFormat.SampleRate):
					default:
					}
				},
			})
			p.resumeNoLock()
			prevSongId = p.curMusic.Id
			started = true

		nextLoop:
			p.l.Unlock()
			// speaker.* 必须在不持有 p.l 时调用：拉取路径持 speaker 锁后再取
			// p.l，此处反向取锁会与拉取互等造成死锁。失败路径同样要 Clear，
			// 否则 speaker 仍拉取已被关闭的旧链，streamer 会空指针。
			speaker.Clear()
			if started {
				speaker.Play(p.volume)
			}
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
	p.l.Lock()
	defer p.l.Unlock()
	return p.curMusic
}

func (p *beepPlayer) setState(state types.State) {
	p.state.Store(uint32(state))
	select {
	case p.stateChan <- state:
	case <-time.After(time.Second * 2):
	}
}

// State 当前状态
func (p *beepPlayer) State() types.State {
	return types.State(p.state.Load())
}

// StateChan 状态发生变更
func (p *beepPlayer) StateChan() <-chan types.State {
	return p.stateChan
}

func (p *beepPlayer) PassedTime() time.Duration {
	p.l.Lock()
	defer p.l.Unlock()
	if p.curStreamer == nil {
		return 0
	}
	return time.Duration(p.curStreamer.Position()) * time.Second / time.Duration(p.curFormat.SampleRate)
}

func (p *beepPlayer) PlayedTime() time.Duration {
	p.l.Lock()
	defer p.l.Unlock()
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
	if duration < 0 {
		return
	}
	// 锁序与拉取路径一致（speaker mu → p.l）：切歌路径（reset 置 nil、关闭
	// 旧 streamer）只在 p.l 下进行，此处必须持 p.l 全程校验，否则切歌瞬间
	// 跳转会对已关闭/已置 nil 的 streamer 操作（空指针或 use-after-close）。
	speaker.Lock()
	defer speaker.Unlock()
	p.l.Lock()
	defer p.l.Unlock()

	// FIXME: 暂时仅对MP3格式提供跳转功能
	// FLAC格式(其他未测)跳转会占用大量CPU资源，比特率越高占用越高
	// 导致Seek方法卡住20-40秒的时间，之后方可随意跳转
	// minimp3未实现Seek
	if !p.cacheDownloaded || p.curStreamer == nil || p.curMusic.Type != Mp3 || configs.AppConfig.Player.Beep.Mp3Decoder == types.BeepMiniMp3Decoder {
		return
	}
	if types.State(p.state.Load()) == types.Playing || types.State(p.state.Load()) == types.Paused {
		newPos := p.curFormat.SampleRate.N(duration)

		if newPos < 0 {
			newPos = 0
		}
		if newPos >= p.curStreamer.Len() {
			newPos = p.curStreamer.Len() - 1
		}
		err := p.curStreamer.Seek(newPos)
		if err != nil {
			slog.Error("seek error", slogx.Error(err))
		}
		if p.timer != nil {
			p.timer.SetPassed(duration)
		}
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
	if types.State(p.state.Load()) != types.Playing {
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
	if types.State(p.state.Load()) == types.Playing {
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
	if types.State(p.state.Load()) == types.Stopped {
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

	if p.timer != nil {
		p.timer.Stop()
	}
	if p.close != nil {
		close(p.close)
		p.close = nil
	}
	if p.spectrum != nil {
		p.spectrum.Close()
	}
	p.l.Unlock()

	// speaker.* 必须在 p.l 外调用（锁序纪律：拉取路径持 speaker 锁后再取 p.l）
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
	p.spectrumConsumer = nil
	// 注意：不在此处调用 speaker.Clear()——speaker.* 必须在 p.l 外调用
	// （锁序纪律见 listen 的 nextLoop 处），调用方负责。
}

func (p *beepPlayer) streamer(samples [][2]float64) (n int, ok bool) {
	p.l.Lock()

	pos := p.curStreamer.Position()
	n, ok = p.curStreamer.Stream(samples)

	// Spectrum: feed PCM samples to analyzer。
	// 复用预分配缓冲（p.spectrumSamplesL/R）：每次拉取（约 10-50ms 一次、
	// 持锁期间）make 两个切片是纯 GC 压力。缓冲仅本函数使用，且本函数
	// 全程持 p.l 与 speaker 锁，无并发访问。
	if p.spectrumConsumer != nil && n > 0 {
		if cap(p.spectrumSamplesL) < n {
			p.spectrumSamplesL = make([]float32, n)
			p.spectrumSamplesR = make([]float32, n)
		}
		samplesL := p.spectrumSamplesL[:n]
		samplesR := p.spectrumSamplesR[:n]
		for i := 0; i < n; i++ {
			samplesL[i] = float32(samples[i][0])
			samplesR[i] = float32(samples[i][1])
		}
		p.spectrumConsumer(float64(sampleRate), samplesL, samplesR)
	}

	err := p.curStreamer.Err()
	if err == nil && (ok || p.cacheDownloaded) {
		p.l.Unlock()
		return
	}
	p.pausedNoLock()

	// 重试期间不持 p.l：拉取路径已持有 speaker 锁，若再持 p.l 睡眠，会同时
	// 冻结音频管线与 UI（PassedTime/CurMusic 均需 p.l），最长 20 秒。
	myStreamer := p.curStreamer
	isFlac := p.curMusic.Type == Flac
	p.l.Unlock()

	retry := 4
	for !ok && retry > 0 {
		if isFlac {
			if err = myStreamer.Seek(pos); err != nil {
				return
			}
		}
		errorx.ResetError(myStreamer)

		select {
		case <-time.After(time.Second * 5):
			p.l.Lock()
			// 等待期间可能已切歌：旧 streamer 已被 reset 关闭，直接放弃重试
			if p.curStreamer != myStreamer {
				p.l.Unlock()
				return
			}
			n, ok = myStreamer.Stream(samples)
			err = myStreamer.Err()
			p.l.Unlock()
		case <-p.close:
			return
		}
		retry--
	}
	p.l.Lock()
	p.resumeNoLock()
	p.l.Unlock()
	return
}

func (p *beepPlayer) resampleStreamer(old beep.SampleRate) beep.Streamer {
	if old == sampleRate {
		return beep.StreamerFunc(p.streamer)
	}
	return beep.Resample(resampleQuiality, old, sampleRate, beep.StreamerFunc(p.streamer))
}

func (p *beepPlayer) Spectrum() SpectrumFrame {
	if p.spectrum == nil {
		return SpectrumFrame{}
	}
	return p.spectrum.Spectrum()
}

func (p *beepPlayer) RawSamples() RawSampleFrame {
	if p.spectrum == nil {
		return RawSampleFrame{}
	}
	return p.spectrum.RawSamples()
}
