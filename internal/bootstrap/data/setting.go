package data

import (
	"github.com/alist-org/alist/v3/internal/conf"
	"github.com/alist-org/alist/v3/internal/db"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var initialSettingItems = []model.SettingItem{
	// site settings
	{Key: "version", Value: conf.Version, Type: conf.TypeString, Group: model.SITE, Flag: model.READONLY},
	{Key: "site_title", Value: "AList", Type: conf.TypeString, Group: model.SITE},
	{Key: "site_logo", Value: "https://cdn.jsdelivr.net/gh/alist-org/logo@main/logo.svg", Type: conf.TypeString, Group: model.SITE},
	{Key: "favicon", Value: "https://cdn.jsdelivr.net/gh/alist-org/logo@main/logo.svg", Type: conf.TypeString, Group: model.SITE},
	{Key: "announcement", Value: "https://github.com/alist-org/alist", Type: conf.TypeString, Group: model.SITE},
	// style settings
	{Key: "icon_color", Value: "#1890ff", Type: conf.TypeString, Group: model.STYLE},
	// preview settings
	{Key: "text_types", Value: "txt,htm,html,xml,java,properties,sql,js,md,json,conf,ini,vue,php,py,bat,gitignore,yml,go,sh,c,cpp,h,hpp,tsx,vtt,srt,ass,rs,lrc", Type: conf.TypeText, Group: model.PREVIEW, Flag: model.PRIVATE},
	{Key: "audio_types", Value: "mp3,flac,ogg,m4a,wav,opus", Type: conf.TypeText, Group: model.PREVIEW, Flag: model.PRIVATE},
	{Key: "video_types", Value: "mp4,mkv,avi,mov,rmvb,webm,flv", Type: conf.TypeText, Group: model.PREVIEW, Flag: model.PRIVATE},
	{Key: "proxy_types", Value: "m3u8", Type: conf.TypeText, Group: model.PREVIEW, Flag: model.PRIVATE},
	{Key: "pdf_viewer_url", Value: "https://alist-org.github.io/pdf.js/web/viewer.html?file=$url", Type: conf.TypeString, Group: model.PREVIEW},
	{Key: "audio_autoplay", Value: "true", Type: conf.TypeBool, Group: model.PREVIEW},
	{Key: "video_autoplay", Value: "true", Type: conf.TypeBool, Group: model.PREVIEW},
	// global settings
	{Key: "hide_files", Value: "/\\/README.md/i", Type: conf.TypeText, Group: model.GLOBAL},
	{Key: "global_readme", Value: "This is global readme", Type: conf.TypeText, Group: model.GLOBAL},
	{Key: "customize_head", Type: conf.TypeText, Group: model.GLOBAL, Flag: model.PRIVATE},
	{Key: "customize_body", Type: conf.TypeText, Group: model.GLOBAL, Flag: model.PRIVATE},
}

func initSettings() {
	// check deprecated
	settings, err := db.GetSettings()
	if err != nil {
		log.Fatalf("failed get settings: %+v", err)
	}
	for i := range settings {
		if !isActive(settings[i].Key) {
			settings[i].Flag = model.DEPRECATED
		}
	}
	if settings != nil && len(settings) > 0 {
		err = db.SaveSettings(settings)
		if err != nil {
			log.Fatalf("failed save settings: %+v", err)
		}
	}
	// insert new items
	for i, _ := range initialSettingItems {
		v := initialSettingItems[i]
		_, err := db.GetSettingByKey(v.Key)
		if err == nil {
			continue
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = db.SaveSetting(v)
			if err != nil {
				log.Fatalf("failed create setting: %+v", err)
			}
		} else {
			log.Fatalf("failed get setting: %+v", err)
		}
	}
}

func isActive(key string) bool {
	for _, item := range initialSettingItems {
		if item.Key == key {
			return true
		}
	}
	return false
}
