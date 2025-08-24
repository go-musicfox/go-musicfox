package app

import (
	"path/filepath"

	"github.com/skip2/go-qrcode"
)

func GenQRCode(filename, content string) (string, error) {
	localDir := RuntimeDir()
	qrcodePath := filepath.Join(localDir, filename)
	if err := qrcode.WriteFile(content, qrcode.Medium, 256, qrcodePath); err != nil {
		return "", err
	}
	return qrcodePath, nil
}
