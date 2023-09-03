package lastfm

import (
	"fmt"
	"time"

	"github.com/go-musicfox/go-musicfox/internal/constants"
	"github.com/go-musicfox/go-musicfox/internal/structs"
	"github.com/go-musicfox/go-musicfox/utils"

	"github.com/pkg/errors"
	lastfmgo "github.com/shkh/lastfm-go"
)

type AuthInvalid struct {
	error
}

type Client struct {
	api *lastfmgo.Api
}

func NewClient() *Client {
	client := &Client{}
	if constants.LastfmKey == "" || constants.LastfmSecret == "" {
		err := errors.New("lastfm key或secret为空")
		_, _ = client.errorHandle(err)
	} else {
		client.api = lastfmgo.New(constants.LastfmKey, constants.LastfmSecret)
	}
	return client
}

func (c *Client) errorHandle(e error) (bool, error) {
	if e == nil {
		return false, nil
	}
	if lastfmErr, ok := e.(*lastfmgo.LastfmError); ok {
		switch lastfmErr.Code {
		case 9: // invalid session key
			return false, AuthInvalid{lastfmErr}
		case 11, 16: // server error
			return true, e
		default:
			utils.Logger().Printf("[ERROR] Lastfm request err: %+v", lastfmErr)
			return false, e
		}
	}
	utils.Logger().Printf("[ERROR] Lastfm other err: %+v", e)
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

	url = fmt.Sprintf(constants.LastfmAuthUrl, constants.LastfmKey, token)
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

func (c *Client) UpdateNowPlaying(args map[string]interface{}) error {
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

func (c *Client) Scrobble(args map[string]interface{}) error {
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

func (c *Client) GetUserInfo(args map[string]interface{}) (lastfmgo.UserGetInfo, error) {
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
			_ = client.UpdateNowPlaying(map[string]interface{}{
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
				_ = client.Scrobble(map[string]interface{}{
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
