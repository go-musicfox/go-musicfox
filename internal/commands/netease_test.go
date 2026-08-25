package commands

import (
	"testing"

	"github.com/go-musicfox/go-musicfox/internal/configs"
)

// TestHeadlessEnabled verifies the headless branch condition: the CLI flag
// (GlobalOptions.Headless) wins over the [main] headless config value, and
// headless mode is off when neither is set.
func TestHeadlessEnabled(t *testing.T) {
	origAppConfig := configs.AppConfig
	origHeadless := GlobalOptions.Headless
	t.Cleanup(func() {
		configs.AppConfig = origAppConfig
		GlobalOptions.Headless = origHeadless
	})

	configs.AppConfig = &configs.Config{Main: configs.MainConfig{Headless: false}}

	GlobalOptions.Headless = false
	if headlessEnabled() {
		t.Error("headlessEnabled() = true, want false when neither flag nor config is set")
	}

	GlobalOptions.Headless = true
	if !headlessEnabled() {
		t.Error("headlessEnabled() = false, want true when CLI flag is set")
	}

	GlobalOptions.Headless = false
	configs.AppConfig.Main.Headless = true
	if !headlessEnabled() {
		t.Error("headlessEnabled() = false, want true when [main] headless config is set")
	}

	GlobalOptions.Headless = true
	if !headlessEnabled() {
		t.Error("headlessEnabled() = false, want true when CLI flag overrides a true config")
	}
}
