package server

import (
	"fmt"
	"github.com/Xhofe/alist/model"
	"github.com/Xhofe/alist/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"path/filepath"
)

func Auth(c *gin.Context) {
	token := c.GetHeader("Authorization")
	password, err := model.GetSettingByKey("password")
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			ErrorResp(c, fmt.Errorf("password not set"), 400)
			return
		}
		ErrorResp(c, err, 500)
		return
	}
	if token != utils.GetMD5Encode(password.Value) {
		ErrorResp(c, fmt.Errorf("wrong password"), 401)
		return
	}
	c.Next()
}

func Login(c *gin.Context) {
	SuccessResp(c)
}

func CheckAccount(c *gin.Context) {
	if model.AccountsCount() == 0 {
		ErrorResp(c, fmt.Errorf("no accounts,please add one first"), 1001)
		return
	}
	c.Next()
}

func CheckParent(path string, password string) bool {
	meta, err := model.GetMetaByPath(path)
	if err == nil {
		if meta.Password != "" && meta.Password != password {
			return false
		}
		return true
	} else {
		if path == "/" {
			return true
		}
		return CheckParent(filepath.Dir(path), password)
	}
}
