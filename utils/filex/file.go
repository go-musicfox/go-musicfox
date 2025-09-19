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
func LoadConfig() {
	configFile := app.ConfigFilePath()

	var cfg *configs.Config
	var err error
	cfg, err = configs.NewConfigFromTomlFile(configFile)
	if err != nil {
		panic(fmt.Sprintf("fatal: failed to load configuration: %v", err))
	}
	configs.AppConfig = cfg
}

// LoadIniConfig 加载ini配置信息
func LoadIniConfig() {
	configFile := app.ConfigFilePath()
	if !FileOrDirExists(configFile) {
		_ = CopyFileFromEmbed("embed/go-musicfox.ini", configFile)
	}
	configs.ConfigRegistry = configs.NewRegistryFromIniFile(configFile)
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
