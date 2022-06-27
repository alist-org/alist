package db

import (
	"fmt"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/pkg/errors"
)

func GetSettings() ([]model.SettingItem, error) {
	var items []model.SettingItem
	if err := db.Find(&items).Error; err != nil {
		return nil, errors.WithStack(err)
	}
	return items, nil
}

func GetSettingByKey(key string) (*model.SettingItem, error) {
	var item model.SettingItem
	if err := db.Where(fmt.Sprintf("%s = ?", columnName("key")), key).First(&item).Error; err != nil {
		return nil, errors.WithStack(err)
	}
	return &item, nil
}

func GetPublicSettings() ([]model.SettingItem, error) {
	var items []model.SettingItem
	if err := db.Where(fmt.Sprintf("%s in ?", columnName("flag")), []int{0, 2}).Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func GetSettingsByGroup(group int) ([]model.SettingItem, error) {
	var items []model.SettingItem
	if err := db.Where(fmt.Sprintf("%s = ?", columnName("group")), group).Find(&items).Error; err != nil {
		return nil, errors.WithStack(err)
	}
	return items, nil
}

func SaveSettings(items []model.SettingItem) error {
	return errors.WithStack(db.Save(items).Error)
}

func SaveSetting(item model.SettingItem) error {
	return errors.WithStack(db.Save(item).Error)
}

func DeleteSettingByKey(key string) error {
	setting := model.SettingItem{
		Key: key,
	}
	return errors.WithStack(db.Delete(&setting).Error)
}
