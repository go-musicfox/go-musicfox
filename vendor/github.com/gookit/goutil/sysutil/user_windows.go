//go:build windows
// +build windows

package sysutil

// ChangeUserByName change work user by new username.
func ChangeUserByName(newUname string) (err error) {
	return ChangeUserUidGid(0, 0)
}

// ChangeUserUidGid change work user by new username uid,gid
func ChangeUserUidGid(newUid int, newGid int) (err error) {
	return nil
}
