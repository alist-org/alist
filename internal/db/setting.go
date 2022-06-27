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
		settings, err := GetPublicSettings()
		if err != nil {
			log.Errorf("failed to get settings: %+v", err)
		}
		for _, setting := range settings {
			publicSettingsMap[setting.Key] = setting.Value
		}
	}
	return publicSettingsMap
}

func GetSettingsMap() map[string]string {
	if settingsMap == nil {
		settingsMap = make(map[string]string)
		settings, err := GetSettings()
		if err != nil {
			log.Errorf("failed to get settings: %+v", err)
		}
		for _, setting := range settings {
			settingsMap[setting.Key] = setting.Value
		}
	}
	return settingsMap
}

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
	settingsMap = nil
	return errors.WithStack(db.Save(items).Error)
}

func SaveSetting(item model.SettingItem) error {
	settingsMap = nil
	return errors.WithStack(db.Save(item).Error)
}

func DeleteSettingByKey(key string) error {
	setting := model.SettingItem{
		Key: key,
	}
	old, err := GetSettingByKey(key)
	if err != nil {
		return errors.WithMessage(err, "failed to get setting")
	}
	if !old.IsDeprecated() {
		return errors.Errorf("setting [%s] is not deprecated", key)
	}
	settingsMap = nil
	return errors.WithStack(db.Delete(&setting).Error)
}
