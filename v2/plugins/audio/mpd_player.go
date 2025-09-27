//go:build linux || unix
// +build linux unix

package audio

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

// MPDPlayer MPD播放器后端实现
type MPDPlayer struct {
	*BasePlayer
	conn        net.Conn
	host        string
	port        string
	password    string
	mutex       sync.RWMutex
	config      map[string]interface{}
	initialized bool
	connected   bool
}

// NewMPDPlayer 创建MPD播放器
func NewMPDPlayer(config map[string]interface{}) *MPDPlayer {
	formats := []string{"mp3", "wav", "flac", "ogg", "m4a", "aac", "wma", "ape"}
	player := &MPDPlayer{
		BasePlayer: NewBasePlayerWithInfo("MPD Player", "1.0.0", formats),
		config:     config,
		host:       "localhost",
		port:       "6600",
	}

	// 从配置中读取连接参数
	if config != nil {
		if host, ok := config["host"].(string); ok {
			player.host = host
		}
		if port, ok := config["port"].(string); ok {
			player.port = port
		}
		if password, ok := config["password"].(string); ok {
			player.password = password
		}
	}

	return player
}

// Initialize 初始化播放器
func (p *MPDPlayer) Initialize(ctx context.Context) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.initialized {
		return nil
	}

	if !p.IsAvailable() {
		return fmt.Errorf("MPD is not available on this system")
	}

	// 连接到MPD服务器
	if err := p.connect(); err != nil {
		return fmt.Errorf("failed to connect to MPD: %w", err)
	}

	p.initialized = true
	return nil
}

// Cleanup 清理资源
func (p *MPDPlayer) Cleanup() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.conn != nil {
		p.sendCommand("close")
		p.conn.Close()
		p.conn = nil
		p.connected = false
	}

	p.setPlaying(false)
	p.initialized = false

	return nil
}

// Play 播放音频
func (p *MPDPlayer) Play(url string) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if !p.connected {
		return fmt.Errorf("not connected to MPD")
	}

	// 清空播放列表
	if err := p.sendCommand("clear"); err != nil {
		return fmt.Errorf("failed to clear playlist: %w", err)
	}

	// 添加音频到播放列表
	if err := p.sendCommand(fmt.Sprintf(`add "%s"`, url)); err != nil {
		return fmt.Errorf("failed to add audio to playlist: %w", err)
	}

	// 开始播放
	if err := p.sendCommand("play 0"); err != nil {
		return fmt.Errorf("failed to start playback: %w", err)
	}

	p.setPlaying(true)
	p.setPosition(0)

	// 启动状态更新协程
	go p.updateStatus()

	return nil
}

// Pause 暂停播放
func (p *MPDPlayer) Pause() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if !p.connected {
		return fmt.Errorf("not connected to MPD")
	}

	if err := p.sendCommand("pause 1"); err != nil {
		return fmt.Errorf("failed to pause: %w", err)
	}

	p.setPlaying(false)
	return nil
}

// Resume 恢复播放
func (p *MPDPlayer) Resume() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if !p.connected {
		return fmt.Errorf("not connected to MPD")
	}

	if err := p.sendCommand("pause 0"); err != nil {
		return fmt.Errorf("failed to resume: %w", err)
	}

	p.setPlaying(true)
	return nil
}

// Stop 停止播放
func (p *MPDPlayer) Stop() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if !p.connected {
		return fmt.Errorf("not connected to MPD")
	}

	if err := p.sendCommand("stop"); err != nil {
		return fmt.Errorf("failed to stop: %w", err)
	}

	p.setPlaying(false)
	p.setPosition(0)

	return nil
}

// Seek 跳转到指定位置
func (p *MPDPlayer) Seek(position time.Duration) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if !p.connected {
		return fmt.Errorf("not connected to MPD")
	}

	positionSeconds := int(position.Seconds())
	cmd := fmt.Sprintf("seekcur %d", positionSeconds)
	if err := p.sendCommand(cmd); err != nil {
		return fmt.Errorf("failed to seek: %w", err)
	}

	p.setPosition(position)
	return nil
}

// SetVolume 设置音量
func (p *MPDPlayer) SetVolume(volume float64) error {
	if err := p.BasePlayer.SetVolume(volume); err != nil {
		return err
	}

	p.mutex.Lock()
	defer p.mutex.Unlock()

	if !p.connected {
		return fmt.Errorf("not connected to MPD")
	}

	// MPD音量范围是0-100
	volumeLevel := int(volume * 100)
	cmd := fmt.Sprintf("setvol %d", volumeLevel)
	if err := p.sendCommand(cmd); err != nil {
		return fmt.Errorf("failed to set volume: %w", err)
	}

	return nil
}

// IsAvailable 检查播放器是否可用
func (p *MPDPlayer) IsAvailable() bool {
	// 检查是否是Linux/Unix系统
	if runtime.GOOS != "linux" && runtime.GOOS != "unix" {
		return false
	}

	// 尝试连接MPD服务器
	conn, err := net.DialTimeout("tcp", p.host+":"+p.port, 2*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// connect 连接到MPD服务器
func (p *MPDPlayer) connect() error {
	conn, err := net.Dial("tcp", p.host+":"+p.port)
	if err != nil {
		return err
	}

	p.conn = conn

	// 读取欢迎消息
	reader := bufio.NewReader(conn)
	welcome, _, err := reader.ReadLine()
	if err != nil {
		conn.Close()
		return err
	}

	if !strings.HasPrefix(string(welcome), "OK MPD") {
		conn.Close()
		return fmt.Errorf("invalid MPD welcome message: %s", welcome)
	}

	// 如果有密码，进行认证
	if p.password != "" {
		if err := p.sendCommand(fmt.Sprintf(`password "%s"`, p.password)); err != nil {
			conn.Close()
			return fmt.Errorf("authentication failed: %w", err)
		}
	}

	p.connected = true
	return nil
}

// sendCommand 发送MPD命令
func (p *MPDPlayer) sendCommand(command string) error {
	if p.conn == nil {
		return fmt.Errorf("not connected")
	}

	// 发送命令
	_, err := p.conn.Write([]byte(command + "\n"))
	if err != nil {
		return err
	}

	// 读取响应
	reader := bufio.NewReader(p.conn)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return err
		}

		line = strings.TrimSpace(line)
		if line == "OK" {
			break
		}
		if strings.HasPrefix(line, "ACK") {
			return fmt.Errorf("MPD error: %s", line)
		}
	}

	return nil
}

// getStatus 获取MPD状态
func (p *MPDPlayer) getStatus() (map[string]string, error) {
	if p.conn == nil {
		return nil, fmt.Errorf("not connected")
	}

	// 发送status命令
	_, err := p.conn.Write([]byte("status\n"))
	if err != nil {
		return nil, err
	}

	// 读取响应
	status := make(map[string]string)
	reader := bufio.NewReader(p.conn)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}

		line = strings.TrimSpace(line)
		if line == "OK" {
			break
		}
		if strings.HasPrefix(line, "ACK") {
			return nil, fmt.Errorf("MPD error: %s", line)
		}

		// 解析状态行
		parts := strings.SplitN(line, ": ", 2)
		if len(parts) == 2 {
			status[parts[0]] = parts[1]
		}
	}

	return status, nil
}

// updateStatus 更新播放状态
func (p *MPDPlayer) updateStatus() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for range ticker.C {
		p.mutex.RLock()
		if !p.connected {
			p.mutex.RUnlock()
			break
		}
		p.mutex.RUnlock()

		status, err := p.getStatus()
		if err != nil {
			continue
		}

		// 更新播放状态
		if state, ok := status["state"]; ok {
			p.setPlaying(state == "play")
		}

		// 更新播放位置
		if elapsed, ok := status["elapsed"]; ok {
			if seconds, err := strconv.ParseFloat(elapsed, 64); err == nil {
				position := time.Duration(seconds * float64(time.Second))
				p.setPosition(position)
			}
		}

		// 更新音乐时长
		if duration, ok := status["duration"]; ok {
			if seconds, err := strconv.ParseFloat(duration, 64); err == nil {
				dur := time.Duration(seconds * float64(time.Second))
				p.setDuration(dur)
			}
		}
	}
}
