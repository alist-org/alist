package data

import (
	"github.com/alist-org/alist/v3/cmd/flags"
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
	for i := range initialSettingItems {
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
	if flags.Dev {
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
		{Key: conf.Announcement, Value: "https://github.com/alist-org/alist", Type: conf.TypeString, Group: model.SITE},
		// style settings
		{Key: conf.Logo, Value: "https://cdn.jsdelivr.net/gh/alist-org/logo@main/logo.svg", Type: conf.TypeString, Group: model.STYLE},
		{Key: conf.Favicon, Value: "https://cdn.jsdelivr.net/gh/alist-org/logo@main/logo.svg", Type: conf.TypeString, Group: model.STYLE},
		{Key: conf.IconColor, Value: "#1890ff", Type: conf.TypeString, Group: model.STYLE},
		{Key: "home_icon", Value: "🏠", Type: conf.TypeString, Group: model.STYLE},
		// preview settings
		{Key: conf.TextTypes, Value: "txt,htm,html,xml,java,properties,sql,js,md,json,conf,ini,vue,php,py,bat,gitignore,yml,go,sh,c,cpp,h,hpp,tsx,vtt,srt,ass,rs,lrc", Type: conf.TypeText, Group: model.PREVIEW, Flag: model.PRIVATE},
		{Key: conf.AudioTypes, Value: "mp3,flac,ogg,m4a,wav,opus", Type: conf.TypeText, Group: model.PREVIEW, Flag: model.PRIVATE},
		{Key: conf.VideoTypes, Value: "mp4,mkv,avi,mov,rmvb,webm,flv", Type: conf.TypeText, Group: model.PREVIEW, Flag: model.PRIVATE},
		{Key: conf.ImageTypes, Value: "jpg,tiff,jpeg,png,gif,bmp,svg,ico,swf,webp", Type: conf.TypeText, Group: model.PREVIEW, Flag: model.PRIVATE},
		{Key: conf.OfficeTypes, Value: "doc,docx,xls,xlsx,ppt,pptx,pdf", Type: conf.TypeText, Group: model.PREVIEW, Flag: model.PRIVATE},
		{Key: conf.ProxyTypes, Value: "m3u8", Type: conf.TypeText, Group: model.PREVIEW, Flag: model.PRIVATE},
		{Key: conf.OfficeViewers, Value: `{
	"Microsoft":"https://view.officeapps.live.com/op/view.aspx?src=$url",
	"Google":"https://docs.google.com/gview?url=$url&embedded=true",
}`, Type: conf.TypeText, Group: model.PREVIEW},
		{Key: conf.PdfViewers, Value: `{
	"pdf.js":"https://alist-org.github.io/pdf.js/web/viewer.html?file=$url"
}`, Type: conf.TypeText, Group: model.PREVIEW},
		{Key: conf.AudioAutoplay, Value: "true", Type: conf.TypeBool, Group: model.PREVIEW},
		{Key: conf.VideoAutoplay, Value: "true", Type: conf.TypeBool, Group: model.PREVIEW},
		// global settings
		{Key: conf.HideFiles, Value: "/\\/README.md/i", Type: conf.TypeText, Group: model.GLOBAL},
		{Key: conf.GlobalReadme, Value: "This is global readme", Type: conf.TypeText, Group: model.GLOBAL},
		{Key: conf.CustomizeHead, Type: conf.TypeText, Group: model.GLOBAL, Flag: model.PRIVATE},
		{Key: conf.CustomizeBody, Type: conf.TypeText, Group: model.GLOBAL, Flag: model.PRIVATE},
		{Key: conf.LinkExpiration, Value: "0", Type: conf.TypeNumber, Group: model.GLOBAL, Flag: model.PRIVATE},
		{Key: conf.PrivacyRegs, Value: `(?:(?:\d|[1-9]\d|1\d\d|2[0-4]\d|25[0-5])\.){3}(?:\d|[1-9]\d|1\d\d|2[0-4]\d|25[0-5])
([[:xdigit:]]{1,4}(?::[[:xdigit:]]{1,4}){7}|::|:(?::[[:xdigit:]]{1,4}){1,6}|[[:xdigit:]]{1,4}:(?::[[:xdigit:]]{1,4}){1,5}|(?:[[:xdigit:]]{1,4}:){2}(?::[[:xdigit:]]{1,4}){1,4}|(?:[[:xdigit:]]{1,4}:){3}(?::[[:xdigit:]]{1,4}){1,3}|(?:[[:xdigit:]]{1,4}:){4}(?::[[:xdigit:]]{1,4}){1,2}|(?:[[:xdigit:]]{1,4}:){5}:[[:xdigit:]]{1,4}|(?:[[:xdigit:]]{1,4}:){1,6}:)`,
			Type: conf.TypeText, Group: model.GLOBAL, Flag: model.PRIVATE},
		// aria2 settings
		{Key: conf.Aria2Uri, Value: "http://localhost:6800/jsonrpc", Type: conf.TypeString, Group: model.ARIA2, Flag: model.PRIVATE},
		{Key: conf.Aria2Secret, Value: "", Type: conf.TypeString, Group: model.ARIA2, Flag: model.PRIVATE},
		// single settings
		{Key: conf.Token, Value: token, Type: conf.TypeString, Group: model.SINGLE, Flag: model.PRIVATE},
	}
	if flags.Dev {
		initialSettingItems = append(initialSettingItems, model.SettingItem{Key: "test_deprecated", Value: "test_value", Type: conf.TypeString, Flag: model.DEPRECATED})
	}
}
