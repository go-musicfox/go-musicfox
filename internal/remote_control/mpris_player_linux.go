//go:build linux

package remote_control

import (
	"fmt"
	"log"
	"math"
	"time"

	"github.com/go-musicfox/go-musicfox/internal/types"
	"github.com/go-musicfox/go-musicfox/utils/app"
	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/prop"
	"github.com/pkg/errors"
)

// Player is a DBus object satisfying the `org.mpris.MediaPlayer2.Player` interface.
// https://specifications.freedesktop.org/mpris-spec/latest/Player_Interface.html
type Player struct {
	*RemoteControl

	props map[string]*prop.Prop
}

// PlaybackStatus is a playback state.
// https://specifications.freedesktop.org/mpris-spec/latest/Player_Interface.html#Enum:Playback_Status
type PlaybackStatus string

// Defined PlaybackStatuses.
const (
	PlaybackStatusPlaying PlaybackStatus = "Playing"
	PlaybackStatusPaused  PlaybackStatus = "Paused"
	PlaybackStatusStopped PlaybackStatus = "Stopped"
)

func PlaybackStatusFromPlayer(state types.State) (PlaybackStatus, error) {
	switch state {
	case types.Playing:
		return PlaybackStatusPlaying, nil
	case types.Paused:
		return PlaybackStatusPaused, nil
	case types.Stopped:
		return PlaybackStatusStopped, nil
	}
	return "", errors.Errorf("unknown playback status: %d", state)
}

// TimeInUs is time in microseconds.
// https://specifications.freedesktop.org/mpris-spec/latest/Player_Interface.html#Simple-Type:Time_In_Us
type TimeInUs int64

// UsFromDuration returns the type from a time.Duration
func UsFromDuration(t time.Duration) TimeInUs {
	return TimeInUs(t / time.Microsecond)
}

// Duration returns the type in time.Duration
func (t TimeInUs) Duration() time.Duration { return time.Duration(t) * time.Microsecond }

// ============================================================================

func notImplemented(c *prop.Change) *dbus.Error {
	return dbus.MakeFailedError(errors.New("Not implemented"))
}

// OnVolume handles volume changes.
func (p *Player) OnVolume(c *prop.Change) *dbus.Error {
	val := int(math.Round(c.Value.(float64) * 100))

	p.RemoteControl.player.CtrlSetVolume(val)
	return nil
}

func (p *Player) createStatus(info PlayingInfo) {
	playStatus, _ := PlaybackStatusFromPlayer(info.State)
	volume := math.Max(0, float64(info.Volume)/100.0)

	p.props = map[string]*prop.Prop{
		"PlaybackStatus": newProp(playStatus, nil),
		"LoopStatus":     newProp("None", nil),
		"Rate":           newProp(1.0, nil),
		"Shuffle":        newProp(false, nil),
		"Metadata":       newProp(MapFromPlayingInfo(info), nil),
		"Volume":         newProp(volume, p.OnVolume),
		"Position": {
			Value:    UsFromDuration(info.PassedDuration),
			Writable: true,
			Emit:     prop.EmitFalse,
			Callback: nil,
		},
		"MinimumRate":   newProp(1.0, nil),
		"MaximumRate":   newProp(1.0, nil),
		"CanGoNext":     newProp(true, nil),
		"CanGoPrevious": newProp(true, nil),
		"CanPlay":       newProp(true, nil),
		"CanPause":      newProp(true, nil),
		"CanSeek":       newProp(false, nil),
		"CanControl":    newProp(true, nil),
	}
}

// ============================================================================

// Next skips to the next track in the tracklist.
// https://specifications.freedesktop.org/mpris-spec/latest/Player_Interface.html#Method:Next
func (p *Player) Next() *dbus.Error {
	log.Printf("Next requested\n")
	p.RemoteControl.player.CtrlNext()
	return nil
}

// Previous skips to the previous track in the tracklist.
// https://specifications.freedesktop.org/mpris-spec/latest/Player_Interface.html#Method:Previous
func (p *Player) Previous() *dbus.Error {
	log.Printf("Previous requested\n")
	p.RemoteControl.player.CtrlPrevious()
	return nil
}

// Pause pauses playback.
// https://specifications.freedesktop.org/mpris-spec/latest/Player_Interface.html#Method:Pause
func (p *Player) Pause() *dbus.Error {
	log.Printf("Pause requested\n")
	p.RemoteControl.player.CtrlPause()
	return nil
}

// Play starts or resumes playback.
// https://specifications.freedesktop.org/mpris-spec/latest/Player_Interface.html#Method:Play
func (p *Player) Play() *dbus.Error {
	log.Printf("Play requested\n")
	p.RemoteControl.player.CtrlResume()
	return nil
}

// Stop stops playback.
// https://specifications.freedesktop.org/mpris-spec/latest/Player_Interface.html#Method:Stop
func (p *Player) Stop() *dbus.Error {
	log.Printf("Stop requested\n")
	p.RemoteControl.player.CtrlPause() // 只暂停
	return nil
}

// PlayPause toggles playback.
// If playback is already paused, resumes playback.
// If playback is stopped, starts playback.
// https://specifications.freedesktop.org/mpris-spec/latest/Player_Interface.html#Method:PlayPause
func (p *Player) PlayPause() *dbus.Error {
	log.Printf("Play/Pause requested. Switching context...\n")
	p.RemoteControl.player.CtrlToggle()
	return nil
}

type MetadataMap map[string]any

func (m *MetadataMap) nonEmptyString(field, value string) {
	if value != "" {
		(*m)[field] = value
	}
}

func (m *MetadataMap) nonEmptySlice(field string, values []string) {
	var toAdd []string
	for _, v := range values {
		if v != "" {
			toAdd = append(toAdd, v)
		}
	}
	if len(toAdd) > 0 {
		(*m)[field] = toAdd
	}
}

func MapFromPlayingInfo(info PlayingInfo) MetadataMap {
	if info.TrackID == 0 {
		// No song
		return MetadataMap{
			"mpris:trackid": dbus.ObjectPath("/org/mpris/MediaPlayer2/TrackList/NoTrack"),
		}
	}

	m := &MetadataMap{
		"mpris:trackid": dbus.ObjectPath(fmt.Sprintf("/org/mpd/Tracks/%d", info.TrackID)),
		"mpris:length":  info.TotalDuration / time.Microsecond,
		"mpris:artUrl":  app.AddResizeParamForPicUrl(info.PicUrl, 1024),
	}

	m.nonEmptyString("xesam:album", info.Album)
	m.nonEmptyString("xesam:title", info.Name)
	m.nonEmptyString("xesam:asText", info.LRCText)
	m.nonEmptySlice("xesam:albumArtist", []string{info.AlbumArtist})
	m.nonEmptySlice("xesam:artist", []string{info.Artist})

	return *m
}
