package template

import "time"

// write all struct here

// TYPE_CX_FILE type=1 文件
var TYPE_CX_FILE = int64(1)

// TYPE_CX_FOLDER type=2 文件夹
var TYPE_CX_FOLDER = int64(2)

// TYPE_CX_SHARED_ROOT type=4 共享文件的根目录
var TYPE_CX_SHARED_ROOT = int64(4)

type Resp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type File struct {
	Id        string     `json:"id"`
	FileName  string     `json:"file_name"`
	Size      int64      `json:"size"`
	File      bool       `json:"file"`
	UpdatedAt *time.Time `json:"updated_at"`
}
