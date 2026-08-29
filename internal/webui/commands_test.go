package webui

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-musicfox/go-musicfox/internal/configs"
	"github.com/go-musicfox/go-musicfox/internal/frontend"
)

// Test command keys are prefixed with testCommandPrefix and suffixed per test
// so they never collide with real (T4/T8) registrations. The frontend registry
// is package-global and intentionally left populated after the tests (existing
// convention); duplicate keys would panic, hence the per-test suffixes.
const testCommandPrefix = "t6test-"

// registerCommandForTest registers a command under a derived key and returns
// the key. PluginID is left empty unless the test sets it explicitly (empty =
// not subject to [plugins] disabled filtering).
func registerCommandForTest(t *testing.T, suffix string, cmd frontend.Command) string {
	t.Helper()
	key := testCommandPrefix + suffix
	cmd.Key = key
	frontend.RegisterCommand(cmd)
	return key
}

// doAuthedPost issues a POST against path with an optional token cookie.
func doAuthedPost(t *testing.T, ts *httptest.Server, path, cookie string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodPost, ts.URL+path, nil)
	if err != nil {
		t.Fatalf("NewRequest(%s): %v", path, err)
	}
	if cookie != "" {
		req.AddCookie(&http.Cookie{Name: tokenCookieName, Value: cookie})
	}
	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		t.Fatalf("POST %s: %v", path, err)
	}
	t.Cleanup(func() { _ = resp.Body.Close() })
	return resp
}

// TestCommandListRequiresAuth verifies GET /api/commands is behind the auth
// middleware (no cookie → 401).
func TestCommandListRequiresAuth(t *testing.T) {
	_, ts := newWSServer(t)
	resp := doAuthedRequest(t, ts, "/api/commands", "", "", "")
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", resp.StatusCode)
	}
}

// TestCommandList includes a registered command (PluginID empty = enabled).
func TestCommandList(t *testing.T) {
	key := registerCommandForTest(t, "list-visible", frontend.Command{
		Title: "测试命令",
		Run: func(frontend.CommandContext) frontend.CommandResult {
			return frontend.CommandResult{Action: "toast", Message: "hello"}
		},
	})
	s, ts := newWSServer(t)
	resp := doAuthedRequest(t, ts, "/api/commands", s.token, "", "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	var body struct {
		OK   bool          `json:"ok"`
		Data []commandItem `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if !body.OK {
		t.Fatal("ok = false, want true")
	}
	var found bool
	for _, item := range body.Data {
		if item.Key == key {
			found = true
			if item.Title != "测试命令" {
				t.Fatalf("title = %q, want 测试命令", item.Title)
			}
		}
	}
	if !found {
		t.Fatalf("list missing command %q: %+v", key, body.Data)
	}
}

// TestCommandListHidesWhenShowFalse verifies a command whose Show gate reports
// false is absent from the list.
func TestCommandListHidesWhenShowFalse(t *testing.T) {
	key := registerCommandForTest(t, "list-hidden", frontend.Command{
		Title: "隐藏命令",
		Show:  func(frontend.CommandContext) bool { return false },
		Run:   func(frontend.CommandContext) frontend.CommandResult { return frontend.CommandResult{} },
	})
	s, ts := newWSServer(t)
	resp := doAuthedRequest(t, ts, "/api/commands", s.token, "", "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	var body struct {
		OK   bool          `json:"ok"`
		Data []commandItem `json:"data"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&body)
	for _, item := range body.Data {
		if item.Key == key {
			t.Fatalf("list contains hidden command %q", key)
		}
	}
}

// TestCommandExecToast verifies POST runs the command and echoes its result
// fields back in the {ok:true, action, title, message, level} envelope.
func TestCommandExecToast(t *testing.T) {
	key := registerCommandForTest(t, "exec-toast", frontend.Command{
		Run: func(frontend.CommandContext) frontend.CommandResult {
			return frontend.CommandResult{Action: "toast", Title: "T", Message: "hi", Level: "success"}
		},
	})
	s, ts := newWSServer(t)
	resp := doAuthedPost(t, ts, "/api/commands/"+key, s.token)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	var body struct {
		OK      bool   `json:"ok"`
		Action  string `json:"action"`
		Title   string `json:"title"`
		Message string `json:"message"`
		Level   string `json:"level"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&body)
	if !body.OK || body.Action != "toast" || body.Title != "T" || body.Message != "hi" || body.Level != "success" {
		t.Fatalf("body = %+v", body)
	}
}

// TestCommandExecExecForbidden verifies an "exec" result is rejected with 403
// by the WebUI side-effect policy (commandExecAllowed = false).
func TestCommandExecExecForbidden(t *testing.T) {
	key := registerCommandForTest(t, "exec-exec", frontend.Command{
		Run: func(frontend.CommandContext) frontend.CommandResult {
			return frontend.CommandResult{Action: "exec", Command: "shutdown"}
		},
	})
	s, ts := newWSServer(t)
	resp := doAuthedPost(t, ts, "/api/commands/"+key, s.token)
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", resp.StatusCode)
	}
}

// TestCommandExecShowFalse verifies the Show gate is enforced at execution
// time: a hidden command answers 403 rather than running.
func TestCommandExecShowFalse(t *testing.T) {
	key := registerCommandForTest(t, "exec-showfalse", frontend.Command{
		Show: func(frontend.CommandContext) bool { return false },
		Run:  func(frontend.CommandContext) frontend.CommandResult { return frontend.CommandResult{Action: "toast"} },
	})
	s, ts := newWSServer(t)
	resp := doAuthedPost(t, ts, "/api/commands/"+key, s.token)
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", resp.StatusCode)
	}
}

// TestCommandExecUnknownKey verifies an unregistered key answers 404.
func TestCommandExecUnknownKey(t *testing.T) {
	s, ts := newWSServer(t)
	resp := doAuthedPost(t, ts, "/api/commands/"+testCommandPrefix+"no-such-key", s.token)
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

// TestCommandDisabledPlugin verifies a command attributed to a disabled plugin
// (via [plugins] disabled) is hidden from the list and answers 404 on exec, so
// the endpoint does not leak the existence of disabled commands.
func TestCommandDisabledPlugin(t *testing.T) {
	pluginID := testCommandPrefix + "disabled-plugin"
	key := registerCommandForTest(t, "exec-disabled", frontend.Command{
		PluginID: pluginID,
		Title:    "被禁用命令",
		Run:      func(frontend.CommandContext) frontend.CommandResult { return frontend.CommandResult{Action: "toast"} },
	})

	prev := configs.AppConfig
	configs.AppConfig = &configs.Config{Plugins: configs.PluginsConfig{Disabled: []string{pluginID}}}
	t.Cleanup(func() { configs.AppConfig = prev })

	s, ts := newWSServer(t)
	resp := doAuthedRequest(t, ts, "/api/commands", s.token, "", "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("list status = %d, want 200", resp.StatusCode)
	}
	var body struct {
		OK   bool          `json:"ok"`
		Data []commandItem `json:"data"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&body)
	for _, item := range body.Data {
		if item.Key == key {
			t.Fatalf("list contains disabled command %q", key)
		}
	}

	execResp := doAuthedPost(t, ts, "/api/commands/"+key, s.token)
	if execResp.StatusCode != http.StatusNotFound {
		t.Fatalf("exec status = %d, want 404", execResp.StatusCode)
	}
}
