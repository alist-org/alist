package file

import (
	"github.com/Xhofe/alist/drivers/operate"
	"github.com/Xhofe/alist/server/common"
	"github.com/gin-gonic/gin"
)

type MkdirReq struct {
	Path string `json:"path"`
}

func Mkdir(c *gin.Context) {
	var req MkdirReq
	if err := c.ShouldBind(&req); err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	account, path_, driver, err := common.ParsePath(req.Path)
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	if path_ == "/" {
		common.ErrorStrResp(c, "Folder name can't be empty", 400)
		return
	}
	err = operate.MakeDir(driver, account, path_, true)
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	common.SuccessResp(c)
}
