package model

import "github.com/Xhofe/alist/conf"

type Meta struct {
	Path     string `json:"path" gorm:"primaryKey"`
	Password string `json:"password"`
	Hide     bool   `json:"hide"`
	Ignore   bool   `json:"ignore"`
}

func GetMetaByPath(path string) (*Meta,error) {
	var meta Meta
	meta.Path = path
	err := conf.DB.First(&meta).Error
	if err != nil {
		return nil, err
	}
	return &meta, nil
}