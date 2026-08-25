package headless

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"sync/atomic"
	"time"
)

// ctrlID is a package-level atomic counter that supplies fresh Request IDs to
// the CtrlClient.
var ctrlID atomic.Int64

// CtrlClient connects to a running headless daemon's control channel. The
// server reads exactly one request per connection and closes it, so Call opens
// a fresh connection for every invocation.
type CtrlClient struct {
	network string
	addr    string
}

// Dial resolves the control-channel address and probes it for a live daemon.
// It returns an error (reported as "无头模式未在运行" by the CLI) when no
// daemon is running.
func Dial() (*CtrlClient, error) {
	network, addr := DialAddr()
	if network == "" || addr == "" {
		return nil, errors.New("headless daemon is not running")
	}
	conn, err := net.DialTimeout(network, addr, time.Second)
	if err != nil {
		return nil, errors.New("headless daemon is not running")
	}
	_ = conn.Close()
	return &CtrlClient{network: network, addr: addr}, nil
}

// Close releases the client. The current implementation keeps no persistent
// connection (each Call dials fresh), so this is a no-op retained for API
// symmetry.
func (c *CtrlClient) Close() error {
	return nil
}

// Call writes a Request with a fresh id on a new connection, then reads the
// correlated Response. The read is bounded by a 3s deadline (or an earlier ctx
// deadline).
func (c *CtrlClient) Call(ctx context.Context, cmd string, args map[string]any) (*Response, error) {
	conn, err := net.DialTimeout(c.network, c.addr, time.Second)
	if err != nil {
		return nil, errors.New("headless daemon is not running")
	}
	defer func() { _ = conn.Close() }()

	req := Request{V: ProtocolVersion, ID: ctrlID.Add(1), Cmd: cmd, Args: args}
	if err := json.NewEncoder(conn).Encode(req); err != nil {
		return nil, err
	}

	deadline := time.Now().Add(3 * time.Second)
	if d, ok := ctx.Deadline(); ok && d.Before(deadline) {
		deadline = d
	}
	_ = conn.SetReadDeadline(deadline)

	var resp Response
	if err := json.NewDecoder(conn).Decode(&resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
