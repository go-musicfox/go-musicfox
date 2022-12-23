package utils

import (
	"io"
	"log"
	"os"
	"path"
)

var (
	logger    *log.Logger
	logWriter io.Writer
)

func SetLogger(l *log.Logger) {
	logger = l
}

func Logger() *log.Logger {
	if logger != nil {
		return logger
	}

	dir := GetLocalDataDir()
	var err error
	logWriter, err = os.OpenFile(path.Join(dir, "musicfox.log"), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	if err != nil {
		return log.Default()
	}

	logger = log.New(logWriter, "", log.LstdFlags)
	return logger
}

func LogWriter() io.Writer {
	if logWriter == nil {
		Logger()
	}
	return logWriter
}
