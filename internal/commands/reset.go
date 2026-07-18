package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gookit/gcli/v2"

	"github.com/go-musicfox/go-musicfox/utils/app"
)

var resetOpts struct {
	withConfigFile bool
}

func NewResetCommand() *gcli.Command {
	cmd := &gcli.Command{
		Name:   "reset",
		UseFor: "Clear all cache files (not config by default)",
		Examples: "{$binName} {$cmd}               # Clear cache only\n" +
			"  {$binName} {$cmd} --with-config-file  # Clear cache and config file",
		Config: func(c *gcli.Command) {
			c.Flags.BoolOpt(&resetOpts.withConfigFile, "with-config-file", "", false, "Also delete the config file (config.toml)")
		},
		Func: runReset,
	}
	return cmd
}

func runReset(_ *gcli.Command, _ []string) error {
	dirs := map[string]string{
		"Cache":   app.CacheDir(),
		"Runtime": app.RuntimeDir(),
		"Log":     app.LogDir(),
	}

	removed := make(map[string]bool)

	// 收集需要清理的唯一路径
	uniqueDirs := make(map[string]string)
	for label, dir := range dirs {
		if _, exists := uniqueDirs[dir]; !exists {
			uniqueDirs[dir] = label
		}
	}

	// 清理缓存目录
	for dir, label := range uniqueDirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			fmt.Printf("  %-10s %s (does not exist, skip)\n", label+":", dir)
			continue
		}

		fmt.Printf("  %-10s %s\n", label+":", dir)
		if err := os.RemoveAll(dir); err != nil {
			return fmt.Errorf("failed to remove %s directory: %w", label, err)
		}

		// 重新创建目录，因为应用期望它们存在
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to recreate %s directory: %w", label, err)
		}

		removed[label] = true
	}

	fmt.Println()

	// 如果指定了 --with-config-file，也删除配置文件
	if resetOpts.withConfigFile {
		configPath := app.ConfigFilePath()
		fmt.Printf("  %-10s %s\n", "Config:", configPath)

		if _, err := os.Stat(configPath); err == nil {
			if err := os.Remove(configPath); err != nil {
				return fmt.Errorf("failed to remove config file: %w", err)
			}
			removed["Config"] = true
		} else {
			fmt.Printf("    (config file does not exist, skip)\n")
		}
	}

	// 报告清理结果
	if len(removed) == 0 {
		fmt.Println("Nothing to reset.")
	} else {
		var names []string
		for name := range removed {
			names = append(names, name)
		}
		fmt.Printf("Reset complete. Cleared: %v\n", names)
	}

	// 如果删除了配置文件，提示用户需要重新配置
	if removed["Config"] {
		configDir := app.ConfigDir()
		fmt.Printf("\nConfig file has been removed. The default config will be regenerated on next launch.\n")
		fmt.Printf("You can find it at: %s\n", filepath.Join(configDir, "config.toml"))
	}

	return nil
}
