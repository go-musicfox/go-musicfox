package filex

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

func FileOrDirExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

func ReadFileFromEmbed(src string) ([]byte, error) {
	return embedDir.ReadFile(src)
}

func ReadDirFromEmbed(src string) ([]fs.DirEntry, error) {
	return fs.ReadDir(embedDir, src)
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

	// 0600：该文件（config.toml）会写入 neteaseCookie、lastfm key/secret 等
	// 凭据，0766 在默认 umask 下实际为 0744（group/other 可读），同机其他
	// 用户可读登录 cookie；与 bbolt 数据库（0600）保持一致
	if dstfd, err = os.OpenFile(dst, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600); err != nil {
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

	if err = os.MkdirAll(dst, 0700); err != nil {
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
