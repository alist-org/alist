package _89

import (
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/utils"
	"path"
)

type Cloud189Error struct {
	ErrorCode string `json:"errorCode"`
	ErrorMsg  string `json:"errorMsg"`
}

type Cloud189File struct {
	Id         int64  `json:"id"`
	LastOpTime string `json:"lastOpTime"`
	Name       string `json:"name"`
	Size       int64  `json:"size"`
	Icon       struct {
		SmallUrl string `json:"smallUrl"`
		//LargeUrl string `json:"largeUrl"`
	} `json:"icon"`
	Url string `json:"url"`
}

func (f Cloud189File) GetSize() uint64 {
	if f.Size == -1 {
		return 0
	}
	return uint64(f.Size)
}

func (f Cloud189File) GetName() string {
	return f.Name
}

func (f Cloud189File) GetType() int {
	if f.Size == -1 {
		return conf.FOLDER
	}
	return utils.GetFileType(path.Ext(f.Name))
}

type Cloud189Folder struct {
	Id         int64  `json:"id"`
	LastOpTime string `json:"lastOpTime"`
	Name       string `json:"name"`
}

type Cloud189Files struct {
	ResCode    int    `json:"res_code"`
	ResMessage string `json:"res_message"`
	FileListAO struct {
		Count      int              `json:"count"`
		FileList   []Cloud189File   `json:"fileList"`
		FolderList []Cloud189Folder `json:"folderList"`
	} `json:"fileListAO"`
}

type UploadUrlsResp struct {
	Code       string          `json:"code"`
	UploadUrls map[string]Part `json:"uploadUrls"`
}

type Part struct {
	RequestURL    string `json:"requestURL"`
	RequestHeader string `json:"requestHeader"`
}

//type Info struct {
//	SessionKey string
//	Rsa        Rsa
//}

type Rsa struct {
	Expire int64  `json:"expire"`
	PkId   string `json:"pkId"`
	PubKey string `json:"pubKey"`
}

type Cloud189Down struct {
	ResCode         int    `json:"res_code"`
	ResMessage      string `json:"res_message"`
	FileDownloadUrl string `json:"fileDownloadUrl"`
}

type DownResp struct {
	ResCode         int    `json:"res_code"`
	ResMessage      string `json:"res_message"`
	FileDownloadUrl string `json:"downloadUrl"`
}
