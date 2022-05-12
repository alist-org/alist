package _23

import (
	"path"
	"time"

	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/utils"
)

type File struct {
	FileName    string     `json:"FileName"`
	Size        int64      `json:"Size"`
	UpdateAt    *time.Time `json:"UpdateAt"`
	FileId      int64      `json:"FileId"`
	Type        int        `json:"Type"`
	Etag        string     `json:"Etag"`
	S3KeyFlag   string     `json:"S3KeyFlag"`
	DownloadUrl string     `json:"DownloadUrl"`
}

func (f File) GetSize() uint64 {
	return uint64(f.Size)
}

func (f File) GetName() string {
	return f.FileName
}

func (f File) GetType() int {
	if f.Type == 1 {
		return conf.FOLDER
	}
	return utils.GetFileType(path.Ext(f.FileName))
}

type BaseResp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type Pan123TokenResp struct {
	BaseResp
	Data struct {
		Token string `json:"token"`
	} `json:"data"`
}

type Pan123Files struct {
	BaseResp
	Data struct {
		InfoList []File `json:"InfoList"`
		Next     string `json:"Next"`
	} `json:"data"`
}

type Pan123DownResp struct {
	BaseResp
	Data struct {
		DownloadUrl string `json:"DownloadUrl"`
	} `json:"data"`
}

type UploadResp struct {
	BaseResp
	Data struct {
		AccessKeyId     string `json:"AccessKeyId"`
		Bucket          string `json:"Bucket"`
		Key             string `json:"Key"`
		SecretAccessKey string `json:"SecretAccessKey"`
		SessionToken    string `json:"SessionToken"`
		FileId          int64  `json:"FileId"`
	} `json:"data"`
}
