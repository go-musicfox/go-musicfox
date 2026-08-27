package framework

import (
	"errors"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
)

// trackingPlugin records each lifecycle call into a shared history slice.
type trackingPlugin struct {
	name    string
	history *[]string
}

func (p *trackingPlugin) Start(ctx *Context) error {
	*p.history = append(*p.history, "start:"+p.name)
	return nil
}

func (p *trackingPlugin) Stop() error {
	*p.history = append(*p.history, "stop:"+p.name)
	return nil
}

func (p *trackingPlugin) Dispose() error {
	*p.history = append(*p.history, "dispose:"+p.name)
	return nil
}

// depsPlugin implements PluginWithDeps: Deps resolves the "greeting" service
// from the context before Start runs.
type depsPlugin struct {
	name     string
	history  *[]string
	depValue any
}

func (p *depsPlugin) Start(ctx *Context) error {
	*p.history = append(*p.history, "start:"+p.name)
	return nil
}

func (p *depsPlugin) Stop() error {
	*p.history = append(*p.history, "stop:"+p.name)
	return nil
}

func (p *depsPlugin) Dispose() error {
	*p.history = append(*p.history, "dispose:"+p.name)
	return nil
}

func (p *depsPlugin) Deps(ctx *Context) error {
	*p.history = append(*p.history, "deps:"+p.name)
	p.depValue = ctx.Service("greeting")
	return nil
}

// failingPlugin fails at a configurable phase: "deps" or "start".
type failingPlugin struct {
	name    string
	history *[]string
	errOn   string
}

func (p *failingPlugin) Start(ctx *Context) error {
	*p.history = append(*p.history, "start:"+p.name)
	if p.errOn == "start" {
		return errors.New(p.name + " start failed")
	}
	return nil
}

func (p *failingPlugin) Stop() error {
	return nil
}

func (p *failingPlugin) Dispose() error {
	return nil
}

func (p *failingPlugin) Deps(ctx *Context) error {
	*p.history = append(*p.history, "deps:"+p.name)
	if p.errOn == "deps" {
		return errors.New(p.name + " deps failed")
	}
	return nil
}

// failStopPlugin records lifecycle calls and fails Stop with errMsg when set.
type failStopPlugin struct {
	name    string
	history *[]string
	errMsg  string
}

func (p *failStopPlugin) Start(ctx *Context) error {
	*p.history = append(*p.history, "start:"+p.name)
	return nil
}

func (p *failStopPlugin) Stop() error {
	*p.history = append(*p.history, "stop:"+p.name)
	if p.errMsg != "" {
		return errors.New(p.errMsg)
	}
	return nil
}

func (p *failStopPlugin) Dispose() error {
	*p.history = append(*p.history, "dispose:"+p.name)
	return nil
}

// failDisposePlugin records lifecycle calls and fails Dispose with errMsg when
// set.
type failDisposePlugin struct {
	name    string
	history *[]string
	errMsg  string
}

func (p *failDisposePlugin) Start(ctx *Context) error {
	*p.history = append(*p.history, "start:"+p.name)
	return nil
}

func (p *failDisposePlugin) Stop() error {
	*p.history = append(*p.history, "stop:"+p.name)
	return nil
}

func (p *failDisposePlugin) Dispose() error {
	*p.history = append(*p.history, "dispose:"+p.name)
	if p.errMsg != "" {
		return errors.New(p.errMsg)
	}
	return nil
}

// flakyStartPlugin fails Start the first failTimes calls and succeeds after.
type flakyStartPlugin struct {
	name       string
	history    *[]string
	startCalls int
	failTimes  int
}

func (p *flakyStartPlugin) Start(ctx *Context) error {
	p.startCalls++
	*p.history = append(*p.history, "start:"+p.name)
	if p.startCalls <= p.failTimes {
		return errors.New(p.name + " start failed")
	}
	return nil
}

func (p *flakyStartPlugin) Stop() error {
	return nil
}

func (p *flakyStartPlugin) Dispose() error {
	return nil
}

func assertHistory(t *testing.T, got, want []string) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("history = %v, want %v", got, want)
	}
}

func TestScopeStartForwardOrder(t *testing.T) {
	var history []string
	scope := &Scope{}
	for _, name := range []string{"p1", "p2", "p3"} {
		_ = scope.Add(&trackingPlugin{name: name, history: &history})
	}
	if err := scope.Start(&Context{}); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	assertHistory(t, history, []string{"start:p1", "start:p2", "start:p3"})
}

func TestScopeStartCallsDepsBeforeStart(t *testing.T) {
	var history []string
	ctx := &Context{}
	ctx.Provide("greeting", "hello")
	p := &depsPlugin{name: "p", history: &history}

	scope := &Scope{}
	_ = scope.Add(p)
	if err := scope.Start(ctx); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	assertHistory(t, history, []string{"deps:p", "start:p"})
	if p.depValue != "hello" {
		t.Fatalf("depValue = %v, want %q (dependency not injected)", p.depValue, "hello")
	}
}

func TestScopeStopReverseOrder(t *testing.T) {
	var history []string
	scope := &Scope{}
	for _, name := range []string{"p1", "p2", "p3"} {
		_ = scope.Add(&trackingPlugin{name: name, history: &history})
	}
	if err := scope.Start(&Context{}); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	if err := scope.Stop(); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
	assertHistory(t, history, []string{"start:p1", "start:p2", "start:p3", "stop:p3", "stop:p2", "stop:p1"})
}

func TestScopeDisposeChildBeforeParent(t *testing.T) {
	var history []string
	parent := &Scope{}
	child := parent.NewScope()
	_ = parent.Add(&trackingPlugin{name: "parent", history: &history})
	_ = child.Add(&trackingPlugin{name: "child", history: &history})

	if err := parent.Start(&Context{}); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	if err := parent.Dispose(); err != nil {
		t.Fatalf("Dispose() error = %v", err)
	}
	// Dispose of a started scope implicitly stops it first: child and parent
	// plugins receive Stop then Dispose (cordis dispose semantics).
	assertHistory(t, history, []string{"start:parent", "start:child", "stop:child", "stop:parent", "dispose:child", "dispose:parent"})
}

func TestChildScopeFollowsParentLifecycle(t *testing.T) {
	var history []string
	parent := &Scope{}
	child := parent.NewScope()
	_ = parent.Add(&trackingPlugin{name: "parent", history: &history})
	_ = child.Add(&trackingPlugin{name: "child", history: &history})

	if err := parent.Start(&Context{}); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	assertHistory(t, history, []string{"start:parent", "start:child"})

	if err := parent.Stop(); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
	assertHistory(t, history, []string{"start:parent", "start:child", "stop:child", "stop:parent"})

	if err := parent.Dispose(); err != nil {
		t.Fatalf("Dispose() error = %v", err)
	}
	assertHistory(t, history, []string{"start:parent", "start:child", "stop:child", "stop:parent", "dispose:child", "dispose:parent"})
}

func TestScopeDisposeRecursesNestedScopes(t *testing.T) {
	var history []string
	root := &Scope{}
	mid := root.NewScope()
	leaf := mid.NewScope()
	_ = root.Add(&trackingPlugin{name: "root", history: &history})
	_ = mid.Add(&trackingPlugin{name: "mid", history: &history})
	_ = leaf.Add(&trackingPlugin{name: "leaf", history: &history})

	if err := root.Start(&Context{}); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	assertHistory(t, history, []string{"start:root", "start:mid", "start:leaf"})

	if err := root.Dispose(); err != nil {
		t.Fatalf("Dispose() error = %v", err)
	}
	// Dispose of a started scope implicitly stops it first (recursively):
	// leaf, mid and root all receive Stop before Dispose.
	assertHistory(t, history, []string{
		"start:root", "start:mid", "start:leaf",
		"stop:leaf", "stop:mid", "stop:root",
		"dispose:leaf", "dispose:mid", "dispose:root",
	})
}

func TestScopeDisposeDetachesChildrenAndIsIdempotent(t *testing.T) {
	var history []string
	parent := &Scope{}
	child := parent.NewScope()
	_ = parent.Add(&trackingPlugin{name: "parent", history: &history})
	_ = child.Add(&trackingPlugin{name: "child", history: &history})

	if err := parent.Dispose(); err != nil {
		t.Fatalf("Dispose() error = %v", err)
	}
	if len(parent.children) != 0 {
		t.Fatalf("parent still holds %d children after Dispose", len(parent.children))
	}
	if child.parent != nil {
		t.Fatal("child still references parent after Dispose")
	}
	// A second Dispose must be a no-op: no double cleanup.
	if err := parent.Dispose(); err != nil {
		t.Fatalf("second Dispose() error = %v", err)
	}
	assertHistory(t, history, []string{"dispose:child", "dispose:parent"})
}

func TestScopeStartErrorRollsBackStartedPlugins(t *testing.T) {
	var history []string
	scope := &Scope{}
	_ = scope.Add(&trackingPlugin{name: "p1", history: &history})
	_ = scope.Add(&failingPlugin{name: "bad", history: &history, errOn: "start"})
	_ = scope.Add(&trackingPlugin{name: "p3", history: &history})

	err := scope.Start(&Context{})
	if err == nil {
		t.Fatal("Start() error = nil, want non-nil")
	}
	// p1 started before bad's Start failed; it must be rolled back (stopped in
	// reverse order). p3 never starts.
	assertHistory(t, history, []string{"start:p1", "deps:bad", "start:bad", "stop:p1"})
}

func TestScopeDepsErrorRollsBackStartedPlugins(t *testing.T) {
	var history []string
	scope := &Scope{}
	_ = scope.Add(&trackingPlugin{name: "p1", history: &history})
	_ = scope.Add(&failingPlugin{name: "bad", history: &history, errOn: "deps"})
	_ = scope.Add(&trackingPlugin{name: "p3", history: &history})

	err := scope.Start(&Context{})
	if err == nil {
		t.Fatal("Start() error = nil, want non-nil")
	}
	// bad's Deps fails, so its Start and p3 must never run; p1 is rolled back.
	assertHistory(t, history, []string{"start:p1", "deps:bad", "stop:p1"})
}

func TestScopeStartErrorRollsBackStartedChildScopes(t *testing.T) {
	var history []string
	scope := &Scope{}
	_ = scope.Add(&trackingPlugin{name: "root", history: &history})
	child1 := scope.NewScope()
	_ = child1.Add(&trackingPlugin{name: "c1", history: &history})
	child2 := scope.NewScope()
	_ = child2.Add(&failingPlugin{name: "c2bad", history: &history, errOn: "start"})
	child3 := scope.NewScope()
	_ = child3.Add(&trackingPlugin{name: "c3", history: &history})

	err := scope.Start(&Context{})
	if err == nil {
		t.Fatal("Start() error = nil, want non-nil")
	}
	// root plugin and child1 started before child2's plugin failed. Rollback
	// stops child1 then root in reverse order; child3 never starts.
	assertHistory(t, history, []string{
		"start:root", "start:c1", "deps:c2bad", "start:c2bad",
		"stop:c1", "stop:root",
	})
}

func TestScopeStartErrorRollsBackDeepChildScope(t *testing.T) {
	var history []string
	scope := &Scope{}
	child := scope.NewScope()
	_ = child.Add(&trackingPlugin{name: "child", history: &history})
	deep := child.NewScope()
	_ = deep.Add(&failingPlugin{name: "deepbad", history: &history, errOn: "start"})

	err := scope.Start(&Context{})
	if err == nil {
		t.Fatal("Start() error = nil, want non-nil")
	}
	// child started, then its nested deep scope failed; child must be stopped.
	assertHistory(t, history, []string{"start:child", "deps:deepbad", "start:deepbad", "stop:child"})
}

// --- lifecycle state machine ---

func TestScopeDoubleStartReturnsError(t *testing.T) {
	var history []string
	scope := &Scope{}
	_ = scope.Add(&trackingPlugin{name: "p", history: &history})
	if err := scope.Start(&Context{}); err != nil {
		t.Fatalf("first Start() error = %v", err)
	}
	err := scope.Start(&Context{})
	if err == nil {
		t.Fatal("second Start() error = nil, want explicit double-start error")
	}
	if !strings.Contains(err.Error(), "already started") {
		t.Fatalf("second Start() error = %v, want %q", err, "already started")
	}
	// The first Start must not have been re-run.
	assertHistory(t, history, []string{"start:p"})
}

func TestScopeStopBeforeStartIsNoop(t *testing.T) {
	var history []string
	scope := &Scope{}
	_ = scope.Add(&trackingPlugin{name: "p", history: &history})
	if err := scope.Stop(); err != nil {
		t.Fatalf("Stop() before Start error = %v, want nil", err)
	}
	if len(history) != 0 {
		t.Fatalf("history = %v, want empty (no lifecycle calls before Start)", history)
	}
}

func TestScopeDisposeImplicitlyStopsStartedScope(t *testing.T) {
	var history []string
	scope := &Scope{}
	_ = scope.Add(&trackingPlugin{name: "p", history: &history})
	if err := scope.Start(&Context{}); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	if err := scope.Dispose(); err != nil {
		t.Fatalf("Dispose() error = %v", err)
	}
	// The started scope is stopped before being disposed, so the plugin
	// receives both Stop and Dispose even though only Dispose was called.
	assertHistory(t, history, []string{"start:p", "stop:p", "dispose:p"})
}

func TestScopeStartAfterDisposeReturnsError(t *testing.T) {
	scope := &Scope{}
	if err := scope.Dispose(); err != nil {
		t.Fatalf("Dispose() error = %v", err)
	}
	if err := scope.Start(&Context{}); err == nil {
		t.Fatal("Start() after Dispose error = nil, want explicit error (disposed scope is final)")
	}
}

func TestScopeAddAfterDisposeReturnsError(t *testing.T) {
	scope := &Scope{}
	if err := scope.Dispose(); err != nil {
		t.Fatalf("Dispose() error = %v", err)
	}
	if err := scope.Add(&trackingPlugin{name: "p", history: &[]string{}}); err == nil {
		t.Fatal("Add() after Dispose error = nil, want explicit error (disposed scope is final)")
	}
	if len(scope.plugins) != 0 {
		t.Fatalf("scope holds %d plugins after rejected Add", len(scope.plugins))
	}
}

func TestScopeFailedStartStaysUnstartedAndCanRestart(t *testing.T) {
	var history []string
	scope := &Scope{}
	_ = scope.Add(&flakyStartPlugin{name: "p", history: &history, failTimes: 1})
	if err := scope.Start(&Context{}); err == nil {
		t.Fatal("first Start() error = nil, want non-nil")
	}
	// The scope never started: Stop is a no-op.
	if err := scope.Stop(); err != nil {
		t.Fatalf("Stop() after failed Start error = %v, want nil", err)
	}
	// A retried Start succeeds, proving the scope was not left started.
	if err := scope.Start(&Context{}); err != nil {
		t.Fatalf("retry Start() error = %v", err)
	}
	assertHistory(t, history, []string{"start:p", "start:p"})
}

// --- Stop/Dispose error aggregation ---

func TestScopeStopAggregatesErrorsAndStopsAll(t *testing.T) {
	var history []string
	scope := &Scope{}
	_ = scope.Add(&failStopPlugin{name: "p1", history: &history, errMsg: "p1 stop failed"})
	_ = scope.Add(&failStopPlugin{name: "p2", history: &history, errMsg: "p2 stop failed"})
	if err := scope.Start(&Context{}); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	err := scope.Stop()
	if err == nil {
		t.Fatal("Stop() error = nil, want non-nil")
	}
	// Both failures are joined into a single error.
	for _, msg := range []string{"p1 stop failed", "p2 stop failed"} {
		if !strings.Contains(err.Error(), msg) {
			t.Fatalf("Stop() error = %v, want it to join %q", err, msg)
		}
	}
	// Both plugins were stopped despite the failures, in reverse order.
	assertHistory(t, history, []string{"start:p1", "start:p2", "stop:p2", "stop:p1"})
	// The scope is marked stopped even with failures: a second Stop is a no-op.
	if err := scope.Stop(); err != nil {
		t.Fatalf("second Stop() error = %v, want nil", err)
	}
}

func TestScopeStopAggregatesChildErrorsAndContinues(t *testing.T) {
	var history []string
	scope := &Scope{}
	child1 := scope.NewScope()
	_ = child1.Add(&failStopPlugin{name: "c1", history: &history, errMsg: "c1 stop failed"})
	child2 := scope.NewScope()
	_ = child2.Add(&failStopPlugin{name: "c2", history: &history, errMsg: "c2 stop failed"})
	_ = scope.Add(&trackingPlugin{name: "p", history: &history})

	if err := scope.Start(&Context{}); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	err := scope.Stop()
	if err == nil {
		t.Fatal("Stop() error = nil, want non-nil")
	}
	// Both child failures are joined and the parent plugin is still stopped.
	for _, msg := range []string{"c1 stop failed", "c2 stop failed"} {
		if !strings.Contains(err.Error(), msg) {
			t.Fatalf("Stop() error = %v, want it to join %q", err, msg)
		}
	}
	assertHistory(t, history, []string{
		"start:p", "start:c1", "start:c2",
		"stop:c2", "stop:c1", "stop:p",
	})
}

func TestScopeDisposeAggregatesErrorsAndClearsState(t *testing.T) {
	var history []string
	scope := &Scope{}
	child := scope.NewScope()
	_ = child.Add(&failDisposePlugin{name: "c1", history: &history, errMsg: "c1 dispose failed"})
	_ = scope.Add(&failDisposePlugin{name: "p1", history: &history, errMsg: "p1 dispose failed"})

	err := scope.Dispose()
	if err == nil {
		t.Fatal("Dispose() error = nil, want non-nil")
	}
	for _, msg := range []string{"c1 dispose failed", "p1 dispose failed"} {
		if !strings.Contains(err.Error(), msg) {
			t.Fatalf("Dispose() error = %v, want it to join %q", err, msg)
		}
	}
	// State cleanup still happened despite the failures.
	if len(scope.children) != 0 || len(scope.plugins) != 0 {
		t.Fatalf("scope state not cleared: children=%d plugins=%d", len(scope.children), len(scope.plugins))
	}
	if child.parent != nil {
		t.Fatal("child still references parent after Dispose")
	}
	// A second Dispose is a no-op returning nil (idempotent).
	if err := scope.Dispose(); err != nil {
		t.Fatalf("second Dispose() error = %v, want nil", err)
	}
}

func TestScopeDisposeStartedScopeAggregatesStopAndDisposeErrors(t *testing.T) {
	var history []string
	scope := &Scope{}
	_ = scope.Add(&failStopPlugin{name: "p", history: &history, errMsg: "p stop failed"})
	if err := scope.Start(&Context{}); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	err := scope.Dispose()
	if err == nil {
		t.Fatal("Dispose() error = nil, want non-nil")
	}
	if !strings.Contains(err.Error(), "p stop failed") {
		t.Fatalf("Dispose() error = %v, want it to include the implicit Stop failure", err)
	}
	// Dispose still ran to completion: the plugin's Dispose was invoked.
	assertHistory(t, history, []string{"start:p", "stop:p", "dispose:p"})
	if len(scope.plugins) != 0 {
		t.Fatalf("scope holds %d plugins after Dispose", len(scope.plugins))
	}
}

// --- enabled registration (AddWithEnabled / AddAndStart / Plugins) ---

// basePlugin embeds NoopPlugin (lifecycle) and PluginBase (enabled state).
type basePlugin struct {
	NoopPlugin
	PluginBase
}

func TestScopeAddWithEnabledDisabledSkipsStartButDisposes(t *testing.T) {
	var history []string
	scope := &Scope{}
	_ = scope.AddWithEnabled(&trackingPlugin{name: "p1", history: &history}, true)
	_ = scope.AddWithEnabled(&trackingPlugin{name: "disabled", history: &history}, false)
	_ = scope.AddWithEnabled(&trackingPlugin{name: "p2", history: &history}, true)

	if err := scope.Start(&Context{}); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	// The disabled plugin is skipped by Start but stays in the slice.
	assertHistory(t, history, []string{"start:p1", "start:p2"})

	if err := scope.Dispose(); err != nil {
		t.Fatalf("Dispose() error = %v", err)
	}
	// The disabled plugin never started, yet its Stop and Dispose still run in
	// order (registered plugins are finalized regardless of enabled state).
	assertHistory(t, history, []string{
		"start:p1", "start:p2",
		"stop:p2", "stop:disabled", "stop:p1",
		"dispose:p2", "dispose:disabled", "dispose:p1",
	})
	// Dispose stays idempotent with a disabled plugin present.
	if err := scope.Dispose(); err != nil {
		t.Fatalf("second Dispose() error = %v, want nil", err)
	}
}

func TestScopeAddWithEnabledAfterDisposeReturnsError(t *testing.T) {
	scope := &Scope{}
	if err := scope.Dispose(); err != nil {
		t.Fatalf("Dispose() error = %v", err)
	}
	if err := scope.AddWithEnabled(&trackingPlugin{name: "p", history: &[]string{}}, true); err == nil {
		t.Fatal("AddWithEnabled() after Dispose error = nil, want explicit error (disposed scope is final)")
	}
}

func TestScopeStartRollbackSkipsDisabledPlugins(t *testing.T) {
	var history []string
	scope := &Scope{}
	_ = scope.AddWithEnabled(&trackingPlugin{name: "p1", history: &history}, true)
	_ = scope.AddWithEnabled(&trackingPlugin{name: "disabled", history: &history}, false)
	_ = scope.AddWithEnabled(&failingPlugin{name: "bad", history: &history, errOn: "start"}, true)

	err := scope.Start(&Context{})
	if err == nil {
		t.Fatal("Start() error = nil, want non-nil")
	}
	// p1 started, the disabled plugin was skipped, then bad failed: rollback
	// stops only the plugins that were actually started (p1), never the
	// disabled one.
	assertHistory(t, history, []string{"start:p1", "deps:bad", "start:bad", "stop:p1"})
}

func TestScopeAddWithEnabledWritesPluginBaseEnabled(t *testing.T) {
	scope := &Scope{}
	enabled := &basePlugin{}
	disabled := &basePlugin{}
	_ = scope.AddWithEnabled(enabled, true)
	_ = scope.AddWithEnabled(disabled, false)
	if !enabled.Enabled {
		t.Fatal("enabled plugin's PluginBase.Enabled = false, want true")
	}
	if disabled.Enabled {
		t.Fatal("disabled plugin's PluginBase.Enabled = true, want false")
	}
}

func TestScopeAddAndStartOnStartedScope(t *testing.T) {
	var history []string
	scope := &Scope{}
	_ = scope.Add(&trackingPlugin{name: "p1", history: &history})
	if err := scope.Start(&Context{}); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	if err := scope.AddAndStart(&Context{}, &trackingPlugin{name: "dyn", history: &history}); err != nil {
		t.Fatalf("AddAndStart() error = %v", err)
	}
	// The dynamic plugin started immediately after registration.
	assertHistory(t, history, []string{"start:p1", "start:dyn"})
	if got := len(scope.Plugins()); got != 2 {
		t.Fatalf("Plugins() len = %d, want 2", got)
	}
	// A later scope Stop finalizes both plugins in reverse order.
	if err := scope.Stop(); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
	assertHistory(t, history, []string{"start:p1", "start:dyn", "stop:dyn", "stop:p1"})
}

func TestScopeAddAndStartBeforeStartDegradesToAdd(t *testing.T) {
	var history []string
	scope := &Scope{}
	if err := scope.AddAndStart(&Context{}, &trackingPlugin{name: "dyn", history: &history}); err != nil {
		t.Fatalf("AddAndStart() error = %v", err)
	}
	// The scope was not started: nothing ran yet (degraded to plain Add).
	if len(history) != 0 {
		t.Fatalf("history = %v, want empty (degraded to Add)", history)
	}
	if err := scope.Start(&Context{}); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	assertHistory(t, history, []string{"start:dyn"})
}

func TestScopeAddAndStartAfterDisposeReturnsError(t *testing.T) {
	scope := &Scope{}
	if err := scope.Dispose(); err != nil {
		t.Fatalf("Dispose() error = %v", err)
	}
	if err := scope.AddAndStart(&Context{}, &trackingPlugin{name: "p", history: &[]string{}}); err == nil {
		t.Fatal("AddAndStart() after Dispose error = nil, want explicit error (disposed scope is final)")
	}
}

// failMountPlugin fails a dynamic mount at a configurable phase and records
// Stop/Dispose, for asserting the AddAndStart rollback (Stop + remove).
type failMountPlugin struct {
	name    string
	history *[]string
	errOn   string // "deps" or "start"
}

func (p *failMountPlugin) Deps(*Context) error {
	*p.history = append(*p.history, "deps:"+p.name)
	if p.errOn == "deps" {
		return errors.New(p.name + " deps failed")
	}
	return nil
}

func (p *failMountPlugin) Start(*Context) error {
	*p.history = append(*p.history, "start:"+p.name)
	if p.errOn == "start" {
		return errors.New(p.name + " start failed")
	}
	return nil
}

func (p *failMountPlugin) Stop() error {
	*p.history = append(*p.history, "stop:"+p.name)
	return nil
}

func (p *failMountPlugin) Dispose() error {
	*p.history = append(*p.history, "dispose:"+p.name)
	return nil
}

func TestScopeAddAndStartFailureRollsBack(t *testing.T) {
	for _, phase := range []string{"deps", "start"} {
		t.Run(phase, func(t *testing.T) {
			var history []string
			scope := &Scope{}
			_ = scope.Add(&trackingPlugin{name: "p1", history: &history})
			if err := scope.Start(&Context{}); err != nil {
				t.Fatalf("Start() error = %v", err)
			}
			bad := &failMountPlugin{name: "bad", history: &history, errOn: phase}
			err := scope.AddAndStart(&Context{}, bad)
			if err == nil {
				t.Fatal("AddAndStart() error = nil, want non-nil")
			}
			// The failed plugin was rolled back (Stop) and removed from the
			// slice; the previously started p1 is untouched.
			var want []string
			if phase == "deps" {
				want = []string{"start:p1", "deps:bad", "stop:bad"}
			} else {
				want = []string{"start:p1", "deps:bad", "start:bad", "stop:bad"}
			}
			assertHistory(t, history, want)
			if got := len(scope.Plugins()); got != 1 {
				t.Fatalf("Plugins() len = %d, want 1 (failed plugin removed)", got)
			}
		})
	}
}

// countStartPlugin increments an atomic counter in Start; used by the
// concurrent AddAndStart test.
type countStartPlugin struct {
	started *atomic.Int32
}

func (p *countStartPlugin) Start(*Context) error {
	p.started.Add(1)
	return nil
}

func (p *countStartPlugin) Stop() error { return nil }

func (p *countStartPlugin) Dispose() error { return nil }

func TestScopeAddAndStartConcurrent(t *testing.T) {
	scope := &Scope{}
	if err := scope.Start(&Context{}); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	var started atomic.Int32
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := scope.AddAndStart(&Context{}, &countStartPlugin{started: &started}); err != nil {
				t.Errorf("AddAndStart() error = %v", err)
			}
		}()
	}
	wg.Wait()
	if got := started.Load(); got != 8 {
		t.Fatalf("started = %d, want 8", got)
	}
	if got := len(scope.Plugins()); got != 8 {
		t.Fatalf("Plugins() len = %d, want 8", got)
	}
}

func TestScopePluginsReturnsSnapshot(t *testing.T) {
	scope := &Scope{}
	p1 := &trackingPlugin{name: "p1", history: &[]string{}}
	_ = scope.Add(p1)
	_ = scope.AddWithEnabled(&trackingPlugin{name: "p2", history: &[]string{}}, false)
	snapshot := scope.Plugins()
	if len(snapshot) != 2 {
		t.Fatalf("Plugins() len = %d, want 2", len(snapshot))
	}
	if snapshot[0] != p1 {
		t.Fatalf("Plugins()[0] = %v, want p1", snapshot[0])
	}
	// The snapshot must not alias the live registry.
	snapshot[0] = nil
	if got := scope.Plugins()[0]; got != p1 {
		t.Fatal("Plugins() snapshot aliases the live slice")
	}
}

func TestNoopPluginLifecycleIsNoop(t *testing.T) {
	var p NoopPlugin
	var _ Plugin = p // compile-time assertion: NoopPlugin satisfies Plugin
	if err := p.Start(&Context{}); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	if err := p.Stop(); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
	if err := p.Dispose(); err != nil {
		t.Fatalf("Dispose() error = %v", err)
	}
}
