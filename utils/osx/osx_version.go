//go:build darwin

package osx

import (
	"strconv"
	"strings"

	"golang.org/x/sys/unix"
)

var versions = map[int]string{
	4: "10.0.{decimal}",
	5: "10.1.{decimal}",
	6: "10.2.{decimal}",
	7: "10.3.{decimal}",
	8 : "10.4.{decimal}",
	9: "10.5.{decimal}",
	10: "10.6.{decimal}",
	11: "10.7.{decimal}",
	12: "10.8.{decimal}",
	13: "10.9.{decimal}",
	14: "10.10.{decimal}",
	15: "10.11.{decimal}",
	16: "10.12.{decimal}",
	17: "10.13.{decimal}",
	18: "10.14.{decimal}",
	19: "10.15.{decimal}",
	20: "11.{decimal}",
	21: "12.{decimal}",
	22: "13.{decimal}",
	23: "14.{decimal}",
	24: "15.{decimal}",
	25: "25.{decimal}",
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

	// if major < 4 || major > 24 {
	// 	return
	// }
	
	versionStr, ok := versions[major-4]
	if !ok {
		return
	}

	decimal := strconv.Itoa(patch)
	if major >= 18 {
		decimal = strconv.Itoa(minor - 1)
	}

	versionStr = strings.Replace(versionStr, "{decimal}", decimal, 1)
	version := strings.Split(versionStr, ".")

	v.Major, _ = strconv.Atoi(version[0])
	v.Minor, _ = strconv.Atoi(version[1])
	if len(version) >= 3 {
		v.Patch, _ = strconv.Atoi(version[2])
	}
	return
}
