package mathutil

import "math"

// MaxFloat compare and return max value
func MaxFloat(x, y float64) float64 {
	return math.Max(x, y)
}

// MaxInt compare and return max value
func MaxInt(x, y int) int {
	if x > y {
		return x
	}
	return y
}

// SwapMaxInt compare and return max, min value
func SwapMaxInt(x, y int) (int, int) {
	if x > y {
		return x, y
	}
	return y, x
}

// MaxI64 compare and return max value
func MaxI64(x, y int64) int64 {
	if x > y {
		return x
	}
	return y
}

// SwapMaxI64 compare and return max, min value
func SwapMaxI64(x, y int64) (int64, int64) {
	if x > y {
		return x, y
	}
	return y, x
}
