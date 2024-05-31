package db

import (
	"fmt"

	"github.com/alist-org/alist/v3/internal/model"
	"github.com/pkg/errors"
)

func GetSettingItems() ([]model.SettingItem, error) {
	var settingItems []model.SettingItem
	if err := db.Find(&settingItems).Error; err != nil {
		return nil, errors.WithStack(err)
	}
	return settingItems, nil
}

func GetSettingItemByKey(key string) (*model.SettingItem, error) {
	var settingItem model.SettingItem
	if err := db.Where(fmt.Sprintf("%s = ?", columnName("key")), key).First(&settingItem).Error; err != nil {
		return nil, errors.WithStack(err)
	}
	return &settingItem, nil
}

// func GetSettingItemInKeys(keys []string) ([]model.SettingItem, error) {
// 	var settingItem []model.SettingItem
// 	if err := db.Where(fmt.Sprintf("%s in ?", columnName("key")), keys).Find(&settingItem).Error; err != nil {
// 		return nil, errors.WithStack(err)
// 	}
// 	return settingItem, nil
// }

func GetPublicSettingItems() ([]model.SettingItem, error) {
	var settingItems []model.SettingItem
	if err := db.Where(fmt.Sprintf("%s in ?", columnName("flag")), []int{model.PUBLIC, model.READONLY}).Find(&settingItems).Error; err != nil {
		return nil, errors.WithStack(err)
	}
	return settingItems, nil
}

func GetSettingItemsByGroup(group int) ([]model.SettingItem, error) {
	var settingItems []model.SettingItem
	if err := db.Where(fmt.Sprintf("%s = ?", columnName("group")), group).Find(&settingItems).Error; err != nil {
		return nil, errors.WithStack(err)
	}
	return settingItems, nil
}

func GetSettingItemsInGroups(groups []int) ([]model.SettingItem, error) {
	var settingItems []model.SettingItem
	err := db.Order(columnName("index")).Where(fmt.Sprintf("%s in ?", columnName("group")), groups).Find(&settingItems).Error
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return settingItems, nil
}

func SaveSettingItems(items []model.SettingItem) (err error) {
	return errors.WithStack(db.Save(items).Error)
}

func SaveSettingItem(item *model.SettingItem) error {
	return errors.WithStack(db.Save(item).Error)
}

func DeleteSettingItemByKey(key string) error {
	return errors.WithStack(db.Delete(&model.SettingItem{Key: key}).Error)
}
