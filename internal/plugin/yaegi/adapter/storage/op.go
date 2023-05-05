package yaegi_storage

import (
	"context"

	"github.com/alist-org/alist/v3/internal/db"
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/op"
)

func RegisterPluginDriver(driver op.New) {
	op.RegisterDriver(driver)
}

func UnRegisterPluginDriver(driver op.New) {
	op.UnRegisterDriver(driver)
}

func DropPluginStorage(driverName string) {
	for _, storage := range op.GetAllStorages() {
		if storage.GetStorage().Driver == driverName {
			op.DropStorage(context.TODO(), *storage.GetStorage())
		}
	}
}

func LoadPluginStorage(driverName string) {
	storages, _ := db.GetEnabledStorages()
	for _, storage := range storages {
		if storage.Driver == driverName {
			go func(storage model.Storage) {
				op.LoadStorage(context.TODO(), storage)
			}(storage)
		}
	}
}

func MustSaveDriverStorage(driver driver.Driver) {
	op.MustSaveDriverStorage(driver)
}
