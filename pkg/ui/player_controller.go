package ui

import (
	"time"

	"github.com/go-musicfox/go-musicfox/pkg/state_handler"
)

func (p *Player) Paused() {
	// NOTICE: 提供给state_handler调用，因为有GC panic问题，这里使用chan传递
	p.ctrl <- CtrlSignal{Type: CtrlPaused}
}

func (p *Player) Resume() {
	// NOTICE: 提供给state_handler调用，因为有GC panic问题，这里使用chan传递
	p.ctrl <- CtrlSignal{Type: CtrlResume}
}

func (p *Player) Stop() {
	// NOTICE: 提供给state_handler调用，因为有GC panic问题，这里使用chan传递
	p.ctrl <- CtrlSignal{Type: CtrlStop}
}

func (p *Player) Toggle() {
	// NOTICE: 提供给state_handler调用，因为有GC panic问题，这里使用chan传递
	p.ctrl <- CtrlSignal{Type: CtrlToggle}
}

func (p *Player) Next() {
	// NOTICE: 提供给state_handler调用，因为有GC panic问题，这里使用chan传递
	p.ctrl <- CtrlSignal{Type: CtrlNext}
}

func (p *Player) Previous() {
	// NOTICE: 提供给state_handler调用，因为有GC panic问题，这里使用chan传递
	p.ctrl <- CtrlSignal{Type: CtrlPrevious}
}

func (p *Player) Rerender() {
	p.ctrl <- CtrlSignal{Type: CtrlRerender}
}

func (p *Player) Seek(duration time.Duration) {
	p.ctrl <- CtrlSignal{
		Type:     CtrlSeek,
		Duration: duration,
	}
}

func (p *Player) SetVolumeByExternalCtrl(volume int) {
	// 不更新playingInfo
	p.Player.SetVolume(volume)
}

func (p *Player) PlayingInfo() state_handler.PlayingInfo {
	music := p.curSong
	return state_handler.PlayingInfo{
		TotalDuration:  music.Duration,
		PassedDuration: p.PassedTime(),
		State:          p.State(),
		Volume:         p.Volume(),
		TrackID:        music.Id,
		PicUrl:         music.PicUrl,
		Name:           music.Name,
		Album:          music.Album.Name,
		Artist:         music.ArtistName(),
		AlbumArtist:    music.Album.ArtistName(),
	}
}
