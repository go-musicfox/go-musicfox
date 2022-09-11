//go:build darwin
// +build darwin

package core

import "fmt"

type CMTime struct {
	Value     int64
	Timescale int32
	Flags     uint32
	Epoch     int64
}

func (time CMTime) String() string {
	return fmt.Sprintf("(value: %v, timescale: %v, flags: %v, epoch: %v)", time.Value, time.Timescale, time.Flags, time.Epoch)
}
