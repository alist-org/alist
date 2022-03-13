package bootstrap

import (
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/model"
	"github.com/Xhofe/alist/utils"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"strings"
)

func InitSettings() {
	log.Infof("init settings...")

	err := model.SaveSetting(model.Version)
	if err != nil {
		log.Fatalf("failed write setting: %s", err.Error())
	}

	settings := []model.SettingItem{
		{
			Key:         "title",
			Value:       "Alist",
			Description: "title",
			Type:        "string",
			Access:      model.PUBLIC,
			Group:       model.FRONT,
		},
		{
			Key:         "password",
			Value:       utils.RandomStr(8),
			Description: "password",
			Type:        "string",
			Access:      model.PRIVATE,
			Group:       model.BACK,
		},
		{
			Key:         "logo",
			Value:       "https://cdn.jsdelivr.net/gh/alist-org/logo@main/can_circle.svg",
			Description: "logo",
			Type:        "string",
			Access:      model.PUBLIC,
			Group:       model.FRONT,
		},
		{
			Key:         "favicon",
			Value:       "https://cdn.jsdelivr.net/gh/alist-org/logo@main/logo.svg",
			Description: "favicon",
			Type:        "string",
			Access:      model.PUBLIC,
			Group:       model.FRONT,
		},
		{
			Key:         "icon color",
			Value:       "#1890ff",
			Description: "icon's color",
			Type:        "string",
			Access:      model.PUBLIC,
			Group:       model.FRONT,
		},
		{
			Key:         "announcement",
			Value:       "This is a test announcement.",
			Description: "announcement message (support markdown)",
			Type:        "text",
			Access:      model.PUBLIC,
			Group:       model.FRONT,
		},
		{
			Key:         "text types",
			Value:       strings.Join(conf.TextTypes, ","),
			Type:        "string",
			Description: "text type extensions",
			Group:       model.FRONT,
		},
		{
			Key:         "d_proxy types",
			Value:       strings.Join(conf.DProxyTypes, ","),
			Type:        "string",
			Description: "/d but proxy",
			Access:      model.PRIVATE,
			Group:       model.BACK,
		},
		{
			Key:         "hide files",
			Value:       "/\\/README.md/i",
			Type:        "text",
			Description: "hide files, support RegExp, one per line",
			Group:       model.FRONT,
		},
		{
			Key:         "music cover",
			Value:       "https://cdn.jsdelivr.net/gh/alist-org/logo@main/circle_center.svg",
			Description: "music cover image",
			Type:        "string",
			Access:      model.PUBLIC,
			Group:       model.FRONT,
		},
		{
			Key:         "site beian",
			Description: "chinese beian info",
			Type:        "string",
			Access:      model.PUBLIC,
			Group:       model.FRONT,
		},
		{
			Key:         "home readme url",
			Description: "when have multiple, the readme file to show",
			Type:        "string",
			Access:      model.PUBLIC,
			Group:       model.FRONT,
		},
		{
			Key:    "autoplay video",
			Value:  "false",
			Type:   "bool",
			Access: model.PUBLIC,
			Group:  model.FRONT,
		},
		{
			Key:    "autoplay audio",
			Value:  "false",
			Type:   "bool",
			Access: model.PUBLIC,
			Group:  model.FRONT,
		},
		{
			Key:         "check parent folder",
			Value:       "false",
			Type:        "bool",
			Description: "check parent folder password",
			Access:      model.PRIVATE,
			Group:       model.BACK,
		},
		{
			Key:         "customize head",
			Value:       "",
			Type:        "text",
			Description: "Customize head, placed at the beginning of the head",
			Access:      model.PRIVATE,
			Group:       model.FRONT,
		},
		{
			Key:         "customize body",
			Value:       "",
			Type:        "text",
			Description: "Customize script, placed at the end of the body",
			Access:      model.PRIVATE,
			Group:       model.FRONT,
		},
		{
			Key:         "home emoji",
			Value:       "🏠",
			Type:        "string",
			Description: "emoji in front of home in nav",
			Access:      model.PUBLIC,
			Group:       model.FRONT,
		},
		{
			Key:         "animation",
			Value:       "true",
			Type:        "bool",
			Description: "when there are a lot of files, the animation will freeze when opening",
			Access:      model.PUBLIC,
			Group:       model.FRONT,
		},
		{
			Key:         "check down link",
			Value:       "false",
			Type:        "bool",
			Description: "check down link password, your link will be 'https://alist.com/d/filename?pw=xxx'",
			Access:      model.PUBLIC,
			Group:       model.BACK,
		},
		{
			Key:         "WebDAV username",
			Value:       "admin",
			Description: "WebDAV username",
			Type:        "string",
			Access:      model.PRIVATE,
			Group:       model.BACK,
		},
		{
			Key:         "WebDAV password",
			Value:       utils.RandomStr(8),
			Description: "WebDAV password",
			Type:        "string",
			Access:      model.PRIVATE,
			Group:       model.BACK,
		},
		{
			Key:         "artplayer whitelist",
			Value:       "*",
			Description: "refer to https://artplayer.org/document/options#whitelist",
			Type:        "string",
			Access:      model.PUBLIC,
			Group:       model.FRONT,
		},
		{
			Key:         "artplayer autoSize",
			Value:       "true",
			Description: "refer to https://artplayer.org/document/options#autosize",
			Type:        "bool",
			Access:      model.PUBLIC,
			Group:       model.FRONT,
		},
		{
			Key:         "Visitor WebDAV username",
			Value:       "guest",
			Description: "Visitor WebDAV username",
			Type:        "string",
			Access:      model.PRIVATE,
			Group:       model.BACK,
		},
		{
			Key:         "Visitor WebDAV password",
			Value:       "guest",
			Description: "Visitor WebDAV password",
			Type:        "string",
			Access:      model.PRIVATE,
			Group:       model.BACK,
		},
		{
			Key:         "load type",
			Value:       "all",
			Type:        "select",
			Values:      "all,load more,auto load more,pagination",
			Description: "Not recommended to choose to auto load more, it has bugs now",
			Access:      model.PUBLIC,
			Group:       model.FRONT,
		},
		{
			Key:    "default page size",
			Value:  "30",
			Type:   "number",
			Access: model.PUBLIC,
			Group:  model.FRONT,
		},
		{
			Key:         "ocr api",
			Value:       "https://api.xhofe.top/ocr/file/json",
			Description: "Used to identify verification codes",
			Type:        "string",
			Access:      model.PRIVATE,
			Group:       model.BACK,
		},
	}
	for i, _ := range settings {
		v := settings[i]
		v.Version = conf.GitTag
		o, err := model.GetSettingByKey(v.Key)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				err = model.SaveSetting(v)
				if v.Key == "password" {
					log.Infof("Initial password: %s", conf.C.Sprintf(v.Value))
				}
				if err != nil {
					log.Fatalf("failed write setting: %s", err.Error())
				}
			} else {
				log.Fatalf("can't get setting: %s", err.Error())
			}
		} else {
			//o.Version = conf.GitTag
			//err = model.SaveSetting(*o)
			v.Value = o.Value
			err = model.SaveSetting(v)
			if err != nil {
				log.Fatalf("failed write setting: %s", err.Error())
			}
			if v.Key == "password" {
				log.Infof("Your password: %s", conf.C.Sprintf(v.Value))
			}
		}
	}
	model.LoadSettings()
}
