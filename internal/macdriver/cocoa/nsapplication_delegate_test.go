//go:build darwin

package cocoa

import "testing"

func TestApplicationStaysAliveAfterLastWindowClosed(t *testing.T) {
	if applicationShouldTerminateAfterLastWindowClosed(0, 0, 0) {
		t.Fatal("closing an auxiliary window must not terminate the Go application")
	}
}
