package store

import (
	"github.com/alist-org/alist/v3/internal/model"
	"gorm.io/gorm"
	"log"
)

var db gorm.DB

func Init(d *gorm.DB) {
	db = *d
	err := db.AutoMigrate(&model.Account{})
	if err != nil {
		log.Fatalf("failed migrate database: %s", err.Error())
	}
}
