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
	msg := formatCrashReport(errors.New("boom"))
	for _, want := range []string{
		"musicfox 发生未预期的错误",
		"错误: boom",
		"issues/new?template=1_bug_report.yml",
		"日志",
	} {
		if !strings.Contains(msg, want) {
			t.Errorf("crash report missing %q, got: %s", want, msg)
		}
	}
}
