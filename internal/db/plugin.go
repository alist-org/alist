package db

import (
	"fmt"

	"github.com/alist-org/alist/v3/internal/model"
	"github.com/pkg/errors"
)

func GetPlugins(pageIndex, pageSize int) ([]model.Plugin, int64, error) {
	pluginDB := db.Model(&model.Plugin{})
	var count int64
	if err := pluginDB.Count(&count).Error; err != nil {
		return nil, 0, errors.Wrapf(err, "failed get storages count")
	}
	var plugins []model.Plugin
	if err := pluginDB.Order(columnName("order")).Offset((pageIndex - 1) * pageSize).Limit(pageSize).Find(&plugins).Error; err != nil {
		return nil, 0, errors.WithStack(err)
	}
	return plugins, count, nil
}

func GetAllPlugin() ([]model.Plugin, error) {
	var plugins []model.Plugin
	if err := db.Find(&plugins).Error; err != nil {
		return nil, errors.WithStack(err)
	}
	return plugins, nil
}

func GetEnabledPlugin() ([]model.Plugin, error) {
	var plugins []model.Plugin
	if err := db.Where(fmt.Sprintf("%s = ?", columnName("disabled")), false).Find(&plugins).Error; err != nil {
		return nil, errors.WithStack(err)
	}
	return plugins, nil
}

func GetPluginById(id uint) (*model.Plugin, error) {
	var plugin model.Plugin
	plugin.ID = id
	if err := db.First(&plugin).Error; err != nil {
		return nil, errors.WithStack(err)
	}
	return &plugin, nil
}

func CreatePlugin(plugin *model.Plugin) error {
	return errors.WithStack(db.Create(plugin).Error)
}

func UpdatePlugin(plugin *model.Plugin) error {
	return errors.WithStack(db.Save(plugin).Error)
}

func DeleteStorageByPlugin(id uint) error {
	return errors.WithStack(db.Delete(&model.Plugin{}, id).Error)
}
