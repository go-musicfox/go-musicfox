package player

import (
	"os/exec"
	"time"

	"go-musicfox/pkg/configs"
)

type Player interface {
	Play(music UrlMusic)
	CurMusic() UrlMusic
	Paused()
	Resume()
	Stop()
	Toggle()
	Seek(duration time.Duration)
	PassedTime() time.Duration
	TimeChan() <-chan time.Duration
	State() State
	StateChan() <-chan State
	UpVolume()
	DownVolume()
	SetVolume(volume int)
	Volume() int
	Close()
}

func NewPlayerFromConfig() Player {
	registry := configs.ConfigRegistry
	var player Player
	switch registry.PlayerEngine {
	case "beep":
		player = NewBeepPlayer()
	case "mpd":
		if registry.PlayerMpdNetwork == "" || registry.PlayerMpdAddr == "" ||
			registry.PlayerBin == "" {
			panic("缺少MPD配置")
		}
		cmd := exec.Command(registry.PlayerBin, "--version")
		if err := cmd.Run(); err != nil {
			panic(err)
		}
		player = NewMpdPlayer(registry.PlayerBin, registry.PlayerConfigFile, registry.PlayerMpdNetwork, registry.PlayerMpdAddr)
	case "osx":
		player = NewOsxPlayer()
	default:
		panic("unknown player engine")
	}

	return player
}
