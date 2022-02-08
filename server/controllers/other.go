package controllers

import (
	"github.com/Xhofe/alist/conf"
	"github.com/gin-gonic/gin"
)

func Favicon(c *gin.Context) {
	c.Redirect(302, conf.GetStr("favicon"))
}
