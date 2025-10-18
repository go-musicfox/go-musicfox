package types

// Mode 播放模式
type Mode uint8

const (
	PmUnknown Mode = iota
	PmListLoop
	PmOrdered
	PmSingleLoop
	PmListRandom
	PmInfRandom
	PmIntelligent
)

// String implements the fmt.Stringer interface for the Mode type.
func (m Mode) String() string {
	switch m {
	case PmListLoop:
		return "列表循环"
	case PmOrdered:
		return "顺序播放"
	case PmSingleLoop:
		return "单曲循环"
	case PmListRandom:
		return "列表随机"
	case PmIntelligent:
		return "心动模式"
	case PmInfRandom:
		return "无限随机"
	default:
		return "未知模式"
	}
}

// Name returns the human-readable name of the play mode.
func (m Mode) Name() string {
	return m.String()
}

type State uint8

const (
	Unknown State = iota
	Playing
	Paused
	Stopped
	Interrupted
)
