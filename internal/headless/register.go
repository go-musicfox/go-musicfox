package headless

import (
	"context"

	"github.com/go-musicfox/go-musicfox/internal/frontend"
)

// headlessFrontend adapts the headless frontend to the frontend.Frontend
// contract. Its Run delegates to the existing headless.Run(once string)
// (internal/headless/run.go).
type headlessFrontend struct{}

func (headlessFrontend) ID() string   { return "headless" }
func (headlessFrontend) Name() string { return "Headless" }

func (headlessFrontend) Run(_ context.Context, opts frontend.LaunchOptions) error {
	return Run(opts.Once)
}

func init() {
	frontend.Register(headlessFrontend{})
}
