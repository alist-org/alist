package model

import (
	"fmt"
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/utils"
	"strings"
)

const (
	PUBLIC = iota
	PRIVATE
	CONST
)

const (
	FRONT = iota
	BACK
	OTHER
)

type SettingItem struct {
	Key         string `json:"key" gorm:"primaryKey" binding:"required"`
	Value       string `json:"value"`
	Description string `json:"description"`
	Type        string `json:"type"`
	Group       int    `json:"group"`
	Access      int    `json:"access"`
	Values      string `json:"values"`
	Version     string `json:"version"`
}

var Version = SettingItem{
	Key:         "version",
	Value:       conf.GitTag,
	Description: "version",
	Type:        "string",
	Access:      CONST,
	Version:     conf.GitTag,
	Group:       OTHER,
}

func SaveSettings(items []SettingItem) error {
	return conf.DB.Save(items).Error
}

func SaveSetting(item SettingItem) error {
	return conf.DB.Save(item).Error
}

func GetSettingsPublic() ([]SettingItem, error) {
	var items []SettingItem
	if err := conf.DB.Where("`access` <> ?", 1).Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func GetSettingsByGroup(group int) ([]SettingItem, error) {
	var items []SettingItem
	if err := conf.DB.Where("`group` = ?", group).Find(&items).Error; err != nil {
		return nil, err
	}
	items = append([]SettingItem{Version}, items...)
	return items, nil
}

func GetSettings() ([]SettingItem, error) {
	var items []SettingItem
	if err := conf.DB.Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func DeleteSetting(key string) error {
	setting := SettingItem{
		Key: key,
	}
	return conf.DB.Delete(&setting).Error
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
	// html
	favicon, err := GetSettingByKey("favicon")
	if err == nil {
		//conf.Favicon = favicon.Value
		conf.ManageHtml = strings.Replace(conf.RawIndexHtml, "https://store.heytapimage.com/cdo-portal/feedback/202110/30/d43c41c5d257c9bc36366e310374fb19.png", favicon.Value, 1)
	}
	title, err := GetSettingByKey("title")
	if err == nil {
		conf.ManageHtml = strings.Replace(conf.ManageHtml, "Loading...", title.Value, 1)
	}
	customizeHead, err := GetSettingByKey("customize head")
	if err == nil {
		conf.IndexHtml = strings.Replace(conf.ManageHtml, "<!-- customize head -->", customizeHead.Value, 1)
	}
	customizeBody, err := GetSettingByKey("customize body")
	if err == nil {
		conf.IndexHtml = strings.Replace(conf.IndexHtml, "<!-- customize body -->", customizeBody.Value, 1)
	}
	// token
	adminPassword, err := GetSettingByKey("password")
	if err == nil {
		conf.Token = utils.GetMD5Encode(fmt.Sprintf("https://github.com/Xhofe/alist-%s", adminPassword.Value))
	}
	// load settings
	for _, key := range conf.LoadSettings {
		vm, err := GetSettingByKey(key)
		if err == nil {
			conf.Set(key, vm.Value)
		}
	}
}
