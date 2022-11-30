package utils

import (
	"strings"

	"github.com/alist-org/alist/v3/internal/conf"
)

func MappingName(name string) string {
	for k, v := range conf.FilenameCharMap {
		name = strings.ReplaceAll(name, k, v)
	}
	return name
}
