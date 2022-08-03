package setting

import (
	"strconv"

	"github.com/alist-org/alist/v3/internal/db"
)

func GetByKey(key string, defaultValue ...string) string {
	val, ok := db.GetSettingsMap()[key]
	if !ok {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return ""
	}
	return val
}

func GetIntSetting(key string, defaultVal int) int {
	i, err := strconv.Atoi(GetByKey(key))
	if err != nil {
		return defaultVal
	}
	return i
}

func IsTrue(key string) bool {
	return GetByKey(key) == "true" || GetByKey(key) == "1"
}
