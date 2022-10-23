package _115

import (
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/deadblue/elevengo"
	"time"
)

var (
	UploadMaxSize       = 5 * 1024 * 1024 * 1024
	UploadSimplyMaxSize = 200 * 1024 * 1024
)

type File elevengo.File

func (f File) GetPath() string {
	return ""
}

func (f File) GetSize() int64 {
	return f.Size
}

func (f File) GetName() string {
	return f.Name
}

func (f File) ModTime() time.Time {
	return f.UpdateTime
}

func (f File) IsDir() bool {
	return f.IsDirectory
}

func (f File) GetID() string {
	return f.FileId
}

var _ model.Obj = (*File)(nil)

// func fileToObj(f File) *model.ObjThumb {
// 	return &model.ObjThumb{
// 		Object: model.Object{
// 			ID:       f.FileId,
// 			Name:     f.Name,
// 			Size:     f.Size,
// 			Modified: f.UpdateTime,
// 			IsFolder: f.IsDirectory,
// 		},
// 	}
// }
