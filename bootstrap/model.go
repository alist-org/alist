package bootstrap

import (
	"fmt"
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/server/models"
	"github.com/Xhofe/alist/utils"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"strings"
)

func InitModel() bool {
	log.Infof("初始化数据库...")
	dbConfig := conf.Conf.Database
	switch dbConfig.Type {
	case "sqlite3":
		{
			if !(strings.HasSuffix(dbConfig.DBFile, ".db") && len(dbConfig.DBFile) > 3) {
				log.Errorf("db名称不正确.")
				return false
			}
			needMigrate := !utils.Exists(dbConfig.DBFile)
			db, err := gorm.Open(sqlite.Open(dbConfig.DBFile), &gorm.Config{
				NamingStrategy: schema.NamingStrategy{
					TablePrefix: dbConfig.TablePrefix,
				},
			})
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
				//models.BuildTreeAll()
			}
			return true
		}
	case "mysql":
		{
			dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
				dbConfig.User, dbConfig.Password, dbConfig.Host, dbConfig.Port, dbConfig.Name)
			db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
				NamingStrategy: schema.NamingStrategy{
					TablePrefix: dbConfig.TablePrefix,
				},
			})
			if err != nil {
				log.Errorf("连接数据库出现错误:%s", err.Error())
				return false
			}
			conf.DB = db
			log.Infof("迁移数据库...")
			err = conf.DB.AutoMigrate(&models.File{})
			if err != nil {
				log.Errorf("数据库迁移失败:%s", err.Error())
				return false
			}
			return true
		}
	default:
		log.Errorf("不支持的数据库类型:%s", dbConfig.Type)
		return false
	}
}
