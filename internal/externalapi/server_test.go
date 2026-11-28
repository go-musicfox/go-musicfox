package externalapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

type fakeProvider struct {
	mu       sync.Mutex
	snapshot Snapshot
	calls    []string
}

func (p *fakeProvider) ExternalAPISnapshot() Snapshot {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.snapshot
}

func (p *fakeProvider) record(name string) {
	p.mu.Lock()
	p.calls = append(p.calls, name)
	p.mu.Unlock()
}

func (p *fakeProvider) ExternalAPIPlay()     { p.record("play") }
func (p *fakeProvider) ExternalAPIPause()    { p.record("pause") }
func (p *fakeProvider) ExternalAPIStop()     { p.record("stop") }
func (p *fakeProvider) ExternalAPINext()     { p.record("next") }
func (p *fakeProvider) ExternalAPIPrevious() { p.record("prev") }
func (p *fakeProvider) ExternalAPISeek(positionMs int64) {
	p.record(fmt.Sprintf("seek:%d", positionMs))
}
func (p *fakeProvider) ExternalAPISetVolume(volume float64) {
	p.record(fmt.Sprintf("volume:%.2f", volume))
}

func fixtureProvider() *fakeProvider {
	track := &Track{
		ID:       "123",
		Source:   "netease",
		Title:    "Example",
		Artists:  []Artist{{ID: "1", Name: "Artist"}},
		Album:    &Album{ID: "2", Name: "Album", Cover: "https://example.test/cover.jpg"},
		Duration: 180000,
		Cover:    "https://example.test/cover.jpg",
	}
	lyrics := LyricSnapshot{
		TrackID: "123",
		Source:  &LyricSource{Source: "online", Format: "lrc"},
		Lyric:   []LyricLine{{StartTime: 0, EndTime: 4000, Text: "Hello"}},
	}
	status := Status{State: StatePaused, Position: 1000, Duration: 180000, Volume: 0.5, Speed: 1}
	return &fakeProvider{snapshot: Snapshot{
		NowPlaying: NowPlaying{Track: track, Position: 1000, State: StatePaused, Speed: 1},
		Status:     status,
		Lyrics:     lyrics,
	}}
}

func freePort(t *testing.T) int {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port
}

func startTestServer(t *testing.T, provider Provider) (*Server, string) {
	t.Helper()
	port := freePort(t)
	server, err := Start(provider, Options{Enable: true, EnableWS: true, Port: port})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = server.Close(ctx)
	})
	return server, fmt.Sprintf("http://127.0.0.1:%d", port)
}

func TestHTTPStatusAndCommands(t *testing.T) {
	provider := fixtureProvider()
	_, baseURL := startTestServer(t, provider)

	response, err := http.Get(baseURL + "/api/status")
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK || response.Header.Get("Access-Control-Allow-Origin") != "*" {
		t.Fatalf("unexpected status response: %d, cors=%q", response.StatusCode, response.Header.Get("Access-Control-Allow-Origin"))
	}
	var status Status
	if err := json.NewDecoder(response.Body).Decode(&status); err != nil {
		t.Fatal(err)
	}
	if status.State != StatePaused || status.Position != 1000 || status.Duration != 180000 {
		t.Fatalf("unexpected status: %+v", status)
	}
	response, err = http.Get(baseURL + "/api/volume")
	if err != nil {
		t.Fatal(err)
	}
	var volume struct {
		Volume float64 `json:"volume"`
	}
	if err := json.NewDecoder(response.Body).Decode(&volume); err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if volume.Volume != 0.5 {
		t.Fatalf("unexpected volume: %+v", volume)
	}

	response, err = http.Post(baseURL+"/api/seek", "application/json", strings.NewReader(`{}`))
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("missing seek position status = %d", response.StatusCode)
	}

	response, err = http.Post(strings.TrimRight(baseURL, "/")+"/api/seek", "application/json", strings.NewReader(`{"positionMs":2500}`))
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("seek status = %d", response.StatusCode)
	}
	response, err = http.Post(baseURL+"/api/volume", "application/json", strings.NewReader(`{"volume":0.42}`))
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("volume status = %d", response.StatusCode)
	}
	response, err = http.Post(baseURL+"/api/volume", "application/json", strings.NewReader(`{}`))
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("missing volume status = %d", response.StatusCode)
	}

	request, _ := http.NewRequest(http.MethodPost, baseURL+"/api/play", nil)
	request.Header.Set("Origin", "https://evil.example")
	response, err = http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusForbidden {
		t.Fatalf("cross-origin control status = %d", response.StatusCode)
	}

	provider.mu.Lock()
	defer provider.mu.Unlock()
	if len(provider.calls) != 2 || provider.calls[0] != "seek:2500" || provider.calls[1] != "volume:0.42" {
		t.Fatalf("unexpected calls: %v", provider.calls)
	}
}

func TestWebSocketEventsAndControl(t *testing.T) {
	provider := fixtureProvider()
	_, baseURL := startTestServer(t, provider)
	headers := http.Header{"Origin": []string{"file://"}}
	conn, _, err := websocket.DefaultDialer.Dial(
		strings.Replace(baseURL, "http://", "ws://", 1)+"/ws",
		headers,
	)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	_ = conn.SetReadDeadline(time.Now().Add(3 * time.Second))

	seen := map[string]bool{}
	for len(seen) < 4 {
		var message map[string]any
		if err := conn.ReadJSON(&message); err != nil {
			t.Fatal(err)
		}
		switch message["kind"] {
		case "hello":
			seen["hello"] = true
		case "event":
			seen[fmt.Sprint(message["type"])] = true
		}
	}
	for _, expected := range []string{"hello", "track", "lyric", "status"} {
		if !seen[expected] {
			t.Fatalf("missing websocket event %s: %v", expected, seen)
		}
	}

	if err := conn.WriteJSON(map[string]any{"op": "seek"}); err != nil {
		t.Fatal(err)
	}
	for {
		var message map[string]any
		if err := conn.ReadJSON(&message); err != nil {
			t.Fatal(err)
		}
		if message["kind"] == "error" {
			if message["op"] != "seek" {
				t.Fatalf("unexpected error ack: %v", message)
			}
			break
		}
	}

	if err := conn.WriteJSON(map[string]any{"op": "setVolume", "volume": 0.42}); err != nil {
		t.Fatal(err)
	}
	for {
		var message map[string]any
		if err := conn.ReadJSON(&message); err != nil {
			t.Fatal(err)
		}
		if message["kind"] == "ack" {
			if message["op"] != "setVolume" {
				t.Fatalf("unexpected ack: %v", message)
			}
			break
		}
	}
	provider.mu.Lock()
	defer provider.mu.Unlock()
	if len(provider.calls) != 1 || provider.calls[0] != "volume:0.42" {
		t.Fatalf("unexpected calls: %v", provider.calls)
	}
}

func TestWebSocketDisabled(t *testing.T) {
	port := freePort(t)
	server, err := Start(fixtureProvider(), Options{Enable: true, EnableWS: false, Port: port})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = server.Close(ctx)
	}()
	response, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/ws", port))
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusForbidden {
		t.Fatalf("ws status = %d", response.StatusCode)
	}
}

func TestAllowedOrigin(t *testing.T) {
	tests := []struct {
		origin string
		want   bool
	}{
		{origin: "", want: true},
		{origin: "null", want: true},
		{origin: "file://", want: true},
		{origin: "file:///D:/TMP/splayer-obs/index.html", want: true},
		{origin: "http://127.0.0.1:6161", want: true},
		{origin: "http://localhost:6161", want: true},
		{origin: "https://evil.example", want: false},
	}
	for _, test := range tests {
		if got := allowedOrigin(test.origin); got != test.want {
			t.Errorf("allowedOrigin(%q) = %v, want %v", test.origin, got, test.want)
		}
	}
}
