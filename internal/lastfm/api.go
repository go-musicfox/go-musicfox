package lastfm

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/pkg/errors"
	lastfmgo "github.com/shkh/lastfm-go"

	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/internal/types"
	"github.com/go-musicfox/go-musicfox/utils/slogx"
)

type AuthInvalid struct {
	error
}

type Client struct {
	api *lastfmgo.Api
}

func NewClient() *Client {
	client := &Client{}
	if configs.ConfigRegistry.Lastfm.Key == "" || configs.ConfigRegistry.Lastfm.Secret == "" {
		err := errors.New("lastfm key或secret为空")
		_, _ = client.errorHandle(err)
	} else {
		client.api = lastfmgo.New(configs.ConfigRegistry.Lastfm.Key, configs.ConfigRegistry.Lastfm.Secret)
	}
	return client
}

func (c *Client) errorHandle(e error) (bool, error) {
	if e == nil {
		return false, nil
	}
	var lastfmErr *lastfmgo.LastfmError
	if errors.As(e, &lastfmErr) {
		switch lastfmErr.Code {
		case 9: // invalid session key
			return false, AuthInvalid{lastfmErr}
		case 11, 16: // server error
			return true, e
		default:
			slog.Error("Lastfm request failed", slogx.Error(lastfmErr))
			return false, e
		}
	}
	slog.Error("Lastfm other err", slogx.Error(e))
	return false, e
}

func (c *Client) GetAuthUrlWithToken() (token, url string, err error) {
	if c.api == nil {
		return "", "", errors.New("lastfm key或secret为空")
	}
	token, err = c.api.GetToken()
	if _, err = c.errorHandle(err); err != nil {
		return
	}

	url = fmt.Sprintf(types.LastfmAuthUrl, configs.ConfigRegistry.Lastfm.Key, token)
	return
}

func (c *Client) SetSession(session string) {
	if c.api != nil {
		c.api.SetSession(session)
	}
}

func (c *Client) GetSession(token string) (sessionKey string, err error) {
	if c.api == nil {
		return "", errors.New("lastfm key或secret为空")
	}
	err = c.api.LoginWithToken(token)
	if _, err = c.errorHandle(err); err != nil {
		return
	}
	sessionKey = c.api.GetSessionKey()
	return
}

func (c *Client) UpdateNowPlaying(args map[string]any) error {
	if c.api == nil {
		return errors.New("lastfm key或secret为空")
	}
	if c.api.GetSessionKey() == "" {
		_, err := c.errorHandle(errors.New("empty session key"))
		return err
	}
	_, err := c.api.Track.UpdateNowPlaying(args)

	var retry bool
	if retry, err = c.errorHandle(err); retry {
		return c.UpdateNowPlaying(args)
	}
	return err
}

func (c *Client) Scrobble(args map[string]any) error {
	if c.api == nil {
		return errors.New("lastfm key或secret为空")
	}
	if c.api.GetSessionKey() == "" {
		_, err := c.errorHandle(errors.New("empty session key"))
		return err
	}
	_, err := c.api.Track.Scrobble(args)

	var retry bool
	if retry, err = c.errorHandle(err); retry {
		return c.Scrobble(args)
	}
	return err
}

func (c *Client) GetUserInfo(args map[string]any) (lastfmgo.UserGetInfo, error) {
	if c.api == nil {
		return lastfmgo.UserGetInfo{}, errors.New("lastfm key或secret为空")
	}
	if c.api.GetSessionKey() == "" {
		_, err := c.errorHandle(errors.New("empty session key"))
		return lastfmgo.UserGetInfo{}, err
	}
	userInfo, err := c.api.User.GetInfo(args)

	var retry bool
	if retry, err = c.errorHandle(err); retry {
		return c.GetUserInfo(args)
	}
	return userInfo, err
}

type ReportPhase uint8

const (
	ReportPhaseStart ReportPhase = iota
	ReportPhaseComplete
)

func Report(client *Client, phase ReportPhase, song structs.Song, passedTime time.Duration) {
	switch phase {
	case ReportPhaseStart:
		go func(song structs.Song) {
			_ = client.UpdateNowPlaying(map[string]any{
				"artist":   song.ArtistName(),
				"track":    song.Name,
				"album":    song.Album.Name,
				"duration": song.Duration,
			})
		}(song)
	case ReportPhaseComplete:
		duration := song.Duration.Seconds()
		passedSeconds := passedTime.Seconds()
		if passedSeconds >= duration/2 {
			go func(song structs.Song, passed time.Duration) {
				_ = client.Scrobble(map[string]any{
					"artist":    song.ArtistName(),
					"track":     song.Name,
					"album":     song.Album.Name,
					"timestamp": time.Now().Unix(),
					"duration":  song.Duration.Seconds(),
				})
			}(song, passedTime)
		}
	}
}
