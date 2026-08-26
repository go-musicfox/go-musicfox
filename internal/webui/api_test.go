package webui

import (
	"encoding/json"
	"net/http"
	"testing"
)

// TestAPIStatus verifies /api/status returns the dispatcher snapshot wrapped in
// {"ok":true,"data":{...}} (same shape as the WS snapshot frame).
func TestAPIStatus(t *testing.T) {
	s, ts := newWSServer(t)
	resp := doAuthedRequest(t, ts, "/api/status", s.token, "", "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	var body struct {
		OK   bool           `json:"ok"`
		Data map[string]any `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if !body.OK {
		t.Fatal("ok = false, want true")
	}
	if body.Data == nil {
		t.Fatal("data missing")
	}
	if _, ok := body.Data["song"]; !ok {
		t.Fatalf("data missing song: %v", body.Data)
	}
}

// TestAPIStatusUnauthorized verifies a missing token cookie answers 401.
func TestAPIStatusUnauthorized(t *testing.T) {
	_, ts := newWSServer(t)
	resp := doAuthedRequest(t, ts, "/api/status", "", "", "")
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", resp.StatusCode)
	}
}

// TestAPILyricsEmpty verifies /api/lyrics with no song returns an empty
// structure (200, not 500).
func TestAPILyricsEmpty(t *testing.T) {
	s, ts := newWSServer(t)
	resp := doAuthedRequest(t, ts, "/api/lyrics", s.token, "", "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	for _, key := range []string{"fragments", "translatedFragments", "currentIndex", "offsetMs"} {
		if _, ok := body[key]; !ok {
			t.Fatalf("lyrics body missing key %q: %v", key, body)
		}
	}
	if frags, ok := body["fragments"].([]any); ok && len(frags) != 0 {
		t.Fatalf("fragments = %v, want empty", frags)
	}
}

// TestAPIAlbumArtNoSong verifies /api/albumart with no current song answers
// 404 (no upstream fetch is attempted).
func TestAPIAlbumArtNoSong(t *testing.T) {
	s, ts := newWSServer(t)
	resp := doAuthedRequest(t, ts, "/api/albumart", s.token, "", "")
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}
