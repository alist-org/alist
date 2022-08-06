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
	if err := conf.DB.Where(fmt.Sprintf("%s <> ?", columnName("access")), 1).Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func GetSettingsByGroup(group int) ([]SettingItem, error) {
	var items []SettingItem
	if err := conf.DB.Where(fmt.Sprintf("%s = ?", columnName("group")), group).Find(&items).Error; err != nil {
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
	if err := conf.DB.Where(fmt.Sprintf("%s = ?", columnName("key")), key).First(&items).Error; err != nil {
		return nil, err
	}
	return &items, nil
}

func LoadSettings() {
	textTypes, err := GetSettingByKey("text types")
	if err == nil {
		conf.TextTypes = strings.Split(textTypes.Value, ",")
	}
	audioTypes, err := GetSettingByKey("audio types")
	if err == nil {
		conf.AudioTypes = strings.Split(audioTypes.Value, ",")
	}
	videoTypes, err := GetSettingByKey("video types")
	if err == nil {
		conf.VideoTypes = strings.Split(videoTypes.Value, ",")
	}
	dProxyTypes, err := GetSettingByKey("d_proxy types")
	if err == nil {
		conf.DProxyTypes = strings.Split(dProxyTypes.Value, ",")
	}
	// html
	favicon, err := GetSettingByKey("favicon")
	if err == nil {
		//conf.Favicon = favicon.Value
		conf.ManageHtml = strings.Replace(conf.RawIndexHtml, "https://cdn.jsdelivr.net/gh/alist-org/logo@main/logo.svg", favicon.Value, 1)
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
		if adminPassword.Value != "" {
			conf.Token = utils.GetMD5Encode(fmt.Sprintf("https://github.com/Xhofe/alist-%s", adminPassword.Value))
		} else {
			conf.Token = ""
		}
	}
	// load settings
	for _, key := range conf.LoadSettings {
		vm, err := GetSettingByKey(key)
		if err == nil {
			conf.Set(key, vm.Value)
		}
	}
}
