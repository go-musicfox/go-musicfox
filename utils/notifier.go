package utils

import (
	"os"
	"runtime"

	"go-musicfox/pkg/constants"

	"github.com/anhoder/notificator"
	"go-musicfox/pkg/configs"
)

type NotifyContent struct {
	Title  string
	Text   string
	Url    string
	Icon   string
	Sender string
}

func Notify(content NotifyContent) {
	if !configs.ConfigRegistry.MainShowNotify {
		return
	}

	if content.Url != "" {
		content.Sender = ""
	} else if content.Sender == "" {
		content.Sender = configs.ConfigRegistry.MainNotifySender
	}

	notify := notificator.New(notificator.Options{
		AppName:   constants.AppName,
		OSXSender: content.Sender,
	})

	if content.Icon == "" || runtime.GOOS != "darwin" {
		localDir := GetLocalDataDir()
		content.Icon = localDir + "/logo.png"
		if _, err := os.Stat(content.Icon); os.IsNotExist(err) {
			// 写入logo文件
			logoContent, _ := embedDir.ReadFile("embed/logo.png")
			_ = os.WriteFile(content.Icon, logoContent, 0644)
		}
	}

	_ = notify.Push(notificator.UrNormal, content.Title, content.Text, content.Icon, content.Url)
}
