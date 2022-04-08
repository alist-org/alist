package lanzou

import (
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/utils"
	"path"
)

type LanZouFile struct {
	Name    string `json:"name"`
	NameAll string `json:"name_all"`
	Id      string `json:"id"`
	FolId   string `json:"fol_id"`
	Size    string `json:"size"`
	Time    string `json:"time"`
	Folder  bool
}

func (f LanZouFile) GetSize() uint64 {
	return 0
}

func (f LanZouFile) GetName() string {
	if f.Folder {
		return f.Name
	}
	return f.NameAll
}

func (f LanZouFile) GetType() int {
	if f.Folder {
		return conf.FOLDER
	}
	return utils.GetFileType(path.Ext(f.NameAll))
}

type DownPageResp struct {
	Zt   int `json:"zt"`
	Info struct {
		Pwd    string `json:"pwd"`
		Onof   string `json:"onof"`
		FId    string `json:"f_id"`
		Taoc   string `json:"taoc"`
		IsNewd string `json:"is_newd"`
	} `json:"info"`
	Text interface{} `json:"text"`
}
