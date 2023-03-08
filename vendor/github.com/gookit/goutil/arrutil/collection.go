package arrutil

import (
	"errors"
	"reflect"
)

// ErrElementNotFound is the error returned when the element is not found.
const ErrElementNotFound = "element not found"

// Comparer Function to compare two elements.
type Comparer func(a, b any) int

// Predicate Function to predicate a struct/value satisfies a condition.
type Predicate func(a any) bool

var (
	// StringEqualsComparer Comparer for string. It will compare the string by their value.
	// returns: 0 if equal, -1 if a != b
	StringEqualsComparer Comparer = func(a, b any) int {
		typeOfA := reflect.TypeOf(a)
		if typeOfA.Kind() == reflect.Ptr {
			typeOfA = typeOfA.Elem()
		}

		typeOfB := reflect.TypeOf(b)
		if typeOfB.Kind() == reflect.Ptr {
			typeOfB = typeOfB.Elem()
		}

		if typeOfA != typeOfB {
			return -1
		}

		strA := ""
		strB := ""

		if val, ok := a.(string); ok {
			strA = val
		} else if val, ok := a.(*string); ok {
			strA = *val
		} else {
			return -1
		}

		if val, ok := b.(string); ok {
			strB = val
		} else if val, ok := b.(*string); ok {
			strB = *val
		} else {
			return -1
		}

		if strA == strB {
			return 0
		}
		return -1
	}

	// ReferenceEqualsComparer Comparer for strcut ptr. It will compare the struct by their ptr addr.
	// returns: 0 if equal, -1 if a != b
	ReferenceEqualsComparer Comparer = func(a, b any) int {
		if a == b {
			return 0
		}
		return -1
	}

	// ElemTypeEqualsComparer Comparer for struct/value. It will compare the struct by their element type (reflect.Type.Elem()).
	// returns: 0 if same type, -1 if not.
	ElemTypeEqualsComparer Comparer = func(a, b any) int {
		at := reflect.TypeOf(a)
		bt := reflect.TypeOf(b)
		if at.Kind() == reflect.Ptr {
			at = at.Elem()
		}

		if bt.Kind() == reflect.Ptr {
			bt = bt.Elem()
		}

		if at == bt {
			return 0
		}
		return -1
	}
)

// TwowaySearch Find specialized element in a slice forward and backward in the same time, should be more quickly.
//
//	data: the slice to search in. MUST BE A SLICE.
//	item: the element to search.
//	fn: the comparer function.
//	return: the index of the element, or -1 if not found.
func TwowaySearch(data any, item any, fn Comparer) (int, error) {
	if data == nil {
		return -1, errors.New("collections.TwowaySearch: data is nil")
	}
	if fn == nil {
		return -1, errors.New("collections.TwowaySearch: fn is nil")
	}

	dataType := reflect.TypeOf(data)
	if dataType.Kind() != reflect.Slice {
		return -1, errors.New("collections.TwowaySearch: data is not a slice")
	}

	dataVal := reflect.ValueOf(data)
	if dataVal.Len() == 0 {
		return -1, errors.New("collections.TwowaySearch: data is empty")
	}
	itemType := dataType.Elem()
	if itemType.Kind() == reflect.Ptr {
		itemType = itemType.Elem()
	}

	if itemType != dataVal.Index(0).Type() {
		return -1, errors.New("collections.TwowaySearch: item type is not the same as data type")
	}

	forward := 0
	backward := dataVal.Len() - 1

	for forward <= backward {
		forwardVal := dataVal.Index(forward).Interface()
		if fn(forwardVal, item) == 0 {
			return forward, nil
		}

		backwardVal := dataVal.Index(backward).Interface()
		if fn(backwardVal, item) == 0 {
			return backward, nil
		}

		forward++
		backward--
	}

	return -1, errors.New(ErrElementNotFound)
}

// MakeEmptySlice Create a new slice with the elements of the source that satisfy the predicate.
//
// itemType: the type of the elements in the source.
// returns: the new slice.
func MakeEmptySlice(itemType reflect.Type) any {
	ret := reflect.MakeSlice(reflect.SliceOf(itemType), 0, 0).Interface()
	return ret
}

// CloneSlice Clone a slice.
//
//	data: the slice to clone.
//	returns: the cloned slice.
func CloneSlice(data any) any {
	typeOfData := reflect.TypeOf(data)
	if typeOfData.Kind() != reflect.Slice {
		panic("collections.CloneSlice: data must be a slice")
	}
	return reflect.AppendSlice(reflect.New(reflect.SliceOf(typeOfData.Elem())).Elem(), reflect.ValueOf(data)).Interface()
}

// Differences Produces the set difference of two slice according to a comparer function.
//
//	first: the first slice. MUST BE A SLICE.
//	second: the second slice. MUST BE A SLICE.
//	fn: the comparer function.
//	returns: the difference of the two slices.
func Differences[T any](first, second []T, fn Comparer) []T {
	typeOfFirst := reflect.TypeOf(first)
	if typeOfFirst.Kind() != reflect.Slice {
		panic("collections.Excepts: first must be a slice")
	}

	typeOfSecond := reflect.TypeOf(second)
	if typeOfSecond.Kind() != reflect.Slice {
		panic("collections.Excepts: second must be a slice")
	}

	firstLen := len(first)
	if firstLen == 0 {
		return CloneSlice(second).([]T)
	}

	secondLen := len(second)
	if secondLen == 0 {
		return CloneSlice(first).([]T)
	}

	max := firstLen
	if secondLen > firstLen {
		max = secondLen
	}

	result := make([]T, 0)
	for i := 0; i < max; i++ {
		if i < firstLen {
			s := first[i]
			if i, _ := TwowaySearch(second, s, fn); i < 0 {
				result = append(result, s)
			}
		}

		if i < secondLen {
			t := second[i]
			if i, _ := TwowaySearch(first, t, fn); i < 0 {
				result = append(result, t)
			}
		}
	}

	return result
}

// Excepts Produces the set difference of two slice according to a comparer function.
//
//	first: the first slice. MUST BE A SLICE.
//	second: the second slice. MUST BE A SLICE.
//	fn: the comparer function.
//	returns: the difference of the two slices.
func Excepts(first, second any, fn Comparer) any {
	typeOfFirst := reflect.TypeOf(first)
	if typeOfFirst.Kind() != reflect.Slice {
		panic("collections.Excepts: first must be a slice")
	}
	valOfFirst := reflect.ValueOf(first)
	if valOfFirst.Len() == 0 {
		return MakeEmptySlice(typeOfFirst.Elem())
	}

	typeOfSecond := reflect.TypeOf(second)
	if typeOfSecond.Kind() != reflect.Slice {
		panic("collections.Excepts: second must be a slice")
	}

	valOfSecond := reflect.ValueOf(second)
	if valOfSecond.Len() == 0 {
		return CloneSlice(first)
	}

	result := reflect.New(reflect.SliceOf(typeOfFirst.Elem())).Elem()
	for i := 0; i < valOfFirst.Len(); i++ {
		s := valOfFirst.Index(i).Interface()
		if i, _ := TwowaySearch(second, s, fn); i < 0 {
			result = reflect.Append(result, reflect.ValueOf(s))
		}
	}

	return result.Interface()
}

// Intersects Produces to intersect of two slice according to a comparer function.
//
//	first: the first slice. MUST BE A SLICE.
//	second: the second slice. MUST BE A SLICE.
//	fn: the comparer function.
//	returns: to intersect of the two slices.
func Intersects(first any, second any, fn Comparer) any {
	typeOfFirst := reflect.TypeOf(first)
	if typeOfFirst.Kind() != reflect.Slice {
		panic("collections.Intersects: first must be a slice")
	}
	valOfFirst := reflect.ValueOf(first)
	if valOfFirst.Len() == 0 {
		return MakeEmptySlice(typeOfFirst.Elem())
	}

	typeOfSecond := reflect.TypeOf(second)
	if typeOfSecond.Kind() != reflect.Slice {
		panic("collections.Intersects: second must be a slice")
	}

	valOfSecond := reflect.ValueOf(second)
	if valOfSecond.Len() == 0 {
		return MakeEmptySlice(typeOfFirst.Elem())
	}

	result := reflect.New(reflect.SliceOf(typeOfFirst.Elem())).Elem()
	for i := 0; i < valOfFirst.Len(); i++ {
		s := valOfFirst.Index(i).Interface()
		if i, _ := TwowaySearch(second, s, fn); i >= 0 {
			result = reflect.Append(result, reflect.ValueOf(s))
		}
	}

	return result.Interface()
}

// Union Produces the set union of two slice according to a comparer function
//
//	first: the first slice. MUST BE A SLICE.
//	second: the second slice. MUST BE A SLICE.
//	fn: the comparer function.
//	returns: the union of the two slices.
func Union(first, second any, fn Comparer) any {
	excepts := Excepts(second, first, fn)

	typeOfFirst := reflect.TypeOf(first)
	if typeOfFirst.Kind() != reflect.Slice {
		panic("collections.Intersects: first must be a slice")
	}
	valOfFirst := reflect.ValueOf(first)
	if valOfFirst.Len() == 0 {
		return CloneSlice(second)
	}

	result := reflect.AppendSlice(reflect.New(reflect.SliceOf(typeOfFirst.Elem())).Elem(), valOfFirst)
	result = reflect.AppendSlice(result, reflect.ValueOf(excepts))
	return result.Interface()
}

// Find Produces the struct/value of a slice according to a predicate function.
//
//	source: the slice. MUST BE A SLICE.
//	fn: the predicate function.
//	returns: the struct/value of the slice.
func Find(source any, fn Predicate) (any, error) {
	aType := reflect.TypeOf(source)
	if aType.Kind() != reflect.Slice {
		panic("collections.Find: source must be a slice")
	}

	sourceVal := reflect.ValueOf(source)
	if sourceVal.Len() == 0 {
		return nil, errors.New(ErrElementNotFound)
	}

	for i := 0; i < sourceVal.Len(); i++ {
		s := sourceVal.Index(i).Interface()
		if fn(s) {
			return s, nil
		}
	}
	return nil, errors.New(ErrElementNotFound)
}

// FindOrDefault Produce the struct/value f a slice to a predicate function,
// Produce default value when predicate function not found.
//
//	source: the slice. MUST BE A SLICE.
//	fn: the predicate function.
//	defaultValue: the default value.
//	returns: the struct/value of the slice.
func FindOrDefault(source any, fn Predicate, defaultValue any) any {
	item, err := Find(source, fn)
	if err != nil {
		if err.Error() == ErrElementNotFound {
			return defaultValue
		}
	}
	return item
}

// TakeWhile Produce the set of a slice according to a predicate function,
// Produce empty slice when predicate function not matched.
//
//	data: the slice. MUST BE A SLICE.
//	fn: the predicate function.
//	returns: the set of the slice.
func TakeWhile(data any, fn Predicate) any {
	aType := reflect.TypeOf(data)
	if aType.Kind() != reflect.Slice {
		panic("collections.TakeWhile: data must be a slice")
	}

	sourceVal := reflect.ValueOf(data)
	if sourceVal.Len() == 0 {
		return MakeEmptySlice(aType.Elem())
	}

	result := reflect.New(reflect.SliceOf(aType.Elem())).Elem()
	for i := 0; i < sourceVal.Len(); i++ {
		s := sourceVal.Index(i).Interface()
		if fn(s) {
			result = reflect.Append(result, reflect.ValueOf(s))
		}
	}
	return result.Interface()
}

// ExceptWhile Produce the set of a slice except with a predicate function,
// Produce original slice when predicate function not match.
//
//	data: the slice. MUST BE A SLICE.
//	fn: the predicate function.
//	returns: the set of the slice.
func ExceptWhile(data any, fn Predicate) any {
	aType := reflect.TypeOf(data)
	if aType.Kind() != reflect.Slice {
		panic("collections.ExceptWhile: data must be a slice")
	}

	sourceVal := reflect.ValueOf(data)
	if sourceVal.Len() == 0 {
		return MakeEmptySlice(aType.Elem())
	}

	result := reflect.New(reflect.SliceOf(aType.Elem())).Elem()
	for i := 0; i < sourceVal.Len(); i++ {
		s := sourceVal.Index(i).Interface()
		if !fn(s) {
			result = reflect.Append(result, reflect.ValueOf(s))
		}
	}
	return result.Interface()
}
