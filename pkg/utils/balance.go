package utils

import "strings"

var balance = ".balance"

func IsBalance(str string) bool {
	return strings.Contains(str, balance)
}

// GetActualMountPath remove balance and cache suffix
func GetActualMountPath(mountPath string) string {
	bIndex := strings.LastIndex(mountPath, ".balance")
	if bIndex != -1 {
		mountPath = mountPath[:bIndex]
	}
	// copy from the front, should be optimized
	bIndex = strings.LastIndex(mountPath, ".cache")
	if bIndex != -1 {
		mountPath = mountPath[:bIndex]
	}
	return mountPath
}
