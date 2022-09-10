package google_drive

import (
	"strconv"
	"time"

	"github.com/alist-org/alist/v3/internal/model"
)

type TokenError struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

type Files struct {
	NextPageToken string `json:"nextPageToken"`
	Files         []File `json:"files"`
}

type File struct {
	Id            string    `json:"id"`
	Name          string    `json:"name"`
	MimeType      string    `json:"mimeType"`
	ModifiedTime  time.Time `json:"modifiedTime"`
	Size          string    `json:"size"`
	ThumbnailLink string    `json:"thumbnailLink"`
}

func fileToObj(f File) *model.ObjThumb {
	size, _ := strconv.ParseInt(f.Size, 10, 64)
	return &model.ObjThumb{
		Object: model.Object{
			ID:       f.Id,
			Name:     f.Name,
			Size:     size,
			Modified: time.Time{},
			IsFolder: f.MimeType == "application/vnd.google-apps.folder",
		},
		Thumbnail: model.Thumbnail{},
	}
}

type Error struct {
	Error struct {
		Errors []struct {
			Domain       string `json:"domain"`
			Reason       string `json:"reason"`
			Message      string `json:"message"`
			LocationType string `json:"location_type"`
			Location     string `json:"location"`
		}
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}
