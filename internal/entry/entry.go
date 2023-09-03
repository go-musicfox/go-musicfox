//go:build !darwin

package entry

import "github.com/go-musicfox/go-musicfox/utils"

func AppEntry() {
	defer utils.Recover(false)

	runCLI()
}
