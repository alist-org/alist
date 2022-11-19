package _123

import (
	"strconv"
	"time"

	"github.com/alist-org/alist/v3/internal/model"
)

//type BaseResp struct {
//	Code    interface{} `json:"code"`
//	Message string      `json:"message"`
//}

type TokenResp struct {
	//BaseResp
	Data struct {
		Token string `json:"token"`
	} `json:"data"`
}

type File struct {
	FileName    string    `json:"FileName"`
	Size        int64     `json:"Size"`
	UpdateAt    time.Time `json:"UpdateAt"`
	FileId      int64     `json:"FileId"`
	Type        int       `json:"Type"`
	Etag        string    `json:"Etag"`
	S3KeyFlag   string    `json:"S3KeyFlag"`
	DownloadUrl string    `json:"DownloadUrl"`
}

func (f File) GetPath() string {
	return ""
}

func (f File) GetSize() int64 {
	return f.Size
}

func (f File) GetName() string {
	return f.FileName
}

func (f File) ModTime() time.Time {
	return f.UpdateAt
}

func (f File) IsDir() bool {
	return f.Type == 1
}

func (f File) GetID() string {
	return strconv.FormatInt(f.FileId, 10)
}

var _ model.Obj = (*File)(nil)

//func (f File) Thumb() string {
//
//}
//var _ model.Thumb = (*File)(nil)

type Files struct {
	//BaseResp
	Data struct {
		InfoList []File `json:"InfoList"`
		Next     string `json:"Next"`
	} `json:"data"`
}

//type DownResp struct {
//	//BaseResp
//	Data struct {
//		DownloadUrl string `json:"DownloadUrl"`
//	} `json:"data"`
//}

type UploadResp struct {
	//BaseResp
	Data struct {
		AccessKeyId     string `json:"AccessKeyId"`
		Bucket          string `json:"Bucket"`
		Key             string `json:"Key"`
		SecretAccessKey string `json:"SecretAccessKey"`
		SessionToken    string `json:"SessionToken"`
		FileId          int64  `json:"FileId"`
		Reuse           bool   `json:"Reuse"`
	} `json:"data"`
}
