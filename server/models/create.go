package models

import (
	"fmt"
	"github.com/Xhofe/alist/alidrive"
	"github.com/Xhofe/alist/conf"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"strings"
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
	rootFile := File{
		Dir:    "",
		FileId: conf.Conf.AliDrive.RootFolder,
		Name:   "root",
		Type:   "folder",
	}
	if err := tx.Create(&rootFile).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := BuildOne(conf.Conf.AliDrive.RootFolder, "root/", tx, ""); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

func BuildOne(parent string, path string, tx *gorm.DB, parentPassword string) error {
	files, err := alidrive.GetList(parent, conf.Conf.AliDrive.MaxFilesCount, "", "", "")
	if err != nil {
		return err
	}
	for _, file := range files.Items {
		name := file.Name
		if strings.HasSuffix(name, ".hide") {
			continue
		}
		password := parentPassword
		if strings.Contains(name, ".password-") {
			index := strings.Index(name, ".password-")
			name = file.Name[:index]
			password = file.Name[index+10:]
		}
		newFile := File{
			Dir:           path,
			FileExtension: file.FileExtension,
			FileId:        file.FileId,
			Name:          name,
			Type:          file.Type,
			UpdatedAt:     file.UpdatedAt,
			Category:      file.Category,
			ContentType:   file.ContentType,
			Size:          file.Size,
			Password:      password,
		}
		log.Debugf("插入file:%+v", newFile)
		if err := tx.Create(&newFile).Error; err != nil {
			return err
		}
		if file.Type == "folder" {
			if err := BuildOne(file.FileId, fmt.Sprintf("%s%s/", path, file.Name), tx, password); err != nil {
				return err
			}
		}
	}
	return nil
}
