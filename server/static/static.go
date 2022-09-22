package static

import (
	"fmt"
	"io/fs"
	"net/http"
	"net/http/pprof"
	"strings"

	"github.com/alist-org/alist/v3/cmd/flags"
	"github.com/alist-org/alist/v3/internal/conf"
	"github.com/alist-org/alist/v3/internal/setting"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/alist-org/alist/v3/public"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func InitIndex() {
	index, err := public.Public.ReadFile("dist/index.html")
	if err != nil {
		log.Fatalf("failed to read index.html: %v", err)
	}
	conf.RawIndexHtml = string(index)
	UpdateIndex()
}

func UpdateIndex() {
	cdn := strings.TrimSuffix(conf.Conf.Cdn, "/")
	cdn = strings.ReplaceAll(cdn, "$version", conf.WebVersion)
	basePath := setting.GetStr(conf.BasePath)
	if basePath != "" {
		basePath = utils.StandardizePath(basePath)
	}
	if cdn == "" {
		cdn = basePath
	}
	apiUrl := setting.GetStr(conf.ApiUrl)
	favicon := setting.GetStr(conf.Favicon)
	title := setting.GetStr(conf.SiteTitle)
	customizeHead := setting.GetStr(conf.CustomizeHead)
	customizeBody := setting.GetStr(conf.CustomizeBody)
	mainColor := setting.GetStr(conf.MainColor)
	conf.ManageHtml = conf.RawIndexHtml
	replaceMap1 := map[string]string{
		"https://jsd.nn.ci/gh/alist-org/logo@main/logo.svg": favicon,
		"Loading...":            title,
		"cdn: undefined":        fmt.Sprintf("cdn: '%s'", cdn),
		"base_path: undefined":  fmt.Sprintf("base_path: '%s'", basePath),
		"api: undefined":        fmt.Sprintf("api: '%s'", apiUrl),
		"main_color: undefined": fmt.Sprintf("main_color: '%s'", mainColor),
	}
	for k, v := range replaceMap1 {
		conf.ManageHtml = strings.Replace(conf.ManageHtml, k, v, 1)
	}
	conf.IndexHtml = conf.ManageHtml
	replaceMap2 := map[string]string{
		"<!-- customize head -->": customizeHead,
		"<!-- customize body -->": customizeBody,
	}
	for k, v := range replaceMap2 {
		conf.IndexHtml = strings.Replace(conf.IndexHtml, k, v, 1)
	}
}

func Static(r *gin.Engine) {
	InitIndex()
	folders := []string{"assets", "images", "streamer"}
	for i, folder := range folders {
		folder = "dist/" + folder
		sub, err := fs.Sub(public.Public, folder)
		if err != nil {
			log.Fatalf("can't find folder: %s", folder)
		}
		r.StaticFS(fmt.Sprintf("/%s/", folders[i]), http.FS(sub))
	}

	r.NoRoute(func(c *gin.Context) {
		c.Header("Content-Type", "text/html")
		c.Status(200)
		if strings.HasPrefix(c.Request.URL.Path, "/@manage") {
			_, _ = c.Writer.WriteString(conf.ManageHtml)
		} else if strings.HasPrefix(c.Request.URL.Path, "/debug/pprof") && flags.Debug {
			pprof.Index(c.Writer, c.Request)
		} else {
			_, _ = c.Writer.WriteString(conf.IndexHtml)
		}
		c.Writer.Flush()
		c.Writer.WriteHeaderNow()
	})
}
