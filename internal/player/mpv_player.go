package player

import (
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/types"
	"github.com/go-musicfox/go-musicfox/utils/slogx"
)

var (
	tmpdir = "/tmp"
)

// mpvPlayer 实现基于MPV的播放器
type mpvPlayer struct {
	binPath string // MPV可执行文件路径

	cmd   *exec.Cmd
	mutex sync.Mutex

	curMusic URLMusic

	volume        int
	state         types.State
	timeChan      chan time.Duration
	stateChan     chan types.State
	close         chan struct{}
	cachedTimePos time.Duration // 缓存的播放位置，避免频繁查询
	lastSyncTime  time.Time     // 上次同步时间
	ticker        *time.Ticker  // 定期同步播放进度
	tickerDone    chan bool     // ticker 停止信号
	playStartTime time.Time     // 播放开始时间
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

	p := &mpvPlayer{
		binPath:   binPath, // 保存自定义路径
		volume:    50,      // 默认音量
		state:     types.Stopped,
		timeChan:  make(chan time.Duration, 1),
		stateChan: make(chan types.State, 10),
		close:     make(chan struct{}),
	}

	go p.listenMpvEvent()

	return p
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

// 启动MPV进程
func (p *mpvPlayer) startMpv(url, mediaTitle string) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if IsTermux() {
		tmpdir = "/data/data/com.termux/files/usr/tmp"
	}
	// 如果已有进程在运行，先停止
	if p.cmd != nil && p.cmd.Process != nil {
		_ = p.cmd.Process.Kill()
		p.cmd = nil
	}

	// 准备MPV命令
	args := []string{
		"--no-video",    // 无视频模式
		"--no-terminal", // 不使用终端
		"--input-ipc-server=" + tmpdir + "/mpvsocket", // IPC套接字，用于控制
		"--idle",                       // 空闲模式
		"--cache=yes",                  // 启用缓存
		"--demuxer-max-bytes=120MiB",   // 增大缓存容量
		"--demuxer-readahead-secs=120", // 增加预读时间
		"--log-file=" + tmpdir + "/mpvipc.log",
		"--audio-device=auto",                // 自动选择音频设备
		fmt.Sprintf("--volume=%d", p.volume), // 设置音量
	}

	if mt := strings.TrimSpace(mediaTitle); mt != "" {
		args = append(args, "--force-media-title="+mt)
	}

	if url != "" {
		args = append(args, url)
	}

	p.cmd = exec.Command(p.binPath, args...)

	// 启动进程
	if err := p.cmd.Start(); err != nil {
		return fmt.Errorf("启动MPV失败: %v", err)
	}

	return nil
}

// 监听mpv事件
func (p *mpvPlayer) listenMpvEvent() {
	for {
		conn, err := net.DialUnix("unix", nil, &net.UnixAddr{
			Name: tmpdir + "/mpvsocket",
			Net:  "unix",
		})
		if err != nil {
			time.Sleep(time.Second)
			continue
		}
		defer conn.Close()

		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if err != nil {
				break
			}
			msg := string(buf[:n])
			if strings.Contains(msg, `"event":"end-file"`) {
				// 歌曲播放结束
				p.setState(types.Stopped)
			}
		}
		time.Sleep(time.Second)
	}
}

// 向MPV发送命令
func (p *mpvPlayer) sendCommand(cmd string) error {
	conn, err := net.DialUnix("unix", nil, &net.UnixAddr{
		Name: tmpdir + "/mpvsocket",
		Net:  "unix",
	})
	if err != nil {
		return fmt.Errorf("连接MPV失败: %v", err)
	}
	defer conn.Close()

	_ = conn.SetWriteDeadline(time.Now().Add(2 * time.Second))
	_, err = conn.Write([]byte(cmd + "\n"))
	if err != nil {
		return fmt.Errorf("发送命令失败: %v", err)
	}
	return nil
}

// getProperty 从MPV获取属性值
func (p *mpvPlayer) getProperty(property string) (string, error) {
	conn, err := net.DialUnix("unix", nil, &net.UnixAddr{
		Name: tmpdir + "/mpvsocket",
		Net:  "unix",
	})
	if err != nil {
		return "", fmt.Errorf("连接MPV失败: %v", err)
	}
	defer conn.Close()

	// 发送获取属性命令
	cmd := fmt.Sprintf(`{ "command": ["get_property", "%s"] }`+"\n", property)
	_ = conn.SetWriteDeadline(time.Now().Add(time.Second))
	if _, err := conn.Write([]byte(cmd)); err != nil {
		return "", fmt.Errorf("发送命令失败: %v", err)
	}

	// 读取响应
	_ = conn.SetReadDeadline(time.Now().Add(time.Second))
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %v", err)
	}

	return string(buf[:n]), nil
}

// getMpvTimePos 从MPV获取当前播放位置（秒）
func (p *mpvPlayer) getMpvTimePos() (time.Duration, error) {
	resp, err := p.getProperty("time-pos")
	if err != nil {
		return 0, err
	}
	var seconds float64
	if _, err := fmt.Sscanf(resp, `{"data":%f`, &seconds); err != nil {
		if strings.Contains(resp, "error") {
			return 0, fmt.Errorf("mpv返回错误: %s", resp)
		}
		return 0, err
	}

	return time.Duration(seconds * float64(time.Second)), nil
}

// Play 播放指定音乐
func (p *mpvPlayer) Play(music URLMusic) {
	// 重置播放状态（保护 curMusic、ticker 和 state）
	p.mutex.Lock()
	p.curMusic = music
	p.stopTicker()
	p.lastSyncTime = time.Now()
	p.playStartTime = time.Now()
	p.cachedTimePos = 0
	p.startTicker()
	p.state = types.Playing
	p.mutex.Unlock()

	// 启动MPV播放音乐（在解锁后，避免长时间持有锁）
	if err := p.startMpv(music.URL, buildMpvMediaTitle(music)); err != nil {
		slog.Error("MPV播放失败", slogx.Error(err))
		// 如果启动失败，需要恢复状态
		p.mutex.Lock()
		p.state = types.Stopped
		p.mutex.Unlock()
		return
	}

	// 通知状态变化
	select {
	case p.stateChan <- types.Playing:
	case <-time.After(time.Second * 2):
	}
}

// startTicker 启动定期同步 ticker
func (p *mpvPlayer) startTicker() {
	p.ticker = time.NewTicker(configs.AppConfig.Main.FrameRate.Interval())
	p.tickerDone = make(chan bool)

	go func() {
		for {
			select {
			case <-p.ticker.C:
				// 每秒从 mpv 同步一次实际播放位置
				if time.Since(p.lastSyncTime) >= time.Second {
					if timePos, err := p.getMpvTimePos(); err == nil {
						p.mutex.Lock()
						p.cachedTimePos = timePos
						p.lastSyncTime = time.Now()
						p.mutex.Unlock()
					}
				}
				// 发送当前播放时间
				select {
				case p.timeChan <- p.PassedTime():
				default:
				}
			case <-p.tickerDone:
				return
			}
		}
	}()
}

// stopTicker 停止 ticker
func (p *mpvPlayer) stopTicker() {
	if p.ticker != nil {
		p.ticker.Stop()
		if p.tickerDone != nil {
			close(p.tickerDone)
			p.tickerDone = nil
		}
		p.ticker = nil
	}
}

// CurMusic 获取当前播放的音乐
func (p *mpvPlayer) CurMusic() URLMusic {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	return p.curMusic
}

// Pause 暂停播放
func (p *mpvPlayer) Pause() {
	p.mutex.Lock()
	if p.state != types.Playing {
		p.mutex.Unlock()
		return
	}
	p.mutex.Unlock()

	_ = p.sendCommand("{ \"command\": [\"set_property\", \"pause\", true] }")
	if timePos, err := p.getMpvTimePos(); err == nil {
		p.mutex.Lock()
		p.cachedTimePos = timePos
		p.mutex.Unlock()
	}

	p.mutex.Lock()
	p.state = types.Paused
	p.mutex.Unlock()
	select {
	case p.stateChan <- types.Paused:
	case <-time.After(time.Second * 2):
	}
}

// Resume 恢复播放
func (p *mpvPlayer) Resume() {
	p.mutex.Lock()
	if p.state != types.Paused && p.state != types.Stopped {
		p.mutex.Unlock()
		return
	}
	p.mutex.Unlock()

	_ = p.sendCommand("{ \"command\": [\"set_property\", \"pause\", false] }")
	p.mutex.Lock()
	p.state = types.Playing
	p.mutex.Unlock()
	select {
	case p.stateChan <- types.Playing:
	case <-time.After(time.Second * 2):
	}
}

// Stop 停止播放
func (p *mpvPlayer) Stop() {
	_ = p.sendCommand("{ \"command\": [\"stop\"] }")
	p.mutex.Lock()
	p.stopTicker()
	p.state = types.Stopped
	p.mutex.Unlock()
	select {
	case p.stateChan <- types.Stopped:
	case <-time.After(time.Second * 2):
	}
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
	p.mutex.Lock()
	if p.state != types.Playing && p.state != types.Paused {
		p.mutex.Unlock()
		return
	}
	p.mutex.Unlock()

	cmd := fmt.Sprintf(`{ "command": ["set_property", "time-pos", %f] }`, duration.Seconds())
	if err := p.sendCommand(cmd); err != nil {
		slog.Error("跳转命令发送失败", slogx.Error(err))
		return
	}

	// 更新缓存位置
	p.mutex.Lock()
	p.cachedTimePos = duration
	p.lastSyncTime = time.Now()
	p.mutex.Unlock()
}

// PassedTime 获取已播放时间（基于 MPV 的实际播放时间）
func (p *mpvPlayer) PassedTime() time.Duration {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// 如果不在播放状态，返回缓存位置
	if p.state != types.Playing {
		return p.cachedTimePos
	}

	// 如果最近2秒内同步过，使用缓存值加上估算的增量
	if time.Since(p.lastSyncTime) < 2*time.Second {
		elapsed := time.Since(p.lastSyncTime)
		return p.cachedTimePos + elapsed
	}

	// 否则返回缓存值（可能已经过时，但下一次 tick 会更新）
	return p.cachedTimePos
}

// PlayedTime 获取从播放开始到现在的时间
func (p *mpvPlayer) PlayedTime() time.Duration {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.playStartTime.IsZero() {
		return 0
	}
	return time.Since(p.playStartTime)
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
	p.mutex.Lock()
	p.state = state
	p.mutex.Unlock()
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
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if volume > 100 {
		volume = 100
	}
	if volume < 0 {
		volume = 0
	}

	p.volume = volume
	_ = p.sendCommand(fmt.Sprintf("{ \"command\": [\"set_property\", \"volume\", %d] }", volume))
}

// UpVolume 增加音量
func (p *mpvPlayer) UpVolume() {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.volume+5 >= 100 {
		p.volume = 100
	} else {
		p.volume += 5
	}

	_ = p.sendCommand(fmt.Sprintf("{ \"command\": [\"set_property\", \"volume\", %d] }", p.volume))
}

// DownVolume 降低音量
func (p *mpvPlayer) DownVolume() {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.volume-5 <= 0 {
		p.volume = 0
	} else {
		p.volume -= 5
	}

	_ = p.sendCommand(fmt.Sprintf("{ \"command\": [\"set_property\", \"volume\", %d] }", p.volume))
}

// Close 关闭播放器
func (p *mpvPlayer) Close() {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.stopTicker()

	if p.cmd != nil && p.cmd.Process != nil {
		_ = p.cmd.Process.Kill()
		p.cmd = nil
	}

	if p.close != nil {
		close(p.close)
		p.close = nil
	}
}

// IsTermux 检查当前是否在 Termux 环境中运行
func IsTermux() bool {
	// 方法1：检查特定环境变量
	if path, ok := os.LookupEnv("PREFIX"); ok {
		if strings.Contains(path, "com.termux") {
			return true
		}
	}

	// 方法2：检查特定目录是否存在
	termuxPaths := []string{
		"/data/data/com.termux/files/usr",
		"/data/data/com.termux/files/home",
	}

	for _, path := range termuxPaths {
		if _, err := os.Stat(path); err == nil {
			return true
		}
	}

	// 方法3：检查可执行文件路径
	if exe, err := os.Executable(); err == nil {
		if strings.Contains(filepath.Dir(exe), "com.termux") {
			return true
		}
	}

	return false
}
