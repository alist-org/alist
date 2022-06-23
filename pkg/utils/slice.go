package utils

func SliceEqual[T comparable](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

func SliceContains[T comparable](arr []T, v T) bool {
	for _, vv := range arr {
		if vv == v {
			return true
		}
	}
	return false
}
