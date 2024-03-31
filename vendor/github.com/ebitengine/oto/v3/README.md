# Oto (v3)

[![Go Reference](https://pkg.go.dev/badge/github.com/ebitengine/oto/v3.svg)](https://pkg.go.dev/github.com/ebitengine/oto/v3)
[![Build Status](https://github.com/ebitengine/oto/actions/workflows/test.yml/badge.svg)](https://github.com/ebitengine/oto/actions?query=workflow%3Atest)

A low-level library to play sound.

- [Oto (v3)](#oto-v3)
  - [Platforms](#platforms)
  - [Prerequisite](#prerequisite)
    - [macOS](#macos)
    - [iOS](#ios)
    - [Linux](#linux)
    - [FreeBSD, OpenBSD](#freebsd-openbsd)
  - [Usage](#usage)
    - [Playing sounds from memory](#playing-sounds-from-memory)
    - [Playing sounds by file streaming](#playing-sounds-by-file-streaming)
    - [Advanced usage](#advanced-usage)
  - [Crosscompiling](#crosscompiling)

## Platforms

- Windows (no Cgo required!)
- macOS (no Cgo required!)
- Linux
- FreeBSD
- OpenBSD
- Android
- iOS
- WebAssembly
- Nintendo Switch
- Xbox

## Prerequisite

On some platforms you will need a C/C++ compiler in your path that Go can use.

- iOS: On newer macOS versions type `clang` on your terminal and a dialog with installation instructions will appear if you don't have it
  - If you get an error with clang use xcode instead `xcode-select --install`
- Linux and other Unix systems: Should be installed by default, but if not try [GCC](https://gcc.gnu.org/) or [Clang](https://releases.llvm.org/download.html)

### macOS

Oto requires `AudioToolbox.framework`, but this is automatically linked.

### iOS

Oto requires these frameworks:

- `AVFoundation.framework`
- `AudioToolbox.framework`

Add them to "Linked Frameworks and Libraries" on your Xcode project.

### Linux

ALSA is required. On Ubuntu or Debian, run this command:

```sh
apt install libasound2-dev
```

On RedHat-based linux distributions, run:

```sh
dnf install alsa-lib-devel
```

In most cases this command must be run by root user or through `sudo` command.

### FreeBSD, OpenBSD

BSD systems are not tested well. If ALSA works, Oto should work.

## Usage

The two main components of Oto are a `Context` and `Players`. The context handles interactions with
the OS and audio drivers, and as such there can only be **one** context in your program.

From a context you can create any number of different players, where each player is given an `io.Reader` that
it reads bytes representing sounds from and plays.

Note that a single `io.Reader` must **not** be used by multiple players.

### Playing sounds from memory

The following is an example of loading and playing an MP3 file:

```go
package main

import (
    "time"
    "os"

    "github.com/ebitengine/oto/v3"
    "github.com/hajimehoshi/go-mp3"
)

func main() {
    // Read the mp3 file into memory
    fileBytes, err := os.ReadFile("./my-file.mp3")
    if err != nil {
        panic("reading my-file.mp3 failed: " + err.Error())
    }

    // Convert the pure bytes into a reader object that can be used with the mp3 decoder
    fileBytesReader := bytes.NewReader(fileBytes)

    // Decode file
    decodedMp3, err := mp3.NewDecoder(fileBytesReader)
    if err != nil {
        panic("mp3.NewDecoder failed: " + err.Error())
    }

    // Prepare an Oto context (this will use your default audio device) that will
    // play all our sounds. Its configuration can't be changed later.

    op := &oto.NewContextOptions{}

    // Usually 44100 or 48000. Other values might cause distortions in Oto
    op.SampleRate = 44100

    // Number of channels (aka locations) to play sounds from. Either 1 or 2.
    // 1 is mono sound, and 2 is stereo (most speakers are stereo). 
    op.ChannelCount = 2

    // Format of the source. go-mp3's format is signed 16bit integers.
    op.Format = oto.FormatSignedInt16LE

    // Remember that you should **not** create more than one context
    otoCtx, readyChan, err := oto.NewContext(op)
    if err != nil {
        panic("oto.NewContext failed: " + err.Error())
    }
    // It might take a bit for the hardware audio devices to be ready, so we wait on the channel.
    <-readyChan

    // Create a new 'player' that will handle our sound. Paused by default.
    player := otoCtx.NewPlayer(decodedMp3)
    
    // Play starts playing the sound and returns without waiting for it (Play() is async).
    player.Play()

    // We can wait for the sound to finish playing using something like this
    for player.IsPlaying() {
        time.Sleep(time.Millisecond)
    }

    // Now that the sound finished playing, we can restart from the beginning (or go to any location in the sound) using seek
    // newPos, err := player.(io.Seeker).Seek(0, io.SeekStart)
    // if err != nil{
    //     panic("player.Seek failed: " + err.Error())
    // }
    // println("Player is now at position:", newPos)
    // player.Play()

    // If you don't want the player/sound anymore simply close
    err = player.Close()
    if err != nil {
        panic("player.Close failed: " + err.Error())
    }
}
```

### Playing sounds by file streaming

The above example loads the entire file into memory and then plays it. This is great for smaller files
but might be an issue if you are playing a long song since it would take too much memory and too long to load.

In such cases you might want to stream the file. Luckily this is very simple, just use `os.Open`:

```go
package main

import (
    "bytes"
    "os"
    "time"

    "github.com/hajimehoshi/go-mp3"
    "github.com/hajimehoshi/oto/v3"
)

func main() {
    // Open the file for reading. Do NOT close before you finish playing!
    file, err := os.Open("./my-file.mp3")
    if err != nil {
        panic("opening my-file.mp3 failed: " + err.Error())
    }

    // Decode file. This process is done as the file plays so it won't
    // load the whole thing into memory.
    decodedMp3, err := mp3.NewDecoder(file)
    if err != nil {
        panic("mp3.NewDecoder failed: " + err.Error())
    }

    // Rest is the same...

    // Close file only after you finish playing
    file.Close()
}
```

The only thing to note about streaming is that the *file* object must be kept alive, otherwise
you might just play static.

To keep it alive not only must you be careful about when you close it, but you might need to keep a reference
to the original file object alive (by for example keeping it in a struct).

### Advanced usage

Players have their own internal audio data buffer, so while for example 200 bytes have been read from the `io.Reader` that
doesn't mean they were all played from the audio device.

Data is moved from io.Reader->internal buffer->audio device, and when the internal buffer moves data to the audio device
is not guaranteed, so there might be a small delay. The amount of data in the buffer can be retrieved
using `Player.UnplayedBufferSize()`.

The size of the underlying buffer of a player can also be set by type-asserting the player object:

```go
myPlayer.(oto.BufferSizeSetter).SetBufferSize(newBufferSize)
```

This works because players implement a `Player` interface and a `BufferSizeSetter` interface.

## Crosscompiling

Crosscompiling to macOS or Windows is as easy as setting `GOOS=darwin` or `GOOS=windows`, respectively.

To crosscompile for other platforms, make sure the libraries for the target architecture are installed, and set 
`CGO_ENABLED=1` as Go disables [Cgo](https://golang.org/cmd/cgo/#hdr-Using_cgo_with_the_go_command) on crosscompiles by default.
