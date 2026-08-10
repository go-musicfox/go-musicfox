//go:build linux

package ui

import (
	"testing"

	"github.com/go-musicfox/go-musicfox/internal/webkitgtk"
)

func TestUsesLegacyWebKitAPI(t *testing.T) {
	tests := []struct {
		name    string
		version int
		want    bool
	}{
		{name: "6.0 uses network session", version: webkitgtk.Version6, want: false},
		{name: "4.1 uses web context", version: webkitgtk.Version41, want: true},
		{name: "4.0 uses web context", version: webkitgtk.Version40, want: true},
		{name: "unavailable uses neither", version: 0, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := usesLegacyWebKitAPI(tt.version); got != tt.want {
				t.Fatalf("usesLegacyWebKitAPI(%d) = %t, want %t", tt.version, got, tt.want)
			}
		})
	}
}
