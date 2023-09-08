package types

// Mode 播放模式
type Mode uint8

const (
	PmUnknown Mode = iota
	PmListLoop
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

type State uint8

const (
	Unknown State = iota
	Playing
	Paused
	Stopped
	Interrupted
)
