package core

import (
	"testing"

	"github.com/go-musicfox/go-musicfox/internal/composer"
	"github.com/go-musicfox/go-musicfox/internal/framework"
	"github.com/go-musicfox/go-musicfox/internal/lastfm"
)

// TestServicePluginsRegisterShareAndLastfm verifies the scope-provided
// services (shareSvc/lastfm, formerly wired by newAppScope in
// services_scope.go) are registered by the new service constructor plugins
// after the root scope starts.
func TestServicePluginsRegisterShareAndLastfm(t *testing.T) {
	e := testEngine()
	_, ctx := startTestScope(e)

	if svc, ok := framework.ServiceOf[*composer.ShareService](ctx, ServiceShareSvc); !ok || svc != e.shareSvc {
		t.Fatalf("ServiceOf(shareSvc) = %v, %v; want %T, true", svc, ok, e.shareSvc)
	}
	if svc, ok := framework.ServiceOf[*lastfm.Client](ctx, ServiceLastfm); !ok || svc != e.lastfm {
		t.Fatalf("ServiceOf(lastfm) = %v, %v; want %T, true", svc, ok, e.lastfm)
	}
}
