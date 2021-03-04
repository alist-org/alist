package models

import (
	"fmt"
	"github.com/Xhofe/alist/conf"
	"time"
)

type File struct {
	ParentPath    string     `json:"parent_path" gorm:"index"`
	FileExtension string     `json:"file_extension"`
	FileId        string     `json:"file_id"`
	Name          string     `json:"name" gorm:"index"`
	Type          string     `json:"type"`
	UpdatedAt     *time.Time `json:"updated_at"`
	Category      string     `json:"category"`
	ContentType   string     `json:"content_type"`
	Size          int64      `json:"size"`
}

func (file *File) Create() error {
	return conf.DB.Create(file).Error
}

func GetFilesByParentPath(parentPath string) (*[]File,error) {
	var files []File
	if err := conf.DB.Where("parent_path = ?", parentPath).Find(&files).Error; err != nil {
		return nil,err
	}
	return &files, nil
}

func SearchByNameGlobal(keyword string) (*[]File, error) {
	var files []File
	if err := conf.DB.Where("name LIKE ?", fmt.Sprintf("%%%s%%", keyword)).Find(&files).Error; err != nil {
		return nil,err
	}
	return &files, nil
}

func SearchByNameInPath(keyword string, parentPath string) (*[]File, error) {
	var files []File
	if err := conf.DB.Where("parent_path LIKE ? AND name LIKE ?",fmt.Sprintf("%s%%", parentPath), fmt.Sprintf("%%%s%%", keyword)).Find(&files).Error; err != nil {
		return nil,err
	}
	return &files, nil
}