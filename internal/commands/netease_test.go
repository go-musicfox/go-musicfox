package commands

import (
	"testing"

	"github.com/go-musicfox/go-musicfox/internal/configs"
)

// TestResolveFrontend verifies the frontend resolution priority chain:
// --frontend CLI ＞ --headless（legacy 别名）＞ [main] frontend
// ＞ [main] headless（legacy 配置）＞ 缺省 tui。
func TestResolveFrontend(t *testing.T) {
	origAppConfig := configs.AppConfig
	origHeadless := GlobalOptions.Headless
	origFrontend := GlobalOptions.Frontend
	t.Cleanup(func() {
		configs.AppConfig = origAppConfig
		GlobalOptions.Headless = origHeadless
		GlobalOptions.Frontend = origFrontend
	})

	tests := []struct {
		name     string
		cliFront string
		cliHead  bool
		cfgFront string
		cfgHead  bool
		want     string
	}{
		{name: "default is tui", want: "tui"},
		{name: "[main] frontend=headless", cfgFront: "headless", want: "headless"},
		{name: "[main] headless=true", cfgHead: true, want: "headless"},
		{name: "--headless", cliHead: true, want: "headless"},
		{name: "--frontend=headless", cliFront: "headless", want: "headless"},
		{name: "--frontend overrides --headless", cliFront: "tui", cliHead: true, want: "tui"},
		{name: "--headless overrides [main] frontend", cfgFront: "tui", cliHead: true, want: "headless"},
		{name: "--frontend overrides [main] headless", cfgHead: true, cliFront: "tui", want: "tui"},
		{name: "[main] frontend overrides [main] headless", cfgFront: "tui", cfgHead: true, want: "tui"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configs.AppConfig = &configs.Config{Main: configs.MainConfig{Frontend: tt.cfgFront, Headless: tt.cfgHead}}
			GlobalOptions.Frontend = tt.cliFront
			GlobalOptions.Headless = tt.cliHead
			if got := resolveFrontendID(); got != tt.want {
				t.Errorf("resolveFrontendID() = %q, want %q", got, tt.want)
			}
		})
	}
}
