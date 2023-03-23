package sign

import (
	"github.com/alist-org/alist/v3/internal/op"
)

func IsStorageSigned(rawPath string) bool {
	storage := op.GetBalancedStorage(rawPath).GetStorage()
	if storage.EnableSign == true {
		return true
	}
	return false
}
