package utils

import (
	"log"
	"os"
)

var logger *log.Logger

func SetLogger(l *log.Logger) {
	logger = l
}

func Logger() *log.Logger {
	if logger != nil {
		return logger
	}

	dir := GetLocalDataDir()
	logFile, err := os.OpenFile(dir+"/musicfox.log", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	if err != nil {
		return log.Default()
	}

	logger = log.New(logFile, "", log.LstdFlags)
	return logger
}
