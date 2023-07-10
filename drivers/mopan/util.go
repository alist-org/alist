package mopan

import (
	"time"

	"github.com/alist-org/alist/v3/internal/model"
	"github.com/foxxorcat/mopan-sdk-go"
)

func fileToObj(f mopan.File) model.Obj {
	return &model.ObjThumb{
		Object: model.Object{
			ID:       string(f.ID),
			Name:     f.Name,
			Size:     int64(f.Size),
			Modified: time.Time(f.LastOpTime),
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
		},
		Thumbnail: model.Thumbnail{
			Thumbnail: thumb,
		},
	}
}
