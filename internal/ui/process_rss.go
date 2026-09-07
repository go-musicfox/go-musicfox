package ui

import (
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

// processRSSBytes returns the process resident set size in bytes.
// Darwin uses `ps` (rss is KiB); Linux reads /proc/self/statm. Returns 0
// when the platform is unsupported or the probe fails.
func processRSSBytes() uint64 {
	switch runtime.GOOS {
	case "darwin":
		return rssFromPS()
	case "linux":
		return rssFromProcStatm()
	default:
		return 0
	}
}

func rssFromPS() uint64 {
	out, err := exec.Command("ps", "-o", "rss=", "-p", strconv.Itoa(os.Getpid())).Output()
	if err != nil {
		return 0
	}
	kb, err := strconv.ParseUint(strings.TrimSpace(string(out)), 10, 64)
	if err != nil {
		return 0
	}
	return kb * 1024
}

func rssFromProcStatm() uint64 {
	data, err := os.ReadFile("/proc/self/statm")
	if err != nil {
		return 0
	}
	fields := strings.Fields(string(data))
	if len(fields) < 2 {
		return 0
	}
	pages, err := strconv.ParseUint(fields[1], 10, 64)
	if err != nil {
		return 0
	}
	return pages * uint64(os.Getpagesize())
}
