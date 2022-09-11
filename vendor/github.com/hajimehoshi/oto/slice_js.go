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

//go:build go1.13
// +build go1.13

package oto

import (
	"reflect"
	"runtime"
	"syscall/js"
	"unsafe"
)

func float32SliceToTypedArray(s []float32) (js.Value, func()) {
	h := (*reflect.SliceHeader)(unsafe.Pointer(&s))
	h.Len *= 4
	h.Cap *= 4
	bs := *(*[]byte)(unsafe.Pointer(h))

	a := js.Global().Get("Uint8Array").New(len(bs))
	js.CopyBytesToJS(a, bs)
	runtime.KeepAlive(s)
	buf := a.Get("buffer")
	return js.Global().Get("Float32Array").New(buf, a.Get("byteOffset"), a.Get("byteLength").Int()/4), func() {}
}

func copyFloat32sToJS(v js.Value, s []float32) {
	h := (*reflect.SliceHeader)(unsafe.Pointer(&s))
	h.Len *= 4
	h.Cap *= 4
	bs := *(*[]byte)(unsafe.Pointer(h))

	js.CopyBytesToJS(v, bs)
	runtime.KeepAlive(s)
}
