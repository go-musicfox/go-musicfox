package slogx

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/go-musicfox/go-musicfox/utils/app"
)

func init() {
	dir := app.LogDir()

	f, err := os.OpenFile(filepath.Join(dir, "musicfox.log"), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	if err != nil {
		panic(fmt.Sprintf("failed to open log file, err: %v", err))
	}

	logger := slog.New(slog.NewTextHandler(f, &slog.HandlerOptions{AddSource: true}))

	log.SetOutput(f)
	slog.SetDefault(logger)
}

func Error(err any) slog.Attr {
	if err == nil {
		return slog.Attr{}
	}

	return slog.String("error", fmt.Sprintf("%+v", err))
}

func Bytes(k string, b []byte) slog.Attr {
	return slog.String(k, string(b))
}
