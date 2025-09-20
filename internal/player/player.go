package player

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/types"
)

type Player interface {
	Play(music URLMusic)
	CurMusic() URLMusic
	Pause()
	Resume()
	Stop()
	Toggle()
	Seek(duration time.Duration)
	PassedTime() time.Duration
	PlayedTime() time.Duration
	TimeChan() <-chan time.Duration
	State() types.State
	StateChan() <-chan types.State
	Volume() int
	SetVolume(volume int)
	UpVolume()
	DownVolume()
	Close()
}

func NewPlayerFromConfig() Player {
	cfg := configs.AppConfig
	var player Player
	switch cfg.Player.Engine {
	case types.BeepPlayer:
		player = NewBeepPlayer()
	case types.MpdPlayer:
		if cfg.Player.Mpd.Network == "" || cfg.Player.Mpd.Addr == "" ||
			cfg.Player.Mpd.Bin == "" {
			panic("缺少MPD配置")
		}
		cmd := exec.Command(cfg.Player.Mpd.Bin, "--version")
		if err := cmd.Run(); err != nil {
			panic(err)
		}
		player = NewMpdPlayer(&MpdConfig{
			Bin:        cfg.Player.Mpd.Bin,
			ConfigFile: cfg.Player.Mpd.ConfigFile,
			Network:    cfg.Player.Mpd.Network,
			Address:    cfg.Player.Mpd.Addr,
			AutoStart:  cfg.Player.Mpd.AutoStart,
		})
	case types.OsxPlayer:
		player = NewOsxPlayer()
	case types.WinMediaPlayer:
		player = NewWinMediaPlayer()
	case types.MpvPlayer:
		cmd := exec.Command(cfg.Player.Mpv.Bin, "--version")
		output, err := cmd.CombinedOutput()
		if err != nil || !strings.Contains(string(output), "mpv") {
			panic(fmt.Sprintf("MPV不可用: %v, 输出: %s", err, string(output)))
		}
		player = NewMpvPlayer(&MpvConfig{
			BinPath: cfg.Player.Mpv.Bin, // 使用配置文件中的mpv路径
		})
	default:
		panic("unknown player engine")
	}

	return player
}
