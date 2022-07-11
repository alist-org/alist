package data

import (
	"github.com/alist-org/alist/v3/cmd/args"
	"github.com/alist-org/alist/v3/internal/conf"
	"github.com/alist-org/alist/v3/internal/db"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/utils/random"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var initialSettingItems []model.SettingItem

func initSettings() {
	initialSettings()
	// check deprecated
	settings, err := db.GetSettingItems()
	if err != nil {
		log.Fatalf("failed get settings: %+v", err)
	}
	for i := range settings {
		if !isActive(settings[i].Key) {
			settings[i].Flag = model.DEPRECATED
		}
	}
	if settings != nil && len(settings) > 0 {
		err = db.SaveSettingItems(settings)
		if err != nil {
			log.Fatalf("failed save settings: %+v", err)
		}
	}
	// insert new items
	for i, _ := range initialSettingItems {
		v := initialSettingItems[i]
		_, err := db.GetSettingItemByKey(v.Key)
		if err == nil {
			continue
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = db.SaveSettingItem(v)
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

func initialSettings() {
	var token string
	if args.Dev {
		token = "dev_token"
	} else {
		token = random.Token()
	}
	initialSettingItems = []model.SettingItem{
		// site settings
		{Key: conf.VERSION, Value: conf.Version, Type: conf.TypeString, Group: model.SITE, Flag: model.READONLY},
		{Key: conf.ApiUrl, Value: "", Type: conf.TypeString, Group: model.SITE},
		{Key: conf.BasePath, Value: "", Type: conf.TypeString, Group: model.SITE},
		{Key: conf.SiteTitle, Value: "AList", Type: conf.TypeString, Group: model.SITE},
		{Key: conf.SiteLogo, Value: "https://cdn.jsdelivr.net/gh/alist-org/logo@main/logo.svg", Type: conf.TypeString, Group: model.SITE},
		{Key: conf.Favicon, Value: "https://cdn.jsdelivr.net/gh/alist-org/logo@main/logo.svg", Type: conf.TypeString, Group: model.SITE},
		{Key: conf.Announcement, Value: "https://github.com/alist-org/alist", Type: conf.TypeString, Group: model.SITE},
		// style settings
		{Key: conf.IconColor, Value: "#1890ff", Type: conf.TypeString, Group: model.STYLE},
		// preview settings
		{Key: conf.TextTypes, Value: "txt,htm,html,xml,java,properties,sql,js,md,json,conf,ini,vue,php,py,bat,gitignore,yml,go,sh,c,cpp,h,hpp,tsx,vtt,srt,ass,rs,lrc", Type: conf.TypeText, Group: model.PREVIEW, Flag: model.PRIVATE},
		{Key: conf.AudioTypes, Value: "mp3,flac,ogg,m4a,wav,opus", Type: conf.TypeText, Group: model.PREVIEW, Flag: model.PRIVATE},
		{Key: conf.VideoTypes, Value: "mp4,mkv,avi,mov,rmvb,webm,flv", Type: conf.TypeText, Group: model.PREVIEW, Flag: model.PRIVATE},
		{Key: conf.ProxyTypes, Value: "m3u8", Type: conf.TypeText, Group: model.PREVIEW, Flag: model.PRIVATE},
		{Key: conf.PdfViewerUrl, Value: "https://alist-org.github.io/pdf.js/web/viewer.html?file=$url", Type: conf.TypeString, Group: model.PREVIEW},
		{Key: conf.AudioAutoplay, Value: "true", Type: conf.TypeBool, Group: model.PREVIEW},
		{Key: conf.VideoAutoplay, Value: "true", Type: conf.TypeBool, Group: model.PREVIEW},
		// global settings
		{Key: conf.HideFiles, Value: "/\\/README.md/i", Type: conf.TypeText, Group: model.GLOBAL},
		{Key: conf.GlobalReadme, Value: "This is global readme", Type: conf.TypeText, Group: model.GLOBAL},
		{Key: conf.CustomizeHead, Type: conf.TypeText, Group: model.GLOBAL, Flag: model.PRIVATE},
		{Key: conf.CustomizeBody, Type: conf.TypeText, Group: model.GLOBAL, Flag: model.PRIVATE},
		{Key: conf.LinkExpiration, Value: "0", Type: conf.TypeNumber, Group: model.GLOBAL, Flag: model.PRIVATE},
		// aria2 settings
		{Key: conf.Aria2Uri, Value: "http://localhost:6800/jsonrpc", Type: conf.TypeString, Group: model.ARIA2, Flag: model.PRIVATE},
		{Key: conf.Aria2Secret, Value: "", Type: conf.TypeString, Group: model.ARIA2, Flag: model.PRIVATE},
		// single settings
		{Key: conf.Token, Value: token, Type: conf.TypeString, Group: model.SINGLE, Flag: model.PRIVATE},
	}
	if args.Dev {
		initialSettingItems = append(initialSettingItems, model.SettingItem{Key: "test_deprecated", Value: "test_value", Type: conf.TypeString, Flag: model.DEPRECATED})
	}
}
