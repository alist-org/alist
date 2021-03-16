package bootstrap

import (
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/server/models"
	"github.com/Xhofe/alist/utils"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"strings"
)

func InitModel() bool {
	log.Infof("初始化数据库...")
	switch conf.Conf.Database.Type {
	case "sqlite3":
		{
			if !(strings.HasSuffix(conf.Conf.Database.DBFile, ".db") && len(conf.Conf.Database.DBFile) > 3) {
				log.Errorf("db名称不正确.")
				return false
			}
			needMigrate := !utils.Exists(conf.Conf.Database.DBFile)
			db, err := gorm.Open(sqlite.Open(conf.Conf.Database.DBFile), &gorm.Config{})
			if err != nil {
				log.Errorf("连接数据库出现错误:%s", err.Error())
				return false
			}
			conf.DB = db
			if needMigrate {
				log.Infof("迁移数据库...")
				err = conf.DB.AutoMigrate(&models.File{})
				if err != nil {
					log.Errorf("数据库迁移失败:%s", err.Error())
					return false
				}
				models.BuildTreeAll()
			}
			return true
		}
	default:
		log.Errorf("不支持的数据库类型:%s", conf.Conf.Database.Type)
		return false
	}
}
