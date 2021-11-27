package native

import (
	"github.com/Xhofe/alist/drivers"
)

func init() {
	drivers.RegisterDriver(driverName, &Native{})
}
