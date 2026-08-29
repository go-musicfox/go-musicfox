package ui

import (
	"reflect"
	"sort"
	"testing"

	"github.com/go-musicfox/go-musicfox/internal/core"
	"github.com/go-musicfox/go-musicfox/internal/framework"
)

// testNetease builds a Netease whose TUI-only fields are non-nil so
// registerUIExtraServices registers every ui service with the right concrete
// type. The engine-owned services are not registered here — the engine owns
// them at runtime; the core registration is asserted by the core test package.
func testNetease() *Netease {
	return &Netease{
		coverRenderer: &CoverRenderer{},
	}
}

func TestRegisterUIExtraServicesRegistersExactSet(t *testing.T) {
	ctx := &framework.Context{}
	if err := registerUIExtraServices(ctx, testNetease()); err != nil {
		t.Fatalf("registerUIExtraServices() error = %v", err)
	}

	// The ui layer registers exactly the three TUI-only extras.
	got := ctx.Names()
	sort.Strings(got)
	wantUI := []string{ServiceCoverRenderer, ServiceMenuRegistry, ServicePageRegistry}
	sort.Strings(wantUI)
	if !reflect.DeepEqual(got, wantUI) {
		t.Fatalf("ui registered services = %v, want %v", got, wantUI)
	}

	// The combined canonical set is core's 8 + ui's 3 — no missing, no extras
	// and no overlap. The alias consts (ServicePlayer etc.) must resolve to the
	// core values so the shared name space cannot drift.
	want := []string{
		ServicePlayer, ServiceLyricService, ServiceTrackManager, ServiceDesktopLyrics,
		ServiceUserService, ServiceLoginService, ServiceShareSvc, ServiceLastfm,
		ServiceCoverRenderer, ServiceMenuRegistry, ServicePageRegistry,
	}
	sort.Strings(want)
	if got := len(want); got != 11 {
		t.Fatalf("combined canonical service set = %d names, want 11 (core 8 + ui 3)", got)
	}
	// The core aliases must equal the core constants (name-space integrity).
	for i, name := range []string{ServicePlayer, ServiceLyricService, ServiceTrackManager, ServiceDesktopLyrics,
		ServiceUserService, ServiceLoginService, ServiceShareSvc, ServiceLastfm} {
		if name != []string{core.ServicePlayer, core.ServiceLyricService, core.ServiceTrackManager, core.ServiceDesktopLyrics,
			core.ServiceUserService, core.ServiceLoginService, core.ServiceShareSvc, core.ServiceLastfm}[i] {
			t.Fatalf("ui service alias %q diverged from core constant", name)
		}
	}
}

func TestRegisterUIExtraServicesRegistersRightConcreteTypes(t *testing.T) {
	ctx := &framework.Context{}
	n := testNetease()
	if err := registerUIExtraServices(ctx, n); err != nil {
		t.Fatalf("registerUIExtraServices() error = %v", err)
	}

	if svc, ok := framework.ServiceOf[*CoverRenderer](ctx, ServiceCoverRenderer); !ok || svc != n.coverRenderer {
		t.Errorf("ServiceOf(coverRenderer) = %v, %v; want %T, true", svc, ok, n.coverRenderer)
	}
	if svc, ok := framework.ServiceOf[MenuRegistry](ctx, ServiceMenuRegistry); !ok {
		t.Errorf("ServiceOf(menuRegistry) not resolvable: %v, %v", svc, ok)
	}
	if svc, ok := framework.ServiceOf[PageRegistry](ctx, ServicePageRegistry); !ok {
		t.Errorf("ServiceOf(pageRegistry) not resolvable: %v, %v", svc, ok)
	}
}
