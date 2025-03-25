package utils

import (
	"fmt"
	"reflect"
)

// Must asserts that a given error *MUST* be nil at runtime, otherwise a panic
// *WILL* occur.
func Must[T any](value T, err error) T {
	if err != nil {
		panic(fmt.Sprintf("Must: err was not nil: %v\n", err))
	}

	return value
}

// MustCast asserts that a given interface argument *MUST* be down-castable to
// T at runtime, otherwise a panic *WILL* occur.
func MustCast[T any](i any) *T {
	casted, ok := i.(*T)
	if !ok {
		panic(fmt.Sprintf("MustCast: given interface was not down-castable to *%s", reflect.TypeFor[T]().Name()))
	}

	return casted
}
