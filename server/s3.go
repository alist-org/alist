package server

import (
	"context"
	"path"
	"strings"

	"github.com/alist-org/alist/v3/internal/conf"
	"github.com/alist-org/alist/v3/internal/setting"
	"github.com/alist-org/alist/v3/server/common"
	"github.com/alist-org/alist/v3/server/s3"
	"github.com/gin-gonic/gin"
)

func S3(g *gin.RouterGroup) {
	if !setting.GetBool(conf.S3Enabled) {
		g.Any("/*path", func(c *gin.Context) {
			common.ErrorStrResp(c, "S3 server is not enabled", 403)
		})
		return
	}
	h, _ := s3.NewServer(context.Background(), []string{setting.GetStr(conf.S3AccessKeyId) + "," + setting.GetStr(conf.S3SecretAccessKey)})

	g.Any("/*path", func(c *gin.Context) {
		adjustedPath := strings.TrimPrefix(c.Request.URL.Path, path.Join(conf.URL.Path, "/s3"))
		c.Request.URL.Path = adjustedPath
		gin.WrapH(h)(c)
	})
}
