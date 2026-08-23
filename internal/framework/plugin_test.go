package framework

import (
	"errors"
	"reflect"
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
		scope.Add(&trackingPlugin{name: name, history: &history})
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
	scope.Add(p)
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
		scope.Add(&trackingPlugin{name: name, history: &history})
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
	parent.Add(&trackingPlugin{name: "parent", history: &history})
	child.Add(&trackingPlugin{name: "child", history: &history})

	if err := parent.Start(&Context{}); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	if err := parent.Dispose(); err != nil {
		t.Fatalf("Dispose() error = %v", err)
	}
	assertHistory(t, history, []string{"start:parent", "start:child", "dispose:child", "dispose:parent"})
}

func TestChildScopeFollowsParentLifecycle(t *testing.T) {
	var history []string
	parent := &Scope{}
	child := parent.NewScope()
	parent.Add(&trackingPlugin{name: "parent", history: &history})
	child.Add(&trackingPlugin{name: "child", history: &history})

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
	root.Add(&trackingPlugin{name: "root", history: &history})
	mid.Add(&trackingPlugin{name: "mid", history: &history})
	leaf.Add(&trackingPlugin{name: "leaf", history: &history})

	if err := root.Start(&Context{}); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	assertHistory(t, history, []string{"start:root", "start:mid", "start:leaf"})

	if err := root.Dispose(); err != nil {
		t.Fatalf("Dispose() error = %v", err)
	}
	assertHistory(t, history, []string{"start:root", "start:mid", "start:leaf", "dispose:leaf", "dispose:mid", "dispose:root"})
}

func TestScopeDisposeDetachesChildrenAndIsIdempotent(t *testing.T) {
	var history []string
	parent := &Scope{}
	child := parent.NewScope()
	parent.Add(&trackingPlugin{name: "parent", history: &history})
	child.Add(&trackingPlugin{name: "child", history: &history})

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
	scope.Add(&trackingPlugin{name: "p1", history: &history})
	scope.Add(&failingPlugin{name: "bad", history: &history, errOn: "start"})
	scope.Add(&trackingPlugin{name: "p3", history: &history})

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
	scope.Add(&trackingPlugin{name: "p1", history: &history})
	scope.Add(&failingPlugin{name: "bad", history: &history, errOn: "deps"})
	scope.Add(&trackingPlugin{name: "p3", history: &history})

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
	scope.Add(&trackingPlugin{name: "root", history: &history})
	child1 := scope.NewScope()
	child1.Add(&trackingPlugin{name: "c1", history: &history})
	child2 := scope.NewScope()
	child2.Add(&failingPlugin{name: "c2bad", history: &history, errOn: "start"})
	child3 := scope.NewScope()
	child3.Add(&trackingPlugin{name: "c3", history: &history})

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
	child.Add(&trackingPlugin{name: "child", history: &history})
	deep := child.NewScope()
	deep.Add(&failingPlugin{name: "deepbad", history: &history, errOn: "start"})

	err := scope.Start(&Context{})
	if err == nil {
		t.Fatal("Start() error = nil, want non-nil")
	}
	// child started, then its nested deep scope failed; child must be stopped.
	assertHistory(t, history, []string{"start:child", "deps:deepbad", "start:deepbad", "stop:child"})
}
