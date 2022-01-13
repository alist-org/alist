package file

import (
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
		common.ErrorStrResp(c, "Empty file names", 400)
		return
	}
	for i, name := range req.Names {
		account, path_, driver, err := common.ParsePath(utils.Join(req.Path, name))
		if err != nil {
			common.ErrorResp(c, err, 500)
			return
		}
		if path_ == "/" {
			common.ErrorStrResp(c, "Delete root folder is not allowed", 400)
			return
		}
		clearCache := false
		if i == len(req.Names)-1 {
			clearCache = true
		}
		err = operate.Delete(driver, account, path_, clearCache)
		if err != nil {
			if i == 0 {
				_ = base.DeleteCache(utils.Dir(path_), account)
			}
			common.ErrorResp(c, err, 500)
			return
		}
	}
	common.SuccessResp(c)
}
