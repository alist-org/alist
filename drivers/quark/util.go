package quark

import (
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/model"
	"github.com/Xhofe/alist/utils"
	"path"
	"time"
)

func getTime(t int64) *time.Time {
	tm := time.UnixMilli(t)
	//log.Debugln(tm)
	return &tm
}

func (driver Quark) formatFile(f *File) *model.File {
	file := model.File{
		Id:        f.Fid,
		Name:      f.FileName,
		Size:      f.Size,
		Driver:    driver.Config().Name,
		UpdatedAt: getTime(f.UpdatedAt),
	}
	if f.File {
		file.Type = utils.GetFileType(path.Ext(f.FileName))
	} else {
		file.Type = conf.FOLDER
	}
	return &file
}
