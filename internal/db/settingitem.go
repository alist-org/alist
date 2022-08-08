package db

import (
	"fmt"

	"github.com/alist-org/alist/v3/internal/model"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

var settingsMap map[string]string
var publicSettingsMap map[string]string

func settingsUpdate() {
	settingsMap = nil
	publicSettingsMap = nil
}

func GetPublicSettingsMap() map[string]string {
	if publicSettingsMap == nil {
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

func GetSettingItemInKeys(keys []string) ([]model.SettingItem, error) {
	var settingItem []model.SettingItem
	if err := db.Where(fmt.Sprintf("%s in ?", columnName("key")), keys).Find(&settingItem).Error; err != nil {
		return nil, errors.WithStack(err)
	}
	return settingItem, nil
}

func GetPublicSettingItems() ([]model.SettingItem, error) {
	var settingItems []model.SettingItem
	if err := db.Where(fmt.Sprintf("%s in ?", columnName("flag")), []int{model.PUBLIC, model.READONLY}).Find(&settingItems).Error; err != nil {
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

func GetSettingItemsInGroups(groups []int) ([]model.SettingItem, error) {
	var settingItems []model.SettingItem
	if err := db.Where(fmt.Sprintf("%s in ?", columnName("group")), groups).Find(&settingItems).Error; err != nil {
		return nil, errors.WithStack(err)
	}
	return settingItems, nil
}

func SaveSettingItems(items []model.SettingItem) error {
	others := make([]model.SettingItem, 0)
	for i := range items {
		if ok, err := HandleSettingItem(&items[i]); ok {
			if err != nil {
				return err
			} else {
				err = db.Save(items[i]).Error
				if err != nil {
					return errors.WithStack(err)
				}
			}
		} else {
			others = append(others, items[i])
		}
	}
	err := db.Save(others).Error
	if err == nil {
		settingsUpdate()
	}
	return err
}

func SaveSettingItem(item model.SettingItem) error {
	_, err := HandleSettingItem(&item)
	if err != nil {
		return err
	}
	err = db.Save(item).Error
	if err == nil {
		settingsUpdate()
	}
	return errors.WithStack(err)
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
	settingsUpdate()
	return errors.WithStack(db.Delete(&settingItem).Error)
}
