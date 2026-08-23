package ui

import (
	"errors"

	"github.com/go-musicfox/go-musicfox/internal/framework"
)

// servicePlugin adapts a simple provide-function into a framework.PluginWithDeps.
// It is the minimal adapter used to bring existing Netease-owned service
// instances into the framework scope lifecycle: Deps resolves prerequisites,
// Start registers the instance into the context, Stop/Dispose are no-ops until
// the owned instances gain real cleanup (Phase 3.1.5 migration).
type servicePlugin struct {
	name    string
	provide func(ctx *framework.Context) error
}

// Deps resolves prerequisites before Start. None of the current slice services
// have prerequisites yet.
func (p *servicePlugin) Deps(_ *framework.Context) error {
	return nil
}

// Start registers the service into the context via the provider.
func (p *servicePlugin) Start(ctx *framework.Context) error {
	if p.provide == nil {
		return nil
	}
	return p.provide(ctx)
}

// Stop stops the service; no-op for the current slice.
func (p *servicePlugin) Stop() error {
	return nil
}

// Dispose performs recursive cleanup; no-op for the current slice.
func (p *servicePlugin) Dispose() error {
	return nil
}

// newAppScope builds the app-wide framework scope and wires the first
// production slice of framework-managed services into it. The instances stay
// owned by Netease; each plugin's Start registers the existing instance into
// the shared context. Ownership migration happens in Phase 3.1.5.
func newAppScope(n *Netease) *framework.Scope {
	scope := framework.NewScope()

	if err := scope.Add(&servicePlugin{
		name: ServiceShareSvc,
		provide: func(ctx *framework.Context) error {
			if n.shareSvc == nil {
				return errors.New("shareSvc not initialized before scope start")
			}
			provideIfAbsent(ctx, ServiceShareSvc, n.shareSvc)
			return nil
		},
	}); err != nil {
		panic(err) // unreachable for a fresh scope
	}

	if err := scope.Add(&servicePlugin{
		name: ServiceLastfm,
		provide: func(ctx *framework.Context) error {
			if n.lastfm == nil {
				return errors.New("lastfm not initialized before scope start")
			}
			provideIfAbsent(ctx, ServiceLastfm, n.lastfm)
			return nil
		},
	}); err != nil {
		panic(err) // unreachable for a fresh scope
	}

	return scope
}
