package yaegi_storage

import (
	"github.com/alist-org/alist/v3/internal/driver"
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
