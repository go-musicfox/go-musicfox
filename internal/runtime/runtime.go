//go:build !darwin

package runtime

import "github.com/go-musicfox/go-musicfox/utils"

func Run(f func()) {
	defer utils.Recover(false)

	f()
}
