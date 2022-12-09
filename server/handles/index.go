package handles

import (
	"context"

	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/search"
	"github.com/alist-org/alist/v3/server/common"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type BuildIndexReq struct {
	Paths       []string `json:"paths"`
	MaxDepth    int      `json:"max_depth"`
	IgnorePaths []string `json:"ignore_paths"`
	Clear       bool     `json:"clear"`
}

func BuildIndex(c *gin.Context) {
	var req BuildIndexReq
	if err := c.ShouldBind(&req); err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	if search.Running.Load() {
		common.ErrorStrResp(c, "index is running", 400)
		return
	}
	indexPaths := search.GetIndexPaths()
	indexPaths = append(indexPaths, req.Paths...)
	indexPathsSet := mapset.NewSet[string]()
	for _, indexPath := range indexPaths {
		indexPathsSet.Add(indexPath)
	}
	indexPaths = indexPathsSet.ToSlice()
	ignorePaths, err := search.GetIgnorePaths()
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	ignorePaths = append(ignorePaths, req.IgnorePaths...)
	go func() {
		ctx := context.Background()
		var err error
		if req.Clear {
			err = search.Clear(ctx)
			if err != nil {
				log.Errorf("clear index error: %+v", err)
				return
			}
		} else {
			for _, path := range req.Paths {
				err = search.Del(ctx, path)
				if err != nil {
					log.Errorf("delete index on %s error: %+v", path, err)
					return
				}
			}
		}
		err = search.BuildIndex(context.Background(), indexPaths, ignorePaths, req.MaxDepth, true)
		if err != nil {
			log.Errorf("build index error: %+v", err)
		}
	}()
	common.SuccessResp(c)
}

func StopIndex(c *gin.Context) {
	if !search.Running.Load() {
		common.ErrorStrResp(c, "index is not running", 400)
		return
	}
	search.Quit <- struct{}{}
	common.SuccessResp(c)
}

func ClearIndex(c *gin.Context) {
	if search.Running.Load() {
		common.ErrorStrResp(c, "index is running", 400)
		return
	}
	search.Clear(c)
	search.WriteProgress(&model.IndexProgress{
		ObjCount:     0,
		IsDone:       false,
		LastDoneTime: nil,
		Error:        "",
	})
	common.SuccessResp(c)
}

func GetProgress(c *gin.Context) {
	progress, err := search.Progress()
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	common.SuccessResp(c, progress)
}
