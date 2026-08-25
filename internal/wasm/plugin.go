package wasm

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	wazero "github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
)

// DefaultTimeout is how long a single plugin export call may run before the
// instance is closed and ErrTimeout is returned.
const DefaultTimeout = 5 * time.Second

// ErrTimeout is returned when a plugin export call exceeds its timeout. The
// plugin instance is closed after a timeout and cannot be reused.
var ErrTimeout = errors.New("wasm: plugin call timed out")

// ErrPluginClosed is returned when Run is called on a plugin whose instance
// was closed, either explicitly or by the timeout watchdog.
var ErrPluginClosed = errors.New("wasm: plugin instance is closed")

// Plugin is a loaded WASM plugin instance. It is safe for concurrent use: the
// internal mutex serializes export calls, as api.Function.Call is not
// goroutine-safe per instance.
type Plugin struct {
	ID    string
	Name  string
	Menus []MenuDecl

	timeout time.Duration
	runtime wazero.Runtime // owned by the Manager; the plugin holds a reference
	mod     api.Module
	alloc   api.Function            // guest-exported alloc(size uint32) uint32
	dealloc api.Function            // guest-exported dealloc(ptr, size uint32)
	exports map[string]api.Function // per-menu export lookup
	mu      sync.Mutex              // serializes calls
}

// newPlugin compiles and instantiates wasmBytes as a reactor module named by
// the manifest id, then resolves the memory allocator and every menu export.
func newPlugin(runtime wazero.Runtime, manifest *Manifest, wasmBytes []byte) (*Plugin, error) {
	return newPluginWithTimeout(runtime, manifest, wasmBytes, DefaultTimeout)
}

// newPluginWithTimeout is newPlugin with a caller-supplied per-call timeout,
// used by tests to exercise the watchdog without waiting DefaultTimeout.
func newPluginWithTimeout(runtime wazero.Runtime, manifest *Manifest, wasmBytes []byte, timeout time.Duration) (*Plugin, error) {
	ctx := context.Background()

	compiled, err := runtime.CompileModule(ctx, wasmBytes)
	if err != nil {
		return nil, fmt.Errorf("wasm: compile plugin %q: %w", manifest.ID, err)
	}

	// Go reactor builds export _initialize as their runtime entrypoint. wazero
	// does not call it automatically: without it, exported calls trap with
	// "wasmexport function called before runtime initialization".
	mod, err := runtime.InstantiateModule(ctx, compiled, wazero.NewModuleConfig().
		WithName(manifest.ID).
		WithStartFunctions("_initialize").
		WithSysWalltime().
		WithSysNanotime())
	if err != nil {
		return nil, fmt.Errorf("wasm: instantiate plugin %q: %w", manifest.ID, err)
	}

	p := &Plugin{
		ID:      manifest.ID,
		Name:    manifest.Name,
		Menus:   manifest.Menus,
		timeout: timeout,
		runtime: runtime,
		mod:     mod,
		exports: make(map[string]api.Function, len(manifest.Menus)),
	}

	if p.alloc = mod.ExportedFunction("alloc"); p.alloc == nil {
		_ = mod.Close(ctx)
		return nil, fmt.Errorf("wasm: plugin %q does not export alloc", manifest.ID)
	}
	if p.dealloc = mod.ExportedFunction("dealloc"); p.dealloc == nil {
		_ = mod.Close(ctx)
		return nil, fmt.Errorf("wasm: plugin %q does not export dealloc", manifest.ID)
	}
	for _, menu := range manifest.Menus {
		export := mod.ExportedFunction(menu.Export)
		if export == nil {
			_ = mod.Close(ctx)
			return nil, fmt.Errorf("wasm: plugin %q menu %q references missing export %q", manifest.ID, menu.Key, menu.Export)
		}
		p.exports[menu.Export] = export
	}

	return p, nil
}

// Run invokes the export declared by menu with reqJSON, following the wazero
// allocation protocol:
//
//  1. alloc a guest buffer and write the request into it;
//  2. call the export with (ptr, len);
//  3. the guest packs (outPtr<<32)|outLen into its single uint64 result;
//  4. copy the response out of guest memory, then dealloc both buffers.
//
// Go's wasmexport ABI supports at most one result value, hence the packed
// uint64. A hung plugin is interrupted after p.timeout and the instance is
// closed.
func (p *Plugin) Run(ctx context.Context, menu MenuDecl, reqJSON []byte) ([]byte, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.mod == nil || p.mod.IsClosed() {
		return nil, fmt.Errorf("%w: plugin %q", ErrPluginClosed, p.ID)
	}

	type outcome struct {
		out []byte
		err error
	}
	done := make(chan outcome, 1)
	go func() {
		out, err := p.callLocked(ctx, menu, reqJSON)
		done <- outcome{out: out, err: err}
	}()

	timeout := p.timeout
	if timeout <= 0 {
		timeout = DefaultTimeout
	}
	select {
	case o := <-done:
		return o.out, o.err
	case <-time.After(timeout):
		// Interrupt the in-flight call (it returns a sys.ExitError to the
		// caller goroutine) and mark the instance closed. This is best-effort
		// per wazero's concurrency model: with WithCloseOnContextDone the
		// close terminates running executions.
		_ = p.mod.Close(ctx)
		return nil, fmt.Errorf("%w: plugin %q", ErrTimeout, p.ID)
	}
}

// Close releases the plugin instance. It is idempotent.
func (p *Plugin) Close(ctx context.Context) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.mod != nil && !p.mod.IsClosed() {
		_ = p.mod.Close(ctx)
	}
}

// callLocked performs the actual host<->guest call. p.mu must be held.
func (p *Plugin) callLocked(ctx context.Context, menu MenuDecl, reqJSON []byte) ([]byte, error) {
	export := p.exports[menu.Export]
	if export == nil {
		return nil, fmt.Errorf("wasm: plugin %q has no export %q", p.ID, menu.Export)
	}

	size := uint32(len(reqJSON))
	res, err := p.alloc.Call(ctx, uint64(size))
	if err != nil {
		return nil, fmt.Errorf("wasm: plugin %q alloc: %w", p.ID, err)
	}
	if len(res) != 1 {
		return nil, fmt.Errorf("wasm: plugin %q alloc returned %d results, want 1", p.ID, len(res))
	}
	reqPtr := uint32(res[0])

	// Always release the request buffer, on every path below.
	defer func() {
		if _, err := p.dealloc.Call(ctx, uint64(reqPtr), uint64(size)); err != nil {
			// Best-effort: the plugin may already be closed.
		}
	}()

	if !p.mod.Memory().Write(reqPtr, reqJSON) {
		return nil, fmt.Errorf("wasm: plugin %q memory write out of range: ptr=%d size=%d", p.ID, reqPtr, size)
	}

	res, err = export.Call(ctx, uint64(reqPtr), uint64(size))
	if err != nil {
		return nil, fmt.Errorf("wasm: plugin %q export %q failed: %w", p.ID, menu.Export, err)
	}
	if len(res) != 1 {
		return nil, fmt.Errorf("wasm: plugin %q export %q returned %d results, want 1", p.ID, menu.Export, len(res))
	}

	packed := res[0]
	outPtr := uint32(packed >> 32)
	outLen := uint32(packed)

	out, ok := p.mod.Memory().Read(outPtr, outLen)
	if !ok {
		return nil, fmt.Errorf("wasm: plugin %q memory read out of range: ptr=%d size=%d", p.ID, outPtr, outLen)
	}
	// Read returns a write-through view of guest memory; copy it into a fresh
	// slice before releasing the buffers.
	copied := make([]byte, len(out))
	copy(copied, out)

	if _, err := p.dealloc.Call(ctx, uint64(outPtr), uint64(outLen)); err != nil {
		// Best-effort: the plugin may already be closed.
	}

	return copied, nil
}
