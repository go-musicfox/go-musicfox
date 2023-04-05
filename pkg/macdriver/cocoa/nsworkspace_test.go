//go:build darwin

package cocoa

import (
	"testing"
)

func TestNSWorkspace(t *testing.T) {
	workspace := NSWorkspace_sharedWorkspace()
	if workspace.ID == 0 {
		panic("get shared workspace failed")
	}

	nc := workspace.NotificationCenter()
	if nc.ID == 0 {
		panic("get notification center failed")
	}
}
