package db

import (
	"fmt"
	"log"

	"github.com/alist-org/alist/v3/internal/conf"
	"github.com/alist-org/alist/v3/internal/model"
	"gorm.io/gorm"
)

var db *gorm.DB

func Init(d *gorm.DB) {
	db = d
	err := AutoMigrate(new(model.Storage), new(model.User), new(model.Meta), new(model.SettingItem), new(model.SearchNode))
	switch conf.Conf.Database.Type {
	case "sqlite3":
	case "mysql":
		if err == nil {
			tableName := fmt.Sprintf("%ssearch_nodes", conf.Conf.Database.TablePrefix)
			db.Exec(fmt.Sprintf("CREATE FULLTEXT INDEX idx_%s_name_fulltext ON %s(name);", tableName, tableName))
		}
	case "postgres":
		if err == nil {
			db.Exec("CREATE EXTENSION pg_trgm;")
			db.Exec("CREATE EXTENSION btree_gin;")
			tableName := fmt.Sprintf("%ssearch_nodes", conf.Conf.Database.TablePrefix)
			db.Exec(fmt.Sprintf("CREATE INDEX idx_%s_name ON %s USING GIN (name);", tableName, tableName))
		}
	}
	if err != nil {
		log.Fatalf("failed migrate database: %s", err.Error())
	}
}

func AutoMigrate(dst ...interface{}) error {
	var err error
	if conf.Conf.Database.Type == "mysql" {
		err = db.Set("gorm:table_options", "ENGINE=InnoDB CHARSET=utf8mb4").AutoMigrate(dst...)
	} else {
		err = db.AutoMigrate(dst...)
	}
	return err
}
