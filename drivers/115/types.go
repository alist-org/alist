package _115

import (
	"time"

	"github.com/SheltonZhu/115driver/pkg/driver"
	"github.com/alist-org/alist/v3/internal2/model"
	"github.com/alist-org/alist/v3/pkg/utils"
)

var _ model.Obj = (*FileObj)(nil)

type FileObj struct {
	driver.File
}

func (f *FileObj) CreateTime() time.Time {
	return f.File.CreateTime
}

func (f *FileObj) GetHash() utils.HashInfo {
	return utils.NewHashInfo(utils.SHA1, f.Sha1)
}
