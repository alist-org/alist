package template

import (
	"github.com/Xhofe/wopan-sdk-go"
	"github.com/alist-org/alist/v3/internal/model"
)

type Object struct {
	model.ObjThumb
	FID string
}

func fileToObj(file wopan.File) (model.Obj, error) {
	t, err := getTime(file.CreateTime)
	if err != nil {
		return nil, err
	}
	return &Object{
		ObjThumb: model.ObjThumb{
			Object: model.Object{
				ID: file.Id,
				//Path:     "",
				Name:     file.Name,
				Size:     file.Size,
				Modified: t,
				IsFolder: file.Type == 0,
			},
			Thumbnail: model.Thumbnail{
				Thumbnail: file.ThumbUrl,
			},
		},
		FID: file.Fid,
	}, nil
}
