package model

import (
	"github.com/Xhofe/alist/conf"
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
	Type        int    `json:"type"`
}

func SaveSettings(items []SettingItem) error {
	return conf.DB.Save(items).Error
}

func GetSettingByType(t int) (*[]SettingItem, error) {
	var items []SettingItem
	if err := conf.DB.Where("type = ?", t).Find(&items).Error; err != nil {
		return nil, err
	}
	return &items, nil
}

func GetSettingByKey(key string) (*SettingItem, error) {
	var items SettingItem
	if err := conf.DB.Where("key = ?", key).First(&items).Error; err != nil {
		return nil, err
	}
	return &items, nil
}

