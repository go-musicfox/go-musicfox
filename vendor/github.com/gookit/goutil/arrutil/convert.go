package arrutil

import (
	"errors"
	"reflect"
	"strconv"
	"strings"

	"github.com/gookit/goutil/comdef"
	"github.com/gookit/goutil/mathutil"
	"github.com/gookit/goutil/reflects"
	"github.com/gookit/goutil/strutil"
)

// ErrInvalidType error
var ErrInvalidType = errors.New("the input param type is invalid")

/*************************************************************
 * Join func for slice
 *************************************************************/

// JoinStrings alias of strings.Join
func JoinStrings(sep string, ss ...string) string {
	return strings.Join(ss, sep)
}

// StringsJoin alias of strings.Join
func StringsJoin(sep string, ss ...string) string {
	return strings.Join(ss, sep)
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
		sb.WriteString(strutil.QuietString(v))
	}

	return sb.String()
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
		i64s[i] = mathutil.QuietInt64(v)
	}
	return i64s
}

// StringsAsInts convert and ignore error
func StringsAsInts(ss []string) []int {
	ints, _ := StringsTryInts(ss)
	return ints
}

// StringsToInts string slice to int slice
func StringsToInts(ss []string) (ints []int, err error) {
	return StringsTryInts(ss)
}

// StringsTryInts string slice to int slice
func StringsTryInts(ss []string) (ints []int, err error) {
	for _, str := range ss {
		iVal, err := strconv.Atoi(str)
		if err != nil {
			return nil, err
		}

		ints = append(ints, iVal)
	}
	return
}

// AnyToSlice convert any(allow: array,slice) to []any
func AnyToSlice(sl any) (ls []any, err error) {
	rfKeys := reflect.ValueOf(sl)
	if rfKeys.Kind() != reflect.Slice && rfKeys.Kind() != reflect.Array {
		return nil, ErrInvalidType
	}

	for i := 0; i < rfKeys.Len(); i++ {
		ls = append(ls, rfKeys.Index(i).Interface())
	}
	return
}

// AnyToStrings convert array or slice to []string
func AnyToStrings(arr any) []string {
	ret, _ := ToStrings(arr)
	return ret
}

// MustToStrings convert array or slice to []string
func MustToStrings(arr any) []string {
	ret, err := ToStrings(arr)
	if err != nil {
		panic(err)
	}
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

// ToStrings convert any(allow: array,slice) to []string
func ToStrings(arr any) (ret []string, err error) {
	rv := reflect.ValueOf(arr)
	if rv.Kind() == reflect.String {
		return []string{rv.String()}, nil
	}

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
	return QuietStrings(arr)
}

// QuietStrings convert []any to []string
func QuietStrings(arr []any) []string {
	ss := make([]string, len(arr))
	for i, v := range arr {
		ss[i] = strutil.QuietString(v)
	}
	return ss
}

// ConvType convert type of slice elements to new type slice, by the given newElemTyp type.
//
// Supports conversion between []string, []intX, []uintX, []floatX.
//
// Usage:
//
//	ints, _ := arrutil.ConvType([]string{"12", "23"}, 1) // []int{12, 23}
func ConvType[T any, R any](arr []T, newElemTyp R) ([]R, error) {
	newArr := make([]R, len(arr))
	elemTyp := reflect.TypeOf(newElemTyp)

	for i, elem := range arr {
		var anyElem any = elem
		// type is same.
		if _, ok := anyElem.(R); ok {
			newArr[i] = anyElem.(R)
			continue
		}

		// need conv type.
		rfVal, err := reflects.ValueByType(elem, elemTyp)
		if err != nil {
			return nil, err
		}
		newArr[i] = rfVal.Interface().(R)
	}
	return newArr, nil
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
		sb.WriteString(strutil.QuietString(v))
	}

	sb.WriteByte(']')
	return sb.String()
}

// CombineToMap combine two slice to map[K]V.
//
// If keys length is greater than values, the extra keys will be ignored.
func CombineToMap[K comdef.SortedType, V any](keys []K, values []V) map[K]V {
	ln := len(values)
	mp := make(map[K]V, len(keys))

	for i, key := range keys {
		if i >= ln {
			break
		}
		mp[key] = values[i]
	}
	return mp
}

// CombineToSMap combine two string-slice to map[string]string
func CombineToSMap(keys, values []string) map[string]string {
	ln := len(values)
	mp := make(map[string]string, len(keys))

	for i, key := range keys {
		if ln > i {
			mp[key] = values[i]
		} else {
			mp[key] = ""
		}
	}
	return mp
}
