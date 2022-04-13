package model

import (
	"fmt"
	"github.com/Xhofe/alist/conf"
)

type ISearchFile interface {
	GetName() string
	GetSize() uint64
	GetType() int
}

type SearchFile struct {
	Path string `json:"path" gorm:"index"`
	Name string `json:"name"`
	Size uint64 `json:"size"`
	Type int    `json:"type"`
}

func CreateSearchFiles(files []SearchFile) error {
	return conf.DB.Create(files).Error
}

func DeleteSearchFilesByPath(path string) error {
	return conf.DB.Where(fmt.Sprintf("%s = ?", columnName("path")), path).Delete(&SearchFile{}).Error
}

func SearchByNameAndPath(path, keyword string) ([]SearchFile, error) {
	var files []SearchFile
	if err := conf.DB.Where(fmt.Sprintf("%s LIKE ? AND %s LIKE ?", columnName("path"), columnName("name")), fmt.Sprintf("%s%%", path), fmt.Sprintf("%%%s%%", keyword)).Find(&files).Error; err != nil {
		return nil, err
	}
	return files, nil
}
