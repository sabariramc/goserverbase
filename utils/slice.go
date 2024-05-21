package utils

// Prepend adds an element to the beginning of a slice.
//
// This function takes a slice of any type and an element of the same type,
// and returns a new slice with the element prepended.
func Prepend[T any](a []T, b T) []T {
	var c T
	a = append(a, c)
	copy(a[1:], a)
	a[0] = b
	return a
}
