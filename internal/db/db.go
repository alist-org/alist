package db

import (
	"log"

	"github.com/alist-org/alist/v3/internal/conf"
	"github.com/alist-org/alist/v3/internal/model"
	"gorm.io/gorm"
)

var db *gorm.DB

func Init(d *gorm.DB) {
	db = d
	var err error
	if conf.Conf.Database.Type == "mysql" {
		err = db.Set("gorm:table_options", "ENGINE=InnoDB CHARSET=utf8mb4").AutoMigrate(new(model.Storage), new(model.User), new(model.Meta), new(model.SettingItem))
	} else {
		err = db.AutoMigrate(new(model.Storage), new(model.User), new(model.Meta), new(model.SettingItem))
	}
	if err != nil {
		log.Fatalf("failed migrate database: %s", err.Error())
	}
}
