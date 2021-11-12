package model

import (
	"github.com/Xhofe/alist/conf"
	"strings"
)

const (
	PUBLIC = iota
	PRIVATE
	CONST
)

type SettingItem struct {
	Key         string `json:"key" gorm:"primaryKey" validate:"required"`
	Value       string `json:"value"`
	Description string `json:"description"`
	Type        string `json:"type"`
	Group       int    `json:"group"`
	Values      string `json:"values"`
}

func SaveSettings(items []SettingItem) error {
	return conf.DB.Save(items).Error
}

func SaveSetting(item SettingItem) error {
	return conf.DB.Save(item).Error
}

func GetSettingsPublic() (*[]SettingItem, error) {
	var items []SettingItem
	if err := conf.DB.Where("`group` <> ?", 1).Find(&items).Error; err != nil {
		return nil, err
	}
	return &items, nil
}

func GetSettings() (*[]SettingItem, error) {
	var items []SettingItem
	if err := conf.DB.Find(&items).Error; err != nil {
		return nil, err
	}
	return &items, nil
}

func GetSettingByKey(key string) (*SettingItem, error) {
	var items SettingItem
	if err := conf.DB.Where("`key` = ?", key).First(&items).Error; err != nil {
		return nil, err
	}
	return &items, nil
}

func LoadSettings() {
	textTypes, err := GetSettingByKey("text types")
	if err == nil {
		conf.TextTypes = strings.Split(textTypes.Value, ",")
	}
	checkParent, err := GetSettingByKey("check parent folder")
	if err == nil {
		conf.CheckParent = checkParent.Value == "true"
	}
}