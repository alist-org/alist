package handles

import (
	"fmt"
	"io"
	stdpath "path"
	"strings"

	"github.com/alist-org/alist/v3/internal/conf"
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/fs"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/setting"
	"github.com/alist-org/alist/v3/internal/sign"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/alist-org/alist/v3/server/common"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func Down(c *gin.Context) {
	rawPath := c.MustGet("path").(string)
	filename := stdpath.Base(rawPath)
	storage, err := fs.GetStorage(rawPath, &fs.GetStoragesArgs{})
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	if common.ShouldProxy(storage, filename) {
		Proxy(c)
		return
	} else {
		link, _, err := fs.Link(c, rawPath, model.LinkArgs{
			IP:      c.ClientIP(),
			Header:  c.Request.Header,
			Type:    c.Query("type"),
			HttpReq: c.Request,
		})
		if err != nil {
			common.ErrorResp(c, err, 500)
			return
		}
		if link.MFile != nil {
			defer func(ReadSeekCloser io.ReadCloser) {
				err := ReadSeekCloser.Close()
				if err != nil {
					log.Errorf("close data error: %s", err)
				}
			}(link.MFile)
		}
		c.Header("Referrer-Policy", "no-referrer")
		c.Header("Cache-Control", "max-age=0, no-cache, no-store, must-revalidate")
		if setting.GetBool(conf.ForwardDirectLinkParams) {
			query := c.Request.URL.Query()
			for _, v := range conf.SlicesMap[conf.IgnoreDirectLinkParams] {
				query.Del(v)
			}
			link.URL, err = utils.InjectQuery(link.URL, query)
			if err != nil {
				common.ErrorResp(c, err, 500)
				return
			}
		}
		c.Redirect(302, link.URL)
	}
}

func Proxy(c *gin.Context) {
	rawPath := c.MustGet("path").(string)
	filename := stdpath.Base(rawPath)
	storage, err := fs.GetStorage(rawPath, &fs.GetStoragesArgs{})
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	if canProxy(storage, filename) {
		downProxyUrl := storage.GetStorage().DownProxyUrl
		if downProxyUrl != "" {
			_, ok := c.GetQuery("d")
			if !ok {
				URL := fmt.Sprintf("%s%s?sign=%s",
					strings.Split(downProxyUrl, "\n")[0],
					utils.EncodePath(rawPath, true),
					sign.Sign(rawPath))
				c.Redirect(302, URL)
				return
			}
		}
		link, file, err := fs.Link(c, rawPath, model.LinkArgs{
			Header:  c.Request.Header,
			Type:    c.Query("type"),
			HttpReq: c.Request,
		})
		if err != nil {
			common.ErrorResp(c, err, 500)
			return
		}
		if link.URL != "" && setting.GetBool(conf.ForwardDirectLinkParams) {
			query := c.Request.URL.Query()
			for _, v := range conf.SlicesMap[conf.IgnoreDirectLinkParams] {
				query.Del(v)
			}
			link.URL, err = utils.InjectQuery(link.URL, query)
			if err != nil {
				common.ErrorResp(c, err, 500)
				return
			}
		}
		if storage.GetStorage().ProxyRange {
			common.ProxyRange(link, file.GetSize())
		}
		err = common.Proxy(c.Writer, c.Request, link, file)
		if err != nil {
			common.ErrorResp(c, err, 500, true)
			return
		}
	} else {
		common.ErrorStrResp(c, "proxy not allowed", 403)
		return
	}
}

// TODO need optimize
// when can be proxy?
// 1. text file
// 2. config.MustProxy()
// 3. storage.WebProxy
// 4. proxy_types
// solution: text_file + shouldProxy()
func canProxy(storage driver.Driver, filename string) bool {
	if storage.Config().MustProxy() || storage.GetStorage().WebProxy || storage.GetStorage().WebdavProxy() {
		return true
	}
	if utils.SliceContains(conf.SlicesMap[conf.ProxyTypes], utils.Ext(filename)) {
		return true
	}
	if utils.SliceContains(conf.SlicesMap[conf.TextTypes], utils.Ext(filename)) {
		return true
	}
	return false
}
