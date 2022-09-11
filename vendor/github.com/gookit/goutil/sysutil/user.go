package sysutil

import (
	"os"
	"os/user"

	"github.com/mitchellh/go-homedir"
)

// MustFindUser must find an system user by name
func MustFindUser(uname string) *user.User {
	u, err := user.Lookup(uname)
	if err != nil {
		panic(err)
	}
	return u
}

// LoginUser must get current user
func LoginUser() *user.User {
	return CurrentUser()
}

// CurrentUser must get current user
func CurrentUser() *user.User {
	// check $HOME/.terminfo
	u, err := user.Current()
	if err != nil {
		panic(err)
	}
	return u
}

// UserHomeDir is alias of os.UserHomeDir, but ignore error
func UserHomeDir() string {
	dir, _ := os.UserHomeDir()
	return dir
}

// UHomeDir get user home dir path.
func UHomeDir() string {
	// check $HOME/.terminfo
	u, err := user.Current()
	if err != nil {
		return ""
	}
	return u.HomeDir
}

// HomeDir get user home dir path.
func HomeDir() string {
	dir, _ := homedir.Dir()
	return dir
}

// UserDir will prepend user home dir to subPath
func UserDir(subPath string) string {
	dir, _ := homedir.Dir()

	return dir + "/" + subPath
}

// UserCacheDir will prepend user `$HOME/.cache` to subPath
func UserCacheDir(subPath string) string {
	dir, _ := homedir.Dir()

	return dir + "/.cache/" + subPath
}

// UserConfigDir will prepend user `$HOME/.config` to subPath
func UserConfigDir(subPath string) string {
	dir, _ := homedir.Dir()

	return dir + "/.config/" + subPath
}

// ExpandPath will parse `~` as user home dir path.
func ExpandPath(path string) string {
	path, _ = homedir.Expand(path)
	return path
}
