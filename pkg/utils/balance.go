package utils

import "strings"

var balance = ".balance"

func IsBalance(str string) bool {
	return strings.Contains(str, balance)
}

// GetActualVirtualPath remove balance suffix
func GetActualVirtualPath(virtualPath string) string {
	bIndex := strings.LastIndex(virtualPath, ".balance")
	if bIndex != -1 {
		virtualPath = virtualPath[:bIndex]
	}
	return virtualPath
}
