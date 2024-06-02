package mathx

import (
	"fmt"
	"math"
	"strconv"
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

// FormatBytes returns a string representing the size in bytes with a suffix.
func FormatBytes(size int64) string {
	const unit = 1000
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	s := float64(size)
	units := []string{"kB", "MB", "GB", "TB", "PB", "EB"}

	e := math.Floor(math.Log10(s/unit) / math.Log10(unit))
	n := s / math.Pow(unit, e)
	return strconv.FormatFloat(n, 'f', -1, 64) + " " + units[int(e)]
}
