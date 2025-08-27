package netease

import (
	"fmt"
	"net/url"
	"strings"
)

var bannedLinkFeatures = []string{
	"/resource/n2/73/84/3759149332.mp3",
}

// HasBannedPathSuffix 检查是否匹配需要排除的特征
func HasBannedPathSuffix(rawURL string) (bool, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return false, fmt.Errorf("无法解析 URL: %w", err)
	}

	for _, suffix := range bannedLinkFeatures {
		if strings.HasSuffix(parsedURL.Path, suffix) {
			return true, nil
		}
	}

	return false, nil
}
