package utils

func IsBool(bs ...bool) bool {
	return len(bs) > 0 && bs[0]
}
