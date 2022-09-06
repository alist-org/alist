package setting

import (
	"strconv"

	"github.com/alist-org/alist/v3/internal/db"
)

func GetStr(key string, defaultValue ...string) string {
	val, ok := db.GetSettingsMap()[key]
	if !ok {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return ""
	}
	return val
}

func GetInt(key string, defaultVal int) int {
	i, err := strconv.Atoi(GetStr(key))
	if err != nil {
		return defaultVal
	}
	return i
}

func GetBool(key string) bool {
	return GetStr(key) == "true" || GetStr(key) == "1"
}
