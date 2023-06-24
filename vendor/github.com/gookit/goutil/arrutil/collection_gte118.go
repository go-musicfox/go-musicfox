package arrutil

// type MapFn func(obj T) (target V, find bool)

// Map a list to new list
//
// eg: mapping [object0{},object1{},...] to flatten list [object0.someKey, object1.someKey, ...]
func Map[T any, V any](list []T, mapFn func(obj T) (val V, find bool)) []V {
	flatArr := make([]V, 0, len(list))

	for _, obj := range list {
		if target, ok := mapFn(obj); ok {
			flatArr = append(flatArr, target)
		}
	}
	return flatArr
}

// Column alias of Map func
func Column[T any, V any](list []T, mapFn func(obj T) (val V, find bool)) []V {
	return Map(list, mapFn)
}
