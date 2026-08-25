// Package main is a runnable example WASM plugin for go-musicfox.
//
// It is a standard Go wasip1 reactor. Build it with:
//
//	GOOS=wasip1 GOARCH=wasm go build -buildmode=c-shared -o main.wasm .
//
// The host (internal/wasm) instantiates the module with wazero, calls the
// _initialize reactor entrypoint, then invokes the exported "run" function
// with a JSON request and receives a JSON response.
//
// The memory protocol follows the official wazero allocation example:
//
//   - alloc(size) returns a linear-memory offset for a size-byte buffer.
//   - run(reqPtr, reqLen) packs (outPtr<<32)|outLen into its single uint64
//     result (Go's wasmexport ABI allows at most one result value).
//   - dealloc(ptr, size) releases a buffer previously returned by alloc.
package main

import (
	"encoding/json"
	"fmt"
	"sync"
	"unsafe"
)

// allocs keeps every allocation alive until dealloc is called. Go's GC is
// non-moving (also on wasm), so the uint32 linear-memory offset stays valid as
// long as the backing slice is referenced here.
var (
	allocsMu sync.Mutex
	allocs   = map[uint32][]byte{}
)

// alloc exports a size -> offset allocator for host-provided buffers.
//
//go:wasmexport alloc
func alloc(size uint32) uint32 {
	if size == 0 {
		size = 1 // &buf[0] below must be addressable
	}
	buf := make([]byte, size)
	ptr := uint32(uintptr(unsafe.Pointer(&buf[0])))
	allocsMu.Lock()
	allocs[ptr] = buf
	allocsMu.Unlock()
	return ptr
}

// dealloc releases a buffer previously returned by alloc.
//
//go:wasmexport dealloc
func dealloc(ptr uint32, _ uint32) {
	allocsMu.Lock()
	delete(allocs, ptr)
	allocsMu.Unlock()
}

// The JSON contract shapes below are duplicated from internal/wasm/contract.go
// so the example stays a standalone plugin source tree that does not import
// the host package.
type request struct {
	Version int            `json:"version"`
	Action  string         `json:"action"`
	Args    map[string]any `json:"args,omitempty"`
	Context requestContext `json:"context,omitempty"`
}

type requestContext struct {
	UserID   int64     `json:"userId,omitempty"`
	UserName string    `json:"userName,omitempty"`
	Playing  bool      `json:"playing"`
	Song     *songInfo `json:"song,omitempty"`
}

type songInfo struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Artist string `json:"artist"`
	Album  string `json:"album"`
}

type response struct {
	Action  string   `json:"action"`
	Title   string   `json:"title,omitempty"`
	Message string   `json:"message,omitempty"`
	Level   string   `json:"level,omitempty"`
	URL     string   `json:"url,omitempty"`
	Command string   `json:"command,omitempty"`
	Args    []string `json:"args,omitempty"`
}

// run is the main plugin entry point. It reads the JSON request from guest
// memory, echoes a greeting built from request.Args["name"] (and the context
// song, if any) and returns the packed response.
//
//go:wasmexport run
func run(reqPtr uint32, reqLen uint32) uint64 {
	reqBytes := unsafe.Slice((*byte)(unsafe.Pointer(uintptr(reqPtr))), reqLen)
	var req request
	if err := json.Unmarshal(reqBytes, &req); err != nil {
		return pack(errorResponse("invalid request JSON: " + err.Error()))
	}

	name, _ := req.Args["name"].(string)
	if name == "" {
		name = "world"
	}
	argsJSON, _ := json.Marshal(req.Args)
	msg := fmt.Sprintf("Hello, %s! args=%s", name, argsJSON)
	if req.Context.Song != nil {
		msg += fmt.Sprintf(" song=%s - %s", req.Context.Song.Name, req.Context.Song.Artist)
	}

	resp := response{
		Action:  "view",
		Title:   "WASM Hello",
		Message: msg,
	}
	if action, _ := req.Args["action"].(string); action == "toast" {
		resp.Action = "toast"
		resp.Level = "info"
	}

	out, err := json.Marshal(resp)
	if err != nil {
		return pack(errorResponse("marshal response: " + err.Error()))
	}
	return pack(out)
}

// hang loops forever and is used by the host's timeout watchdog test. It
// matches the run ABI so the host can invoke it through the standard call
// protocol.
//
//go:wasmexport hang
func hang(_ uint32, _ uint32) uint64 {
	for {
	}
}

func errorResponse(message string) []byte {
	out, _ := json.Marshal(response{Action: "view", Title: "WASM Hello", Message: message})
	return out
}

// pack allocates a guest buffer, copies respBytes into it and returns the
// (ptr<<32)|len packing used as the single wasmexport result.
func pack(respBytes []byte) uint64 {
	if len(respBytes) == 0 {
		return 0
	}
	ptr := alloc(uint32(len(respBytes)))
	allocsMu.Lock()
	buf := allocs[ptr]
	allocsMu.Unlock()
	copy(buf, respBytes)
	return (uint64(ptr) << 32) | uint64(len(respBytes))
}

// main is required for package main to compile; reactor builds ignore it.
func main() {}
