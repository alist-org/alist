package db

import (
	"fmt"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

var settingsMap map[string]string
var publicSettingsMap map[string]string

func GetPublicSettingsMap() map[string]string {
	if settingsMap == nil {
		publicSettingsMap = make(map[string]string)
		settingItems, err := GetPublicSettingItems()
		if err != nil {
			log.Errorf("failed to get settingItems: %+v", err)
		}
		for _, settingItem := range settingItems {
			publicSettingsMap[settingItem.Key] = settingItem.Value
		}
	}
	return publicSettingsMap
}

func GetSettingsMap() map[string]string {
	if settingsMap == nil {
		settingsMap = make(map[string]string)
		settingItems, err := GetSettingItems()
		if err != nil {
			log.Errorf("failed to get settingItems: %+v", err)
		}
		for _, settingItem := range settingItems {
			settingsMap[settingItem.Key] = settingItem.Value
		}
	}
	return settingsMap
}

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

func GetPublicSettingItems() ([]model.SettingItem, error) {
	var settingItems []model.SettingItem
	if err := db.Where(fmt.Sprintf("%s in ?", columnName("flag")), []int{0, 2}).Find(&settingItems).Error; err != nil {
		return nil, err
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

func SaveSettingItems(items []model.SettingItem) error {
	settingsMap = nil
	return errors.WithStack(db.Save(items).Error)
}

func SaveSettingItem(item model.SettingItem) error {
	settingsMap = nil
	return errors.WithStack(db.Save(item).Error)
}

func DeleteSettingItemByKey(key string) error {
	settingItem := model.SettingItem{
		Key: key,
	}
	old, err := GetSettingItemByKey(key)
	if err != nil {
		return errors.WithMessage(err, "failed to get settingItem")
	}
	if !old.IsDeprecated() {
		return errors.Errorf("setting [%s] is not deprecated", key)
	}
	settingsMap = nil
	return errors.WithStack(db.Delete(&settingItem).Error)
}
