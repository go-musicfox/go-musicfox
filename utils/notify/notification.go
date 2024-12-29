package notify

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/go-musicfox/notificator"
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
	localDir := app.DataRootDir()
	notifierPath := filepath.Join(localDir, "musicfox-notifier.app")
	if _, err := os.Stat(notifierPath); os.IsNotExist(err) {
		err = filex.CopyDirFromEmbed("embed/musicfox-notifier.app", notifierPath)
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
	if !configs.ConfigRegistry.Main.ShowNotify {
		return
	}

	notify := NewNotificator(notificator.Options{
		AppName: types.AppName,
	})

	// Hard code flatpak icon directory
	if os.Getenv("container") == "flatpak" {
		content.Icon = "/app/share/icons/hicolor/512x512/apps/io.github.go_musicfox.go-musicfox.png"
	} else if runtime.GOOS != "darwin" {
		localDir := app.DataRootDir()
		content.Icon = filepath.Join(localDir, configs.ConfigRegistry.Main.NotifyIcon)
		if _, err := os.Stat(content.Icon); os.IsNotExist(err) {
			content.Icon = filepath.Join(localDir, types.DefaultNotifyIcon)
			// 写入logo文件
			err = filex.CopyFileFromEmbed("embed/logo.png", content.Icon)
			if err != nil {
				log.Printf("copy logo.png failed, err: %+v", errors.WithStack(err))
			}
		} else if err != nil {
			log.Printf("logo.png status err: %+v", errors.WithStack(err))
		}
	}

	_ = notify.Push(notificator.UrNormal, content.Title, content.Text, content.Icon, content.Url, content.GroupId)
}
