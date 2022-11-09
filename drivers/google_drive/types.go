package google_drive

import (
	"strconv"
	"time"

	"github.com/alist-org/alist/v3/internal/model"
	log "github.com/sirupsen/logrus"
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
	Id              string    `json:"id"`
	Name            string    `json:"name"`
	MimeType        string    `json:"mimeType"`
	ModifiedTime    time.Time `json:"modifiedTime"`
	Size            string    `json:"size"`
	ThumbnailLink   string    `json:"thumbnailLink"`
	ShortcutDetails struct {
		TargetId       string `json:"targetId"`
		TargetMimeType string `json:"targetMimeType"`
	} `json:"shortcutDetails"`
}

func fileToObj(f File) *model.ObjThumb {
	log.Debugf("google file: %+v", f)
	size, _ := strconv.ParseInt(f.Size, 10, 64)
	obj := &model.ObjThumb{
		Object: model.Object{
			ID:       f.Id,
			Name:     f.Name,
			Size:     size,
			Modified: f.ModifiedTime,
			IsFolder: f.MimeType == "application/vnd.google-apps.folder",
		},
		Thumbnail: model.Thumbnail{},
	}
	if f.MimeType == "application/vnd.google-apps.shortcut" {
		obj.ID = f.ShortcutDetails.TargetId
		obj.IsFolder = f.ShortcutDetails.TargetMimeType == "application/vnd.google-apps.folder"
	}
	return obj
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
