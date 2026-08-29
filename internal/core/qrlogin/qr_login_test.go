package qrlogin

import "testing"

// TestGetKeyNilJarNoPanic verifies that GetKey tolerates a nil cookie jar
// without panicking. A nil jar is legitimate (e.g. tourist mode / tests); a
// network error — when the network is unavailable — is acceptable and still
// returned as a normal error, never a panic.
func TestGetKeyNilJarNoPanic(t *testing.T) {
	_, _, _ = GetKey(nil) // must not panic, regardless of network availability
}

// TestCheckStatusEmptyKeyShortCircuit verifies that CheckStatus with an empty
// uniKey short-circuits without any network access.
func TestCheckStatusEmptyKeyShortCircuit(t *testing.T) {
	code, resp, err := CheckStatus("", nil)
	if code != 0 || resp != nil || err != nil {
		t.Fatalf("CheckStatus(\"\") = (%v, %v, %v), want (0, nil, nil)", code, resp, err)
	}
}
