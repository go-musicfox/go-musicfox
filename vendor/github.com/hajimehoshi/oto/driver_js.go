// Copyright 2015 Hajime Hoshi
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

//go:build js
// +build js

package oto

import (
	"encoding/base64"
	"errors"
	"fmt"
	"sync"
	"syscall/js"
)

type driver struct {
	sampleRate      int
	channelNum      int
	bitDepthInBytes int
	nextPos         float64
	tmp             []byte
	bufferSize      int
	context         js.Value
	ready           bool
	callbacks       map[string]js.Func

	// For Audio Worklet
	workletNode     js.Value
	workletNodePost js.Value
	messageArray    js.Value
	transferArray   js.Value
	bufs            [][]js.Value
	cond            *sync.Cond
}

type warn struct {
	msg string
}

func (w *warn) Error() string {
	return w.msg
}

const audioBufferSamples = 3200

func tryAudioWorklet(context js.Value, channelNum int) (js.Value, error) {
	if !js.Global().Get("AudioWorkletNode").Truthy() {
		return js.Undefined(), nil
	}

	worklet := context.Get("audioWorklet")
	if !worklet.Truthy() {
		return js.Undefined(), &warn{
			msg: "AudioWorklet is not available due to the insecure context. See https://developer.mozilla.org/en-US/docs/Web/API/AudioWorklet",
		}
	}

	script := `
class EbitenAudioWorkletProcessor extends AudioWorkletProcessor {
  constructor() {
    super();

    this.buffers_ = [[], []];
    this.offsets_ = [0, 0];
    this.offsetsInArray_ = [0, 0];
    this.consumed_ = [];

    this.port.onmessage = (e) => {
      const bufs = e.data;
      for (let ch = 0; ch < bufs.length; ch++) {
        const buf = bufs[ch];
        this.buffers_[ch].push(new Float32Array(buf.buffer, buf.byteOffset, buf.byteLength / 4));
      }
    };
  }

  bufferTotalLength(ch) {
    const sum = this.buffers_[ch].reduce((total, buf) => total + buf.length, 0);
    return sum - this.offsetsInArray_[ch];
  }

  consume(ch, i) {
    while (this.buffers_[ch][0].length <= i - this.offsets_[ch]) {
      this.offsets_[ch] += this.buffers_[ch][0].length;
      this.offsetsInArray_[ch] = 0;
      const buf = this.buffers_[ch].shift();
      this.appendConsumedBuffer(ch, buf);
    }
    this.offsetsInArray_[ch]++;
    return this.buffers_[ch][0][i - this.offsets_[ch]];
  }

  appendConsumedBuffer(ch, buf) {
    let idx = this.consumed_.length - 1;
    if (idx < 0 || this.consumed_[idx][ch]) {
      this.consumed_.push([]);
      idx++;
    }
    this.consumed_[idx][ch] = new Uint8Array(buf.buffer, buf.byteOffset, buf.byteLength);
  }

  process(inputs, outputs, parameters) {
    const out = outputs[0];

    if (this.bufferTotalLength(0) < out[0].length) {
      for (let ch = 0; ch < out.length; ch++) {
        for (let i = 0; i < out[ch].length; i++) {
          out[ch][i] = 0;
        }
      }
      return true;
    }

    for (let ch = 0; ch < out.length; ch++) {
      const offset = this.offsets_[ch] + this.offsetsInArray_[ch];
      for (let i = 0; i < out[ch].length; i++) {
        out[ch][i] = this.consume(ch, i + offset);
      }
    }

    for (let bufs of this.consumed_) {
      this.port.postMessage(bufs, bufs.map(buf => buf.buffer));
    }
    this.consumed_ = [];

    return true;
  }
}

registerProcessor('ebiten-audio-worklet-processor', EbitenAudioWorkletProcessor);`
	scriptURL := "data:application/javascript;base64," + base64.StdEncoding.EncodeToString([]byte(script))

	ch := make(chan error)
	worklet.Call("addModule", scriptURL).Call("then", js.FuncOf(func(js.Value, []js.Value) interface{} {
		close(ch)
		return nil
	})).Call("catch", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		err := args[0]
		ch <- fmt.Errorf("oto: error at addModule: %s: %s", err.Get("name").String(), err.Get("message").String())
		close(ch)
		return nil
	}))
	if err := <-ch; err != nil {
		return js.Undefined(), err
	}

	options := js.Global().Get("Object").New()
	arr := js.Global().Get("Array").New()
	arr.Call("push", channelNum)
	options.Set("outputChannelCount", arr)

	node := js.Global().Get("AudioWorkletNode").New(context, "ebiten-audio-worklet-processor", options)
	node.Call("connect", context.Get("destination"))

	return node, nil
}

func newDriver(sampleRate, channelNum, bitDepthInBytes, bufferSize int) (tryWriteCloser, error) {
	class := js.Global().Get("AudioContext")
	if !class.Truthy() {
		class = js.Global().Get("webkitAudioContext")
	}
	if !class.Truthy() {
		return nil, errors.New("oto: audio couldn't be initialized")
	}

	options := js.Global().Get("Object").New()
	options.Set("sampleRate", sampleRate)
	context := class.New(options)

	node, err := tryAudioWorklet(context, channelNum)
	if err != nil {
		w, ok := err.(*warn)
		if !ok {
			return nil, err
		}
		js.Global().Get("console").Call("warn", w.Error())
	}

	bs := bufferSize
	if !node.Truthy() {
		bs = max(bufferSize, audioBufferSamples*channelNum*bitDepthInBytes)
	} else {
		bs = max(bufferSize, 4096)
	}

	p := &driver{
		sampleRate:      sampleRate,
		channelNum:      channelNum,
		bitDepthInBytes: bitDepthInBytes,
		context:         context,
		workletNode:     node,
		bufferSize:      bs,
	}

	if node.Truthy() {
		port := node.Get("port")
		p.workletNodePost = port.Get("postMessage").Call("bind", port)
		p.messageArray = js.Global().Get("Array").New(2)
		p.transferArray = js.Global().Get("Array").New(2)
		p.cond = sync.NewCond(&sync.Mutex{})

		s := p.bufferSize / p.channelNum / p.bitDepthInBytes * 4
		p.bufs = [][]js.Value{
			{
				js.Global().Get("Uint8Array").New(s),
				js.Global().Get("Uint8Array").New(s),
			},
			{
				js.Global().Get("Uint8Array").New(s),
				js.Global().Get("Uint8Array").New(s),
			},
		}

		node.Get("port").Set("onmessage", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			p.cond.L.Lock()
			defer p.cond.L.Unlock()

			bufs := args[0].Get("data")
			var arr []js.Value
			for i := 0; i < bufs.Length(); i++ {
				arr = append(arr, bufs.Index(i))
			}

			notify := len(p.bufs) == 0
			p.bufs = append(p.bufs, arr)
			if notify {
				p.cond.Signal()
			}

			return nil
		}))
	}

	setCallback := func(event string) js.Func {
		var f js.Func
		f = js.FuncOf(func(this js.Value, arguments []js.Value) interface{} {
			if !p.ready {
				p.context.Call("resume")
				p.ready = true
			}
			js.Global().Get("document").Call("removeEventListener", event, f)
			return nil
		})
		js.Global().Get("document").Call("addEventListener", event, f)
		p.callbacks[event] = f
		return f
	}

	// Browsers require user interaction to start the audio.
	// https://developers.google.com/web/updates/2017/09/autoplay-policy-changes#webaudio
	p.callbacks = map[string]js.Func{}
	setCallback("touchend")
	setCallback("keyup")
	setCallback("mouseup")
	return p, nil
}

func toLR(data []byte) ([]float32, []float32) {
	const max = 1 << 15

	l := make([]float32, len(data)/4)
	r := make([]float32, len(data)/4)
	for i := 0; i < len(data)/4; i++ {
		l[i] = float32(int16(data[4*i])|int16(data[4*i+1])<<8) / max
		r[i] = float32(int16(data[4*i+2])|int16(data[4*i+3])<<8) / max
	}
	return l, r
}

func (p *driver) TryWrite(data []byte) (int, error) {
	if !p.ready {
		return 0, nil
	}

	if p.workletNode.Truthy() {
		p.cond.L.Lock()
		defer p.cond.L.Unlock()

		n := min(len(data), max(0, p.bufferSize-len(p.tmp)))
		p.tmp = append(p.tmp, data[:n]...)

		if len(p.tmp) < p.bufferSize {
			return n, nil
		}

		for len(p.bufs) == 0 {
			p.cond.Wait()
		}

		l, r := toLR(p.tmp[:p.bufferSize])
		tl := p.bufs[0][0]
		tr := p.bufs[0][1]
		copyFloat32sToJS(tl, l)
		copyFloat32sToJS(tr, r)
		p.tmp = p.tmp[p.bufferSize:]

		bufs := p.messageArray
		bufs.SetIndex(0, tl)
		bufs.SetIndex(1, tr)
		transfers := p.transferArray
		transfers.SetIndex(0, tl.Get("buffer"))
		transfers.SetIndex(1, tr.Get("buffer"))

		p.workletNodePost.Invoke(bufs, transfers)

		p.bufs = p.bufs[1:]

		return n, nil
	}

	n := min(len(data), max(0, p.bufferSize-len(p.tmp)))
	p.tmp = append(p.tmp, data[:n]...)

	c := p.context.Get("currentTime").Float()

	if p.nextPos < c {
		p.nextPos = c
	}

	// It's too early to enqueue a buffer.
	// Highly likely, there are two playing buffers now.
	if c+float64(p.bufferSize/p.bitDepthInBytes/p.channelNum)/float64(p.sampleRate) < p.nextPos {
		return n, nil
	}

	le := audioBufferSamples * p.bitDepthInBytes * p.channelNum
	if len(p.tmp) < le {
		return n, nil
	}

	buf := p.context.Call("createBuffer", p.channelNum, audioBufferSamples, p.sampleRate)
	l, r := toLR(p.tmp[:le])
	tl, freel := float32SliceToTypedArray(l)
	tr, freer := float32SliceToTypedArray(r)
	if buf.Get("copyToChannel").Truthy() {
		buf.Call("copyToChannel", tl, 0, 0)
		buf.Call("copyToChannel", tr, 1, 0)
	} else {
		// copyToChannel is not defined on Safari 11
		buf.Call("getChannelData", 0).Call("set", tl)
		buf.Call("getChannelData", 1).Call("set", tr)
	}
	freel()
	freer()

	s := p.context.Call("createBufferSource")
	s.Set("buffer", buf)
	s.Call("connect", p.context.Get("destination"))
	s.Call("start", p.nextPos)
	p.nextPos += buf.Get("duration").Float()

	p.tmp = p.tmp[le:]
	return n, nil
}

func (p *driver) Close() error {
	for event, f := range p.callbacks {
		// https://developer.mozilla.org/en-US/docs/Web/API/EventTarget/removeEventListener
		// "Calling removeEventListener() with arguments that do not identify any currently registered EventListener on the EventTarget has no effect."
		js.Global().Get("document").Call("removeEventListener", event, f)
		f.Release()
	}
	p.callbacks = nil
	return nil
}

func (d *driver) tryWriteCanReturnWithoutWaiting() bool {
	return true
}
