package errorx

import (
	"errors"
	"strings"
	"testing"
)

func TestRecoverIgnore(t *testing.T) {
	func() {
		// recover 必须由 defer 直接调用的函数执行才有效
		defer Recover(true)
		panic("boom")
	}()
	// 若 recover 未生效，panic 会逃逸导致本测试失败
}

func TestFormatCrashReport(t *testing.T) {
	stack := []byte("goroutine 1 [running]:\nmain.crash()\n\tmain.go:10 +0x1")
	msg := formatCrashReport(errors.New("boom"), stack)
	for _, want := range []string{
		"musicfox 发生未预期的错误",
		"错误: boom",
		"崩溃堆栈（首个 goroutine）",
		"goroutine 1 [running]",
		"main.crash()",
		"issues/new?template=1_bug_report.yml",
		"日志",
	} {
		if !strings.Contains(msg, want) {
			t.Errorf("crash report missing %q, got: %s", want, msg)
		}
	}
}

func TestStackHeadTruncates(t *testing.T) {
	stack := []byte("a\nb\nc\n")
	got := stackHead(stack, 2)
	if len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Fatalf("unexpected stack head: %v", got)
	}
	if head := stackHead(nil, 2); head != nil {
		t.Fatalf("expected nil for empty stack, got %v", head)
	}
}
