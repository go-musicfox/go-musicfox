package lastfm

import (
	"fmt"
	"log/slog"
	"net/url"

	"github.com/pkg/errors"
	lastfmgo "github.com/shkh/lastfm-go"
	"github.com/skratchdot/open-golang/open"

	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/storage"
	"github.com/go-musicfox/go-musicfox/internal/types"
	"github.com/go-musicfox/go-musicfox/utils/slogx"
)

type AuthInvalid struct {
	error
}

var available = true

// IsAvailable lastfm 功能可用性
func IsAvailable() bool {
	return available
}

type Client struct {
	api        *lastfmgo.Api
	user       *storage.LastfmUser
	Tracker    *Tracker
	needAuth   bool
	apiAccount *storage.LastfmApiAccount
}

func NewClient() *Client {
	client := &Client{
		user:       &storage.LastfmUser{},
		needAuth:   true,
		apiAccount: &storage.LastfmApiAccount{},
	}
	client.apiAccount.InitFromStorage()
	client.Tracker = NewTracker(client)

	key, secret := client.getAPIKey()
	if IsAvailable() {
		client.api = lastfmgo.New(key, secret)
		client.user.InitFromStorage() // 获取lastfm用户信息
		if client.user.ApiKey == key && client.user.SessionKey != "" {
			client.SetSession(client.user.SessionKey)
			client.needAuth = false
		}
	} else {
		err := errors.New("lastfm 当前不可用，请检查 lastfm key 或 secret")
		_, _ = client.errorHandle(err)
	}
	return client
}

func (c *Client) getAPIKey() (key, secret string) {
	switch {
	case c.apiAccount.Key != "" && c.apiAccount.Secret != "":
		return c.apiAccount.Key, c.apiAccount.Secret
	case configs.ConfigRegistry.Lastfm.Key != "" && configs.ConfigRegistry.Lastfm.Secret != "":
		return configs.ConfigRegistry.Lastfm.Key, configs.ConfigRegistry.Lastfm.Secret
	case types.LastfmKey != "" && types.LastfmSecret != "":
		return types.LastfmKey, types.LastfmSecret
	default:
		available = false
		return
	}
}

func (c *Client) errorHandle(e error) (bool, error) {
	if e == nil {
		return false, nil
	}
	var lastfmErr *lastfmgo.LastfmError
	if errors.As(e, &lastfmErr) {
		switch lastfmErr.Code {
		case 9: // invalid session key
			c.needAuth = true
			return true, AuthInvalid{lastfmErr}
		case 11, 16: // server error
			return true, e
		case 10, 26: // API key error
			available = false
			return true, e
		default:
			slog.Error("Lastfm request failed", slogx.Error(lastfmErr))
			return false, e
		}
	}
	var networkErr *url.Error
	if errors.As(e, &networkErr) {
		return true, e
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

	key, _ := c.getAPIKey()
	url = fmt.Sprintf(types.LastfmAuthUrl, key, token)
	return
}

func (c *Client) Login(username, password string) (sessionKey string, err error) {
	if c.api == nil {
		return "", errors.New("lastfm key或secret为空")
	}

	if err := c.api.Login(username, password); err != nil {
		_, err = c.errorHandle(err)
		return "", err
	}
	return c.api.GetSessionKey(), nil
}

func (c *Client) SetSession(session string) {
	if c.api != nil {
		c.needAuth = false
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

func (c *Client) Close() {
	c.Tracker.close()
}

func (c *Client) NeedAuth() bool {
	if c.needAuth {
		return true
	}
	if c.user.SessionKey == "" {
		return true
	}
	return false
}

func (c *Client) OpenUserHomePage() {
	url := "https://www.last.fm"
	if c.user.Url != "" {
		url = c.user.Url
	}
	_ = open.Start(url)
}

func (c *Client) InitUserInfo(user *storage.LastfmUser) {
	c.needAuth = false
	key, _ := c.getAPIKey()
	user.ApiKey = key
	c.user = user
	c.user.Store()
}

func (c *Client) ClearUserInfo() {
	c.user = &storage.LastfmUser{}
	c.user.Clear()
}

func (c *Client) UserName() string {
	if c.user != nil {
		return c.user.Name
	}
	return ""
}

func (c *Client) SetApiAccount(key, secret string) {
	c.apiAccount = &storage.LastfmApiAccount{Key: key, Secret: secret}
	c.apiAccount.Store()
	c.api = lastfmgo.New(key, secret)
	available = true // 更新状态
}

func (c *Client) GetApiAccount() (key, secret string) {
	return c.apiAccount.Key, c.apiAccount.Secret
}

func (c *Client) ClearApiAccount() {
	c.apiAccount = &storage.LastfmApiAccount{}
	c.apiAccount.Clear()
	_, _ = c.getAPIKey() // 刷新状态
}
