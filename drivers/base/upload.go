package base

import (
	"fmt"
	"strings"
	"time"

	"github.com/Xhofe/go-cache"
	"github.com/alist-org/alist/v3/internal/driver"
)

// storage upload progress, for upload recovery
var UploadStateCache = cache.NewMemCache(cache.WithShards[any](32))

// Save upload progress for 20 minutes
func SaveUploadProgress(driver driver.Driver, state any, keys ...string) bool {
	return UploadStateCache.Set(
		fmt.Sprint(driver.Config().Name, "-upload-", strings.Join(keys, "-")),
		state,
		cache.WithEx[any](time.Minute*20))
}

// An upload progress can only be made by one process alone,
// so here you need to get it and then delete it.
func GetUploadProgress[T any](driver driver.Driver, keys ...string) (state T, ok bool) {
	v, ok := UploadStateCache.GetDel(fmt.Sprint(driver.Config().Name, "-upload-", strings.Join(keys, "-")))
	if ok {
		state, ok = v.(T)
	}
	return
}
