package model

import (
	"fmt"
	"github.com/Xhofe/alist/conf"
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
	err := conf.DB.AutoMigrate(&SettingItem{},&Account{})
	if err != nil {
		log.Fatalf("failed to auto migrate")
	}

	// TODO init accounts and filetype
	initAccounts()
}

