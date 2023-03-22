package sign

import (
	"github.com/alist-org/alist/v3/internal/op"
	json "github.com/json-iterator/go"
)

type signType struct {
	Sign bool `json:"sign"`
}

func IsStorageSigned(rawPath string) bool {
	var jsonData signType
	storage := op.GetBalancedStorage(rawPath).GetStorage()
	err := json.Unmarshal([]byte(storage.Addition), &jsonData)
	if err != nil {
		return false
	}
	if jsonData.Sign == true {
		return true
	}
	return false
}
