// Copyright 2019 The Oto Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package oto

import (
	"time"
)

type dummyDriver struct {
	sampleRate      int
	channelNum      int
	bitDepthInBytes int

	current int
}

func newDummyDriver(sampleRate, channelNum, bitDepthInBytes int) *dummyDriver {
	return &dummyDriver{
		sampleRate:      sampleRate,
		channelNum:      channelNum,
		bitDepthInBytes: bitDepthInBytes,
	}
}

func (d *dummyDriver) bytes(t time.Duration) int {
	return int(float64(d.sampleRate*d.channelNum*d.bitDepthInBytes) * float64(t) / float64(time.Second))
}

func (d *dummyDriver) TryWrite(buf []byte) (int, error) {
	d.current += len(buf)
	b := d.bytes(100 * time.Millisecond)
	for d.current >= b {
		time.Sleep(time.Second)
		d.current -= b
	}
	return len(buf), nil
}

func (d *dummyDriver) Close() error {
	return nil
}

func (d *dummyDriver) tryWriteCanReturnWithoutWaiting() bool {
	return false
}
