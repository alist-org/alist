package google_photo

import (
	"time"

	"github.com/alist-org/alist/v3/internal/model"
)

type TokenError struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

type Files struct {
	NextPageToken string      `json:"nextPageToken"`
	MediaItems    []MediaItem `json:"mediaItems"`
}

type MediaItem struct {
	Id            string        `json:"id"`
	BaseURL       string        `json:"baseUrl"`
	MimeType      string        `json:"mimeType"`
	FileName      string        `json:"filename"`
	MediaMetadata MediaMetadata `json:"mediaMetadata"`
}

type MediaMetadata struct {
	CreationTime time.Time `json:"creationTime"`
	Width        string    `json:"width"`
	Height       string    `json:"height"`
	Photo        Photo     `json:"photo,omitempty"`
	Video        Video     `json:"video,omitempty"`
}

type Photo struct {
}

type Video struct {
}

func fileToObj(f MediaItem) *model.ObjThumb {
	//size, _ := strconv.ParseInt(f.Size, 10, 64)
	return &model.ObjThumb{
		Object: model.Object{
			ID:       f.Id,
			Name:     f.FileName,
			Size:     0,
			Modified: f.MediaMetadata.CreationTime,
			IsFolder: false,
		},
		Thumbnail: model.Thumbnail{
			Thumbnail: f.BaseURL + "=w100-h100-c",
		},
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
