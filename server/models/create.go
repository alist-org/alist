package models

import (
	"fmt"
	"github.com/Xhofe/alist/alidrive"
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/utils"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"path/filepath"
	"strings"
)

func BuildTreeAll(depth int) {
	for i, _ := range conf.Conf.AliDrive.Drives {
		if err := BuildTree(&conf.Conf.AliDrive.Drives[i], depth); err != nil {
			log.Errorf("盘[%s]构建目录树失败:%s", conf.Conf.AliDrive.Drives[i].Name, err.Error())
		} else {
			log.Infof("盘[%s]构建目录树成功", conf.Conf.AliDrive.Drives[i].Name)
		}
	}
}

// build tree
func BuildTree(drive *conf.Drive, depth int) error {
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
		Dir:      "",
		FileId:   drive.RootFolder,
		Name:     drive.Name,
		Type:     "folder",
		Password: drive.Password,
	}
	if err := tx.Create(&rootFile).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := BuildOne(drive.RootFolder, drive.Name+"/", tx, drive.Password, drive, depth); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

/*
递归构建目录树,插入指定目录下的所有文件
parent 父目录的file_id
path 指定的目录
parentPassword	父目录所携带的密码
drive 要构建的盘
*/
func BuildOne(parent string, path string, tx *gorm.DB, parentPassword string, drive *conf.Drive, depth int) error {
	if depth == 0 {
		return nil
	}
	marker := "first"
	for marker != "" {
		if marker == "first" {
			marker = ""
		}
		files, err := alidrive.GetList(parent, conf.Conf.AliDrive.MaxFilesCount, marker, "", "", drive)
		if err != nil {
			return err
		}
		marker = files.NextMarker
		for _, file := range files.Items {
			name := file.Name
			password := parentPassword
			if strings.HasSuffix(name, ".hide") {
				continue
			}
			if strings.Contains(name, ".ln-") {
				index := strings.Index(name, ".ln-")
				name = file.Name[:index]
				fileId := file.Name[index+4:]
				newFile := File{
					Dir:           path,
					FileExtension: "",
					FileId:        fileId,
					Name:          name,
					Type:          "folder",
					UpdatedAt:     file.UpdatedAt,
					Category:      "",
					ContentType:   "",
					Size:          0,
					Password:      password,
				}
				log.Debugf("插入file:%+v", newFile)
				if err = tx.Create(&newFile).Error; err != nil {
					return err
				}
				if err = BuildOne(fileId, fmt.Sprintf("%s%s/", path, name), tx, password, drive, depth-1); err != nil {
					return err
				}
				continue
			}
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
				ContentHash:   file.ContentHash,
			}
			log.Debugf("插入file:%+v", newFile)
			if err := tx.Create(&newFile).Error; err != nil {
				return err
			}
			if file.Type == "folder" {
				if err := BuildOne(file.FileId, fmt.Sprintf("%s%s/", path, name), tx, password, drive, depth-1); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

//重建指定路径与深度的目录树: 先删除该目录与该目录下所有文件的model，再重新插入
func BuildTreeWithPath(path string, depth int) error {
	dir, name := filepath.Split(path)
	driveName := strings.Split(path, "/")[0]
	drive := utils.GetDriveByName(driveName)
	if drive == nil {
		return fmt.Errorf("找不到drive[%s]", driveName)
	}
	file := &File{
		Dir:      "",
		FileId:   drive.RootFolder,
		Name:     drive.Name,
		Type:     "folder",
		Password: drive.Password,
	}
	var err error
	if dir != "" {
		file, err = GetFileByDirAndName(dir, name)
		if err != nil {
			if file == nil {
				return fmt.Errorf("path not found")
			}
			return err
		}
	}
	tx := conf.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	if err = tx.Error; err != nil {
		tx.Rollback()
		return err
	}
	if err = tx.Where("dir = ? AND name = ?", file.Dir, file.Name).Delete(file).Error; err != nil{
		tx.Rollback()
		return err
	}
	if err = tx.Where("dir like ?", fmt.Sprintf("%s%%", path)).Delete(&File{}).Error; err != nil{
		tx.Rollback()
		return err
	}
	//if dir != "" {
	//	aliFile, err := alidrive.GetFile(file.FileId, drive)
	//	if err != nil {
	//		tx.Rollback()
	//		return err
	//	}
	//	aliName := aliFile.Name
	//	if strings.HasSuffix(aliName, ".hide") {
	//		return nil
	//	}
	//	if strings.Contains(aliName, ".password-") {
	//		index := strings.Index(name, ".password-")
	//		file.Name = aliName[:index]
	//		file.Password = aliName[index+10:]
	//	}
	//}
	if err = tx.Create(&file).Error; err != nil {
		return err
	}
	if err = BuildOne(file.FileId, path+"/", tx, file.Password, drive, depth); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}
