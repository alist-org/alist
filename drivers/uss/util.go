package uss

import "strings"

// do others that not defined in Driver interface

func getKey(path string, dir bool) string {
	path = strings.TrimPrefix(path, "/")
	if dir {
		path += "/"
	}
	return path
}
