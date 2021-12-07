package middlewares

import (
	"fmt"
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/model"
	"github.com/Xhofe/alist/server/common"
	"github.com/Xhofe/alist/utils"
	"github.com/gin-gonic/gin"
)

func PathCheck(c *gin.Context) {
	var req common.PathReq
	if err := c.ShouldBind(&req); err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	req.Path = utils.ParsePath(req.Path)
	c.Set("req",req)
	token := c.GetHeader("Authorization")
	if token == conf.Token {
		c.Next()
		return
	}
	meta, err := model.GetMetaByPath(req.Path)
	if err == nil {
		if meta.Password != "" && meta.Password != req.Password {
			common.ErrorResp(c, fmt.Errorf("wrong password"), 401)
			c.Abort()
			return
		}
	} else if conf.CheckParent {
		if !common.CheckParent(utils.Dir(req.Path), req.Password) {
			common.ErrorResp(c, fmt.Errorf("wrong password"), 401)
			c.Abort()
			return
		}
	}
	c.Next()
}