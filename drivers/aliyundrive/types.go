package aliyundrive

import (
	"time"

	"github.com/alist-org/alist/v3/internal/model"
)

type RespErr struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type Files struct {
	Items      []File `json:"items"`
	NextMarker string `json:"next_marker"`
}

type File struct {
	DriveId       string     `json:"drive_id"`
	CreatedAt     *time.Time `json:"created_at"`
	FileExtension string     `json:"file_extension"`
	FileId        string     `json:"file_id"`
	Type          string     `json:"type"`
	Name          string     `json:"name"`
	Category      string     `json:"category"`
	ParentFileId  string     `json:"parent_file_id"`
	UpdatedAt     time.Time  `json:"updated_at"`
	Size          int64      `json:"size"`
	Thumbnail     string     `json:"thumbnail"`
	Url           string     `json:"url"`
}

func fileToObj(f File) *model.ObjThumb {
	return &model.ObjThumb{
		Object: model.Object{
			ID:       f.FileId,
			Name:     f.Name,
			Size:     f.Size,
			Modified: f.UpdatedAt,
			IsFolder: f.Type == "folder",
		},
		Thumbnail: model.Thumbnail{Thumbnail: f.Thumbnail},
	}
}

type UploadResp struct {
	FileId       string `json:"file_id"`
	UploadId     string `json:"upload_id"`
	PartInfoList []struct {
		UploadUrl         string `json:"upload_url"`
		InternalUploadUrl string `json:"internal_upload_url"`
	} `json:"part_info_list"`

	RapidUpload bool `json:"rapid_upload"`
}
