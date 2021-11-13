package server

import (
	"github.com/Xhofe/alist/drivers"
	"github.com/gin-gonic/gin"
)

func GetDrivers(c *gin.Context) {
	SuccessResp(c, drivers.GetDrivers())
}