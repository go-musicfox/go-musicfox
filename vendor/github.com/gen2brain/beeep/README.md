## beeep
[![Build Status](https://github.com/gen2brain/beeep/actions/workflows/build.yml/badge.svg)](https://github.com/gen2brain/beeep/actions)
[![Go Reference](https://pkg.go.dev/badge/github.com/gen2brain/beeep.svg)](https://pkg.go.dev/github.com/gen2brain/beeep)
[![Go Report Card](https://goreportcard.com/badge/github.com/gen2brain/beeep?branch=master)](https://goreportcard.com/report/github.com/gen2brain/beeep) 

`beeep` provides a cross-platform library for sending desktop notifications, alerts and beeps.

### Installation

    go get -u github.com/gen2brain/beeep

### Build tags

* `nodbus` - disable `godbus/dbus` and use only `notify-send`

### Examples

```go
err := beeep.Beep(beeep.DefaultFreq, beeep.DefaultDuration)
if err != nil {
    panic(err)
}
```

```go
err := beeep.Notify("Title", "Message body", "assets/information.png")
if err != nil {
    panic(err)
}
```

```go
err := beeep.Alert("Title", "Message body", "assets/warning.png")
if err != nil {
    panic(err)
}
```
