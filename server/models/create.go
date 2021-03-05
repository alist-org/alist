package models

import (
	"fmt"
	"github.com/Xhofe/alist/alidrive"
	"github.com/Xhofe/alist/conf"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// build tree
func BuildTree() error {
	log.Infof("开始构建目录树...")
	tx := conf.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	if err := tx.Error; err != nil {
		return err
	}
	if err := BuildOne(conf.Conf.AliDrive.RootFolder, "/root/", tx); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

func BuildOne(parent string, path string, tx *gorm.DB) error {
	files, err := alidrive.GetList(parent, conf.Conf.AliDrive.MaxFilesCount, "", "", "")
	if err != nil {
		return err
	}
	for _, file := range files.Items {
		newFile := File{
			ParentPath:    path,
			FileExtension: file.FileExtension,
			FileId:        file.FileId,
			Name:          file.Name,
			Type:          file.Type,
			UpdatedAt:     file.UpdatedAt,
			Category:      file.Category,
			ContentType:   file.ContentType,
			Size:          file.Size,
		}
		log.Debugf("插入file:%+v", newFile)
		if err := tx.Create(&newFile).Error; err != nil {
			return err
		}
		if file.Type == "folder" {
			if err := BuildOne(file.FileId, fmt.Sprintf("%s%s/", path, file.Name), tx); err != nil {
				return err
			}
		}
	}
	return nil
}
