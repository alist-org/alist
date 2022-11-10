package alist_v2

import (
	"time"
)

type File struct {
	Id        string     `json:"-"`
	Name      string     `json:"name"`
	Size      int64      `json:"size"`
	Type      int        `json:"type"`
	Driver    string     `json:"driver"`
	UpdatedAt *time.Time `json:"updated_at"`
	Thumbnail string     `json:"thumbnail"`
	Url       string     `json:"url"`
	SizeStr   string     `json:"size_str"`
	TimeStr   string     `json:"time_str"`
}

type PathResp struct {
	Type string `json:"type"`
	//Meta  Meta         `json:"meta"`
	Files []File `json:"files"`
}

type PathReq struct {
	PageNum  int    `json:"page_num"`
	PageSize int    `json:"page_size"`
	Password string `json:"password"`
	Path     string `json:"path"`
}
