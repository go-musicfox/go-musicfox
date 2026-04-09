//go:build darwin

package main

import (
	"github.com/go-musicfox/go-musicfox/internal/commands"
	"github.com/gookit/gcli/v2"
)

func registerPlatformCommands(app *gcli.App) {
	app.Add(commands.NewNotifyCommand())
}
