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

	"github.com/go-musicfox/go-musicfox/internal/types"
	"github.com/go-musicfox/go-musicfox/utils/slogx"
	"github.com/go-musicfox/go-musicfox/utils/timex"
)

var (
	tmpdir = "/tmp"
)

// mpvPlayer 实现基于MPV的播放器
type mpvPlayer struct {
	binPath string // MPV可执行文件路径

	cmd    *exec.Cmd
	mutex  sync.Mutex

	curMusic URLMusic
	timer    *timex.Timer

	volume    int
	state     types.State
	timeChan  chan time.Duration
	stateChan chan types.State
	close     chan struct{}
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
		volume:    50, // 默认音量
		state:     types.Stopped,
		timeChan:  make(chan time.Duration),
		stateChan: make(chan types.State),
		close:     make(chan struct{}),
	}

	go p.listenMpvEvent()

	return p
}

// 启动MPV进程
func (p *mpvPlayer) startMpv(url string) error {
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

// Play 播放指定音乐
func (p *mpvPlayer) Play(music URLMusic) {
	p.curMusic = music

	// 停止当前计时器
	if p.timer != nil {
		p.timer.Stop()
	}

	// 启动MPV播放音乐
	if err := p.startMpv(music.URL); err != nil {
		slog.Error("MPV播放失败", slogx.Error(err))
		return
	}

	// 创建计时器
	p.timer = timex.NewTimer(timex.Options{
		Duration:       8760 * time.Hour, // 长时间，实际由歌曲控制
		TickerInternal: 200 * time.Millisecond,
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

	// 启动计时器
	go p.timer.Run()

	// 设置状态为播放中
	p.setState(types.Playing)
}

// CurMusic 获取当前播放的音乐
func (p *mpvPlayer) CurMusic() URLMusic {
	return p.curMusic
}

// Pause 暂停播放
func (p *mpvPlayer) Pause() {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.state == types.Playing {
		_ = p.sendCommand("{ \"command\": [\"set_property\", \"pause\", true] }")
		if p.timer != nil {
			p.timer.Pause()
		}
		p.setState(types.Paused)
	}
}

// Resume 恢复播放
func (p *mpvPlayer) Resume() {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.state == types.Paused || p.state == types.Stopped {
		_ = p.sendCommand("{ \"command\": [\"set_property\", \"pause\", false] }")
		if p.timer != nil {
			go p.timer.Run()
		}
		p.setState(types.Playing)
	}
}

// Stop 停止播放
func (p *mpvPlayer) Stop() {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	_ = p.sendCommand("{ \"command\": [\"stop\"] }")
	if p.timer != nil {
		p.timer.Stop()
	}
	p.setState(types.Stopped)
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
	defer p.mutex.Unlock()

	if p.state != types.Playing && p.state != types.Paused {
		return
	}

	cmd := fmt.Sprintf(`{ "command": ["set_property", "time-pos", %f] }`, duration.Seconds())
	if err := p.sendCommand(cmd); err != nil {
		slog.Error("跳转命令发送失败", slogx.Error(err))
		return
	}

	if p.timer != nil {
		p.timer.SetPassed(duration)
	}
}

// PassedTime 获取已播放时间
func (p *mpvPlayer) PassedTime() time.Duration {
	if p.timer == nil {
		return 0
	}
	return p.timer.Passed()
}

// PlayedTime 获取计时器实际计时时间
func (p *mpvPlayer) PlayedTime() time.Duration {
	if p.timer == nil {
		return 0
	}
	return p.timer.ActualRuntime()
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
	p.state = state
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

	if p.timer != nil {
		p.timer.Stop()
	}

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
