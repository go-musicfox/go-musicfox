package ui

import (
	"reflect"
	"testing"

	"github.com/go-musicfox/go-musicfox/internal/composer"
	"github.com/go-musicfox/go-musicfox/internal/framework"
	"github.com/go-musicfox/go-musicfox/internal/lastfm"
)

// fakePlugin records lifecycle events into a shared history and optionally
// provides a service into the context on Start.
type fakePlugin struct {
	name    string
	history *[]string
	provide func(ctx *framework.Context) error
}

func (p *fakePlugin) Start(ctx *framework.Context) error {
	*p.history = append(*p.history, "start:"+p.name)
	if p.provide != nil {
		return p.provide(ctx)
	}
	return nil
}

func (p *fakePlugin) Stop() error {
	*p.history = append(*p.history, "stop:"+p.name)
	return nil
}

func (p *fakePlugin) Dispose() error {
	*p.history = append(*p.history, "dispose:"+p.name)
	return nil
}

func TestServiceScopeStartResolvesService(t *testing.T) {
	var history []string
	scope := framework.NewScope()
	ctx := &framework.Context{}

	_ = scope.Add(&fakePlugin{
		name:    "provider",
		history: &history,
		provide: func(ctx *framework.Context) error {
			ctx.Provide("fakeSvc", "hello")
			return nil
		},
	})
	_ = scope.Add(&fakePlugin{name: "consumer", history: &history})

	if err := scope.Start(ctx); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	if svc, ok := framework.ServiceOf[string](ctx, "fakeSvc"); !ok || svc != "hello" {
		t.Fatalf("ServiceOf(fakeSvc) = %q, %v; want %q, true", svc, ok, "hello")
	}
	assertHistory(t, history, []string{"start:provider", "start:consumer"})
}

func TestServiceScopeStopDisposeReverseOrder(t *testing.T) {
	var history []string
	scope := framework.NewScope()
	for _, name := range []string{"p1", "p2", "p3"} {
		_ = scope.Add(&fakePlugin{name: name, history: &history})
	}

	if err := scope.Start(&framework.Context{}); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	if err := scope.Stop(); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
	if err := scope.Dispose(); err != nil {
		t.Fatalf("Dispose() error = %v", err)
	}
	// Stop runs plugins in reverse order, Dispose repeats it in reverse order.
	assertHistory(t, history, []string{
		"start:p1", "start:p2", "start:p3",
		"stop:p3", "stop:p2", "stop:p1",
		"dispose:p3", "dispose:p2", "dispose:p1",
	})
}

func TestServiceScopeDisposeChildBeforeParent(t *testing.T) {
	var history []string
	parent := framework.NewScope()
	child := parent.NewScope()
	_ = parent.Add(&fakePlugin{name: "parent", history: &history})
	_ = child.Add(&fakePlugin{name: "child", history: &history})

	if err := parent.Start(&framework.Context{}); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	if err := parent.Dispose(); err != nil {
		t.Fatalf("Dispose() error = %v", err)
	}
	// Dispose of a started scope implicitly stops it first: child and parent
	// plugins receive Stop then Dispose (cordis dispose semantics).
	assertHistory(t, history, []string{"start:parent", "start:child", "stop:child", "stop:parent", "dispose:child", "dispose:parent"})
}

func TestAppScopeRegistersShareAndLastfm(t *testing.T) {
	// Zero-value instances are sufficient: the scope plugins only check
	// non-nil and register the existing instance. Do not call lastfm.NewClient()
	// here — it reads configs.AppConfig / storage which are uninitialized in
	// unit tests.
	n := &Netease{
		shareSvc: &composer.ShareService{},
		lastfm:   &lastfm.Client{},
	}
	scope := newAppScope(n)
	ctx := &framework.Context{}

	if err := scope.Start(ctx); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	if svc, ok := framework.ServiceOf[*composer.ShareService](ctx, ServiceShareSvc); !ok || svc != n.shareSvc {
		t.Fatalf("ServiceOf(shareSvc) = %v, %v; want existing instance, true", svc, ok)
	}
	if svc, ok := framework.ServiceOf[*lastfm.Client](ctx, ServiceLastfm); !ok || svc != n.lastfm {
		t.Fatalf("ServiceOf(lastfm) = %v, %v; want existing instance, true", svc, ok)
	}

	if err := scope.Stop(); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
	if err := scope.Dispose(); err != nil {
		t.Fatalf("Dispose() error = %v", err)
	}
}

func assertHistory(t *testing.T, got, want []string) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("history = %v, want %v", got, want)
	}
}
