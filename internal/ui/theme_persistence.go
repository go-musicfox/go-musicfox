package ui

import (
	"log/slog"

	"github.com/go-musicfox/go-musicfox/internal/configs"
	mfoxapp "github.com/go-musicfox/go-musicfox/utils/app"
)

func writeActiveTheme(path, name string) error {
	return configs.SetTOMLValue(path, []string{"theme", "activeTheme"}, name)
}

func (n *Netease) saveActiveTheme(name string) {
	if name == "" {
		return
	}
	if err := writeActiveTheme(mfoxapp.ConfigFilePath(), name); err != nil {
		slog.Error("Failed to persist active theme", "theme", name, "err", err)
		return
	}
	if configs.AppConfig != nil {
		configs.AppConfig.Theme.ActiveTheme = name
	}
}
