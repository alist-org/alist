package mopan

import (
	"time"

	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/foxxorcat/mopan-sdk-go"
)

func fileToObj(f mopan.File) model.Obj {
	return &model.ObjThumb{
		Object: model.Object{
			ID:       string(f.ID),
			Name:     f.Name,
			Size:     int64(f.Size),
			Modified: time.Time(f.LastOpTime),
			Ctime:    time.Time(f.CreateDate),
			HashInfo: utils.NewHashInfo(utils.MD5, f.Md5),
		},
		Thumbnail: model.Thumbnail{
			Thumbnail: f.Icon.SmallURL,
		},
	}
}

func folderToObj(f mopan.Folder) model.Obj {
	return &model.Object{
		ID:       string(f.ID),
		Name:     f.Name,
		Modified: time.Time(f.LastOpTime),
		Ctime:    time.Time(f.CreateDate),
		IsFolder: true,
	}
}

func CloneObj(o model.Obj, newID, newName string) model.Obj {
	if o.IsDir() {
		return &model.Object{
			ID:       newID,
			Name:     newName,
			IsFolder: true,
			Modified: o.ModTime(),
			Ctime:    o.CreateTime(),
		}
	}

	thumb := ""
	if o, ok := o.(model.Thumb); ok {
		thumb = o.Thumb()
	}
	return &model.ObjThumb{
		Object: model.Object{
			ID:       newID,
			Name:     newName,
			Size:     o.GetSize(),
			Modified: o.ModTime(),
			Ctime:    o.CreateTime(),
			HashInfo: o.GetHash(),
		},
		Thumbnail: model.Thumbnail{
			Thumbnail: thumb,
		},
	}
}
