package player

import (
	"os/exec"
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
	registry := configs.ConfigRegistry
	var player Player
	switch registry.Player.Engine {
	case types.BeepPlayer:
		player = NewBeepPlayer()
	case types.MpdPlayer:
		if registry.Player.MpdNetwork == "" || registry.Player.MpdAddr == "" ||
			registry.Player.MpdBin == "" {
			panic("缺少MPD配置")
		}
		cmd := exec.Command(registry.Player.MpdBin, "--version")
		if err := cmd.Run(); err != nil {
			panic(err)
		}
		player = NewMpdPlayer(&MpdConfig{
			Bin:        registry.Player.MpdBin,
			ConfigFile: registry.Player.MpdConfigFile,
			Network:    registry.Player.MpdNetwork,
			Address:    registry.Player.MpdAddr,
			AutoStart:  registry.Player.MpdAutoStart,
		})
	case types.OsxPlayer:
		player = NewOsxPlayer()
	case types.WinMediaPlayer:
		player = NewWinMediaPlayer()
	default:
		panic("unknown player engine")
	}

	return player
}
