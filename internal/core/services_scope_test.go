package core

import (
	"testing"

	"github.com/go-musicfox/go-musicfox/internal/composer"
	"github.com/go-musicfox/go-musicfox/internal/framework"
	"github.com/go-musicfox/go-musicfox/internal/lastfm"
)

func TestAppScopeRegistersShareAndLastfm(t *testing.T) {
	// Zero-value instances are sufficient: the scope plugins only check
	// non-nil and register the existing instance. Do not call lastfm.NewClient()
	// here — it reads configs.AppConfig / storage which are uninitialized in
	// unit tests.
	e := &Engine{
		shareSvc: &composer.ShareService{},
		lastfm:   &lastfm.Client{},
	}
	scope := newAppScope(e)
	ctx := &framework.Context{}

	if err := scope.Start(ctx); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	if svc, ok := framework.ServiceOf[*composer.ShareService](ctx, ServiceShareSvc); !ok || svc != e.shareSvc {
		t.Fatalf("ServiceOf(shareSvc) = %v, %v; want existing instance, true", svc, ok)
	}
	if svc, ok := framework.ServiceOf[*lastfm.Client](ctx, ServiceLastfm); !ok || svc != e.lastfm {
		t.Fatalf("ServiceOf(lastfm) = %v, %v; want existing instance, true", svc, ok)
	}

	if err := scope.Stop(); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
	if err := scope.Dispose(); err != nil {
		t.Fatalf("Dispose() error = %v", err)
	}
}
