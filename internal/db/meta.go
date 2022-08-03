package db

import (
	stdpath "path"
	"time"

	"github.com/Xhofe/go-cache"
	"github.com/alist-org/alist/v3/internal/errs"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/singleflight"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

var metaCache = cache.NewMemCache(cache.WithShards[*model.Meta](2))

// metaG maybe not needed
var metaG singleflight.Group[*model.Meta]

func GetNearestMeta(path string) (*model.Meta, error) {
	path = utils.StandardizePath(path)
	meta, err := GetMetaByPath(path)
	if err == nil {
		return meta, nil
	}
	if errors.Cause(err) != gorm.ErrRecordNotFound {
		return nil, err
	}
	if path == "/" {
		return nil, errors.WithStack(errs.MetaNotFound)
	}
	return GetNearestMeta(stdpath.Dir(path))
}

func GetMetaByPath(path string) (*model.Meta, error) {
	meta, ok := metaCache.Get(path)
	if ok {
		return meta, nil
	}
	meta, err, _ := metaG.Do(path, func() (*model.Meta, error) {
		meta := model.Meta{Path: path}
		if err := db.Where(meta).First(&meta).Error; err != nil {
			return nil, errors.Wrapf(err, "failed select meta")
		}
		metaCache.Set(path, &meta, cache.WithEx[*model.Meta](time.Hour))
		return &meta, nil
	})
	return meta, err
}

func GetMetaById(id uint) (*model.Meta, error) {
	var u model.Meta
	if err := db.First(&u, id).Error; err != nil {
		return nil, errors.Wrapf(err, "failed get old meta")
	}
	return &u, nil
}

func CreateMeta(u *model.Meta) error {
	return errors.WithStack(db.Create(u).Error)
}

func UpdateMeta(u *model.Meta) error {
	old, err := GetMetaById(u.ID)
	if err != nil {
		return err
	}
	metaCache.Del(old.Path)
	return errors.WithStack(db.Save(u).Error)
}

func GetMetas(pageIndex, pageSize int) ([]model.Meta, int64, error) {
	metaDB := db.Model(&model.Meta{})
	var count int64
	if err := metaDB.Count(&count).Error; err != nil {
		return nil, 0, errors.Wrapf(err, "failed get metas count")
	}
	var metas []model.Meta
	if err := metaDB.Offset((pageIndex - 1) * pageSize).Limit(pageSize).Find(&metas).Error; err != nil {
		return nil, 0, errors.Wrapf(err, "failed get find metas")
	}
	return metas, count, nil
}

func DeleteMetaById(id uint) error {
	old, err := GetMetaById(id)
	if err != nil {
		return err
	}
	metaCache.Del(old.Path)
	return errors.WithStack(db.Delete(&model.Meta{}, id).Error)
}
