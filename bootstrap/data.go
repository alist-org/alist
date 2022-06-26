package bootstrap

import (
	"github.com/alist-org/alist/v3/internal/db"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/utils/random"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func InitData() {
	initUser()
}

func initUser() {
	admin, err := db.GetAdmin()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			admin = &model.User{
				Username: "admin",
				Password: random.RandomStr(8),
				Role:     model.ADMIN,
				BasePath: "/",
				Webdav:   true,
			}
			if err := db.CreateUser(admin); err != nil {
				panic(err)
			}
		} else {
			panic(err)
		}
	}
	guest, err := db.GetGuest()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			guest = &model.User{
				Username: "guest",
				Password: "guest",
				ReadOnly: true,
				Webdav:   true,
				Role:     model.GUEST,
				BasePath: "/",
			}
			if err := db.CreateUser(guest); err != nil {
				panic(err)
			}
		} else {
			panic(err)
		}
	}
	log.Infof("admin password: %+v", admin.Password)
}
