package utils

import (
    "github.com/anhoder/notificator"
    "go-musicfox/config"
    "go-musicfox/constants"
    "io/ioutil"
    "os"
)

func Notify(title, text string) {
    if !config.ConfigRegistry.MainShowNotify {
        return
    }

    notify := notificator.New(notificator.Options{
        AppName: "musicfox",
    })

    localDir := GetLocalDataDir()
    iconPath := localDir + "/logo.png"
    if _, err := os.Stat(iconPath); os.IsNotExist(err) {

        // 写入logo文件
        logoContent, _ := static.ReadFile("static/logo.png")
        _ = ioutil.WriteFile(iconPath, logoContent, 0644)

    }

    _ = notify.Push(notificator.UrNormal, title, text, iconPath, constants.AppGithubUrl)
}
