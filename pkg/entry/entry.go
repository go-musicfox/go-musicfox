//go:build !darwin
// +build !darwin

package entry

import "go-musicfox/utils"

func AppEntry() {
	defer utils.Recover(false)

	runCLI()
}
