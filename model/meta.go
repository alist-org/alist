package model

import "github.com/Xhofe/alist/conf"

type Meta struct {
	Path     string `json:"path" gorm:"primaryKey" binding:"required"`
	Password string `json:"password"`
	Hide     string `json:"hide"`
}

func GetMetaByPath(path string) (*Meta, error) {
	var meta Meta
	meta.Path = path
	err := conf.DB.First(&meta).Error
	if err != nil {
		return nil, err
	}
	return &meta, nil
}

func SaveMeta(meta Meta) error {
	return conf.DB.Save(meta).Error
}

func DeleteMeta(path string) error {
	meta := Meta{Path: path}
	return conf.DB.Delete(&meta).Error
}

func GetMetas() (*[]Meta, error) {
	var metas []Meta
	if err := conf.DB.Find(&metas).Error; err != nil {
		return nil, err
	}
	return &metas, nil
}
