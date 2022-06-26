package controllers

import (
	"github.com/alist-org/alist/v3/internal/db"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/operations"
	"github.com/alist-org/alist/v3/server/common"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"strconv"
)

func ListAccounts(c *gin.Context) {
	var req common.PageReq
	if err := c.ShouldBind(&req); err != nil {
		common.ErrorResp(c, err, 400, true)
		return
	}
	log.Debugf("%+v", req)
	accounts, total, err := db.GetAccounts(req.PageIndex, req.PageSize)
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	common.SuccessResp(c, common.PageResp{
		Content: accounts,
		Total:   total,
	})
}

func CreateAccount(c *gin.Context) {
	var req model.Account
	if err := c.ShouldBind(&req); err != nil {
		common.ErrorResp(c, err, 400, true)
		return
	}
	if err := operations.CreateAccount(c, req); err != nil {
		common.ErrorResp(c, err, 500)
	} else {
		common.SuccessResp(c)
	}
}

func UpdateAccount(c *gin.Context) {
	var req model.Account
	if err := c.ShouldBind(&req); err != nil {
		common.ErrorResp(c, err, 400, true)
		return
	}
	if err := operations.UpdateAccount(c, req); err != nil {
		common.ErrorResp(c, err, 500)
	} else {
		common.SuccessResp(c)
	}
}

func DeleteAccount(c *gin.Context) {
	idStr := c.Query("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		common.ErrorResp(c, err, 400, true)
		return
	}
	if err := operations.DeleteAccountById(c, uint(id)); err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	common.SuccessResp(c)
}
