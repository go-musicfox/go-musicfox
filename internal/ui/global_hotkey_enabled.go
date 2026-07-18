//go:build enable_global_hotkey && !purego

package ui

/*
#cgo CFLAGS: -I${SRCDIR}/../../vendor/github.com/robotn/gohook/hook

#include <stdarg.h>
#include <stdbool.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include "logger.h"

static FILE *gHookLogFile = NULL;
static unsigned int gMinLogLevel = LOG_LEVEL_INFO;

// Format string keywords that will suppress a log message entirely.
// "Mouse" covers mouse move, drag, and wheel events—extremely high
// frequency and rarely useful outside of low-level debugging.
static const char *suppressedKeywords[] = {"Mouse", NULL};

static int strcontains(const char *str, const char *sub) {
	size_t sublen = strlen(sub);
	if (sublen == 0) return 0;
	for (; *str; str++) {
		if (strncmp(str, sub, sublen) == 0) return 1;
	}
	return 0;
}

static const char *levelName(unsigned int level) {
	switch (level) {
	case LOG_LEVEL_DEBUG:
		return "DEBUG";
	case LOG_LEVEL_INFO:
		return "INFO";
	case LOG_LEVEL_WARN:
		return "WARN";
	case LOG_LEVEL_ERROR:
		return "ERROR";
	default:
		return "?????";
	}
}

// goMusicfoxLoggerProc replaces the default loggerProc from gohook,
// redirecting all log output to a file instead of stdout/stderr.
// It prepends the log level to each line and filters messages
// below the minimum level or matching suppressed keywords.
static bool goMusicfoxLoggerProc(unsigned int level, const char *format, ...) {
	if (gHookLogFile == NULL) {
		return false;
	}

	// Suppress messages whose format string contains a blacklisted keyword.
	for (int i = 0; suppressedKeywords[i] != NULL; i++) {
		if (strcontains(format, suppressedKeywords[i])) {
			return true;
		}
	}

	if (level < gMinLogLevel) {
		return true; // suppressed by level filter
	}

	fprintf(gHookLogFile, "[%-5s] ", levelName(level));

	bool status = false;
	va_list args;
	va_start(args, format);
	status = vfprintf(gHookLogFile, format, args) >= 0;
	va_end(args);

	fflush(gHookLogFile);
	return status;
}

// initGohookLogger opens the log file and overrides the global logger
// function pointer so all subsequent hook log output goes to the file.
// minLogLevel uses gohook's log_level enum values (LOG_LEVEL_DEBUG=1, etc.)
static void initGohookLogger(const char *logPath, unsigned int minLogLevel) {
	if (gHookLogFile != NULL) {
		fclose(gHookLogFile);
	}
	gMinLogLevel = minLogLevel;
	gHookLogFile = fopen(logPath, "w");
	if (gHookLogFile != NULL) {
		logger = &goMusicfoxLoggerProc;
	}
}

// closeGohookLogger closes the log file and resets the logger pointer.
static void closeGohookLogger(void) {
	if (gHookLogFile != NULL) {
		fclose(gHookLogFile);
		gHookLogFile = NULL;
	}
	logger = NULL;
}
*/
import "C"

import (
	"log/slog"
	"path/filepath"
	"time"
	"unsafe"

	"github.com/anhoder/foxful-cli/model"
	"github.com/go-musicfox/go-musicfox/utils/app"
	"github.com/go-musicfox/go-musicfox/utils/slogx"
)

// gohookMinLevel maps slog levels to gohook's log_level enum values.
// gohook: DEBUG=1, INFO=2, WARN=3, ERROR=4
// slog:   DEBUG=-4, INFO=0, WARN=4, ERROR=8
func gohookMinLevel(slogLevel slog.Level) C.uint {
	switch {
	case slogLevel <= slog.LevelDebug:
		return 1 // LOG_LEVEL_DEBUG
	case slogLevel <= slog.LevelInfo:
		return 2 // LOG_LEVEL_INFO
	case slogLevel <= slog.LevelWarn:
		return 3 // LOG_LEVEL_WARN
	default:
		return 4 // LOG_LEVEL_ERROR
	}
}

func (h *EventHandler) RegisterGlobalHotkeys(opts *model.Options) {
	h.registerGlobalHotkeyBindings(opts)

	// Override gohook's logger to redirect output to file instead of stdout/stderr.
	// This runs after a short delay so that hook.Start() in ListenGlobalKeys
	// has already set the default loggerProc first.
	go func() {
		time.Sleep(200 * time.Millisecond)

		logPath := C.CString(filepath.Join(app.LogDir(), "gohook.log"))
		defer C.free(unsafe.Pointer(logPath))

		minLevel := gohookMinLevel(slogx.LevelVar().Level())
		C.initGohookLogger(logPath, minLevel)
	}()
}

// CloseGohookLogger closes the gohook log file. Called during app shutdown.
func CloseGohookLogger() {
	C.closeGohookLogger()
}
