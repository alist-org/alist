package baiduphoto

import (
	"fmt"

	"github.com/Xhofe/alist/utils"
)

type TokenErrResp struct {
	ErrorDescription string `json:"error_description"`
	ErrorMsg         string `json:"error"`
}

func (e *TokenErrResp) Error() string {
	return fmt.Sprint(e.ErrorMsg, " : ", e.ErrorDescription)
}

type Erron struct {
	Errno     int `json:"errno"`
	RequestID int `json:"request_id"`
}

type Page struct {
	HasMore int    `json:"has_more"`
	Cursor  string `json:"cursor"`
}

func (p Page) HasNextPage() bool {
	return p.HasMore == 1
}

type (
	FileListResp struct {
		Page
		List []File `json:"list"`
	}

	File struct {
		Fsid     int64    `json:"fsid"` // 文件ID
		Path     string   `json:"path"` // 文件路径
		Size     int64    `json:"size"`
		Ctime    int64    `json:"ctime"` // 创建时间 s
		Mtime    int64    `json:"mtime"` // 修改时间 s
		Thumburl []string `json:"thumburl"`
	}
)

func (f File) Name() string {
	return utils.Base(f.Path)
}

/*相册部分*/
type (
	AlbumListResp struct {
		Page
		List       []Album `json:"list"`
		Reset      int64   `json:"reset"`
		TotalCount int64   `json:"total_count"`
	}

	Album struct {
		AlbumID    string `json:"album_id"`
		Tid        int64  `json:"tid"`
		Title      string `json:"title"`
		JoinTime   int64  `json:"join_time"`
		CreateTime int64  `json:"create_time"`
		Mtime      int64  `json:"mtime"`
	}

	AlbumFileListResp struct {
		Page
		List       []AlbumFile `json:"list"`
		Reset      int64       `json:"reset"`
		TotalCount int64       `json:"total_count"`
	}

	AlbumFile struct {
		File
		Tid int64 `json:"tid"`
		Uk  int64 `json:"uk"`
	}
)

type (
	CopyFileResp struct {
		List []CopyFile `json:"list"`
	}
	CopyFile struct {
		FromFsid  int64  `json:"from_fsid"` // 源ID
		Fsid      int64  `json:"fsid"`      // 目标ID
		Path      string `json:"path"`
		ShootTime int    `json:"shoot_time"`
	}
)

/*上传部分*/
type (
	UploadFile struct {
		FsID           int64  `json:"fs_id"`
		Size           int64  `json:"size"`
		Md5            string `json:"md5"`
		ServerFilename string `json:"server_filename"`
		Path           string `json:"path"`
		Ctime          int    `json:"ctime"`
		Mtime          int    `json:"mtime"`
		Isdir          int    `json:"isdir"`
		Category       int    `json:"category"`
		ServerMd5      string `json:"server_md5"`
		ShootTime      int    `json:"shoot_time"`
	}

	CreateFileResp struct {
		Data UploadFile `json:"data"`
	}

	PrecreateResp struct {
		ReturnType int `json:"return_type"` //存在返回2 不存在返回1 已经保存3
		//存在返回
		CreateFileResp

		//不存在返回
		Path      string  `json:"path"`
		UploadID  string  `json:"uploadid"`
		Blocklist []int64 `json:"block_list"`
	}
)
