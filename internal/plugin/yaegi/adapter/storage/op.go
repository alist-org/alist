package yaegi_storage

import (
	"context"

	"github.com/alist-org/alist/v3/internal/db"
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/op"
)

// yaegi use
type PluginNew func() DriverPlugin
type PluginResultNew func() DriverPluginResult

// yaegi use
func RegisterPluginResultDriver(driverNew PluginResultNew) {
	op.RegisterDriver(func() driver.Driver {
		return driverNew()
	})
}

func UnRegisterPluginResultDriver(driverNew PluginResultNew) {
	op.UnRegisterDriver(func() driver.Driver {
		return driverNew()
	})
}

func RegisterPluginDriver(driverNew PluginNew) {
	op.RegisterDriver(func() driver.Driver {
		return driverNew()
	})
}

func UnRegisterPluginDriver(driverNew PluginNew) {
	op.UnRegisterDriver(func() driver.Driver {
		return driverNew()
	})
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
	if storages != nil {
		for _, storage := range storages {
			if storage.Driver == driverName {
				go func(storage model.Storage) {
					op.LoadStorage(context.TODO(), storage)
				}(storage)
			}
		}
	}
}
