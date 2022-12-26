package mediatrack

import (
	"time"

	"github.com/alist-org/alist/v3/internal/model"
)

type BaseResp struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}
type File struct {
	Category     int           `json:"category"`
	ChildAssets  []interface{} `json:"childAssets"`
	CommentCount int           `json:"comment_count"`
	CoverAsset   interface{}   `json:"cover_asset"`
	CoverAssetID string        `json:"cover_asset_id"`
	CreatedAt    time.Time     `json:"created_at"`
	DeletedAt    string        `json:"deleted_at"`
	Description  string        `json:"description"`
	File         *struct {
		Cover string `json:"cover"`
		Src   string `json:"src"`
	} `json:"file"`
	//FileID string `json:"file_id"`
	ID string `json:"id"`

	Size       string        `json:"size"`
	Thumbnails []interface{} `json:"thumbnails"`
	Title      string        `json:"title"`
	UpdatedAt  time.Time     `json:"updated_at"`
}

type ChildrenResp struct {
	Status string `json:"status"`
	Data   struct {
		Total  int    `json:"total"`
		Assets []File `json:"assets"`
	} `json:"data"`
	Path      string `json:"path"`
	TraceID   string `json:"trace_id"`
	RequestID string `json:"requestId"`
}

type UploadResp struct {
	Status string `json:"status"`
	Data   struct {
		Credentials struct {
			TmpSecretID  string    `json:"TmpSecretId"`
			TmpSecretKey string    `json:"TmpSecretKey"`
			Token        string    `json:"Token"`
			ExpiredTime  int       `json:"ExpiredTime"`
			Expiration   time.Time `json:"Expiration"`
			StartTime    int       `json:"StartTime"`
		} `json:"credentials"`
		Object string `json:"object"`
		Bucket string `json:"bucket"`
		Region string `json:"region"`
		URL    string `json:"url"`
		Size   string `json:"size"`
	} `json:"data"`
	Path      string `json:"path"`
	TraceID   string `json:"trace_id"`
	RequestID string `json:"requestId"`
}

type Object struct {
	model.Object
	model.Thumbnail
	ParentID string
}
