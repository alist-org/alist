package yaegi_storage

import (
	"github.com/alist-org/alist/v3/internal/driver"
)

type DriverPluginResult interface {
	driver.Meta
	driver.Reader
	driver.GetRooter
	driver.MkdirResult
	driver.MoveResult
	driver.RenameResult
	driver.CopyResult
	driver.PutResult
	driver.Remove
	driver.Other
}

type DriverPlugin interface {
	driver.Meta
	driver.Reader
	driver.GetRooter
	driver.Mkdir
	driver.Move
	driver.Rename
	driver.Copy
	driver.Put
	driver.Remove
	driver.Other
}
