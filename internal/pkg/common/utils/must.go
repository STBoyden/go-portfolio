package utils

import "fmt"

func Must[T any](value T, err error) T {
	if err != nil {
		panic(fmt.Sprintf("err was not nil: %v\n", err))
	}

	return value
}
