//go:build !windows
// +build !windows

package sysutil

import (
	"syscall"

	"github.com/gookit/goutil/strutil"
)

// ChangeUserByName change work user by new username.
func ChangeUserByName(newUname string) (err error) {
	u := MustFindUser(newUname)

	// syscall.Setlogin(newUname)
	return ChangeUserUidGid(strutil.IntOrPanic(u.Uid), strutil.IntOrPanic(u.Gid))
}

// ChangeUserUidGid change work user by new username uid,gid
func ChangeUserUidGid(newUid int, newGid int) (err error) {
	if newUid > 0 {
		err = syscall.Setuid(newUid)

		// update group id
		if err == nil && newGid > 0 {
			err = syscall.Setgid(newGid)
		}
	}

	return
}
