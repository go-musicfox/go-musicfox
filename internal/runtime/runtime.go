//go:build !darwin

package runtime

import (
	"github.com/go-musicfox/go-musicfox/utils/errorx"
)

func Run(f func()) {
	defer errorx.Recover(false)

	f()
}
