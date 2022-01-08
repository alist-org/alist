package file

import (
	"errors"
	"github.com/Xhofe/alist/drivers/base"
	"github.com/Xhofe/alist/drivers/operate"
	"github.com/Xhofe/alist/server/common"
	"github.com/Xhofe/alist/utils"
	"github.com/gin-gonic/gin"
)

type DeleteFilesReq struct {
	Path  string   `json:"path"`
	Names []string `json:"names"`
}

func DeleteFiles(c *gin.Context) {
	var req DeleteFilesReq
	if err := c.ShouldBind(&req); err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	if len(req.Names) == 0 {
		common.ErrorResp(c, errors.New("empty file names"), 400)
		return
	}
	for i, name := range req.Names {
		account, path_, driver, err := common.ParsePath(utils.Join(req.Path, name))
		if err != nil {
			common.ErrorResp(c, err, 500)
			return
		}
		clearCache := false
		if i == len(req.Names)-1 {
			clearCache = true
		}
		err = operate.Delete(driver, account, path_, clearCache)
		if err != nil {
			_ = base.DeleteCache(utils.Dir(path_), account)
			common.ErrorResp(c, err, 500)
			return
		}
	}
	common.SuccessResp(c)
}
