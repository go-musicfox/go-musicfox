package player

import (
	"fmt"
	"log/slog"
	"net"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/types"
	"github.com/go-musicfox/go-musicfox/utils/errorx"
	"github.com/go-musicfox/go-musicfox/utils/slogx"
	"github.com/go-musicfox/go-musicfox/utils/timex"
)

// mpvPlayer 实现基于MPV的播放器（守护进程模式）
//
// 架构说明（类似 mpdPlayer）：
//   - mpv 以 --idle 模式在后台运行一次（守护进程）
//   - 切歌通过 IPC 发送 loadfile 命令，不重启进程
//   - timex.Timer 追踪播放位置（不轮询 mpv）
//   - 事件连接 watch() 独立读取 IPC 事件
//   - 命令连接 getIPCConn() 缓存复用
type mpvPlayer struct {
	binPath string
	cmd     *exec.Cmd // mpv 守护进程

	// IPC 命令连接（与事件连接分离）
	ipcConn  net.Conn
	ipcMutex sync.Mutex

	curMusic URLMusic

	volume    int
	state     types.State
	timeChan  chan time.Duration
	stateChan chan types.State
	closeCh   chan struct{}

	timer     *timex.Timer  // 播放位置追踪（参考 mpdPlayer）
	musicChan chan URLMusic // 异步切歌信号（参考 mpdPlayer）

	watchOnce sync.Once // 确保 watch 只启动一次

	// 切歌watchdog（加固）：等待 file-loaded 事件确认切歌成功
	songLoadedCh chan struct{}

	loadSeq      uint64 // 切歌代际计数，作废旧watchdog
	loadSeqLock  sync.Mutex
	loadWDCancel chan struct{} // 取消上一个切歌watchdog
	loadWDMutex  sync.Mutex    // 保护 loadWDCancel
}

// MpvConfig MPV播放器配置
type MpvConfig struct {
	BinPath string // MPV可执行文件路径
}

// NewMpvPlayer 创建新的MPV播放器实例
func NewMpvPlayer(conf *MpvConfig) *mpvPlayer {
	binPath := "mpv"
	if conf != nil && conf.BinPath != "" {
		binPath = conf.BinPath
	}

	// 检查MPV是否可用
	cmd := exec.Command(binPath, "--version")
	if err := cmd.Run(); err != nil {
		panic(fmt.Sprintf("MPV不可用: %v", err))
	}
	slog.Info("mpv daemon: MPV版本检测通过", slog.String("bin", binPath))

	p := &mpvPlayer{
		binPath:      binPath,
		volume:       50,
		state:        types.Stopped,
		timeChan:     make(chan time.Duration, 1),
		stateChan:    make(chan types.State, 10),
		musicChan:    make(chan URLMusic, 1),
		closeCh:      make(chan struct{}),
		songLoadedCh: make(chan struct{}, 1),
	}

	if err := p.startDaemon(); err != nil {
		slog.Error("mpv daemon: 启动守护进程失败", slogx.Error(err))
		panic(fmt.Sprintf("MPV守护进程启动失败: %v", err))
	}

	slog.Info("mpv daemon: 守护进程启动完成", slog.String("ipc", ipcServerPath()))
	return p
}

// startDaemon 启动 mpv 守护进程（仅调用一次）
func (p *mpvPlayer) startDaemon() error {
	args := []string{
		"--no-video",                            // 无视频模式
		"--no-terminal",                         // 不使用终端
		"--input-ipc-server=" + ipcServerPath(), // IPC通道
		"--idle",                                // 空闲驻留
		"--cache=yes",                           // 启用缓存
		"--demuxer-max-bytes=120MiB",            // 增大缓存容量
		"--demuxer-readahead-secs=120",          // 增加预读时间
		"--log-file=" + ipcLogPath(),
		"--audio-device=auto", // 自动选择音频设备
		"--input-media-keys=no",
		fmt.Sprintf("--volume=%d", p.volume), // 设置音量
	}

	p.cmd = exec.Command(p.binPath, args...)
	slog.Info("mpv daemon: 启动mpv进程",
		slog.String("bin", p.binPath),
		slog.Any("args", args),
		slog.String("ipc", ipcServerPath()),
		slog.String("log", ipcLogPath()),
	)

	if err := p.cmd.Start(); err != nil {
		return fmt.Errorf("启动MPV失败: %v", err)
	}
	slog.Info("mpv daemon: mpv进程已启动", slog.Int("pid", p.cmd.Process.Pid))

	// 等待 IPC 就绪
	slog.Info("mpv daemon: 等待IPC就绪...")
	if err := p.waitForIPC(); err != nil {
		_ = p.cmd.Process.Kill()
		_, _ = p.cmd.Process.Wait()
		return err
	}
	slog.Info("mpv daemon: IPC就绪")

	// 启动事件监听
	errorx.WaitGoStart(p.listen)
	p.watchOnce.Do(func() {
		go p.watch()
	})
	slog.Info("mpv daemon: listen/watch goroutines已启动")

	return nil
}

// waitForIPC 等待 mpv IPC 端点就绪，超时 10 秒
func (p *mpvPlayer) waitForIPC() error {
	deadline := time.Now().Add(10 * time.Second)
	attempt := 0
	for time.Now().Before(deadline) {
		attempt++
		conn, err := p.dialIPC()
		if err == nil {
			conn.Close()
			slog.Info("mpv daemon: IPC就绪", slog.Int("attempts", attempt))
			return nil
		}
		if attempt == 1 || attempt%10 == 0 {
			slog.Info("mpv daemon: 等待IPC...", slog.Int("attempt", attempt), slogx.Error(err))
		}
		select {
		case <-time.After(100 * time.Millisecond):
		case <-p.closeCh:
			return fmt.Errorf("播放器已关闭")
		}
	}
	return fmt.Errorf("等待MPV IPC就绪超时（尝试%d次，10s），请检查 mpv 是否正确安装", attempt)
}

// listen 监听切歌信号（异步处理 musicChan，参考 mpdPlayer.listen）
func (p *mpvPlayer) listen() {
	slog.Info("mpv listen: goroutine启动")
	for {
		select {
		case <-p.closeCh:
			slog.Info("mpv listen: 收到关闭信号，退出")
			return
		case music := <-p.musicChan:
			slog.Info("mpv listen: 收到新歌曲",
				slog.String("name", music.Name),
				slog.String("url", music.URL),
				slog.Int64("songId", music.Id),
			)
			p.handleNewSong(music)
		}
	}
}

// handleNewSong 处理新歌曲加载
func (p *mpvPlayer) handleNewSong(music URLMusic) {
	slog.Info("mpv handleNewSong: 开始处理新歌曲",
		slog.String("name", music.Name),
		slog.String("url", music.URL),
	)

	p.curMusic = music

	// 停止旧 timer
	if p.timer != nil {
		slog.Debug("mpv handleNewSong: 停止旧timer")
		p.timer.Stop()
		p.timer = nil
	}

	// 发送 loadfile 命令加载新歌曲
	cmd := fmt.Sprintf(`{ "command": ["loadfile", %s, "replace"] }`, jsonString(music.URL))
	slog.Info("mpv handleNewSong: 发送loadfile命令", slog.String("cmd", cmd))
	if err := p.sendCommand(cmd); err != nil {
		slog.Error("mpv handleNewSong: loadfile失败", slogx.Error(err))
		return
	}
	slog.Info("mpv handleNewSong: loadfile发送成功")

	// 加固：切歌watchdog，超时未收到 file-loaded 则重建连接并重发 loadfile
	p.armSongLoadedWatchdog(cmd)

	// 设置媒体标题
	if title := buildMpvMediaTitle(music); title != "" {
		titleCmd := fmt.Sprintf(`{ "command": ["set_property", "media-title", %s] }`, jsonString(title))
		slog.Debug("mpv handleNewSong: 设置media-title", slog.String("title", title))
		if err := p.sendCommand(titleCmd); err != nil {
			slog.Warn("mpv handleNewSong: 设置media-title失败", slogx.Error(err))
		}
	}
	slog.Info("mpv handleNewSong: media-title设置完成")

	// 创建 timer 追踪播放位置（参考 mpdPlayer）
	// 使用超长 Duration，歌曲结束依赖 mpv 的 end-file 事件
	interval := configs.AppConfig.Main.FrameRate.Interval()
	slog.Info("mpv handleNewSong: 创建timer",
		slog.Duration("interval", interval),
		slog.String("duration", "8760h"),
	)
	p.timer = timex.NewTimer(timex.Options{
		Duration:       8760 * time.Hour,
		TickerInternal: interval,
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
	p.Resume()

	// go p.timer.Run()
	// slog.Info("mpv handleNewSong: timer已启动")

	// p.setState(types.Playing)
	// slog.Info("mpv handleNewSong: 完成，状态已设为Playing")
}

// songLoadTimeout 切歌watchdog超时：loadfile 发出后等待 file-loaded 的最长时限
const songLoadTimeout = 10 * time.Second

// armSongLoadedWatchdog 为最近一次 loadfile 启动超时watchdog（加固）
//
// 若在 songLoadTimeout 内未收到 file-loaded 事件，说明命令连接可能已僵死
// （旧版本只写不读导致 mpv IPC 线程卡死），此时重建命令连接并重发 loadfile，
// 兜住历史版本已僵死场景下 UI 卡在"新歌无声/仍播上一首"的问题。
func (p *mpvPlayer) armSongLoadedWatchdog(loadCmd string) {
	// 取消上一个watchdog，确保快速连续切歌时同时只有一个watchdog在等待
	p.loadWDMutex.Lock()
	if p.loadWDCancel != nil {
		close(p.loadWDCancel)
	}
	cancel := make(chan struct{})
	p.loadWDCancel = cancel
	p.loadWDMutex.Unlock()

	// 递增代际：即使旧watchdog已越过select进入超时分支，也能阻止其干预新歌
	p.loadSeqLock.Lock()
	p.loadSeq++
	seq := p.loadSeq
	p.loadSeqLock.Unlock()

	// 排空可能残留的上一次 file-loaded 事件
	select {
	case <-p.songLoadedCh:
	default:
	}

	go func(cancel <-chan struct{}, seq uint64) {
		select {
		case <-p.songLoadedCh:
			slog.Debug("mpv watchdog: 已收到file-loaded，切歌正常", slog.Uint64("seq", seq))
		case <-cancel:
			return
		case <-p.closeCh:
			return
		case <-time.After(songLoadTimeout):
			// 若已关闭播放器则直接放弃（防止退出过程中的误重发）
			select {
			case <-p.closeCh:
				return
			default:
			}

			p.loadSeqLock.Lock()
			current := p.loadSeq
			p.loadSeqLock.Unlock()
			if seq != current {
				return // 期间已切到新歌，由新watchdog负责
			}

			slog.Error("mpv watchdog: 切歌超时未收到file-loaded，重建命令连接并重发loadfile",
				slog.Uint64("seq", seq),
				slog.Duration("timeout", songLoadTimeout),
			)
			// 关闭旧连接（排空goroutine随之退出），强制重连后再重发
			p.closeIPCConn()
			if err := p.sendCommand(loadCmd); err != nil {
				slog.Error("mpv watchdog: 重发loadfile失败", slogx.Error(err))
			}
		}
	}(cancel, seq)
}

// watch 监听 mpv IPC 事件（end-file 检测）
// 使用独立的事件连接（与命令连接分离）
func (p *mpvPlayer) watch() {
	slog.Info("mpv watch: goroutine启动")
	for {
		select {
		case <-p.closeCh:
			slog.Info("mpv watch: 收到关闭信号，退出")
			return
		default:
		}

		slog.Debug("mpv watch: 准备连接IPC")
		conn, err := p.dialIPC()
		if err != nil {
			slog.Error("mpv watch: dialIPC失败", slogx.Error(err))
			select {
			case <-p.closeCh:
				return
			case <-time.After(time.Second):
			}
			continue
		}
		slog.Info("mpv watch: IPC连接成功（事件连接）")

		buf := make([]byte, 4096)
		readCount := 0
		for {
			n, err := conn.Read(buf)
			if err != nil {
				slog.Warn("mpv watch: 事件连接读取断开",
					slog.Int("readCount", readCount),
					slogx.Error(err),
				)
				break
			}
			readCount++
			msg := string(buf[:n])
			slog.Info("mpv watch: 收到IPC事件",
				slog.Int("readCount", readCount),
				slog.String("raw", msg),
			)

			// 检测歌曲结束
			// end-file 有多种 reason：
			//   eof       — 自然播放结束（需要触发下一首）
			//   stop      — 手动停止（Stop() 已处理）
			//   new-file  — loadfile replace 替换文件（忽略，切歌中）
			//   error     — 播放出错
			// 只对 eof 触发 Stopped → 自动下一首
			if strings.Contains(msg, `"event":"end-file"`) {
				isEOF := strings.Contains(msg, `"reason":"eof"`)
				isError := strings.Contains(msg, `"reason":"error"`)
				slog.Info("mpv watch: 检测到end-file事件",
					slog.Bool("isEOF", isEOF),
					slog.Bool("isError", isError),
				)

				// 只有自然结束(eof)或出错才触发状态变更
				if isEOF || isError {
					slog.Info("mpv watch: 触发歌曲结束", slog.Bool("isEOF", isEOF))
					if p.timer != nil {
						slog.Debug("mpv watch: 停止timer")
						p.timer.Stop()
						p.timer = nil
					}
					p.setState(types.Stopped)
					slog.Info("mpv watch: 状态已设为Stopped")
				} else {
					slog.Info("mpv watch: end-file忽略（非自然结束，可能是切歌中）")
				}
			}

			// 切歌watchdog信号（加固）：file-loaded 表示新歌已成功加载
			if strings.Contains(msg, `"event":"file-loaded"`) {
				select {
				case p.songLoadedCh <- struct{}{}:
				default:
				}
			}
		}
		conn.Close()
		slog.Debug("mpv watch: 事件连接已关闭，1秒后重试")

		// 等待重试
		select {
		case <-p.closeCh:
			return
		case <-time.After(time.Second):
		}
	}
}

func buildMpvMediaTitle(music URLMusic) string {
	name := strings.TrimSpace(music.Name)
	if name == "" {
		return ""
	}

	var artists []string
	for _, a := range music.Artists {
		an := strings.TrimSpace(a.Name)
		if an != "" {
			artists = append(artists, an)
		}
	}
	if len(artists) == 0 {
		return sanitizeMpvTitle(name)
	}
	return sanitizeMpvTitle(name + " - " + strings.Join(artists, ", "))
}

func sanitizeMpvTitle(title string) string {
	title = strings.ReplaceAll(title, "\n", " ")
	title = strings.ReplaceAll(title, "\r", " ")
	return strings.TrimSpace(title)
}

// jsonString 将字符串转为 JSON 安全字符串（用于 IPC 命令参数）
func jsonString(s string) string {
	// 简单的 JSON 字符串转义
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\r", "\\r")
	s = strings.ReplaceAll(s, "\t", "\\t")
	return `"` + s + `"`
}

// ---------------------------------------------------------------------------
// IPC 连接管理
// ---------------------------------------------------------------------------

// getIPCConn 获取或创建持久 IPC 连接（用于发送命令）
func (p *mpvPlayer) getIPCConn() (net.Conn, error) {
	p.ipcMutex.Lock()
	defer p.ipcMutex.Unlock()

	if p.ipcConn != nil {
		slog.Debug("mpv ipc: 复用缓存连接")
		return p.ipcConn, nil
	}

	slog.Debug("mpv ipc: 创建新连接（命令连接）")
	conn, err := p.dialIPC()
	if err != nil {
		slog.Error("mpv ipc: dialIPC失败", slogx.Error(err))
		return nil, fmt.Errorf("连接MPV失败: %v", err)
	}
	p.ipcConn = conn
	// 方案A：启动排空goroutine持续读取命令连接上的事件与命令回包，
	// 防止mpv侧发送缓冲堆积满导致其IPC线程僵死（详见 drainIPCConn 注释）
	go p.drainIPCConn(conn)
	slog.Info("mpv ipc: 命令连接已建立（排空goroutine已启动）")
	return conn, nil
}

// closeIPCConn 关闭持久 IPC 连接
func (p *mpvPlayer) closeIPCConn() {
	p.ipcMutex.Lock()
	defer p.ipcMutex.Unlock()

	if p.ipcConn != nil {
		slog.Debug("mpv ipc: 关闭命令连接")
		p.ipcConn.Close()
		p.ipcConn = nil
	}
}

// drainIPCConn 持续读取并丢弃命令连接上的数据（mpv事件 + 命令回包）。
//
// 根因：mpv 的 JSON IPC（如 input/ipc-unix.c）每个客户端一个线程同时负责读写，
// 若 Go 侧只写不读，mpv 发送缓冲会随播放事件堆积至满；其 ipc_write_str 对 EAGAIN
// 采用自旋重试，导致该客户端线程卡死在写操作、从此不再读取命令，表现为播放十几首
// 歌后没声/切歌仍播上一首。排空goroutine保证 mpv 侧发送缓冲永不满。
//
// 读取失败（EOF/连接异常/被 Close）时，若该连接仍是当前缓存连接则将其置空，
// 使下一次 sendCommand 自动重连，天然获得连接失效检测与自愈。
func (p *mpvPlayer) drainIPCConn(conn net.Conn) {
	slog.Debug("mpv ipc: 排空goroutine启动")
	buf := make([]byte, 4096)
	for {
		if _, err := conn.Read(buf); err != nil {
			slog.Warn("mpv ipc: 命令连接读取结束，丢弃缓存连接", slogx.Error(err))
			p.closeIPCConnIfSame(conn)
			return
		}
	}
}

// closeIPCConnIfSame 仅当指定连接仍是当前缓存连接时关闭并置空（供排空goroutine使用）。
// 通过身份校验，避免旧连接的排空goroutine误清掉新建立的连接。
func (p *mpvPlayer) closeIPCConnIfSame(conn net.Conn) {
	p.ipcMutex.Lock()
	defer p.ipcMutex.Unlock()

	if p.ipcConn == conn {
		slog.Debug("mpv ipc: 关闭失效命令连接")
		p.ipcConn = nil
		_ = conn.Close()
	}
}

// sendCommand 发送 IPC 命令到 mpv
func (p *mpvPlayer) sendCommand(cmd string) error {
	conn, err := p.getIPCConn()
	if err != nil {
		return err
	}

	payload := cmd + "\n"
	slog.Debug("mpv sendCommand: 发送命令",
		slog.String("cmd", cmd),
		slog.Int("bytes", len(payload)),
	)

	_ = conn.SetWriteDeadline(time.Now().Add(2 * time.Second))
	_, err = conn.Write([]byte(payload))
	if err != nil {
		slog.Error("mpv sendCommand: 写入失败", slogx.Error(err))
		p.closeIPCConn()
		return fmt.Errorf("发送命令失败: %v", err)
	}
	slog.Debug("mpv sendCommand: 发送成功")
	return nil
}

// ---------------------------------------------------------------------------
// Player 接口实现
// ---------------------------------------------------------------------------

// Play 播放指定音乐（异步，通过 musicChan 发送信号）
func (p *mpvPlayer) Play(music URLMusic) {
	slog.Info("mpv Play: 收到播放请求",
		slog.String("name", music.Name),
		slog.String("url", music.URL),
		slog.Int64("songId", music.Id),
	)
	select {
	case p.musicChan <- music:
		slog.Debug("mpv Play: 已写入musicChan")
	default:
		slog.Debug("mpv Play: musicChan满，丢弃旧信号")
		select {
		case <-p.musicChan:
		default:
		}
		p.musicChan <- music
		slog.Debug("mpv Play: 丢弃后写入成功")
	}
}

// CurMusic 获取当前播放的音乐
func (p *mpvPlayer) CurMusic() URLMusic {
	return p.curMusic
}

// Pause 暂停播放
func (p *mpvPlayer) Pause() {
	slog.Info("mpv Pause: 请求暂停", slog.Any("state", p.state))
	if p.state != types.Playing {
		slog.Debug("mpv Pause: 当前状态非Playing，忽略")
		return
	}

	_ = p.sendCommand(`{ "command": ["set_property", "pause", true] }`)
	if p.timer != nil {
		slog.Debug("mpv Pause: 暂停timer")
		p.timer.Pause()
	}
	p.setState(types.Paused)
	slog.Info("mpv Pause: 完成")
}

// Resume 恢复播放
func (p *mpvPlayer) Resume() {
	slog.Info("mpv Resume: 请求恢复", slog.Any("state", p.state))
	if p.state != types.Paused && p.state != types.Stopped {
		slog.Debug("mpv Resume: 当前状态不可恢复，忽略")
		return
	}

	_ = p.sendCommand(`{ "command": ["set_property", "pause", false] }`)
	if p.timer != nil {
		slog.Debug("mpv Resume: 恢复timer")
		go p.timer.Run()
	}
	p.setState(types.Playing)
	slog.Info("mpv Resume: 完成")
}

// Stop 停止播放
func (p *mpvPlayer) Stop() {
	slog.Info("mpv Stop: 请求停止")
	_ = p.sendCommand(`{ "command": ["stop"] }`)
	if p.timer != nil {
		slog.Debug("mpv Stop: 停止timer")
		p.timer.Stop()
		p.timer = nil
	}
	p.setState(types.Stopped)
	slog.Info("mpv Stop: 完成")
}

// Toggle 切换播放/暂停状态
func (p *mpvPlayer) Toggle() {
	switch p.State() {
	case types.Paused, types.Stopped:
		p.Resume()
	case types.Playing:
		p.Pause()
	default:
		p.Resume()
	}
}

// Seek 跳转到指定时间
func (p *mpvPlayer) Seek(duration time.Duration) {
	slog.Info("mpv Seek: 跳转", slog.Duration("target", duration))
	cmd := fmt.Sprintf(`{ "command": ["set_property", "time-pos", %f] }`, duration.Seconds())
	if err := p.sendCommand(cmd); err != nil {
		slog.Error("mpv Seek: 跳转命令发送失败", slogx.Error(err))
		return
	}

	if p.timer != nil {
		slog.Debug("mpv Seek: 更新timer位置")
		p.timer.SetPassed(duration)
	}
	slog.Info("mpv Seek: 完成")
}

// PassedTime 获取已播放时间（基于 timer，不轮询 mpv）
func (p *mpvPlayer) PassedTime() time.Duration {
	pt := time.Duration(0)
	if p.timer != nil {
		pt = p.timer.Passed()
	}
	slog.Debug("mpv PassedTime", slog.Duration("passed", pt))
	return pt
}

// PlayedTime 获取从播放开始到现在的时间
func (p *mpvPlayer) PlayedTime() time.Duration {
	art := time.Duration(0)
	if p.timer != nil {
		art = p.timer.ActualRuntime()
	}
	slog.Debug("mpv PlayedTime", slog.Duration("actualRuntime", art))
	return art
}

// TimeChan 获取时间更新通道
func (p *mpvPlayer) TimeChan() <-chan time.Duration {
	return p.timeChan
}

// State 获取当前状态
func (p *mpvPlayer) State() types.State {
	return p.state
}

// StateChan 获取状态更新通道
func (p *mpvPlayer) StateChan() <-chan types.State {
	return p.stateChan
}

// setState 设置状态并通知
func (p *mpvPlayer) setState(state types.State) {
	slog.Info("mpv setState: 状态变更", slog.Any("state", state))
	p.state = state
	p.sendStateToChan(state)
}

func (p *mpvPlayer) sendStateToChan(state types.State) {
	select {
	case p.stateChan <- state:
	case <-time.After(time.Second * 2):
	}
}

// Volume 获取当前音量
func (p *mpvPlayer) Volume() int {
	return p.volume
}

// SetVolume 设置音量
func (p *mpvPlayer) SetVolume(volume int) {
	slog.Debug("mpv SetVolume", slog.Int("volume", volume))
	if volume > 100 {
		volume = 100
	}
	if volume < 0 {
		volume = 0
	}

	p.volume = volume
	_ = p.sendCommand(fmt.Sprintf(`{ "command": ["set_property", "volume", %d] }`, volume))
}

// UpVolume 增加音量
func (p *mpvPlayer) UpVolume() {
	slog.Debug("mpv UpVolume")
	if p.volume+5 >= 100 {
		p.volume = 100
	} else {
		p.volume += 5
	}
	_ = p.sendCommand(fmt.Sprintf(`{ "command": ["set_property", "volume", %d] }`, p.volume))
}

// DownVolume 降低音量
func (p *mpvPlayer) DownVolume() {
	slog.Debug("mpv DownVolume")
	if p.volume-5 <= 0 {
		p.volume = 0
	} else {
		p.volume -= 5
	}
	_ = p.sendCommand(fmt.Sprintf(`{ "command": ["set_property", "volume", %d] }`, p.volume))
}

// Close 关闭播放器，发送 quit 命令并等待 mpv 退出
func (p *mpvPlayer) Close() {
	slog.Info("mpv Close: 开始关闭播放器")

	// 停止 timer
	if p.timer != nil {
		slog.Debug("mpv Close: 停止timer")
		p.timer.Stop()
		p.timer = nil
	}

	// 关闭信号通道，通知 listen/watch 退出
	if p.closeCh != nil {
		slog.Debug("mpv Close: 关闭closeCh信号通道")
		close(p.closeCh)
		p.closeCh = nil
	}

	// 关闭 IPC 命令连接
	p.closeIPCConn()

	// 发送 quit 命令让 mpv 正常退出
	slog.Info("mpv Close: 发送quit命令")
	if err := p.sendCommand(`{ "command": ["quit"] }`); err != nil {
		slog.Warn("mpv Close: quit命令发送失败（可能已断开）", slogx.Error(err))
	}

	// 等待 mpv 退出
	if p.cmd != nil && p.cmd.Process != nil {
		slog.Info("mpv Close: 等待mpv进程退出", slog.Int("pid", p.cmd.Process.Pid))
		done := make(chan struct{})
		go func() {
			_, _ = p.cmd.Process.Wait()
			close(done)
		}()
		select {
		case <-done:
			slog.Info("mpv Close: mpv进程正常退出")
		case <-time.After(3 * time.Second):
			slog.Warn("mpv Close: 超时未退出，强制Kill")
			_ = p.cmd.Process.Kill()
		}
		p.cmd = nil
	}

	slog.Info("mpv Close: 关闭完成")
}
