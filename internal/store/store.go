package store

import (
	"github.com/alist-org/alist/v3/internal/model"
	"gorm.io/gorm"
)

var db gorm.DB

func Init(dbgorm *gorm.DB) {
	db = *dbgorm
	db.AutoMigrate(&model.Account{})
}
