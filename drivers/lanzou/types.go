package lanzou

import (
	"errors"
	"fmt"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/utils"
	"time"
)

var ErrFileShareCancel = errors.New("file sharing cancellation")
var ErrFileNotExist = errors.New("file does not exist")
var ErrCookieExpiration = errors.New("cookie expiration")

type RespText[T any] struct {
	Text T `json:"text"`
}

type RespInfo[T any] struct {
	Info T `json:"info"`
}

var _ model.Obj = (*FileOrFolder)(nil)
var _ model.Obj = (*FileOrFolderByShareUrl)(nil)

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

	// 缓存字段
	size       *int64     `json:"-"`
	time       *time.Time `json:"-"`
	repairFlag bool       `json:"-"`
	shareInfo  *FileShare `json:"-"`
}

func (f *FileOrFolder) CreateTime() time.Time {
	return f.ModTime()
}

func (f *FileOrFolder) GetHash() utils.HashInfo {
	return utils.HashInfo{}
}

func (f *FileOrFolder) GetID() string {
	if f.IsDir() {
		return f.FolID
	}
	return f.ID
}
func (f *FileOrFolder) GetName() string {
	if f.IsDir() {
		return f.Name
	}
	return f.NameAll
}
func (f *FileOrFolder) GetPath() string { return "" }
func (f *FileOrFolder) GetSize() int64 {
	if f.size == nil {
		size := SizeStrToInt64(f.Size)
		f.size = &size
	}
	return *f.size
}
func (f *FileOrFolder) IsDir() bool { return f.FolID != "" }
func (f *FileOrFolder) ModTime() time.Time {
	if f.time == nil {
		time := MustParseTime(f.Time)
		f.time = &time
	}
	return *f.time
}

func (f *FileOrFolder) SetShareInfo(fs *FileShare) {
	f.shareInfo = fs
}
func (f *FileOrFolder) GetShareInfo() *FileShare {
	return f.shareInfo
}

/* 通过ID获取文件/文件夹分享信息 */
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

/* 分享类型为文件夹 */
type FileOrFolderByShareUrlResp struct {
	Text []FileOrFolderByShareUrl `json:"text"`
}
type FileOrFolderByShareUrl struct {
	ID      string `json:"id"`
	NameAll string `json:"name_all"`

	// 文件特有
	Duan string `json:"duan"`
	Size string `json:"size"`
	Time string `json:"time"`
	//Icon          string `json:"icon"`
	//PIco int `json:"p_ico"`
	//T int `json:"t"`

	// 文件夹特有
	IsFloder bool `json:"-"`

	//
	Url string `json:"-"`
	Pwd string `json:"-"`

	// 缓存字段
	size       *int64     `json:"-"`
	time       *time.Time `json:"-"`
	repairFlag bool       `json:"-"`
}

func (f *FileOrFolderByShareUrl) CreateTime() time.Time {
	return f.ModTime()
}

func (f *FileOrFolderByShareUrl) GetHash() utils.HashInfo {
	return utils.HashInfo{}
}

func (f *FileOrFolderByShareUrl) GetID() string   { return f.ID }
func (f *FileOrFolderByShareUrl) GetName() string { return f.NameAll }
func (f *FileOrFolderByShareUrl) GetPath() string { return "" }
func (f *FileOrFolderByShareUrl) GetSize() int64 {
	if f.size == nil {
		size := SizeStrToInt64(f.Size)
		f.size = &size
	}
	return *f.size
}
func (f *FileOrFolderByShareUrl) IsDir() bool { return f.IsFloder }
func (f *FileOrFolderByShareUrl) ModTime() time.Time {
	if f.time == nil {
		time := MustParseTime(f.Time)
		f.time = &time
	}
	return *f.time
}

// 获取下载链接的响应
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
