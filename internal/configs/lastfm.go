package configs

type LastfmOptions struct {
	Enable        bool   // 是否启用
	Key           string // API Key
	Secret        string // API Shared Secret
	ScrobblePoint int    // 上报百分比
}
