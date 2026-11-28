package configs

// ExternalAPIConfig configures the SPlayer-Next compatible HTTP/WebSocket API.
// The service is disabled by default because it exposes playback control.
type ExternalAPIConfig struct {
	Enable   bool `koanf:"enable"`
	EnableWS bool `koanf:"enableWs"`
	AllowLAN bool `koanf:"allowLan"`
	Port     int  `koanf:"port"`
}
