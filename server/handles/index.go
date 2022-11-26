package handles

import (
	"github.com/alist-org/alist/v3/internal/search"
	"github.com/alist-org/alist/v3/server/common"
	"github.com/gin-gonic/gin"
)

type BuildIndexReq struct {
	Paths       []string `json:"paths"`
	MaxDepth    int      `json:"max_depth"`
	IgnorePaths []string `json:"ignore_paths"`
}

func BuildIndex(c *gin.Context) {
	var req BuildIndexReq
	if err := c.ShouldBind(&req); err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	err := search.BuildIndex(c, req.Paths, req.IgnorePaths, req.MaxDepth)
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	common.SuccessResp(c)
}

func GetProgress(c *gin.Context) {
	progress, err := search.Progress(c)
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	common.SuccessResp(c, progress)
}
