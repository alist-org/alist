package handles

import (
	"context"
	"strconv"

	"github.com/alist-org/alist/v3/internal/db"
	"github.com/alist-org/alist/v3/internal/index"
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
		index.BuildIndex(ctx, indexPaths, ignorePaths, maxDepth)
	}()
	common.SuccessResp(c)
}

func GetProgress(c *gin.Context) {
	progress := index.ReadProgress()
	common.SuccessResp(c, progress)
}

func Search(c *gin.Context) {
	results := []string{}
	query, exists := c.GetQuery("query")
	if !exists {
		common.SuccessResp(c, results)
	}
	sizeStr, _ := c.GetQuery("size")
	size, err := strconv.Atoi(sizeStr)
	if err != nil {
		size = 10
	}
	searchResults, err := index.Search(query, size)
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	for _, documentMatch := range searchResults.Hits {
		results = append(results, documentMatch.Fields["Path"].(string))
	}
	common.SuccessResp(c, results)
}
