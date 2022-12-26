package op

import (
	"strings"

	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// GetStorageAndActualPath Get the corresponding storage and actual path
// for path: remove the mount path prefix and join the actual root folder if exists
func GetStorageAndActualPath(rawPath string) (storage driver.Driver, actualPath string, err error) {
	rawPath = utils.FixAndCleanPath(rawPath)
	storage = GetBalancedStorage(rawPath)
	if storage == nil {
		err = errors.Errorf("can't find storage with rawPath: %s", rawPath)
		return
	}
	log.Debugln("use storage: ", storage.GetStorage().MountPath)
	mountPath := utils.GetActualMountPath(storage.GetStorage().MountPath)
	actualPath = utils.FixAndCleanPath(strings.TrimPrefix(rawPath, mountPath))
	return
}
