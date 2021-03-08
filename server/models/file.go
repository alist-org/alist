package models

import (
	"fmt"
	"github.com/Xhofe/alist/conf"
	"time"
)

type File struct {
	Dir           string     `json:"dir" gorm:"index"`
	FileExtension string     `json:"file_extension"`
	FileId        string     `json:"file_id"`
	Name          string     `json:"name" gorm:"index"`
	Type          string     `json:"type"`
	UpdatedAt     *time.Time `json:"updated_at"`
	Category      string     `json:"category"`
	ContentType   string     `json:"content_type"`
	Size          int64      `json:"size"`
	Password      string     `json:"password"`
}

func (file *File) Create() error {
	return conf.DB.Create(file).Error
}

func Clear() error {
	return conf.DB.Where("1 = 1").Delete(&File{}).Error
}

func GetFileByDirAndName(dir, name string) (*File, error) {
	var file File
	if err := conf.DB.Where("dir = ? AND name = ?", dir, name).First(&file).Error; err != nil {
		return nil, err
	}
	return &file, nil
}

func GetFilesByDir(dir string) (*[]File, error) {
	var files []File
	if err := conf.DB.Where("dir = ?", dir).Find(&files).Error; err != nil {
		return nil, err
	}
	return &files, nil
}

func SearchByNameGlobal(keyword string) (*[]File, error) {
	var files []File
	if err := conf.DB.Where("name LIKE ? AND password = ''", fmt.Sprintf("%%%s%%", keyword)).Find(&files).Error; err != nil {
		return nil, err
	}
	return &files, nil
}

func SearchByNameInDir(keyword string, dir string) (*[]File, error) {
	var files []File
	if err := conf.DB.Where("dir LIKE ? AND name LIKE ? AND password = ''", fmt.Sprintf("%s%%", dir), fmt.Sprintf("%%%s%%", keyword)).Find(&files).Error; err != nil {
		return nil, err
	}
	return &files, nil
}
