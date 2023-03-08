package sysutil

import (
	"os"
	"os/user"

	"github.com/gookit/goutil/internal/comfunc"
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

// UHomeDir get user home dir path.
func UHomeDir() string {
	// check $HOME/.terminfo
	u, err := user.Current()
	if err != nil {
		return ""
	}
	return u.HomeDir
}

// homeDir cache
var homeDir string

// UserHomeDir is alias of os.UserHomeDir, but ignore error
func UserHomeDir() string {
	if homeDir == "" {
		homeDir, _ = os.UserHomeDir()
	}
	return homeDir
}

// HomeDir get user home dir path.
func HomeDir() string {
	return UserHomeDir()
}

// UserDir will prepend user home dir to subPath
func UserDir(subPath string) string {
	dir := UserHomeDir()
	return dir + "/" + subPath
}

// UserCacheDir will prepend user `$HOME/.cache` to subPath
func UserCacheDir(subPath string) string {
	dir := UserHomeDir()
	return dir + "/.cache/" + subPath
}

// UserConfigDir will prepend user `$HOME/.config` to subPath
func UserConfigDir(subPath string) string {
	dir := UserHomeDir()
	return dir + "/.config/" + subPath
}

// ExpandPath will parse `~` as user home dir path.
func ExpandPath(path string) string {
	return comfunc.ExpandPath(path)
}
