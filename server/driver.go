package server

import (
	"github.com/Xhofe/alist/drivers/base"
	"github.com/gin-gonic/gin"
)

func GetDrivers(c *gin.Context) {
	SuccessResp(c, base.GetDrivers())
}