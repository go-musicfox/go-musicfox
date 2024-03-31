// Copyright 2022 The Oto Authors
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

//go:build nintendosdk

// The actual implementaiton will be provided by -overlay.

#include <cstddef>

typedef void (*oto_OnReadCallbackType)(float *buf, size_t length);

extern "C" void oto_OpenAudio(int sample_rate, int channel_num,
                              oto_OnReadCallbackType on_read_callback,
                              int buffer_size_in_bytes) {}
