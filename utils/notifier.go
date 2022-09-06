package utils

import (
    "os"

    "github.com/anhoder/notificator"
    "go-musicfox/configs"
)

func Notify(title, text, url string) {
    if !configs.ConfigRegistry.MainShowNotify {
        return
    }

    notify := notificator.New(notificator.Options{
        AppName: "musicfox",
        OSXSender: "com.netease.163music",
    })

    localDir := GetLocalDataDir()
    iconPath := localDir + "/logo.png"
    if _, err := os.Stat(iconPath); os.IsNotExist(err) {

        // 写入logo文件
        logoContent, _ := embedDir.ReadFile("embed/logo.png")
        _ = os.WriteFile(iconPath, logoContent, 0644)

    }

    _ = notify.Push(notificator.UrNormal, title, text, iconPath, url)
}
