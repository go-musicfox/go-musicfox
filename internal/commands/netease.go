// Package commands implements the CLI subcommands (netease player, reset,
// config upgrade, notify test).
package commands

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	_ "net/http/pprof"
	"strconv"
	"strings"

	neteaseutil "github.com/go-musicfox/netease-music/util"
	"github.com/gookit/gcli/v2"
	"github.com/mattn/go-runewidth"

	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/frontend"
	"github.com/go-musicfox/go-musicfox/internal/storage"
	"github.com/go-musicfox/go-musicfox/internal/types"
	"github.com/go-musicfox/go-musicfox/utils/errorx"
	"github.com/go-musicfox/go-musicfox/utils/slogx"
)

func NewPlayerCommand() *gcli.Command {
	cmd := &gcli.Command{
		Name:   "netease",
		UseFor: "Command line player for Netease Cloud Music",
		Func:   runPlayer,
	}
	return cmd
}

func runPlayer(_ *gcli.Command, _ []string) error {
	if err := bootstrap(); err != nil {
		return err
	}
	id := resolveFrontendID()
	fe, ok := frontend.ByID(id)
	if !ok {
		return fmt.Errorf("未知前端 %q（可用: %s）", id, strings.Join(frontend.Registered(), ", "))
	}
	if GlobalOptions.Once != "" && id != "headless" {
		return errors.New("--once 仅支持 headless 前端")
	}
	mode := frontend.Mode(GlobalOptions.Mode)
	switch mode {
	case "", frontend.ModeStandalone, frontend.ModeConnect:
	default:
		return fmt.Errorf("未知模式 %q（可用: standalone|connect）", GlobalOptions.Mode)
	}
	return fe.Run(context.Background(), frontend.LaunchOptions{
		Once:  GlobalOptions.Once,
		Debug: GlobalOptions.DebugMode,
		Pprof: GlobalOptions.PProfMode,
		Mode:  mode,
	})
}

// resolveFrontendID 优先级：--frontend CLI ＞ --headless（legacy 别名）
// ＞ [main] frontend ＞ [main] headless（legacy 配置）＞ 缺省 tui。
func resolveFrontendID() string {
	if GlobalOptions.Frontend != "" {
		return GlobalOptions.Frontend
	}
	if GlobalOptions.Headless {
		return "headless"
	}
	if configs.AppConfig != nil {
		if configs.AppConfig.Main.Frontend != "" {
			return configs.AppConfig.Main.Frontend
		}
		if configs.AppConfig.Main.Headless {
			return "headless"
		}
	}
	return "tui"
}

// bootstrap runs the frontend-independent setup shared by the TUI and the
// headless frontend: pprof, debug log level sync, HTTP timeouts / runewidth
// and the storage DB manager. Order is preserved from the former inline setup.
func bootstrap() error {
	if GlobalOptions.PProfMode {
		errorx.Go(func() {
			panic(http.ListenAndServe(":"+strconv.Itoa(configs.AppConfig.Main.Pprof.Port), nil))
		}, true)
	}

	// Sync CLI --debug flag to AppConfig so it's visible to all packages.
	// Must be done here (inside runPlayer) because gcli parses flags during
	// app.Run(), which happens after cmd/musicfox.go's init-time sync attempt.
	if GlobalOptions.DebugMode {
		configs.AppConfig.Main.Debug = true
	}

	if GlobalOptions.DebugMode || configs.AppConfig.Main.Debug {
		slogx.LevelVar().Set(slog.LevelDebug)
	}

	http.DefaultClient.Timeout = types.AppHttpTimeout
	neteaseutil.HTTPClientTimeout = types.AppHttpTimeout
	runewidth.DefaultCondition.EastAsianWidth = false

	// DBManager 初始化
	storage.DBManager = new(storage.LocalDBManager)

	return nil
}
