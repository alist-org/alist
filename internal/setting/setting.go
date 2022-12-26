package setting

import (
	"strconv"

	"github.com/alist-org/alist/v3/internal/op"
)

func GetStr(key string, defaultValue ...string) string {
	val, _ := op.GetSettingItemByKey(key)
	if val == nil {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return ""
	}
	return val.Value
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
