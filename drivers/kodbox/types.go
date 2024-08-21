package kodbox

type CommonResp struct {
	Code    any    `json:"code"`
	TimeUse string `json:"timeUse"`
	TimeNow string `json:"timeNow"`
	Data    any    `json:"data"`
	Info    any    `json:"info"`
}

type ListPathData struct {
	FolderList []FolderOrFile `json:"folderList"`
	FileList   []FolderOrFile `json:"fileList"`
}

type FolderOrFile struct {
	Name       string `json:"name"`
	Path       string `json:"path"`
	Type       string `json:"type"`
	Ext        string `json:"ext,omitempty"` // 文件特有字段
	Size       int64  `json:"size"`
	CreateTime int64  `json:"createTime"`
	ModifyTime int64  `json:"modifyTime"`
}
