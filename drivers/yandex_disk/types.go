package yandex_disk

import (
	"time"

	"github.com/alist-org/alist/v3/internal/model"
)

type TokenErrResp struct {
	ErrorDescription string `json:"error_description"`
	Error            string `json:"error"`
}

type ErrResp struct {
	Message     string `json:"message"`
	Description string `json:"description"`
	Error       string `json:"error"`
}

type File struct {
	//AntivirusStatus string `json:"antivirus_status"`
	Size int64 `json:"size"`
	//CommentIds      struct {
	//	PrivateResource string `json:"private_resource"`
	//	PublicResource  string `json:"public_resource"`
	//} `json:"comment_ids"`
	Name string `json:"name"`
	//Exif struct {
	//	DateTime time.Time `json:"date_time"`
	//} `json:"exif"`
	//Created    time.Time `json:"created"`
	//ResourceId string    `json:"resource_id"`
	Modified time.Time `json:"modified"`
	//MimeType   string    `json:"mime_type"`
	File string `json:"file"`
	//MediaType  string    `json:"media_type"`
	Preview string `json:"preview"`
	Path    string `json:"path"`
	//Sha256     string    `json:"sha256"`
	Type string `json:"type"`
	//Md5        string    `json:"md5"`
	//Revision   int64     `json:"revision"`
}

func fileToObj(f File) model.Obj {
	return &model.Object{
		Name:     f.Name,
		Size:     f.Size,
		Modified: f.Modified,
		IsFolder: f.Type == "dir",
	}
}

type FilesResp struct {
	Embedded struct {
		Sort   string `json:"sort"`
		Items  []File `json:"items"`
		Limit  int    `json:"limit"`
		Offset int    `json:"offset"`
		Path   string `json:"path"`
		Total  int    `json:"total"`
	} `json:"_embedded"`
	Name string `json:"name"`
	Exif struct {
	} `json:"exif"`
	ResourceId string    `json:"resource_id"`
	Created    time.Time `json:"created"`
	Modified   time.Time `json:"modified"`
	Path       string    `json:"path"`
	CommentIds struct {
	} `json:"comment_ids"`
	Type     string `json:"type"`
	Revision int64  `json:"revision"`
}

type DownResp struct {
	Href      string `json:"href"`
	Method    string `json:"method"`
	Templated bool   `json:"templated"`
}

type UploadResp struct {
	OperationId string `json:"operation_id"`
	Href        string `json:"href"`
	Method      string `json:"method"`
	Templated   bool   `json:"templated"`
}
