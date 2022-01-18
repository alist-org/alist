package file

import (
	"github.com/Xhofe/alist/drivers/base"
	"github.com/Xhofe/alist/drivers/operate"
	"github.com/Xhofe/alist/server/common"
	"github.com/Xhofe/alist/utils"
	"github.com/gin-gonic/gin"
)

type MoveReq struct {
	SrcDir string   `json:"src_dir"`
	DstDir string   `json:"dst_dir"`
	Names  []string `json:"names"`
}

func Move(c *gin.Context) {
	var req MoveReq
	if err := c.ShouldBind(&req); err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	if len(req.Names) == 0 {
		common.ErrorStrResp(c, "Empty file names", 400)
		return
	}
	srcAccount, srcPath, srcDriver, err := common.ParsePath(utils.Join(req.SrcDir, req.Names[0]))
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	dstAccount, dstPath, _, err := common.ParsePath(utils.Join(req.DstDir, req.Names[0]))
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	if srcAccount.Name != dstAccount.Name {
		common.ErrorStrResp(c, "Can't move files between two accounts", 400)
		return
	}
	if srcPath == "/" || dstPath == "/" {
		common.ErrorStrResp(c, "Can't move root folder", 400)
		return
	}
	srcDir, dstDir := utils.Dir(srcPath), utils.Dir(dstPath)
	for i, name := range req.Names {
		clearCache := false
		if i == len(req.Names)-1 {
			clearCache = true
		}
		err := operate.Move(srcDriver, srcAccount, utils.Join(srcDir, name), utils.Join(dstDir, name), clearCache)
		if err != nil {
			if i == 0 {
				_ = base.DeleteCache(srcDir, srcAccount)
				_ = base.DeleteCache(dstDir, dstAccount)
			}
			common.ErrorResp(c, err, 500)
			return
		}
	}
	common.SuccessResp(c)
}
