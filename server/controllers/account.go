package controllers

import (
	"fmt"
	"github.com/Xhofe/alist/drivers/base"
	"github.com/Xhofe/alist/model"
	"github.com/Xhofe/alist/server/common"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"strconv"
	"time"
)

func GetAccounts(c *gin.Context) {
	accounts, err := model.GetAccounts()
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	common.SuccessResp(c, accounts)
}

func CreateAccount(c *gin.Context) {
	var req model.Account
	if err := c.ShouldBind(&req); err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	driver, ok := base.GetDriver(req.Type)
	if !ok {
		common.ErrorStrResp(c, fmt.Sprintf("No [%s] driver", req.Type), 400)
		return
	}
	now := time.Now()
	req.UpdatedAt = &now
	if err := model.CreateAccount(&req); err != nil {
		common.ErrorResp(c, err, 500)
	} else {
		log.Debugf("new account: %+v", req)
		err = driver.Save(&req, nil)
		if err != nil {
			common.ErrorResp(c, err, 500)
			return
		}
		common.SuccessResp(c)
	}
}

func SaveAccount(c *gin.Context) {
	var req model.Account
	if err := c.ShouldBind(&req); err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	driver, ok := base.GetDriver(req.Type)
	if !ok {
		common.ErrorStrResp(c, fmt.Sprintf("No [%s] driver", req.Type), 400)
		return
	}
	old, err := model.GetAccountById(req.ID)
	if err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	now := time.Now()
	req.UpdatedAt = &now
	if old.Name != req.Name {
		model.DeleteAccountFromMap(old.Name)
	}
	if err := model.SaveAccount(&req); err != nil {
		common.ErrorResp(c, err, 500)
	} else {
		log.Debugf("save account: %+v", req)
		err = driver.Save(&req, old)
		if err != nil {
			common.ErrorResp(c, err, 500)
			return
		}
		common.SuccessResp(c)
	}
}

func DeleteAccount(c *gin.Context) {
	idStr := c.Query("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	if account, err := model.DeleteAccount(uint(id)); err != nil {
		common.ErrorResp(c, err, 500)
		return
	} else {
		driver, ok := base.GetDriver(account.Type)
		if ok {
			_ = driver.Save(nil, account)
		} else {
			log.Errorf("no driver: %s", account.Type)
		}
	}
	common.SuccessResp(c)
}
