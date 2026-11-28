package errorx

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTempLog(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "musicfox.log")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write temp log: %v", err)
	}
	return path
}

func TestCrashHeadSignal(t *testing.T) {
	log := "==== musicfox started ====\n" +
		"time=... level=INFO msg=hello\n" +
		"SIGSEGV: segmentation violation\n" +
		"PC=0x793f787416fc m=0 sigcode=1\n" +
		"signal arrived during external code execution\n\n" +
		"goroutine 0 gp=0x1 m=0 mp=0x2 [signal SIGSEGV: segmentation violation code=0x1 addr=0x0 pc=0x793f787416fc]:\n" +
		"runtime: g 0: unknown pc 0x793f787416fc\n" +
		"stack: frame={sp:0x1, fp:0x2}\n"
	path := writeTempLog(t, log)

	lines, reachedEOF := crashHead(path, 30)
	if !reachedEOF {
		t.Fatalf("expected head to reach EOF")
	}
	if len(lines) != 7 {
		t.Fatalf("expected 7 lines, got %d: %v", len(lines), lines)
	}
	if lines[0] != "SIGSEGV: segmentation violation" {
		t.Fatalf("head must start at SIGSEGV, got %q", lines[0])
	}
	if lines[4] != "goroutine 0 gp=0x1 m=0 mp=0x2 [signal SIGSEGV: segmentation violation code=0x1 addr=0x0 pc=0x793f787416fc]:" {
		t.Fatalf("first goroutine line missing, got %q", lines[4])
	}
}

func TestCrashHeadPrefersLastStrongHeader(t *testing.T) {
	log := "panic: old crash\n\ngoroutine 1 [running]:\nold\n" +
		"normal log again\n" +
		"panic: new crash\n\ngoroutine 7 [running]:\nnew\n"
	path := writeTempLog(t, log)

	lines, reachedEOF := crashHead(path, 30)
	if !reachedEOF {
		t.Fatalf("expected head to reach EOF")
	}
	if len(lines) == 0 || lines[0] != "panic: new crash" {
		t.Fatalf("expected last panic header, got %v", lines)
	}
}

func TestCrashHeadTruncatesAndDoesNotReachEOF(t *testing.T) {
	log := "fatal error: boom\n" +
		"goroutine 1 [running]:\n" +
		"frame1\nframe2\nframe3\nframe4\nframe5\n"
	path := writeTempLog(t, log)

	lines, reachedEOF := crashHead(path, 3)
	if reachedEOF {
		t.Fatalf("expected not to reach EOF")
	}
	if len(lines) != 3 || lines[0] != "fatal error: boom" {
		t.Fatalf("unexpected head: %v", lines)
	}
}

func TestCrashHeadNoMarker(t *testing.T) {
	path := writeTempLog(t, "time=... level=INFO msg=hello\n")
	if lines, _ := crashHead(path, 10); lines != nil {
		t.Fatalf("expected nil, got %v", lines)
	}
}

func TestCrashHeadMissingFile(t *testing.T) {
	if lines, _ := crashHead(filepath.Join(t.TempDir(), "nope.log"), 10); lines != nil {
		t.Fatalf("expected nil, got %v", lines)
	}
}
