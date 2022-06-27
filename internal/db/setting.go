package db

import (
	"fmt"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/pkg/errors"
)

func SaveSettings(items []model.SettingItem) error {
	return errors.WithStack(db.Save(items).Error)
}

func SaveSetting(item model.SettingItem) error {
	return errors.WithStack(db.Save(item).Error)
}

func GetSettingsByGroup(group int) ([]model.SettingItem, error) {
	var items []model.SettingItem
	if err := db.Where(fmt.Sprintf("%s = ?", columnName("group")), group).Find(&items).Error; err != nil {
		return nil, errors.WithStack(err)
	}
	return items, nil
}

func DeleteSettingByKey(key string) error {
	setting := model.SettingItem{
		Key: key,
	}
	return errors.WithStack(db.Delete(&setting).Error)
}
