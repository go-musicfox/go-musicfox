package configs

type LastfmOptions struct {
	Key      string // API Key
	Secret   string // API Shared Secret
	Scrobble bool   // 是否启用，默认 true
}
