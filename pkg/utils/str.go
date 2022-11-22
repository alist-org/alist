package utils

import "strings"

func MappingName(name string, m map[string]string) string {
	for k, v := range m {
		name = strings.ReplaceAll(name, k, v)
	}
	return name
}
