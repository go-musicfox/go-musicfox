package core

// Protocol version for the control channel. Bump it whenever the wire format
// (Request/Response) changes incompatibly.
const ProtocolVersion = 1

// Request is a single control request on a control channel.
type Request struct {
	V    int            `json:"v"` // protocol version, must be ProtocolVersion
	ID   int64          `json:"id"`
	Cmd  string         `json:"cmd"`
	Args map[string]any `json:"args,omitempty"`
}

// Response is the single reply correlated with a Request by ID.
type Response struct {
	V     int    `json:"v"` // protocol version
	ID    int64  `json:"id"`
	Ok    bool   `json:"ok"`
	Data  any    `json:"data,omitempty"`
	Error string `json:"error,omitempty"`
}
