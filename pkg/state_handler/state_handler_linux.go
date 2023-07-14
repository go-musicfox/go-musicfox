//go:build linux

package state_handler

import (
	"fmt"
	"log"
	"math"
	"os"
	"sync"

	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/introspect"
	"github.com/godbus/dbus/v5/prop"
	"github.com/pkg/errors"
)

type MediaPlayer2 struct {
	*Handler
}

func (m *MediaPlayer2) properties() map[string]*prop.Prop {
	return map[string]*prop.Prop{
		"CanQuit":             newProp(true, nil),       // https://specifications.freedesktop.org/mpris-spec/latest/Media_Player.html#Property:CanQuit
		"CanRaise":            newProp(false, nil),      // https://specifications.freedesktop.org/mpris-spec/latest/Media_Player.html#Property:CanRaise
		"HasTrackList":        newProp(true, nil),       // https://specifications.freedesktop.org/mpris-spec/latest/Media_Player.html#Property:HasTrackList
		"Identity":            newProp(m.name, nil),     // https://specifications.freedesktop.org/mpris-spec/latest/Media_Player.html#Property:Identity
		"SupportedUriSchemes": newProp([]string{}, nil), // https://specifications.freedesktop.org/mpris-spec/latest/Media_Player.html#Property:SupportedUriSchemes
		"SupportedMimeTypes":  newProp([]string{}, nil), // https://specifications.freedesktop.org/mpris-spec/latest/Media_Player.html#Property:SupportedMimeTypes
	}
}

func (m *MediaPlayer2) Raise() *dbus.Error { return nil }

func (m *MediaPlayer2) Quit() *dbus.Error {
	m.player.CtrlPaused() // 只暂停
	return nil
}

type Handler struct {
	player Controller
	name   string
	dbus   *dbus.Conn
	props  *prop.Properties
	once   sync.Once
}

func NewHandler(p Controller, nowInfo PlayingInfo) *Handler {
	handler := &Handler{
		player: p,
		name:   fmt.Sprintf("org.mpris.MediaPlayer2.musicfox.instance%d", os.Getpid()),
	}

	var err error
	if handler.dbus, err = dbus.SessionBus(); err != nil {
		log.Default().Printf("[MPRIS] init dbus error: %+v", err)
		return handler
	}

	mp2 := &MediaPlayer2{Handler: handler}
	_ = handler.dbus.Export(mp2, "/org/mpris/MediaPlayer2", "org.mpris.MediaPlayer2")

	mprisPlayer := &Player{Handler: handler}
	mprisPlayer.createStatus(nowInfo)
	_ = handler.dbus.Export(mprisPlayer, "/org/mpris/MediaPlayer2", "org.mpris.MediaPlayer2.Player")

	_ = handler.dbus.Export(introspect.NewIntrospectable(handler.IntrospectNode()), "/org/mpris/MediaPlayer2", "org.freedesktop.DBus.Introspectable")

	handler.props, _ = prop.Export(handler.dbus, "/org/mpris/MediaPlayer2", map[string]map[string]*prop.Prop{
		"org.mpris.MediaPlayer2":        mp2.properties(),
		"org.mpris.MediaPlayer2.Player": mprisPlayer.props,
	})

	_, err = handler.dbus.RequestName(handler.name, dbus.NameFlagReplaceExisting)
	if err != nil {
		log.Default().Printf("[MPRIS] dbus request name error: %+v", err)
	}

	return handler
}

func (s *Handler) IntrospectNode() *introspect.Node {
	return &introspect.Node{
		Name: s.name,
		Interfaces: []introspect.Interface{
			introspect.IntrospectData,
			{
				Name: "org.mpris.MediaPlayer2",
				Properties: []introspect.Property{
					{
						Name:   "CanQuit",
						Type:   "b",
						Access: "read",
					},
					{
						Name:   "CanRaise",
						Type:   "b",
						Access: "read",
					},
					{
						Name:   "HasTrackList",
						Type:   "b",
						Access: "read",
					},
					{
						Name:   "Identity",
						Type:   "s",
						Access: "read",
					},
					{
						Name:   "SupportedUriSchemes",
						Type:   "as",
						Access: "read",
					},
					{
						Name:   "SupportedMimeTypes",
						Type:   "as",
						Access: "read",
					},
				},
				Methods: []introspect.Method{
					{
						Name: "Raise",
					},
					{
						Name: "Quit",
					},
				},
			},
			{
				Name: "org.mpris.MediaPlayer2.Player",
				Properties: []introspect.Property{
					{
						Name:   "PlaybackStatus",
						Type:   "s",
						Access: "read",
					},
					{
						Name:   "LoopStatus",
						Type:   "s",
						Access: "readwrite",
					},
					{
						Name:   "Rate",
						Type:   "d",
						Access: "readwrite",
					},
					{
						Name:   "Shuffle",
						Type:   "b",
						Access: "readwrite",
					},
					{
						Name:   "Metadata",
						Type:   "a{sv}",
						Access: "read",
					},
					{
						Name:   "Volume",
						Type:   "d",
						Access: "readwrite",
					},
					{
						Name:   "Position",
						Type:   "x",
						Access: "read",
					},
					{
						Name:   "MinimumRate",
						Type:   "d",
						Access: "read",
					},
					{
						Name:   "MaximumRate",
						Type:   "d",
						Access: "read",
					},
					{
						Name:   "CanGoNext",
						Type:   "b",
						Access: "read",
					},
					{
						Name:   "CanGoPrevious",
						Type:   "b",
						Access: "read",
					},
					{
						Name:   "CanPlay",
						Type:   "b",
						Access: "read",
					},
					{
						Name:   "CanSeek",
						Type:   "b",
						Access: "read",
					},
					{
						Name:   "CanControl",
						Type:   "b",
						Access: "read",
					},
				},
				Signals: []introspect.Signal{
					{
						Name: "Seeked",
						Args: []introspect.Arg{
							{
								Name: "Position",
								Type: "x",
							},
						},
					},
				},
				Methods: []introspect.Method{
					{
						Name: "Next",
					},
					{
						Name: "Previous",
					},
					{
						Name: "Pause",
					},
					{
						Name: "PlayPause",
					},
					{
						Name: "Stop",
					},
					{
						Name: "Play",
					},
					{
						Name: "Seek",
						Args: []introspect.Arg{
							{
								Name:      "Offset",
								Type:      "x",
								Direction: "in",
							},
						},
					},
					{
						Name: "SetPosition",
						Args: []introspect.Arg{
							{
								Name:      "TrackId",
								Type:      "o",
								Direction: "in",
							},
							{
								Name:      "Position",
								Type:      "x",
								Direction: "in",
							},
						},
					},
				},
			},
			// TODO: This interface is not fully implemented.
			// introspect.Interface{
			// 	Name: "org.mpris.MediaPlayer2.TrackList",

			// },
		},
	}
}

func (s *Handler) SetPlayingInfo(info PlayingInfo) {
	if s.props == nil {
		return
	}
	// Playback Status
	go func() {
		playbackStatus, err := PlaybackStatusFromPlayer(info.State)
		if err == nil {
			s.setProp("org.mpris.MediaPlayer2.Player", "PlaybackStatus", dbus.MakeVariant(playbackStatus))
		}

		// Current song metadata
		if info.TrackID != 0 {
			s.setProp("org.mpris.MediaPlayer2.Player", "Metadata", dbus.MakeVariant(MapFromPlayingInfo(info)))
		}

		// Volume
		newVolume := math.Max(0, float64(info.Volume)/100.0)
		s.setProp("org.mpris.MediaPlayer2.Player", "Volume", dbus.MakeVariant(newVolume))
	}()

}

func (s *Handler) setProp(iface, name string, value dbus.Variant) {
	if s.props == nil {
		return
	}
	if err := s.props.Set(iface, name, value); err != nil {
		log.Printf("Setting %s %s failed: %+v\n", iface, name, errors.WithStack(err))
	}
}

func (s *Handler) Release() {
	_ = s.dbus.Close()
}

func newProp(value interface{}, cb func(*prop.Change) *dbus.Error) *prop.Prop {
	return &prop.Prop{
		Value:    value,
		Writable: true,
		Emit:     prop.EmitTrue,
		Callback: cb,
	}
}
