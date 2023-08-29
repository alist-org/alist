package alist_v3

import (
	"time"

	"github.com/alist-org/alist/v3/internal/model"
)

type ListReq struct {
	model.PageReq
	Path     string `json:"path" form:"path"`
	Password string `json:"password" form:"password"`
	Refresh  bool   `json:"refresh"`
}

type ObjResp struct {
	Name     string    `json:"name"`
	Size     int64     `json:"size"`
	IsDir    bool      `json:"is_dir"`
	Modified time.Time `json:"modified"`
	Created  time.Time `json:"created"`
	Sign     string    `json:"sign"`
	Thumb    string    `json:"thumb"`
	Type     int       `json:"type"`
	HashInfo string    `json:"hashinfo"`
}

type FsListResp struct {
	Content  []ObjResp `json:"content"`
	Total    int64     `json:"total"`
	Readme   string    `json:"readme"`
	Write    bool      `json:"write"`
	Provider string    `json:"provider"`
}

type FsGetReq struct {
	Path     string `json:"path" form:"path"`
	Password string `json:"password" form:"password"`
}

type FsGetResp struct {
	ObjResp
	RawURL   string    `json:"raw_url"`
	Readme   string    `json:"readme"`
	Provider string    `json:"provider"`
	Related  []ObjResp `json:"related"`
}

type MkdirOrLinkReq struct {
	Path string `json:"path" form:"path"`
}

type MoveCopyReq struct {
	SrcDir string   `json:"src_dir"`
	DstDir string   `json:"dst_dir"`
	Names  []string `json:"names"`
}

type RenameReq struct {
	Path string `json:"path"`
	Name string `json:"name"`
}

type RemoveReq struct {
	Dir   string   `json:"dir"`
	Names []string `json:"names"`
}

type LoginResp struct {
	Token string `json:"token"`
}

type MeResp struct {
	Id         int    `json:"id"`
	Username   string `json:"username"`
	Password   string `json:"password"`
	BasePath   string `json:"base_path"`
	Role       int    `json:"role"`
	Disabled   bool   `json:"disabled"`
	Permission int    `json:"permission"`
	SsoId      string `json:"sso_id"`
	Otp        bool   `json:"otp"`
}
