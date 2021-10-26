package model

import "github.com/Xhofe/alist/conf"

type SettingItem struct {
	Key         string `json:"key" gorm:"primaryKey" validate:"required"`
	Value       string `json:"value"`
	Description string `json:"description"`
	Type        int    `json:"type"`
}

func SaveSettings(items []SettingItem) error {
	//tx := conf.DB.Begin()
	//for _,item := range items{
	//	tx.Save(item)
	//}
	//tx.Commit()
	return conf.DB.Save(items).Error
}

func GetSettingByType(t int) (*[]SettingItem, error) {
	var items []SettingItem
	if err := conf.DB.Where("type = ?", t).Find(&items).Error; err != nil {
		return nil, err
	}
	return &items, nil
}
