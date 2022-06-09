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
