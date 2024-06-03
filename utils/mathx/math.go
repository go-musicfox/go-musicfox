package mathx

import (
	"fmt"
	"math"
)

type ordinal interface {
	int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64 | float32 | float64
}

func Min[T ordinal](a, b T) T {
	if a > b {
		return b
	}
	return a
}

func Max[T ordinal](a, b T) T {
	if a < b {
		return b
	}
	return a
}

func FormatBytes(size int64) string {
	base := int64(1024)
	if size < base {
		return fmt.Sprintf("%d B", size)
	}

	orders := []string{"", "KB", "MB", "GB", "TB", "PB", "EB"}
	exp := int(math.Floor(math.Log2(float64(size)) / 10))
	if exp >= len(orders) {
		exp = len(orders) - 1
	}

	suffix := orders[exp]
	value := float64(size) / math.Pow(float64(base), float64(exp))
	return fmt.Sprintf("%.2f %s", value, suffix)
}
