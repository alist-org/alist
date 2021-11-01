package bootstrap

import (
	"fmt"
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/drivers"
	"github.com/Xhofe/alist/model"
	"github.com/Xhofe/alist/utils"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"strings"
)

func InitModel() {
	log.Infof("init model...")
	config := conf.Conf.Database
	switch config.Type {
	case "sqlite3":
		{
			if !(strings.HasSuffix(config.DBFile, ".db") && len(config.DBFile) > 3) {
				log.Fatalf("db name error.")
			}
			db, err := gorm.Open(sqlite.Open(config.DBFile), &gorm.Config{
				NamingStrategy: schema.NamingStrategy{
					TablePrefix: config.TablePrefix,
				},
			})
			if err != nil {
				log.Fatalf("failed to connect database:%s", err.Error())
			}
			conf.DB = db
		}
	case "mysql":
		{
			dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
				config.User, config.Password, config.Host, config.Port, config.Name)
			db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
				NamingStrategy: schema.NamingStrategy{
					TablePrefix: config.TablePrefix,
				},
			})
			if err != nil {
				log.Fatalf("failed to connect database:%s", err.Error())
			}
			conf.DB = db
		}
	case "postgres":
		{
			dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=Asia/Shanghai",
				config.Host, config.User, config.Password, config.Name, config.Port)
			db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
				NamingStrategy: schema.NamingStrategy{
					TablePrefix: config.TablePrefix,
				},
			})
			if err != nil {
				log.Errorf("failed to connect database:%s", err.Error())
			}
			conf.DB = db

		}
	default:
		log.Fatalf("not supported database type: %s", config.Type)
	}
	log.Infof("auto migrate model")
	err := conf.DB.AutoMigrate(&model.SettingItem{}, &model.Account{}, &model.Meta{})
	if err != nil {
		log.Fatalf("failed to auto migrate")
	}

	// TODO init filetype
	initAccounts()
	initSettings()
}

func initAccounts() {
	log.Infof("init accounts...")
	var accounts []model.Account
	if err := conf.DB.Find(&accounts).Error; err != nil {
		log.Fatalf("failed sync init accounts")
	}
	for _, account := range accounts {
		model.RegisterAccount(account)
		driver, ok := drivers.GetDriver(account.Type)
		if !ok {
			log.Errorf("no [%s] driver", driver)
		} else {
			err := driver.Save(&account, nil)
			if err != nil {
				log.Errorf("init account [%s] error:[%s]", account.Name, err.Error())
			}
		}
	}
}

func initSettings() {
	log.Infof("init settings...")
	version, err := model.GetSettingByKey("version")
	if err != nil {
		log.Debugf("first run")
		version = &model.SettingItem{
			Key:         "version",
			Value:       "0.0.0",
			Type:        "string",
			Description: "version",
			Group:       model.CONST,
		}
	}
	settingsMap := map[string][]model.SettingItem{
		"2.0.0": {
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
				Type:        "string",
				Description: "password",
				Group:       model.PRIVATE,
			},
			{
				Key:         "version",
				Value:       "2.0.0",
				Type:        "string",
				Description: "version",
				Group:       model.CONST,
			},
			{
				Key:         "logo",
				Value:       "https://store.heytapimage.com/cdo-portal/feedback/202110/30/d43c41c5d257c9bc36366e310374fb19.png",
				Type:        "string",
				Description: "logo",
				Group:       model.PUBLIC,
			},
			{
				Key:         "icon color",
				Value:       "teal.300",
				Type:        "string",
				Description: "icon's color",
				Group:       model.PUBLIC,
			},
			{
				Key:         "text types",
				Value:       "txt,htm,html,xml,java,properties,sql,js,md,json,conf,ini,vue,php,py,bat,gitignore,yml,go,sh,c,cpp,h,hpp",
				Type:        "string",
				Description: "text type extensions",
			},
			{
				Key:         "readme file",
				Value:       "hide",
				Type:        "string",
				Description: "hide readme file? (show/hide)",
			},
			{
				Key:         "music cover",
				Value:       "https://store.heytapimage.com/cdo-portal/feedback/202110/30/d43c41c5d257c9bc36366e310374fb19.png",
				Type:        "string",
				Description: "music cover image",
			},
		},
	}
	for k, v := range settingsMap {
		if utils.VersionCompare(k, version.Value) > 0 {
			log.Infof("writing [v%s] settings", k)
			err = model.SaveSettings(v)
			if err != nil {
				log.Fatalf("save settings error")
			}
		}
	}
	textTypes, err := model.GetSettingByKey("text types")
	if err == nil {
		conf.TextTypes = strings.Split(textTypes.Value, ",")
	}
}
