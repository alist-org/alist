package handles

import (
	"context"
	"strconv"

	"github.com/alist-org/alist/v3/internal/db"
	"github.com/alist-org/alist/v3/internal/search"
	"github.com/alist-org/alist/v3/internal/search/bleve"
	"github.com/alist-org/alist/v3/server/common"
	"github.com/gin-gonic/gin"
)

func BuildIndex(c *gin.Context) {
	go func() {
		// TODO: consider run build index as non-admin
		user, _ := db.GetAdmin()
		ctx := context.WithValue(c.Request.Context(), "user", user)
		maxDepth, err := strconv.Atoi(c.PostForm("max_depth"))
		if err != nil {
			maxDepth = -1
		}
		indexPaths := []string{"/"}
		ignorePaths := c.PostFormArray("ignore_paths")
		bleve.BuildIndex(ctx, indexPaths, ignorePaths, maxDepth)
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
