package types

import (
	"fmt"
	"reflect"
	"strings"
)

// Array cast to and remove elements cannot be casted
func ArrayCastTo[T any](values []any) []T {
	results := []T{}

	for _, value := range values {
		if converted, ok := value.(T); ok {
			results = append(results, converted)
		}
	}
	return results
}

// Creates an Array e.g array := ArrayInit[int](1, 2, 3, 4)
func ArrayInit[T any](arr ...T) Array[T] {
	return Array[T](arr)
}

// Returns Raw type
func (arr Array[T]) ToRaw() []T {
	return []T(arr)
}

// Iterate contents of array
func (arr Array[T]) Foreach(iteratee func(index int, value T)) {
	for index, value := range arr {
		iteratee(index, value)
	}
}

// Iterate and update contents of resulting array
func (arr Array[T]) Map(iteratee func(value T) any) Array[any] {
	result := Array[any]{}

	arr.Foreach(func(_ int, value T) {
		result = append(result, iteratee(value))
	})
	return result
}

// Return reverse array
func (arr Array[T]) Reverse() Array[T] {
	rArray := arr

	for i, j := 0, len(rArray)-1; i < j; i, j = i+1, j-1 {
		rArray[i], rArray[j] = rArray[j], rArray[i]
	}
	return rArray
}

// Join array
func (arr Array[T]) Join(sep string) string {
	items := []string{}

	arr.Foreach(func(index int, value T) {
		items = append(items, fmt.Sprint(value))
	})
	return strings.Join(items, sep)
}

// Check array contains
func (arr Array[T]) Constains(value T) bool {
	for _, aValue := range arr {
		if reflect.DeepEqual(aValue, value) {
			return true
		}
	}
	return false
}

// Check if array is equal
func (arr Array[T]) IsEqualTo(arrC Array[T]) bool {
	if len(arr) != len(arrC) {
		return false
	}
	for i := 0; i < len(arr); i = i + 1 {
		if !reflect.DeepEqual(arr[i], arrC[i]) {
			return false
		}
	}
	return true
}

// Pop index
func (arr *Array[T]) PopIndex(index int) *T {
	aValue := *arr
	aLen := len(aValue)

	if aLen == 0 || index < 0 || index >= aLen {
		return nil
	}
	ret := aValue[index]
	leadAValue := aValue[:index]

	*arr = append(leadAValue, aValue[index+1:]...)
	return &ret
}

// Return last item
func (arr Array[T]) Last() T {
	return arr[len(arr)-1]
}
