package utils

// StreamArray turns a given array into a receive-only chan that can be used to
// stream data to client.
func StreamArray[T any](array []T) <-chan T {
	ch := make(chan T, len(array))

	go func() {
		defer close(ch)

		for _, element := range array {
			ch <- element
		}
	}()

	return ch
}
