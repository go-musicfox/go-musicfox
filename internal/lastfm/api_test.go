package lastfm

import (
	"fmt"
	"testing"
	"time"
)

var client *Client

func TestMain(m *testing.M) {
	client = NewClient()
	m.Run()
}

func TestGetAuthUrlWithToken(t *testing.T) {
	token, url, err := client.GetAuthUrlWithToken()
	if err != nil {
		t.Fatal(err)
	}
	if url == "" || token == "" {
		t.Fatal("url or token is empty")
	}
	fmt.Println(url, token)
}

func TestGetSession(t *testing.T) {
	token := ""
	sessionKey, err := client.GetSession(token)
	fmt.Println(sessionKey, err)
}

func TestUpdateNowPlaying(t *testing.T) {
	client.SetSession("")
	err := client.UpdateNowPlaying(map[string]interface{}{
		"artist":   "薛之谦",
		"track":    "无数",
		"album":    "无数",
		"duration": 330,
	})
	fmt.Println(err)
}

func TestScrobble(t *testing.T) {
	client.SetSession("")
	err := client.Scrobble(map[string]interface{}{
		"artist":    "薛之谦",
		"track":     "无数",
		"album":     "无数",
		"timestamp": time.Now().Unix(),
		"duration":  330,
	})
	fmt.Println(err)
}

func TestGetUserInfo(t *testing.T) {
	client.SetSession("")
	userInfo, err := client.GetUserInfo(map[string]interface{}{})
	fmt.Println(userInfo, err)
}
