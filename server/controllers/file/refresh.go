package file

import (
	"github.com/Xhofe/alist/drivers/base"
	"github.com/Xhofe/alist/server/common"
	"github.com/gin-gonic/gin"
)

type RefreshReq struct {
	Path string `json:"path"`
}

func RefreshFolder(c *gin.Context) {
	var req RefreshReq
	if err := c.ShouldBind(&req); err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	account, path_, _, err := common.ParsePath(req.Path)
	if err != nil {
		if err.Error() == "path not found" && req.Path == "/" {
			common.SuccessResp(c)
			return
		}
		common.ErrorResp(c, err, 500)
		return
	}
	err = base.DeleteCache(path_, account)
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	common.SuccessResp(c)
}
