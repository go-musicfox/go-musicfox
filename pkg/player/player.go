package player

import (
	"os/exec"
	"time"

	"go-musicfox/pkg/constants"

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
	Volume() int
	SetVolume(volume int)
	UpVolume()
	DownVolume()
	Close()
}

func NewPlayerFromConfig() Player {
	registry := configs.ConfigRegistry
	var player Player
	switch registry.PlayerEngine {
	case constants.BeepPlayer:
		player = NewBeepPlayer()
	case constants.MpdPlayer:
		if registry.PlayerMpdNetwork == "" || registry.PlayerMpdAddr == "" ||
			registry.PlayerBin == "" {
			panic("缺少MPD配置")
		}
		cmd := exec.Command(registry.PlayerBin, "--version")
		if err := cmd.Run(); err != nil {
			panic(err)
		}
		player = NewMpdPlayer(registry.PlayerBin, registry.PlayerConfigFile, registry.PlayerMpdNetwork, registry.PlayerMpdAddr)
	case constants.OsxPlayer:
		player = NewOsxPlayer()
	default:
		panic("unknown player engine")
	}

	return player
}

type State uint8

const (
	Unknown State = iota
	Playing
	Paused
	Stopped
	Interrupted
)

// Mode 播放模式
type Mode uint8

const (
	PmListLoop Mode = iota + 1
	PmOrder
	PmSingleLoop
	PmRandom
	PmIntelligent
)

var modeNames = map[Mode]string{
	PmListLoop:    "列表",
	PmOrder:       "顺序",
	PmSingleLoop:  "单曲",
	PmRandom:      "随机",
	PmIntelligent: "心动",
}

func ModeName(mode Mode) string {
	if name, ok := modeNames[mode]; ok {
		return name
	}
	return "未知"
}
