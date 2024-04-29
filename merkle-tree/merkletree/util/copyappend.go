package util

func CopyAppend[T any](arr []T, x ...T) []T {
	ret := make([]T, 0)
	ret = append(ret, arr...)
	ret = append(ret, x...)
	return ret
}
