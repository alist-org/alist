package setting

import (
	"github.com/alist-org/alist/v3/internal/db"
	"strconv"
)

func GetByKey(key string) string {
	return db.GetSettingsMap()[key]
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
