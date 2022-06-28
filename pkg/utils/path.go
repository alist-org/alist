package utils

import (
	stdpath "path"
	"path/filepath"
	"runtime"
	"strings"
)

// StandardizePath convert path like '/' '/root' '/a/b'
func StandardizePath(path string) string {
	path = strings.TrimSuffix(path, "/")
	// abs path
	if filepath.IsAbs(path) && runtime.GOOS == "windows" {
		return path
	}
	// relative path with prefix '..'
	if strings.HasPrefix(path, ".") {
		return path
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return path
}

// PathEqual judge path is equal
func PathEqual(path1, path2 string) bool {
	return StandardizePath(path1) == StandardizePath(path2)
}

func Ext(path string) string {
	ext := stdpath.Ext(path)
	if strings.HasPrefix(ext, ".") {
		return ext[1:]
	}
	return ext
}
