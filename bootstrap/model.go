package bootstrap

import (
	"fmt"
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/model"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	log2 "log"
	"os"
	"strings"
	"time"
)

func InitModel() {
	log.Infof("init model...")
	var err error
	databaseConfig := conf.Conf.Database
	newLogger := logger.New(
		log2.New(os.Stdout, "\r\n", log2.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Silent,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)
	gormConfig := &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix: databaseConfig.TablePrefix,
		},
		Logger: newLogger,
	}
	switch databaseConfig.Type {
	case "sqlite3":
		{
			if !(strings.HasSuffix(databaseConfig.DBFile, ".db") && len(databaseConfig.DBFile) > 3) {
				log.Fatalf("db name error.")
			}
			db, err := gorm.Open(sqlite.Open(databaseConfig.DBFile), gormConfig)
			if err != nil {
				log.Fatalf("failed to connect database:%s", err.Error())
			}
			conf.DB = db
		}
	case "mysql":
		{
			dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
				databaseConfig.User, databaseConfig.Password, databaseConfig.Host, databaseConfig.Port, databaseConfig.Name)
			db, err := gorm.Open(mysql.Open(dsn), gormConfig)
			if err != nil {
				log.Fatalf("failed to connect database:%s", err.Error())
			}
			conf.DB = db
		}
	case "postgres":
		{
			dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=Asia/Shanghai",
				databaseConfig.Host, databaseConfig.User, databaseConfig.Password, databaseConfig.Name, databaseConfig.Port, databaseConfig.SslMode)
			db, err := gorm.Open(postgres.Open(dsn), gormConfig)
			if err != nil {
				log.Errorf("failed to connect database:%s", err.Error())
			}
			conf.DB = db
		}
	default:
		log.Fatalf("not supported database type: %s", databaseConfig.Type)
	}
	log.Infof("auto migrate model...")
	if databaseConfig.Type == "mysql" {
		err = conf.DB.Set("gorm:table_options", "ENGINE=InnoDB CHARSET=utf8mb4").
			AutoMigrate(&model.SettingItem{}, &model.Account{}, &model.Meta{})
	} else {
		err = conf.DB.AutoMigrate(&model.SettingItem{}, &model.Account{}, &model.Meta{})
	}
	if err != nil {
		log.Fatalf("failed to auto migrate")
	}
}
