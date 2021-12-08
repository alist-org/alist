package bootstrap

import (
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/model"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func InitSettings() {
	log.Infof("init settings...")
	version := model.SettingItem{
		Key:         "version",
		Value:       conf.GitTag,
		Description: "version",
		Type:        "string",
		Group:       model.CONST,
		Version:     conf.GitTag,
	}

	err := model.SaveSetting(version)
	if err != nil {
		log.Fatalf("failed write setting: %s", err.Error())
	}

	settings := []model.SettingItem{
		{
			Key:         "title",
			Value:       "Alist",
			Description: "title",
			Type:        "string",
			Group:       model.PUBLIC,
		},
		{
			Key:         "password",
			Value:       "alist",
			Description: "password",
			Type:        "string",
			Group:       model.PRIVATE,
		},
		{
			Key:         "logo",
			Value:       "https://store.heytapimage.com/cdo-portal/feedback/202112/05/1542f45f86b8609495b69c5380753135.png",
			Description: "logo",
			Type:        "string",
			Group:       model.PUBLIC,
		},
		{
			Key:         "favicon",
			Value:       "https://store.heytapimage.com/cdo-portal/feedback/202112/05/1542f45f86b8609495b69c5380753135.png",
			Description: "favicon",
			Type:        "string",
			Group:       model.PUBLIC,
		},
		{
			Key:         "icon color",
			Value:       "teal.300",
			Description: "icon's color",
			Type:        "string",
			Group:       model.PUBLIC,
		},
		{
			Key:         "text types",
			Value:       "txt,htm,html,xml,java,properties,sql,js,md,json,conf,ini,vue,php,py,bat,gitignore,yml,go,sh,c,cpp,h,hpp,tsx",
			Type:        "string",
			Description: "text type extensions",
		},
		{
			Key:         "hide readme file",
			Value:       "true",
			Type:        "bool",
			Description: "hide readme file? ",
		},
		{
			Key:         "music cover",
			Value:       "https://store.heytapimage.com/cdo-portal/feedback/202110/30/d43c41c5d257c9bc36366e310374fb19.png",
			Description: "music cover image",
			Type:        "string",
			Group:       model.PUBLIC,
		},
		{
			Key:         "site beian",
			Description: "chinese beian info",
			Type:        "string",
			Group:       model.PUBLIC,
		},
		{
			Key:         "home readme url",
			Description: "when have multiple, the readme file to show",
			Type:        "string",
			Group:       model.PUBLIC,
		},
		{
			Key:         "markdown theme",
			Value:       "vuepress",
			Description: "default | github | vuepress",
			Group:       model.PUBLIC,
			Type:        "select",
			Values:      "default,github,vuepress",
		},
		{
			Key:   "autoplay video",
			Value: "false",
			Type:  "bool",
			Group: model.PUBLIC,
		},
		{
			Key:   "autoplay audio",
			Value: "false",
			Type:  "bool",
			Group: model.PUBLIC,
		},
		{
			Key:         "check parent folder",
			Value:       "false",
			Type:        "bool",
			Description: "check parent folder password",
			Group:       model.PRIVATE,
		},
		{
			Key: "customize head",
			Value: `<style>
.chakra-ui-light{
  background-image: linear-gradient(120deg,#e0c3fc 0%,#8ec5fc 100%) !important;
  background-attachment: fixed;
}
.main-box {
  border-radius: 15px !important;
}
.chakra-ui-light .main-box {
  background-color: white !important;
}
.chakra-ui-light .readme-box {
  background-color: white !important;
}
.readme-box {
  border-radius: 15px !important;
}
</style>`,
			Type:        "text",
			Description: "Customize head, placed at the beginning of the head",
			Group:       model.PRIVATE,
		},
		{
			Key:         "customize body",
			Value:       "",
			Type:        "text",
			Description: "Customize script, placed at the end of the body",
			Group:       model.PRIVATE,
		},
		{
			Key:         "animation",
			Value:       "true",
			Type:        "bool",
			Description: "when there are a lot of files, the animation will freeze when opening",
			Group:       model.PUBLIC,
		},
		{
			Key:         "check down link",
			Value:       "false",
			Type:        "bool",
			Description: "check down link password, your link will be 'https://alist.com/d/filename?pw=xxx'",
			Group:       model.PUBLIC,
		},
		{
			Key:         "WebDAV username",
			Value:       "alist",
			Description: "WebDAV username",
			Type:        "string",
			Group:       model.PRIVATE,
		},
		{
			Key:         "WebDAV password",
			Value:       "alist",
			Description: "WebDAV password",
			Type:        "string",
			Group:       model.PRIVATE,
		},
	}
	for i, _ := range settings {
		v := settings[i]
		v.Version = conf.GitTag
		o, err := model.GetSettingByKey(v.Key)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				err = model.SaveSetting(v)
				if err != nil {
					log.Fatalf("failed write setting: %s", err.Error())
				}
			} else {
				log.Fatal("can't get setting: %s", err.Error())
			}
		} else {
			o.Version = conf.GitTag
			err = model.SaveSetting(*o)
			if err != nil {
				log.Fatalf("failed write setting: %s", err.Error())
			}
		}
	}
	model.LoadSettings()
}
