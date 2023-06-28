package mathutil

import "github.com/gookit/goutil/comdef"

// Compare any intX,floatX value by given op. returns `srcVal op(=,!=,<,<=,>,>=) dstVal`
//
// Usage:
//
//	mathutil.Compare(2, 3, ">") // false
//	mathutil.Compare(2, 1.3, ">") // true
//	mathutil.Compare(2.2, 1.3, ">") // true
//	mathutil.Compare(2.1, 2, ">") // true
func Compare(srcVal, dstVal any, op string) (ok bool) {
	if srcVal == nil || dstVal == nil {
		return false
	}

	// float
	if srcFlt, ok := srcVal.(float64); ok {
		if dstFlt, err := ToFloat(dstVal); err == nil {
			return CompFloat(srcFlt, dstFlt, op)
		}
		return false
	}

	if srcFlt, ok := srcVal.(float32); ok {
		if dstFlt, err := ToFloat(dstVal); err == nil {
			return CompFloat(float64(srcFlt), dstFlt, op)
		}
		return false
	}

	// as int64
	srcInt, err := ToInt64(srcVal)
	if err != nil {
		return false
	}

	dstInt, err := ToInt64(dstVal)
	if err != nil {
		return false
	}

	return CompInt64(srcInt, dstInt, op)
}

// CompInt compare int,uint value. returns `srcVal op(=,!=,<,<=,>,>=) dstVal`
func CompInt[T comdef.Xint](srcVal, dstVal T, op string) (ok bool) {
	return CompValue(srcVal, dstVal, op)
}

// CompInt64 compare int64 value. returns `srcVal op(=,!=,<,<=,>,>=) dstVal`
func CompInt64(srcVal, dstVal int64, op string) bool {
	return CompValue(srcVal, dstVal, op)
}

// CompFloat compare float64,float32 value. returns `srcVal op(=,!=,<,<=,>,>=) dstVal`
func CompFloat[T comdef.Float](srcVal, dstVal T, op string) (ok bool) {
	return CompValue(srcVal, dstVal, op)
}

// CompValue compare intX,uintX,floatX value. returns `srcVal op(=,!=,<,<=,>,>=) dstVal`
func CompValue[T comdef.XintOrFloat](srcVal, dstVal T, op string) (ok bool) {
	switch op {
	case "<", "lt":
		ok = srcVal < dstVal
	case "<=", "lte":
		ok = srcVal <= dstVal
	case ">", "gt":
		ok = srcVal > dstVal
	case ">=", "gte":
		ok = srcVal >= dstVal
	case "=", "eq":
		ok = srcVal == dstVal
	case "!=", "ne", "neq":
		ok = srcVal != dstVal
	}
	return
}

// InRange check if val in int/float range [min, max]
func InRange[T comdef.IntOrFloat](val, min, max T) bool {
	return val >= min && val <= max
}

// OutRange check if val not in int/float range [min, max]
func OutRange[T comdef.IntOrFloat](val, min, max T) bool {
	return val < min || val > max
}

// InUintRange check if val in unit range [min, max]
func InUintRange[T comdef.Uint](val, min, max T) bool {
	if max == 0 {
		return val >= min
	}
	return val >= min && val <= max
}
