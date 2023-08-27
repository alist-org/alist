package aliyundrive_open

import (
	"github.com/alist-org/alist/v3/pkg/utils"
	"time"

	"github.com/alist-org/alist/v3/internal/model"
)

type ErrResp struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type Files struct {
	Items      []File `json:"items"`
	NextMarker string `json:"next_marker"`
}

type File struct {
	DriveId       string    `json:"drive_id"`
	FileId        string    `json:"file_id"`
	ParentFileId  string    `json:"parent_file_id"`
	Name          string    `json:"name"`
	Size          int64     `json:"size"`
	FileExtension string    `json:"file_extension"`
	ContentHash   string    `json:"content_hash"`
	Category      string    `json:"category"`
	Type          string    `json:"type"`
	Thumbnail     string    `json:"thumbnail"`
	Url           string    `json:"url"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

	// create only
	FileName string `json:"file_name"`
}

func fileToObj(f File) *model.ObjThumb {
	if f.Name == "" {
		f.Name = f.FileName
	}
	return &model.ObjThumb{
		Object: model.Object{
			ID:       f.FileId,
			Name:     f.Name,
			Size:     f.Size,
			Modified: f.UpdatedAt,
			IsFolder: f.Type == "folder",
			Ctime:    f.CreatedAt,
			HashInfo: utils.NewHashInfo(utils.SHA1, f.ContentHash),
		},
		Thumbnail: model.Thumbnail{Thumbnail: f.Thumbnail},
	}
}

type PartInfo struct {
	Etag        interface{} `json:"etag"`
	PartNumber  int         `json:"part_number"`
	PartSize    interface{} `json:"part_size"`
	UploadUrl   string      `json:"upload_url"`
	ContentType string      `json:"content_type"`
}

type CreateResp struct {
	//Type         string `json:"type"`
	//ParentFileId string `json:"parent_file_id"`
	//DriveId      string `json:"drive_id"`
	FileId string `json:"file_id"`
	//RevisionId   string `json:"revision_id"`
	//EncryptMode  string `json:"encrypt_mode"`
	//DomainId     string `json:"domain_id"`
	//FileName     string `json:"file_name"`
	UploadId string `json:"upload_id"`
	//Location     string `json:"location"`
	RapidUpload  bool       `json:"rapid_upload"`
	PartInfoList []PartInfo `json:"part_info_list"`
}

type MoveOrCopyResp struct {
	Exist   bool   `json:"exist"`
	DriveID string `json:"drive_id"`
	FileID  string `json:"file_id"`
}
