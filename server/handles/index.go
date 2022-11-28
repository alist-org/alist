package handles

import (
	"context"

	"github.com/alist-org/alist/v3/internal/search"
	"github.com/alist-org/alist/v3/server/common"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
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
	if search.Running {
		common.ErrorStrResp(c, "index is running", 400)
		return
	}
	go func() {
		ctx := context.Background()
		err := search.Clear(ctx)
		if err != nil {
			log.Errorf("clear index error: %+v", err)
			return
		}
		err = search.BuildIndex(context.Background(), req.Paths, req.IgnorePaths, req.MaxDepth, true)
		if err != nil {
			log.Errorf("build index error: %+v", err)
		}
	}()
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
