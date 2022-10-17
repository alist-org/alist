package google_photo

import (
	"reflect"
	"time"

	"github.com/alist-org/alist/v3/internal/model"
)

type TokenError struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

type Items struct {
	NextPageToken string      `json:"nextPageToken"`
	MediaItems    []MediaItem `json:"mediaItems,omitempty"`
	Albums        []MediaItem `json:"albums,omitempty"`
	SharedAlbums  []MediaItem `json:"sharedAlbums,omitempty"`
}

type MediaItem struct {
	Id                string        `json:"id"`
	Title             string        `json:"title,omitempty"`
	BaseURL           string        `json:"baseUrl,omitempty"`
	CoverPhotoBaseUrl string        `json:"coverPhotoBaseUrl,omitempty"`
	MimeType          string        `json:"mimeType,omitempty"`
	FileName          string        `json:"filename,omitempty"`
	MediaMetadata MediaMetadata     `json:"mediaMetadata,omitempty"`
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
	if !reflect.DeepEqual(f.MediaMetadata, MediaMetadata{}){
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
	return &model.ObjThumb{
		Object: model.Object{
			ID:       f.Id,
			Name:     f.Title,
			Size:     0,
			Modified: time.Time{},
			IsFolder: true,
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
