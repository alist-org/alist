package fs

import (
	"github.com/alist-org/alist/v3/internal/driver"
)

func containsByName(files []driver.FileInfo, file driver.FileInfo) bool {
	for _, f := range files {
		if f.GetName() == file.GetName() {
			return true
		}
	}
	return false
}
