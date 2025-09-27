package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbletea"
)

// Player 播放器界面实现
type Player struct {
	plugin *TUIPlugin
	
	// 播放状态
	isPlaying bool
	currentSong *Song
	position time.Duration
	duration time.Duration
	volume int
	playMode PlayMode
	
	// 歌词相关
	lyrics []LyricLine
	currentLyricIndex int
	showLyrics bool
	
	// 界面状态
	width int
	height int
	renderTicker *time.Ticker
	ctx context.Context
	cancel context.CancelFunc
}

// Song 歌曲信息
type Song struct {
	ID string
	Name string
	Artist string
	Album string
	Duration time.Duration
}

// LyricLine 歌词行
type LyricLine struct {
	Time time.Duration
	Text string
}

// PlayMode 播放模式
type PlayMode int

const (
	PlayModeSequence PlayMode = iota // 顺序播放
	PlayModeLoop                     // 单曲循环
	PlayModeRandom                   // 随机播放
	PlayModeListLoop                 // 列表循环
)

// NewPlayer 创建播放器
func NewPlayer(plugin *TUIPlugin) *Player {
	ctx, cancel := context.WithCancel(context.Background())
	p := &Player{
		plugin: plugin,
		volume: 80,
		playMode: PlayModeSequence,
		showLyrics: true,
		renderTicker: time.NewTicker(100 * time.Millisecond),
		ctx: ctx,
		cancel: cancel,
	}
	
	// 启动渲染循环
	go p.renderLoop()
	
	return p
}

// renderLoop 渲染循环
func (p *Player) renderLoop() {
	for {
		select {
		case <-p.ctx.Done():
			return
		case <-p.renderTicker.C:
			if p.isPlaying {
				// 更新播放位置
				p.position += 100 * time.Millisecond
				// 更新歌词索引
				p.updateLyricIndex()
				// 触发界面更新
				if p.plugin.app != nil {
					// TODO: 添加适当的类型断言来调用Rerender方法
				}
			}
		}
	}
}

// updateLyricIndex 更新歌词索引
func (p *Player) updateLyricIndex() {
	for i, lyric := range p.lyrics {
		if p.position >= lyric.Time {
			p.currentLyricIndex = i
		} else {
			break
		}
	}
}

// Play 播放歌曲
func (p *Player) Play(song *Song) {
	p.currentSong = song
	p.isPlaying = true
	p.position = 0
	p.duration = song.Duration
	p.currentLyricIndex = 0
	// 加载歌词
	p.loadLyrics(song.ID)
}

// Pause 暂停播放
func (p *Player) Pause() {
	p.isPlaying = false
}

// Resume 恢复播放
func (p *Player) Resume() {
	p.isPlaying = true
}

// Stop 停止播放
func (p *Player) Stop() {
	p.isPlaying = false
	p.position = 0
	p.currentSong = nil
	p.lyrics = nil
	p.currentLyricIndex = 0
}

// SetVolume 设置音量
func (p *Player) SetVolume(volume int) {
	if volume < 0 {
		volume = 0
	} else if volume > 100 {
		volume = 100
	}
	p.volume = volume
}

// SetPlayMode 设置播放模式
func (p *Player) SetPlayMode(mode PlayMode) {
	p.playMode = mode
}

// ToggleLyrics 切换歌词显示
func (p *Player) ToggleLyrics() {
	p.showLyrics = !p.showLyrics
}

// loadLyrics 加载歌词
func (p *Player) loadLyrics(songID string) {
	// 从API加载歌词
	go func() {
		// 这里应该调用网易云音乐API获取歌词
		// 暂时使用模拟数据演示歌词解析功能
		lrcContent := p.fetchLyricsFromAPI(songID)
		if lrcContent != "" {
			p.lyrics = p.parseLRC(lrcContent)
		} else {
			// 如果没有歌词，显示默认信息
			p.lyrics = []LyricLine{
				{Time: 0, Text: "暂无歌词"},
			}
		}
	}()
}

// fetchLyricsFromAPI 从API获取歌词
func (p *Player) fetchLyricsFromAPI(songID string) string {
	// 这里应该实现真正的API调用
	// 暂时返回模拟的LRC格式歌词
	if songID == "" {
		return ""
	}
	
	// 模拟LRC格式歌词
	return `[00:00.00]歌词加载中...
[00:05.00]这是一首美妙的歌曲
[00:10.00]让我们一起聆听
[00:15.00]音乐带来的快乐
[00:20.00]感受旋律的魅力
[00:25.00]在这个美好的时刻`
}

// parseLRC 解析LRC格式歌词
func (p *Player) parseLRC(lrcContent string) []LyricLine {
	lines := strings.Split(lrcContent, "\n")
	var lyrics []LyricLine
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		
		// 解析时间标签 [mm:ss.xx]
		if strings.HasPrefix(line, "[") {
			endIndex := strings.Index(line, "]")
			if endIndex > 0 {
				timeStr := line[1:endIndex]
				text := strings.TrimSpace(line[endIndex+1:])
				
				// 解析时间
				duration := p.parseTimeString(timeStr)
				if text != "" {
					lyrics = append(lyrics, LyricLine{
						Time: duration,
						Text: text,
					})
				}
			}
		}
	}
	
	// 按时间排序
	for i := 0; i < len(lyrics)-1; i++ {
		for j := i + 1; j < len(lyrics); j++ {
			if lyrics[i].Time > lyrics[j].Time {
				lyrics[i], lyrics[j] = lyrics[j], lyrics[i]
			}
		}
	}
	
	return lyrics
}

// parseTimeString 解析时间字符串 (mm:ss.xx)
func (p *Player) parseTimeString(timeStr string) time.Duration {
	// 解析 mm:ss.xx 格式
	parts := strings.Split(timeStr, ":")
	if len(parts) != 2 {
		return 0
	}
	
	// 解析分钟
	minutes := 0
	if m, err := fmt.Sscanf(parts[0], "%d", &minutes); err != nil || m != 1 {
		return 0
	}
	
	// 解析秒和毫秒
	secondsParts := strings.Split(parts[1], ".")
	seconds := 0
	milliseconds := 0
	
	if s, err := fmt.Sscanf(secondsParts[0], "%d", &seconds); err != nil || s != 1 {
		return 0
	}
	
	if len(secondsParts) > 1 {
		if ms, err := fmt.Sscanf(secondsParts[1], "%d", &milliseconds); err == nil && ms == 1 {
			// 如果是两位数，需要乘以10变成毫秒
			if len(secondsParts[1]) == 2 {
				milliseconds *= 10
			}
		}
	}
	
	totalMs := minutes*60*1000 + seconds*1000 + milliseconds
	return time.Duration(totalMs) * time.Millisecond
}

// Render 渲染播放器界面
func (p *Player) Render(width, height int) string {
	p.width = width
	p.height = height
	
	if p.currentSong == nil {
		return p.renderEmpty()
	}
	
	var builder strings.Builder
	
	// 渲染歌曲信息
	builder.WriteString(p.renderSongInfo())
	builder.WriteString("\n")
	
	// 渲染进度条
	builder.WriteString(p.renderProgressBar())
	builder.WriteString("\n")
	
	// 渲染控制按钮
	builder.WriteString(p.renderControls())
	builder.WriteString("\n")
	
	// 渲染歌词
	if p.showLyrics {
		builder.WriteString(p.renderLyrics())
	}
	
	return builder.String()
}

// renderEmpty 渲染空状态
func (p *Player) renderEmpty() string {
	return "暂无播放歌曲"
}

// renderSongInfo 渲染歌曲信息
func (p *Player) renderSongInfo() string {
	if p.currentSong == nil {
		return ""
	}
	
	var builder strings.Builder
	builder.WriteString(p.currentSong.Name)
	builder.WriteString(" - ")
	builder.WriteString(p.currentSong.Artist)
	builder.WriteString(" [")
	builder.WriteString(p.currentSong.Album)
	builder.WriteString("]")
	
	return builder.String()
}

// renderProgressBar 渲染进度条
func (p *Player) renderProgressBar() string {
	if p.currentSong == nil || p.duration == 0 {
		return ""
	}
	
	progress := float64(p.position) / float64(p.duration)
	barWidth := p.width - 20 // 留出时间显示的空间
	if barWidth < 10 {
		barWidth = 10
	}
	
	filledWidth := int(progress * float64(barWidth))
	emptyWidth := barWidth - filledWidth
	
	var builder strings.Builder
	builder.WriteString(p.formatDuration(p.position))
	builder.WriteString(" [")
	builder.WriteString(strings.Repeat("=", filledWidth))
	builder.WriteString(strings.Repeat("-", emptyWidth))
	builder.WriteString("] ")
	builder.WriteString(p.formatDuration(p.duration))
	
	return builder.String()
}

// renderControls 渲染控制按钮
func (p *Player) renderControls() string {
	var builder strings.Builder
	
	// 播放/暂停按钮
	if p.isPlaying {
		builder.WriteString("[暂停] ")
	} else {
		builder.WriteString("[播放] ")
	}
	
	// 播放模式
	switch p.playMode {
	case PlayModeSequence:
		builder.WriteString("[顺序] ")
	case PlayModeLoop:
		builder.WriteString("[单曲] ")
	case PlayModeRandom:
		builder.WriteString("[随机] ")
	case PlayModeListLoop:
		builder.WriteString("[列表] ")
	}
	
	// 音量
	builder.WriteString(fmt.Sprintf("[音量:%d%%] ", p.volume))
	
	// 歌词开关
	if p.showLyrics {
		builder.WriteString("[隐藏歌词]")
	} else {
		builder.WriteString("[显示歌词]")
	}
	
	return builder.String()
}

// renderLyrics 渲染歌词
func (p *Player) renderLyrics() string {
	if len(p.lyrics) == 0 {
		return "暂无歌词"
	}
	
	var builder strings.Builder
	builder.WriteString("\n歌词:\n")
	
	// 显示当前歌词及前后几行
	start := p.currentLyricIndex - 2
	if start < 0 {
		start = 0
	}
	end := p.currentLyricIndex + 3
	if end > len(p.lyrics) {
		end = len(p.lyrics)
	}
	
	for i := start; i < end; i++ {
		if i == p.currentLyricIndex {
			// 高亮当前歌词
			builder.WriteString(">>> " + p.lyrics[i].Text + " <<<")
		} else {
			builder.WriteString(p.lyrics[i].Text)
		}
		builder.WriteString("\n")
	}
	
	return builder.String()
}

// formatDuration 格式化时长
func (p *Player) formatDuration(d time.Duration) string {
	minutes := int(d.Minutes())
	seconds := int(d.Seconds()) % 60
	return fmt.Sprintf("%02d:%02d", minutes, seconds)
}

// HandleKeyMsg 处理按键消息
func (p *Player) HandleKeyMsg(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case " ": // 空格键播放/暂停
		if p.isPlaying {
			p.Pause()
		} else {
			p.Resume()
		}
	case "l": // 切换歌词显示
		p.ToggleLyrics()
	case "m": // 切换播放模式
		p.SetPlayMode((p.playMode + 1) % 4)
	case "+", "=": // 增加音量
		p.SetVolume(p.volume + 5)
	case "-": // 减少音量
		p.SetVolume(p.volume - 5)
	}
	return nil
}

// Close 关闭播放器
func (p *Player) Close() {
	if p.renderTicker != nil {
		p.renderTicker.Stop()
	}
	if p.cancel != nil {
		p.cancel()
	}
}