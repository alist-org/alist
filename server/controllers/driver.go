package controllers

import (
	"github.com/Xhofe/alist/drivers/base"
	"github.com/Xhofe/alist/server/common"
	"github.com/gin-gonic/gin"
)

func GetDrivers(c *gin.Context) {
	common.SuccessResp(c, base.GetDrivers())
}