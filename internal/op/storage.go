package op

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/alist-org/alist/v3/internal/db"
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/generic_sync"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// Although the driver type is stored,
// there is a storage in each driver,
// so it should actually be a storage, just wrapped by the driver
var storagesMap generic_sync.MapOf[string, driver.Driver]

func GetStorageByVirtualPath(virtualPath string) (driver.Driver, error) {
	storageDriver, ok := storagesMap.Load(virtualPath)
	if !ok {
		return nil, errors.Errorf("no mount path for an storage is: %s", virtualPath)
	}
	return storageDriver, nil
}

// CreateStorage Save the storage to database so storage can get an id
// then instantiate corresponding driver and save it in memory
func CreateStorage(ctx context.Context, storage model.Storage) error {
	storage.Modified = time.Now()
	storage.MountPath = utils.StandardizePath(storage.MountPath)
	var err error
	// check driver first
	driverName := storage.Driver
	driverNew, err := GetDriverNew(driverName)
	if err != nil {
		return errors.WithMessage(err, "failed get driver new")
	}
	storageDriver := driverNew()
	// insert storage to database
	err = db.CreateStorage(&storage)
	if err != nil {
		return errors.WithMessage(err, "failed create storage in database")
	}
	// already has an id
	err = storageDriver.Init(ctx, storage)
	storagesMap.Store(storage.MountPath, storageDriver)
	if err != nil {
		storageDriver.GetStorage().SetStatus(fmt.Sprintf("%+v", err.Error()))
		MustSaveDriverStorage(storageDriver)
		return errors.Wrapf(err, "failed init storage but storage is already created")
	} else {
		storageDriver.GetStorage().SetStatus(WORK)
		MustSaveDriverStorage(storageDriver)
	}
	log.Debugf("storage %+v is created", storageDriver)
	return nil
}

// LoadStorage load exist storage in db to memory
func LoadStorage(ctx context.Context, storage model.Storage) error {
	storage.MountPath = utils.StandardizePath(storage.MountPath)
	// check driver first
	driverName := storage.Driver
	driverNew, err := GetDriverNew(driverName)
	if err != nil {
		return errors.WithMessage(err, "failed get driver new")
	}
	storageDriver := driverNew()
	err = storageDriver.Init(ctx, storage)
	storagesMap.Store(storage.MountPath, storageDriver)
	if err != nil {
		storageDriver.GetStorage().SetStatus(fmt.Sprintf("%+v", err.Error()))
		MustSaveDriverStorage(storageDriver)
		return errors.Wrapf(err, "failed init storage")
	} else {
		storageDriver.GetStorage().SetStatus(WORK)
		MustSaveDriverStorage(storageDriver)
	}
	log.Debugf("storage %+v is created", storageDriver)
	return nil
}

func EnableStorage(ctx context.Context, id uint) error {
	storage, err := db.GetStorageById(id)
	if err != nil {
		return errors.WithMessage(err, "failed get storage")
	}
	if !storage.Disabled {
		return errors.Errorf("this storage have enabled")
	}
	storage.Disabled = false
	err = db.UpdateStorage(storage)
	if err != nil {
		return errors.WithMessage(err, "failed update storage in db")
	}
	err = LoadStorage(ctx, *storage)
	if err != nil {
		return errors.WithMessage(err, "failed load storage")
	}
	return nil
}

func DisableStorage(ctx context.Context, id uint) error {
	storage, err := db.GetStorageById(id)
	if err != nil {
		return errors.WithMessage(err, "failed get storage")
	}
	if storage.Disabled {
		return errors.Errorf("this storage have disabled")
	}
	storageDriver, err := GetStorageByVirtualPath(storage.MountPath)
	if err != nil {
		return errors.WithMessage(err, "failed get storage driver")
	}
	// drop the storage in the driver
	if err := storageDriver.Drop(ctx); err != nil {
		return errors.Wrapf(err, "failed drop storage")
	}
	// delete the storage in the memory
	storage.Disabled = true
	err = db.UpdateStorage(storage)
	if err != nil {
		return errors.WithMessage(err, "failed update storage in db")
	}
	storagesMap.Delete(storage.MountPath)
	return nil
}

// UpdateStorage update storage
// get old storage first
// drop the storage then reinitialize
func UpdateStorage(ctx context.Context, storage model.Storage) error {
	oldStorage, err := db.GetStorageById(storage.ID)
	if err != nil {
		return errors.WithMessage(err, "failed get old storage")
	}
	if oldStorage.Driver != storage.Driver {
		return errors.Errorf("driver cannot be changed")
	}
	storage.Modified = time.Now()
	storage.MountPath = utils.StandardizePath(storage.MountPath)
	err = db.UpdateStorage(&storage)
	if err != nil {
		return errors.WithMessage(err, "failed update storage in database")
	}
	if storage.Disabled {
		return nil
	}
	storageDriver, err := GetStorageByVirtualPath(oldStorage.MountPath)
	if oldStorage.MountPath != storage.MountPath {
		// mount path renamed, need to drop the storage
		storagesMap.Delete(oldStorage.MountPath)
	}
	if err != nil {
		return errors.WithMessage(err, "failed get storage driver")
	}
	err = storageDriver.Drop(ctx)
	if err != nil {
		return errors.Wrapf(err, "failed drop storage")
	}
	err = storageDriver.Init(ctx, storage)
	storagesMap.Store(storage.MountPath, storageDriver)
	if err != nil {
		storageDriver.GetStorage().SetStatus(fmt.Sprintf("%+v", err.Error()))
		MustSaveDriverStorage(storageDriver)
		return errors.Wrapf(err, "failed init storage")
	} else {
		storageDriver.GetStorage().SetStatus(WORK)
		MustSaveDriverStorage(storageDriver)
	}
	return nil
}

func DeleteStorageById(ctx context.Context, id uint) error {
	storage, err := db.GetStorageById(id)
	if err != nil {
		return errors.WithMessage(err, "failed get storage")
	}
	storageDriver, err := GetStorageByVirtualPath(storage.MountPath)
	if err != nil {
		return errors.WithMessage(err, "failed get storage driver")
	}
	// drop the storage in the driver
	if err := storageDriver.Drop(ctx); err != nil {
		return errors.Wrapf(err, "failed drop storage")
	}
	// delete the storage in the database
	if err := db.DeleteStorageById(id); err != nil {
		return errors.WithMessage(err, "failed delete storage in database")
	}
	// delete the storage in the memory
	storagesMap.Delete(storage.MountPath)
	return nil
}

// MustSaveDriverStorage call from specific driver
func MustSaveDriverStorage(driver driver.Driver) {
	err := saveDriverStorage(driver)
	if err != nil {
		log.Errorf("failed save driver storage: %s", err)
	}
}

func saveDriverStorage(driver driver.Driver) error {
	storage := driver.GetStorage()
	addition := driver.GetAddition()
	bytes, err := utils.Json.Marshal(addition)
	if err != nil {
		return errors.Wrap(err, "error while marshal addition")
	}
	storage.Addition = string(bytes)
	err = db.UpdateStorage(storage)
	if err != nil {
		return errors.WithMessage(err, "failed update storage in database")
	}
	return nil
}

// getStoragesByPath get storage by longest match path, contains balance storage.
// for example, there is /a/b,/a/c,/a/d/e,/a/d/e.balance
// getStoragesByPath(/a/d/e/f) => /a/d/e,/a/d/e.balance
func getStoragesByPath(path string) []driver.Driver {
	storages := make([]driver.Driver, 0)
	curSlashCount := 0
	storagesMap.Range(func(key string, value driver.Driver) bool {
		virtualPath := utils.GetActualVirtualPath(value.GetStorage().MountPath)
		if virtualPath == "/" {
			virtualPath = ""
		}
		// not this
		if path != virtualPath && !strings.HasPrefix(path, virtualPath+"/") {
			return true
		}
		slashCount := strings.Count(virtualPath, "/")
		// not the longest match
		if slashCount < curSlashCount {
			return true
		}
		if slashCount > curSlashCount {
			storages = storages[:0]
			curSlashCount = slashCount
		}
		storages = append(storages, value)
		return true
	})
	// make sure the order is the same for same input
	sort.Slice(storages, func(i, j int) bool {
		return storages[i].GetStorage().MountPath < storages[j].GetStorage().MountPath
	})
	return storages
}

// GetStorageVirtualFilesByPath Obtain the virtual file generated by the storage according to the path
// for example, there are: /a/b,/a/c,/a/d/e,/a/b.balance1,/av
// GetStorageVirtualFilesByPath(/a) => b,c,d
func GetStorageVirtualFilesByPath(prefix string) []model.Obj {
	files := make([]model.Obj, 0)
	storages := storagesMap.Values()
	sort.Slice(storages, func(i, j int) bool {
		if storages[i].GetStorage().Order == storages[j].GetStorage().Order {
			return storages[i].GetStorage().MountPath < storages[j].GetStorage().MountPath
		}
		return storages[i].GetStorage().Order < storages[j].GetStorage().Order
	})
	prefix = utils.StandardizePath(prefix)
	if prefix != "/" {
		prefix += "/"
	}
	set := make(map[string]interface{})
	for _, v := range storages {
		// TODO should save a balanced storage
		// balance storage
		if utils.IsBalance(v.GetStorage().MountPath) {
			continue
		}
		virtualPath := v.GetStorage().MountPath
		if len(virtualPath) <= len(prefix) {
			continue
		}
		// not prefixed with `prefix`
		if !strings.HasPrefix(virtualPath, prefix) {
			continue
		}
		name := strings.Split(strings.TrimPrefix(virtualPath, prefix), "/")[0]
		if _, ok := set[name]; ok {
			continue
		}
		files = append(files, &model.Object{
			Name:     name,
			Size:     0,
			Modified: v.GetStorage().Modified,
			IsFolder: true,
		})
		set[name] = nil
	}
	return files
}

var balanceMap generic_sync.MapOf[string, int]

// GetBalancedStorage get storage by path
func GetBalancedStorage(path string) driver.Driver {
	path = utils.StandardizePath(path)
	storages := getStoragesByPath(path)
	storageNum := len(storages)
	switch storageNum {
	case 0:
		return nil
	case 1:
		return storages[0]
	default:
		virtualPath := utils.GetActualVirtualPath(storages[0].GetStorage().MountPath)
		cur, ok := balanceMap.Load(virtualPath)
		i := 0
		if ok {
			i = cur
			i = (i + 1) % storageNum
			balanceMap.Store(virtualPath, i)
		} else {
			balanceMap.Store(virtualPath, i)
		}
		return storages[i]
	}
}
