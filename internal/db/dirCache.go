package db

import (
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/pkg/errors"
)

func GetDirCachesByPath(startPath string) ([]model.DirCache, error) {
	var dirCaches []model.DirCache
	if err := db.Where("path LIKE ?", startPath+"%").Find(&dirCaches).Error; err != nil {
		return nil, errors.Wrapf(err, "failed select dirCache like "+startPath+"%")
	}
	return dirCaches, nil
}

func CreateDirCache(u *model.DirCache) error {
	return errors.WithStack(db.Create(u).Error)
}

func UpdateDirCache(u *model.DirCache) error {
	return errors.WithStack(db.Where("path=?", u.Path).Save(u).Error)
}
