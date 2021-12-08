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

type SettingItem struct {
	Key         string `json:"key" gorm:"primaryKey" binding:"required"`
	Value       string `json:"value"`
	Description string `json:"description"`
	Type        string `json:"type"`
	Group       int    `json:"group"`
	Values      string `json:"values"`
	Version     string `json:"version"`
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
	checkParent, err := GetSettingByKey("check parent folder")
	if err == nil {
		conf.CheckParent = checkParent.Value == "true"
	}
	checkDown, err := GetSettingByKey("check down link")
	if err == nil {
		conf.CheckDown = checkDown.Value == "true"
	}
	favicon, err := GetSettingByKey("favicon")
	if err == nil {
		//conf.Favicon = favicon.Value
		conf.IndexHtml = strings.Replace(conf.RawIndexHtml, "https://store.heytapimage.com/cdo-portal/feedback/202110/30/d43c41c5d257c9bc36366e310374fb19.png", favicon.Value, 1)
	}
	title, err := GetSettingByKey("title")
	if err == nil {
		conf.IndexHtml = strings.Replace(conf.IndexHtml, "Loading...", title.Value, 1)
	}
	customizeHead, err := GetSettingByKey("customize head")
	if err == nil {
		conf.IndexHtml = strings.Replace(conf.IndexHtml, "<!-- customize head -->", customizeHead.Value, 1)
	}
	customizeBody, err := GetSettingByKey("customize body")
	if err == nil {
		conf.IndexHtml = strings.Replace(conf.IndexHtml, "<!-- customize body -->", customizeBody.Value, 1)
	}

	adminPassword, err := GetSettingByKey("password")
	if err == nil {
		conf.Token = utils.GetMD5Encode(fmt.Sprintf("https://github.com/Xhofe/alist-%s", adminPassword.Value))
	}

	davUsername, err := GetSettingByKey("WebDAV username")
	if err == nil {
		conf.DavUsername = davUsername.Value
	}
	davPassword, err := GetSettingByKey("WebDAV password")
	if err == nil {
		conf.DavPassword = davPassword.Value
	}
}
