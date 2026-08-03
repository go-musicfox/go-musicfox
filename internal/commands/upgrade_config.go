package commands

import (
	"fmt"

	"github.com/gookit/gcli/v2"

	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/utils/app"
)

// NewUpgradeConfigCommand appends missing built-in defaults to the user config.
func NewUpgradeConfigCommand() *gcli.Command {
	return &gcli.Command{
		Name:   "upgrade-config",
		UseFor: "Append missing built-in configuration items",
		Func:   runUpgradeConfig,
	}
}

func runUpgradeConfig(_ *gcli.Command, _ []string) error {
	path := app.ConfigFilePath()
	added, err := configs.UpgradeConfig(path)
	if err != nil {
		return fmt.Errorf("upgrade configuration %s: %w", path, err)
	}
	if added == 0 {
		fmt.Printf("Configuration is already up to date: %s\n", path)
		return nil
	}
	fmt.Printf("Added %d missing configuration item(s): %s\n", added, path)
	return nil
}
