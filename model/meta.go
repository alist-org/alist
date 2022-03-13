package model

import (
	"github.com/Xhofe/alist/conf"
	log "github.com/sirupsen/logrus"
)

type Meta struct {
	ID        uint   `json:"id" gorm:"primaryKey"`
	Path      string `json:"path" gorm:"unique" binding:"required"`
	Password  string `json:"password"`
	Hide      string `json:"hide"`
	Upload    bool   `json:"upload"`
	OnlyShows string `json:"only_shows"`
}

func GetMetaByPath(path string) (*Meta, error) {
	var meta Meta
	err := conf.DB.Where("path = ?", path).First(&meta).Error
	if err != nil {
		return nil, err
	}
	return &meta, nil
}

func SaveMeta(meta Meta) error {
	return conf.DB.Save(&meta).Error
}

func CreateMeta(meta Meta) error {
	return conf.DB.Create(&meta).Error
}

func DeleteMeta(id uint) error {
	meta := Meta{ID: id}
	log.Debugf("delete meta: %+v", meta)
	return conf.DB.Delete(&meta).Error
}

func GetMetas() (*[]Meta, error) {
	var metas []Meta
	if err := conf.DB.Find(&metas).Error; err != nil {
		return nil, err
	}
	return &metas, nil
}
