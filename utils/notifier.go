package utils

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"

	"github.com/go-musicfox/go-musicfox/pkg/configs"
	"github.com/go-musicfox/go-musicfox/pkg/constants"

	"github.com/go-musicfox/notificator"
	"github.com/pkg/errors"
)

type osxNotificator struct {
	appName     string
	defaultIcon string
}

func (o osxNotificator) getNotifierCmd() string {
	localDir := GetLocalDataDir()
	notifierPath := path.Join(localDir, "musicfox-notifier.app")
	if _, err := os.Stat(notifierPath); os.IsNotExist(err) {
		err = CopyDirFromEmbed("embed/musicfox-notifier.app", notifierPath)
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
		var args = []string{"-title", o.appName, "-message", text, "-subtitle", title, "-contentImage", iconPath}
		if redirectUrl != "" {
			args = append(args, "-open", redirectUrl)
		}
		if groupId != "" {
			args = append(args, "-group", groupId)
		}
		return exec.Command(cmdPath, args...)
	} else if notificator.CheckMacOSVersion() {
		title = strings.Replace(title, `"`, `\"`, -1)
		text = strings.Replace(text, `"`, `\"`, -1)

		notification := fmt.Sprintf("display notification \"%s\" with title \"%s\" subtitle \"%s\"", text, o.appName, title)
		return exec.Command("osascript", "-e", notification)
	}

	return exec.Command("growlnotify", "-n", o.appName, "--image", iconPath, "-m", title, "--url", redirectUrl)
}

func (o osxNotificator) pushCritical(title, text, iconPath, redirectUrl, groupId string) *exec.Cmd {
	cmdPath := o.getNotifierCmd()
	if _, err := os.Stat(cmdPath); err == nil {
		var args = []string{"-title", o.appName, "-message", text, "-subtitle", title, "-contentImage", iconPath}
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
	if !configs.ConfigRegistry.MainShowNotify {
		return
	}

	notify := NewNotificator(notificator.Options{
		AppName: constants.AppName,
	})

	if runtime.GOOS != "darwin" {
		localDir := GetLocalDataDir()
		content.Icon = path.Join(localDir, configs.ConfigRegistry.MainNotifyIcon)
		if _, err := os.Stat(content.Icon); os.IsNotExist(err) {
			content.Icon = path.Join(localDir, constants.DefaultNotifyIcon)
			// 写入logo文件
			err = CopyFileFromEmbed("embed/logo.png", content.Icon)
			if err != nil {
				log.Printf("copy logo.png failed, err: %+v", errors.WithStack(err))
			}
		} else if err != nil {
			log.Printf("logo.png status err: %+v", errors.WithStack(err))
		}
	}

	_ = notify.Push(notificator.UrNormal, content.Title, content.Text, content.Icon, content.Url, content.GroupId)
}
