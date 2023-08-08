package _115

import (
	"github.com/SheltonZhu/115driver/pkg/driver"
	"github.com/alist-org/alist/v3/internal/model"
	"time"
)

var _ model.Obj = (*FileObj)(nil)

type FileObj struct {
	driver.File
}

func (f FileObj) CreateTime() time.Time {
	return f.CreateTime()
}

func (f *FileObj) GetHash() (string, string) {
	return f.Sha1, model.SHA1
}
