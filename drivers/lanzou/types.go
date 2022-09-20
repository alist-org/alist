package lanzou

import (
	"fmt"
	"time"

	"github.com/alist-org/alist/v3/internal/model"
)

type FilesOrFoldersResp struct {
	Text []FileOrFolder `json:"text"`
}

type FileOrFolder struct {
	Name string `json:"name"`
	//Onof        string `json:"onof"` // 是否存在提取码
	//IsLock      string `json:"is_lock"`
	//IsCopyright int    `json:"is_copyright"`

	// 文件通用
	ID      string `json:"id"`
	NameAll string `json:"name_all"`
	Size    string `json:"size"`
	Time    string `json:"time"`
	//Icon          string `json:"icon"`
	//Downs         string `json:"downs"`
	//Filelock      string `json:"filelock"`
	//IsBakdownload int    `json:"is_bakdownload"`
	//Bakdownload   string `json:"bakdownload"`
	//IsDes         int    `json:"is_des"` // 是否存在描述
	//IsIco         int    `json:"is_ico"`

	// 文件夹
	FolID string `json:"fol_id"`
	//Folderlock string `json:"folderlock"`
	//FolderDes  string `json:"folder_des"`
}

func (f *FileOrFolder) isFloder() bool {
	return f.FolID != ""
}
func (f *FileOrFolder) ToObj() model.Obj {
	obj := &model.Object{}
	if f.isFloder() {
		obj.ID = f.FolID
		obj.Name = f.Name
		obj.Modified = time.Now()
		obj.IsFolder = true
	} else {
		obj.ID = f.ID
		obj.Name = f.NameAll
		obj.Modified = MustParseTime(f.Time)
		obj.Size = SizeStrToInt64(f.Size)
	}
	return obj
}

type FileShareResp struct {
	Info FileShare `json:"info"`
}
type FileShare struct {
	Pwd    string `json:"pwd"`
	Onof   string `json:"onof"`
	Taoc   string `json:"taoc"`
	IsNewd string `json:"is_newd"`

	// 文件
	FID string `json:"f_id"`

	// 文件夹
	NewUrl string `json:"new_url"`
	Name   string `json:"name"`
	Des    string `json:"des"`
}

type FileOrFolderByShareUrlResp struct {
	Text []FileOrFolderByShareUrl `json:"text"`
}
type FileOrFolderByShareUrl struct {
	ID      string `json:"id"`
	NameAll string `json:"name_all"`
	Size    string `json:"size"`
	Time    string `json:"time"`
	Duan    string `json:"duan"`
	//Icon          string `json:"icon"`
	//PIco int `json:"p_ico"`
	//T int `json:"t"`
	IsFloder bool
}

func (f *FileOrFolderByShareUrl) ToObj() model.Obj {
	return &model.Object{
		ID:       f.ID,
		Name:     f.NameAll,
		Size:     SizeStrToInt64(f.Size),
		Modified: MustParseTime(f.Time),
		IsFolder: f.IsFloder,
	}
}

type FileShareInfoAndUrlResp[T string | int] struct {
	Dom string `json:"dom"`
	URL string `json:"url"`
	Inf T      `json:"inf"`
}

func (u *FileShareInfoAndUrlResp[T]) GetBaseUrl() string {
	return fmt.Sprint(u.Dom, "/file")
}

func (u *FileShareInfoAndUrlResp[T]) GetDownloadUrl() string {
	return fmt.Sprint(u.GetBaseUrl(), "/", u.URL)
}

// 通过分享链接获取文件信息和下载链接
type FileInfoAndUrlByShareUrl struct {
	ID   string
	Name string
	Size string
	Time string
	Url  string
}

func (f *FileInfoAndUrlByShareUrl) ToObj() model.Obj {
	return &model.Object{
		ID:       f.ID,
		Name:     f.Name,
		Size:     SizeStrToInt64(f.Size),
		Modified: MustParseTime(f.Time),
	}
}
