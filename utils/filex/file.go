package filex

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/utils/app"
)

// LoadConfig 加载配置
// 它会自动处理新旧配置文件的选择、创建和加载。
func LoadConfig() {
	configDir := app.ConfigDir()

	resolved := configs.ResolveConfigFile(configDir)

	if !resolved.Exists {
		_ = CopyFileFromEmbed("embed/go-musicfox.toml", resolved.Path)
	}

	var cfg *configs.Config
	var err error
	if resolved.Format == configs.FormatTOML {
		cfg, err = configs.NewConfigFromTomlFile(resolved.Path)
		if err != nil {
			panic(fmt.Sprintf("fatal: failed to load configuration: %v", err))
		}
	} else {
		registry := configs.NewRegistryFromIniFile(resolved.Path)
		cfg = configs.MigrateLegacyRegistry(registry)
	}
	configs.AppConfig = cfg
}

func FileOrDirExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

func CopyFileFromEmbed(src, dst string) error {
	var (
		err   error
		srcfd fs.File
		dstfd *os.File
	)

	if srcfd, err = embedDir.Open(src); err != nil {
		return err
	}
	defer srcfd.Close()

	if dstfd, err = os.OpenFile(dst, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0766); err != nil {
		return err
	}
	defer dstfd.Close()

	if _, err = io.Copy(dstfd, srcfd); err != nil {
		return err
	}
	return nil
}

func CopyDirFromEmbed(src, dst string) error {
	var (
		err error
		fds []fs.DirEntry
	)

	if err = os.MkdirAll(dst, 0766); err != nil {
		return err
	}
	if fds, err = embedDir.ReadDir(src); err != nil {
		return err
	}
	for _, fd := range fds {
		srcfp := filepath.Join(src, fd.Name())
		dstfp := filepath.Join(dst, fd.Name())

		if fd.IsDir() {
			if err = CopyDirFromEmbed(srcfp, dstfp); err != nil {
				return err
			}
		} else {
			if err = CopyFileFromEmbed(srcfp, dstfp); err != nil {
				return err
			}
		}
	}
	return nil
}

func FileURL(filepath string) string {
	return "file://" + filepath
}
