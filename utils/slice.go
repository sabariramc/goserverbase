package utils

/*
Prepend add a element[T] to the beginning of a list[T] it appends
*/
func Prepend[T any](a []T, b T) []T {
	var c T
	a = append(a, c)
	copy(a[1:], a)
	a[0] = b
	return a
}
