package app

import (
	"bytes"
	"path/filepath"

	"github.com/mdp/qrterminal/v3"
	"github.com/skip2/go-qrcode"
)

func GenQRCode(filename, content string) (path string, buffer bytes.Buffer, err error) {
	localDir := RuntimeDir()
	path = filepath.Join(localDir, filename)
	if err = qrcode.WriteFile(content, qrcode.Medium, 256, path); err != nil {
		return "", bytes.Buffer{}, err
	}
	config := qrterminal.Config{
		Level:      qrterminal.L,
		Writer:     &buffer,
		HalfBlocks: true,
		QuietZone:  1,
	}
	qrterminal.GenerateWithConfig(content, config)
	return path, buffer, nil
}
