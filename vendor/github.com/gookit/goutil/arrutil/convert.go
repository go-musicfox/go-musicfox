package arrutil

import (
	"errors"
	"reflect"
	"strconv"
	"strings"

	"github.com/gookit/goutil/mathutil"
	"github.com/gookit/goutil/strutil"
)

// ErrInvalidType error
var ErrInvalidType = errors.New("the input param type is invalid")

/*************************************************************
 * helper func for strings
 *************************************************************/

// JoinStrings alias of strings.Join
func JoinStrings(sep string, ss ...string) string {
	return strings.Join(ss, sep)
}

// StringsJoin alias of strings.Join
func StringsJoin(sep string, ss ...string) string {
	return strings.Join(ss, sep)
}

// StringsToInts string slice to int slice
func StringsToInts(ss []string) (ints []int, err error) {
	for _, str := range ss {
		iVal, err := strconv.Atoi(str)
		if err != nil {
			return nil, err
		}

		ints = append(ints, iVal)
	}
	return
}

// MustToStrings convert array or slice to []string
func MustToStrings(arr any) []string {
	ret, _ := ToStrings(arr)
	return ret
}

// StringsToSlice convert []string to []any
func StringsToSlice(ss []string) []any {
	args := make([]any, len(ss))
	for i, s := range ss {
		args[i] = s
	}
	return args
}

/*************************************************************
 * helper func for slices
 *************************************************************/

// ToInt64s convert any(allow: array,slice) to []int64
func ToInt64s(arr any) (ret []int64, err error) {
	rv := reflect.ValueOf(arr)
	if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
		err = ErrInvalidType
		return
	}

	for i := 0; i < rv.Len(); i++ {
		i64, err := mathutil.Int64(rv.Index(i).Interface())
		if err != nil {
			return []int64{}, err
		}

		ret = append(ret, i64)
	}
	return
}

// MustToInt64s convert any(allow: array,slice) to []int64
func MustToInt64s(arr any) []int64 {
	ret, _ := ToInt64s(arr)
	return ret
}

// SliceToInt64s convert []any to []int64
func SliceToInt64s(arr []any) []int64 {
	i64s := make([]int64, len(arr))
	for i, v := range arr {
		i64s[i] = mathutil.MustInt64(v)
	}
	return i64s
}

// ToStrings convert any(allow: array,slice) to []string
func ToStrings(arr any) (ret []string, err error) {
	rv := reflect.ValueOf(arr)
	if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
		err = ErrInvalidType
		return
	}

	for i := 0; i < rv.Len(); i++ {
		str, err := strutil.ToString(rv.Index(i).Interface())
		if err != nil {
			return []string{}, err
		}

		ret = append(ret, str)
	}
	return
}

// SliceToStrings convert []any to []string
func SliceToStrings(arr []any) []string {
	ss := make([]string, len(arr))
	for i, v := range arr {
		ss[i] = strutil.MustString(v)
	}
	return ss
}

// AnyToString simple and quickly convert any array, slice to string
func AnyToString(arr any) string {
	return NewFormatter(arr).Format()
}

// SliceToString convert []any to string
func SliceToString(arr ...any) string { return ToString(arr) }

// ToString simple and quickly convert []any to string
func ToString(arr []any) string {
	// like fmt.Println([]any(nil))
	if arr == nil {
		return "[]"
	}

	var sb strings.Builder
	sb.WriteByte('[')

	for i, v := range arr {
		if i > 0 {
			sb.WriteByte(',')
		}
		sb.WriteString(strutil.MustString(v))
	}

	sb.WriteByte(']')
	return sb.String()
}

// JoinSlice join []any slice to string.
func JoinSlice(sep string, arr ...any) string {
	if arr == nil {
		return ""
	}

	var sb strings.Builder
	for i, v := range arr {
		if i > 0 {
			sb.WriteString(sep)
		}
		sb.WriteString(strutil.MustString(v))
	}

	return sb.String()
}
