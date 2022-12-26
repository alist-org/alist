package utils

import "strings"

var balance = ".balance"

func IsBalance(str string) bool {
	return strings.Contains(str, balance)
}

// GetActualMountPath remove balance suffix
func GetActualMountPath(virtualPath string) string {
	bIndex := strings.LastIndex(virtualPath, ".balance")
	if bIndex != -1 {
		virtualPath = virtualPath[:bIndex]
	}
	return virtualPath
}
