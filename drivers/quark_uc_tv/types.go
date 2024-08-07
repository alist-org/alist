package quark_uc_tv

import (
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/utils"
	"time"
)

type Resp struct {
	CommonRsp
	Errno     int    `json:"errno"`
	ErrorInfo string `json:"error_info"`
}

type CommonRsp struct {
	Status int    `json:"status"`
	ReqID  string `json:"req_id"`
}

type RefreshTokenAuthResp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Status       int    `json:"status"`
		Errno        int    `json:"errno"`
		ErrorInfo    string `json:"error_info"`
		ReqID        string `json:"req_id"`
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
		Scope        string `json:"scope"`
	} `json:"data"`
}
type Files struct {
	Fid          string `json:"fid"`
	ParentFid    string `json:"parent_fid"`
	Category     int    `json:"category"`
	Filename     string `json:"filename"`
	Size         int64  `json:"size"`
	FileType     string `json:"file_type"`
	SubItems     int    `json:"sub_items,omitempty"`
	Isdir        int    `json:"isdir"`
	Duration     int    `json:"duration"`
	CreatedAt    int64  `json:"created_at"`
	UpdatedAt    int64  `json:"updated_at"`
	IsBackup     int    `json:"is_backup"`
	ThumbnailURL string `json:"thumbnail_url,omitempty"`
}

func (f *Files) GetSize() int64 {
	return f.Size
}

func (f *Files) GetName() string {
	return f.Filename
}

func (f *Files) ModTime() time.Time {
	//return time.Unix(f.UpdatedAt, 0)
	return time.Unix(0, f.UpdatedAt*int64(time.Millisecond))
}

func (f *Files) CreateTime() time.Time {
	//return time.Unix(f.CreatedAt, 0)
	return time.Unix(0, f.CreatedAt*int64(time.Millisecond))
}

func (f *Files) IsDir() bool {
	return f.Isdir == 1
}

func (f *Files) GetHash() utils.HashInfo {
	return utils.HashInfo{}
}

func (f *Files) GetID() string {
	return f.Fid
}

func (f *Files) GetPath() string {
	return ""
}

var _ model.Obj = (*Files)(nil)

type FilesData struct {
	CommonRsp
	Data struct {
		TotalCount int64   `json:"total_count"`
		Files      []Files `json:"files"`
	} `json:"data"`
}

type FileLink struct {
	CommonRsp
	Data struct {
		Fid         string `json:"fid"`
		FileName    string `json:"file_name"`
		Size        int64  `json:"size"`
		DownloadURL string `json:"download_url"`
	} `json:"data"`
}
