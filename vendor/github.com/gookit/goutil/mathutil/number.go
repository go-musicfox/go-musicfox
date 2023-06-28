package mathutil

import (
	"fmt"
	"time"

	"github.com/gookit/goutil/basefn"
)

// IsNumeric returns true if the given character is a numeric, otherwise false.
func IsNumeric(c byte) bool {
	return c >= '0' && c <= '9'
}

// Percent returns a values percent of the total
func Percent(val, total int) float64 {
	if total == 0 {
		return float64(0)
	}
	return (float64(val) / float64(total)) * 100
}

// ElapsedTime calc elapsed time 计算运行时间消耗 单位 ms(毫秒)
func ElapsedTime(startTime time.Time) string {
	return fmt.Sprintf("%.3f", time.Since(startTime).Seconds()*1000)
}

// DataSize format value to data size string. eg: 1024 => 1KB, 1024*1024 => 1MB
// alias format.DataSize()
func DataSize(size uint64) string {
	return basefn.DataSize(size)
}

// HowLongAgo calc time. alias format.HowLongAgo()
func HowLongAgo(sec int64) string {
	return basefn.HowLongAgo(sec)
}
