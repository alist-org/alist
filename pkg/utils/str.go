package utils

import (
	"encoding/base64"
	"strings"

	"github.com/alist-org/alist/v3/internal/conf"
)

func MappingName(name string) string {
	for k, v := range conf.FilenameCharMap {
		name = strings.ReplaceAll(name, k, v)
	}
	return name
}

var DEC = map[string]string{
	"-": "+",
	"_": "/",
	".": "=",
}

func SafeAtob(data string) (string, error) {
	for k, v := range DEC {
		data = strings.ReplaceAll(data, k, v)
	}
	bytes, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return "", err
	}
	return string(bytes), err
}

// GetNoneEmpty returns the first non-empty string, return empty if all empty
func GetNoneEmpty(strArr ...string) string {
	for _, s := range strArr {
		if len(s) > 0 {
			return s
		}
	}
	return ""
}
