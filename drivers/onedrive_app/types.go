package onedrive_app

import (
	"time"

	"github.com/alist-org/alist/v3/internal/model"
)

type Host struct {
	Oauth string
	Api   string
}

type TokenErr struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

type RespErr struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

type File struct {
	Id                   string    `json:"id"`
	Name                 string    `json:"name"`
	Size                 int64     `json:"size"`
	LastModifiedDateTime time.Time `json:"lastModifiedDateTime"`
	Url                  string    `json:"@microsoft.graph.downloadUrl"`
	File                 *struct {
		MimeType string `json:"mimeType"`
	} `json:"file"`
	Thumbnails []struct {
		Medium struct {
			Url string `json:"url"`
		} `json:"medium"`
	} `json:"thumbnails"`
	ParentReference struct {
		DriveId string `json:"driveId"`
	} `json:"parentReference"`
}

type Object struct {
	model.ObjThumb
	ParentID string
}

func fileToObj(f File, parentID string) *Object {
	thumb := ""
	if len(f.Thumbnails) > 0 {
		thumb = f.Thumbnails[0].Medium.Url
	}
	return &Object{
		ObjThumb: model.ObjThumb{
			Object: model.Object{
				ID:       f.Id,
				Name:     f.Name,
				Size:     f.Size,
				Modified: f.LastModifiedDateTime,
				IsFolder: f.File == nil,
			},
			Thumbnail: model.Thumbnail{Thumbnail: thumb},
			//Url:       model.Url{Url: f.Url},
		},
		ParentID: parentID,
	}
}

type Files struct {
	Value    []File `json:"value"`
	NextLink string `json:"@odata.nextLink"`
}
