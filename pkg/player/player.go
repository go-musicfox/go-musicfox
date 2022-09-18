package player

import (
	"go-musicfox/pkg/configs"
	"os/exec"
	"time"
)

type Player interface {
	Play(songType SongType, url string, duration time.Duration)
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
	}

	return player
}
