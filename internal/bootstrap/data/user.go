package data

import (
	"os"

	"github.com/alist-org/alist/v3/cmd/flags"
	"github.com/alist-org/alist/v3/internal/db"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/alist-org/alist/v3/pkg/utils/random"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

func initUser() {
	admin, err := db.GetAdmin()
	adminPassword := random.String(8)
	envpass := os.Getenv("ALIST_ADMIN_PASSWORD")
	if flags.Dev {
		adminPassword = "admin"
	} else if len(envpass) > 0 {
		adminPassword = envpass
	}
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			admin = &model.User{
				Username: "admin",
				Password: adminPassword,
				Role:     model.ADMIN,
				BasePath: "/",
			}
			if err := db.CreateUser(admin); err != nil {
				panic(err)
			} else {
				utils.Log.Infof("Successfully created the admin user and the initial password is: %s", admin.Password)
			}
		} else {
			panic(err)
		}
	}
	guest, err := db.GetGuest()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			guest = &model.User{
				Username:   "guest",
				Password:   "guest",
				Role:       model.GUEST,
				BasePath:   "/",
				Permission: 0,
			}
			if err := db.CreateUser(guest); err != nil {
				panic(err)
			}
		} else {
			panic(err)
		}
	}
}
