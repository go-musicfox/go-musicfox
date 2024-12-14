//go:build darwin

package osx

import (
	"strconv"
	"strings"

	"golang.org/x/sys/unix"
)

var versions = []string{
	"10.0.{decimal}",
	"10.1.{decimal}",
	"10.2.{decimal}",
	"10.3.{decimal}",
	"10.4.{decimal}",
	"10.5.{decimal}",
	"10.6.{decimal}",
	"10.7.{decimal}",
	"10.8.{decimal}",
	"10.9.{decimal}",
	"10.10.{decimal}",
	"10.11.{decimal}",
	"10.12.{decimal}",
	"10.13.{decimal}",
	"10.14.{decimal}",
	"10.15.{decimal}",
	"11.{decimal}",
	"12.{decimal}",
	"13.{decimal}",
	"14.{decimal}",
}

type Version struct {
	Major int
	Minor int
	Patch int
}

func OsVersion() (v Version) {
	var uname unix.Utsname
	if err := unix.Uname(&uname); err != nil {
		return
	}
	kernelVersion := strings.Split(string(uname.Release[:]), ".")
	if len(kernelVersion) < 3 {
		return
	}

	major, _ := strconv.Atoi(kernelVersion[0])
	minor, _ := strconv.Atoi(kernelVersion[1])
	patch, _ := strconv.Atoi(kernelVersion[2])

	if major < 4 /* || major > 23 */ {
		return
	}

	decimal := strconv.Itoa(patch)
	if major >= 18 {
		decimal = strconv.Itoa(minor - 1)
	}

	versionStr := strings.Replace(versions[major-4], "{decimal}", decimal, 1)
	version := strings.Split(versionStr, ".")

	v.Major, _ = strconv.Atoi(version[0])
	v.Minor, _ = strconv.Atoi(version[1])
	if len(version) >= 3 {
		v.Patch, _ = strconv.Atoi(version[2])
	}
	return
}
