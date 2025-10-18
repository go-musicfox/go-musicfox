package notify

import (
	"fmt"
	"hash/fnv"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/adrg/xdg"
	"github.com/gen2brain/beeep"
	"github.com/go-musicfox/notificator"
	"github.com/godbus/dbus/v5"
	"github.com/pkg/errors"

	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/types"
	"github.com/go-musicfox/go-musicfox/utils/app"
	"github.com/go-musicfox/go-musicfox/utils/filex"
)

type osxNotificator struct {
	appName     string
	defaultIcon string
}

func (o osxNotificator) getNotifierCmd() string {
	localDir := app.DataDir()
	notifierPath := filepath.Join(localDir, "musicfox-notifier.app")
	if _, err := os.Stat(notifierPath); os.IsNotExist(err) {
		err = filex.CopyDirFromEmbed("embed_darwin/musicfox-notifier.app", notifierPath)
		if err != nil {
			log.Printf("copy musicfox-notifier.app failed, err: %+v", errors.WithStack(err))
		}
	} else if err != nil {
		log.Printf("musicfox-notifier.app status err: %+v", errors.WithStack(err))
	}

	return notifierPath + "/Contents/MacOS/musicfox-notifier"
}

func (o osxNotificator) push(title, text, iconPath, redirectUrl, groupId string) *exec.Cmd {
	cmdPath := o.getNotifierCmd()
	if _, err := os.Stat(cmdPath); err == nil {
		args := []string{"-title", o.appName, "-message", text, "-subtitle", title, "-contentImage", iconPath}
		if redirectUrl != "" {
			args = append(args, "-open", redirectUrl)
		}
		if groupId != "" {
			args = append(args, "-group", groupId)
		}
		return exec.Command(cmdPath, args...)
	} else if notificator.CheckMacOSVersion() {
		title = strings.ReplaceAll(title, `"`, `\"`)
		text = strings.ReplaceAll(text, `"`, `\"`)

		notification := fmt.Sprintf("display notification \"%s\" with title \"%s\" subtitle \"%s\"", text, o.appName, title)
		return exec.Command("osascript", "-e", notification)
	}

	return exec.Command("growlnotify", "-n", o.appName, "--image", iconPath, "-m", title, "--url", redirectUrl)
}

func (o osxNotificator) pushCritical(title, text, iconPath, redirectUrl, groupId string) *exec.Cmd {
	cmdPath := o.getNotifierCmd()
	if _, err := os.Stat(cmdPath); err == nil {
		args := []string{"-title", o.appName, "-message", text, "-subtitle", title, "-contentImage", iconPath}
		if redirectUrl != "" {
			args = append(args, "-open", redirectUrl)
		}
		if groupId != "" {
			args = append(args, "-group", groupId)
		}
		return exec.Command(cmdPath, args...)
	} else if notificator.CheckMacOSVersion() {
		notification := fmt.Sprintf("display notification \"%s\" with title \"%s\" subtitle \"%s\"", text, o.appName, title)
		return exec.Command("osascript", "-e", notification)
	}

	return exec.Command("growlnotify", "-n", o.appName, "--image", iconPath, "-m", title, "--url", redirectUrl)
}

type Notificator struct {
	osx *osxNotificator
	*notificator.Notificator
}

func NewNotificator(o notificator.Options) *Notificator {
	n := &Notificator{
		Notificator: notificator.New(o),
	}
	if runtime.GOOS == "darwin" {
		n.osx = &osxNotificator{appName: o.AppName, defaultIcon: o.DefaultIcon}
	}
	return n
}

func (n Notificator) Push(urgency, title, text, iconPath, redirectUrl, groupId string) error {
	if runtime.GOOS == "darwin" {
		icon := n.osx.defaultIcon
		if iconPath != "" {
			icon = iconPath
		}
		if urgency == notificator.UrCritical {
			return n.osx.pushCritical(title, text, icon, redirectUrl, groupId).Run()
		}
		return n.osx.push(title, text, icon, redirectUrl, groupId).Run()
	}
	if runtime.GOOS == "windows" {
		return beeep.Notify(title, text, iconPath)
	}
	return n.Notificator.Push(urgency, title, text, iconPath, redirectUrl)
}

type NotifyContent struct {
	Title   string
	Text    string
	Url     string
	Icon    string
	GroupId string
}

func Notify(content NotifyContent) {
	if !configs.AppConfig.Main.Notification.Enable {
		return
	}

	notify := NewNotificator(notificator.Options{
		AppName: types.AppName,
	})

	if runtime.GOOS != "darwin" {
		// On non-macOS operating systems, only one image is allowed per notification.
		// The user can choose to send an album cover in notifications or use the default musicfox icon.
		if configs.AppConfig.Main.Notification.AlbumCover {
			err := useAlbumCover(&content)
			// fall back to default icon
			if err != nil {
				useAppIcon(&content)
			}
		} else {
			useAppIcon(&content)
		}
	}
	if runtime.GOOS == "linux" {
		freedesktop_dbus, err := dbus.ConnectSessionBus()
		if err != nil {
			log.Printf("connect dbus failed: %+v", errors.WithStack(err))
			return
		}
		defer freedesktop_dbus.Close()
		notfiy := freedesktop_dbus.Object("org.freedesktop.Notifications", "/org/freedesktop/Notifications")
		call := notfiy.Call("org.freedesktop.Notifications.Notify", 0, types.AppName, uint32(0),
			content.Icon, content.Title, content.Text, []string{},
			map[string]dbus.Variant{}, int32(5000))
		//Notify spec  https://specifications.freedesktop.org/notification-spec/1.3/protocol.html#command-notify
		if call.Err != nil {
			log.Printf("connect org.freedesktop.Notifications failed: %+v", errors.WithStack(call.Err))
		}
		return
	}
	_ = notify.Push(notificator.UrNormal, content.Title, content.Text, content.Icon, content.Url, content.GroupId)
}

func useAlbumCover(content *NotifyContent) error {
	iconUrl := content.Icon
	localFile, err := getLocalFileName(iconUrl)
	if err != nil {
		return err
	}
	if _, err := os.Stat(localFile); err == nil {
		content.Icon = localFile
		return nil
	} else if errors.Is(err, os.ErrNotExist) {
		response, e := http.Get(iconUrl)
		if e != nil {
			return e
		}
		defer response.Body.Close()
		file, err := os.Create(localFile)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(file, response.Body)
		if err != nil {
			return err
		}
		content.Icon = localFile
		return nil
	} else {
		return err
	}
}

func getLocalFileName(icon string) (string, error) {
	icon = strings.ReplaceAll(icon, "http://", "https://")
	if !strings.Contains(icon, "https://") {
		return "", errors.New("Icon does not seem to be a valid url: " + icon)
	}
	// Hash the url to get a file name
	h := fnv.New64a()
	_, err := h.Write([]byte(icon))
	if err != nil {
		return "", err
	}
	hashed := strconv.FormatUint(h.Sum64(), 36)
	fileName := hashed + ".jpg"
	// Use /tmp/musicfox to temporarily store the album cover
	localDir := filepath.Join(os.TempDir(), "musicfox")
	err = os.MkdirAll(localDir, 0777)
	if err != nil {
		return "", err
	}
	localFile := filepath.Join(localDir, fileName)
	return localFile, nil
}

func useAppIcon(content *NotifyContent) {
	// Flatpak notifications won't work if we use the default logo directory. Use this hard-coded one instead.
	// See https://github.com/flathub/io.github.go_musicfox.go-musicfox/issues/4
	if os.Getenv("container") == "flatpak" {
		content.Icon = "/app/share/icons/hicolor/512x512/apps/io.github.go_musicfox.go-musicfox.png"
		return
	}
	if runtime.GOOS == "linux" {
		content.Icon, _ = xdg.SearchDataFile("pixmaps/musicfox.png") //old systems,power by systems packing
		//app need in XDG_DATA_DIRS hicolor 16 24 32 48 64 96 128 192 256 512
		//icons/hicolor/scalable/apps can use  svg
		//see https://specifications.freedesktop.org/icon-theme-spec/latest/#directory_layout
		//and see https://gitlab.freedesktop.org/xdg/default-icon-theme
		sizeOne := [10]string{"16", "24", "32", "48", "64", "96", "128", "192", "256", "512"}
		for _, size := range sizeOne {
			sizexsize := size + "x" + size
			content.Icon, _ = xdg.SearchDataFile(filepath.Join("icons", "hicolor", sizexsize, "apps", "musicfox.png"))
			if content.Icon != "" {
				content.Icon = "musicfox"
				return
			}
		}
		if content.Icon == "" {
			content.Icon = filepath.Join(xdg.DataHome, "icons", "hicolor", "512x512", "apps", "musicfox.png") //user dir can write
		}
		if _, err := os.Stat(content.Icon); os.IsNotExist(err) {
			err := os.MkdirAll(filepath.Join(xdg.DataHome, "icons", "hicolor", "512x512", "apps"), 0777)
			if err != nil {
				log.Printf("\n filed to mkdir XDG_DATA_DIRS/icons/hicolor/512x512/apps  , err: %+v", errors.WithStack(err))
			}
			// 写入logo文件
			err = filex.CopyFileFromEmbed("embed/logo.png", content.Icon)
			if err != nil {
				log.Printf("copy logo.png to  failed, err: %+v", errors.WithStack(err))
			}
		}
		content.Icon = "musicfox"
		return
	}
	localDir := app.DataDir()
	content.Icon = filepath.Join(localDir, configs.AppConfig.Main.Notification.Icon)
	if _, err := os.Stat(content.Icon); os.IsNotExist(err) {
		if err != nil {
			content.Icon = filepath.Join(localDir, types.DefaultNotifyIcon)
		}
	}
	if _, err := os.Stat(content.Icon); os.IsNotExist(err) {
		// 写入logo文件
		err = filex.CopyFileFromEmbed("embed/logo.png", content.Icon)
		if err != nil {
			log.Printf("copy logo.png failed, err: %+v", errors.WithStack(err))
		}
	} else if err != nil {
		log.Printf("logo.png status err: %+v", errors.WithStack(err))
	}
}
