package model

import (
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/utils"
	log "github.com/sirupsen/logrus"
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

func initSettings() {
	log.Infof("init settings...")
	version, err := GetSettingByKey("version")
	if err != nil {
		log.Debugf("first run")
		version = &SettingItem{
			Key:         "version",
			Value:       "0.0.0",
			Description: "version",
			Type:        CONST,
		}
	}
	settingsMap := map[string][]SettingItem{
		"2.0.0": {
			{
				Key:         "title",
				Value:       "Alist",
				Description: "title",
				Type:        PUBLIC,
			},
			{
				Key:         "password",
				Value:       "alist",
				Description: "password",
				Type:        PRIVATE,
			},
			{
				Key:         "version",
				Value:       "2.0.0",
				Description: "version",
				Type:        CONST,
			},
		},
	}
	for k, v := range settingsMap {
		if utils.VersionCompare(k, version.Value) > 0 {
			log.Infof("writing [v%s] settings",k)
			err = SaveSettings(v)
			if err != nil {
				log.Fatalf("save settings error")
			}
		}
	}
}
