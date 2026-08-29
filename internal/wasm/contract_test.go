package wasm

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestRequestJSONRoundTrip(t *testing.T) {
	req := Request{
		Version: ProtocolVersion,
		Action:  "wasm_hello",
		Args:    map[string]any{"name": "musicfox", "count": float64(3)},
		Context: RequestContext{
			UserID:   123,
			UserName: "musicfox",
			Playing:  true,
			Song:     &SongInfo{ID: 1, Name: "Song", Artist: "Artist", Album: "Album"},
		},
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got Request
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !reflect.DeepEqual(got, req) {
		t.Fatalf("round trip mismatch:\n got: %+v\nwant: %+v", got, req)
	}
}

func TestRequestEmptyContext(t *testing.T) {
	// The omitempty context and its nested fields must survive a round trip
	// as zero values so plugins see a well-formed request.
	req := Request{Version: ProtocolVersion, Action: "x"}
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got Request
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.Context.Playing {
		t.Fatalf("empty context Playing = true, want false")
	}
	if got.Context.Song != nil {
		t.Fatalf("empty context Song = %+v, want nil", got.Context.Song)
	}
}

func TestResponseJSONRoundTrip(t *testing.T) {
	resp := Response{
		Action:  "toast",
		Title:   "Done",
		Message: "All good",
		Level:   "success",
	}
	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got Response
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !reflect.DeepEqual(got, resp) {
		t.Fatalf("round trip mismatch:\n got: %+v\nwant: %+v", got, resp)
	}
}

func TestResponseEmptyFieldsOmitted(t *testing.T) {
	resp := Response{Action: "open_url", URL: "https://example.com"}
	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if _, ok := m["title"]; ok {
		t.Fatalf("empty title field should be omitted, got %v", m)
	}
	if _, ok := m["args"]; ok {
		t.Fatalf("empty args field should be omitted, got %v", m)
	}
}

func TestResponseLevelValid(t *testing.T) {
	cases := []struct {
		level string
		want  bool
	}{
		{"", true},
		{"info", true},
		{"success", true},
		{"warning", true},
		{"error", true},
		{"INFO", false},
		{"bogus", false},
	}
	for _, c := range cases {
		resp := Response{Action: "toast", Level: c.level}
		if got := resp.LevelValid(); got != c.want {
			t.Fatalf("LevelValid(%q) = %v, want %v", c.level, got, c.want)
		}
	}
}
