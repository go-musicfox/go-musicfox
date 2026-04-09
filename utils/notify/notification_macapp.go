//go:build macapp

package notify

import (
	"fmt"
	"hash/fnv"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/macdriver/cocoa"
	"github.com/go-musicfox/go-musicfox/internal/macdriver/core"
)

type NotifyContent struct {
	Title   string
	Text    string
	Url     string
	Icon    string
	GroupId string
}

var notificationDelegate = cocoa.NewUserNotificationDelegate()

func Notify(content NotifyContent) {
	if !configs.AppConfig.Main.Notification.Enable {
		return
	}

	center := cocoa.NSUserNotificationCenter_defaultCenter()
	center.SetDelegate(notificationDelegate.ID())

	notification := cocoa.NewNSUserNotification()

	notification.SetTitle(content.Title)
	notification.SetInformativeText(content.Text)
	notification.SetSoundNameNil()

	if content.Icon != "" {
		localPath, err := downloadIcon(content.Icon)
		if err == nil && localPath != "" {
			notification.SetContentImage(localPath)
		}
	}

	if content.Url != "" {
		userInfo := core.NSMutableDictionary_initWithCapacity(1)
		userInfo.SetValueForKey(core.String("openUrl"), core.NSObject{ID: core.String(content.Url).ID})
		notification.SetUserInfo(userInfo.ID)
	}

	center.DeliverNotification(notification)
}

func downloadIcon(url string) (string, error) {
	url = strings.ReplaceAll(url, "http://", "https://")
	if !strings.Contains(url, "https://") {
		return "", fmt.Errorf("invalid icon URL")
	}

	h := fnvHash(url)
	iconFile := filepath.Join(os.TempDir(), "musicfox", fmt.Sprintf("%s.jpg", h))

	if _, err := os.Stat(iconFile); err == nil {
		return iconFile, nil
	}

	if err := os.MkdirAll(filepath.Dir(iconFile), 0o755); err != nil {
		return "", err
	}

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	file, err := os.Create(iconFile)
	if err != nil {
		return "", err
	}
	defer file.Close()

	if _, err := io.Copy(file, resp.Body); err != nil {
		return "", err
	}

	return iconFile, nil
}

func fnvHash(s string) string {
	h := fnv.New64a()
	h.Write([]byte(s))
	return fmt.Sprintf("%x", h.Sum64())
}
