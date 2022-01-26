package file

import (
	"github.com/Xhofe/alist/drivers/operate"
	"github.com/Xhofe/alist/server/common"
	"github.com/Xhofe/alist/utils"
	"github.com/gin-gonic/gin"
)

type RenameReq struct {
	Path string `json:"path"`
	Name string `json:"name"`
}

func Rename(c *gin.Context) {
	var req RenameReq
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
		common.ErrorStrResp(c, "Can't edit account name here", 400)
		return
	}
	err = operate.Move(driver, account, path_, utils.Join(utils.Dir(path_), req.Name), true)
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	common.SuccessResp(c)
}
