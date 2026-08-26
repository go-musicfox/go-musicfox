package core

import (
	"encoding/json"
	"testing"
)

func TestRequestResponseJSONRoundTrip(t *testing.T) {
	req := Request{V: ProtocolVersion, ID: 42, Cmd: "seek", Args: map[string]any{"seconds": float64(30)}}
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal Request: %v", err)
	}
	var got Request
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal Request: %v", err)
	}
	if got.V != ProtocolVersion || got.ID != 42 || got.Cmd != "seek" {
		t.Fatalf("Request round-trip mismatch: %+v", got)
	}
	if got.Args["seconds"] != float64(30) {
		t.Fatalf("Request args lost: %+v", got.Args)
	}

	resp := Response{V: ProtocolVersion, ID: 42, Ok: true, Data: map[string]any{"volume": float64(60)}}
	data, err = json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal Response: %v", err)
	}
	var gotResp Response
	if err := json.Unmarshal(data, &gotResp); err != nil {
		t.Fatalf("unmarshal Response: %v", err)
	}
	if !gotResp.Ok || gotResp.ID != 42 || gotResp.Error != "" {
		t.Fatalf("Response round-trip mismatch: %+v", gotResp)
	}
	if gotResp.Data.(map[string]any)["volume"] != float64(60) {
		t.Fatalf("Response data lost: %+v", gotResp.Data)
	}
}
