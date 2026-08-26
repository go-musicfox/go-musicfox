package webui

import (
	"context"

	"github.com/go-musicfox/go-musicfox/internal/frontend"
)

// webuiFrontend adapts the WebUI frontend to the frontend.Frontend contract.
// Its Run delegates to webui.Run(ctx) (internal/webui/run.go).
type webuiFrontend struct{}

func (webuiFrontend) ID() string   { return "webui" }
func (webuiFrontend) Name() string { return "WebUI" }

func (webuiFrontend) Run(ctx context.Context, _ frontend.LaunchOptions) error {
	return Run(ctx)
}

func init() {
	frontend.Register(webuiFrontend{})
}
