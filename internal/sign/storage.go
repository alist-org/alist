package sign

import (
	"github.com/alist-org/alist/v3/internal/op"
	json "github.com/json-iterator/go"
)

func IsStorageSigned(rawPath string) bool {
	storage := op.GetBalancedStorage(rawPath).GetStorage()
	var jsonData = map[string]interface{}{}
	err := json.Unmarshal([]byte(storage.Addition), &jsonData)
	if err != nil {
		return false
	}
	if jsonData["sign"] == "true" {
		return true
	}
	return false
}
