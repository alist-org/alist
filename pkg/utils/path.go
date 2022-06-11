package utils

import "strings"

// StandardizationPath convert path like '/' '/root' '/a/b'
func StandardizationPath(path string) string {
	path = strings.TrimSuffix(path, "/")
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return path
}

func PathEqual(path1, path2 string) bool {
	return StandardizationPath(path1) == StandardizationPath(path2)
}
