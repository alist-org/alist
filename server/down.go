package server

import (
	"fmt"
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
	"net/http/httputil"
	"net/url"
	"path/filepath"
	"strings"
)

func Down(c *gin.Context) {
	rawPath, err := url.PathUnescape(c.Param("path"))
	if err != nil {
		ErrorResp(c, err, 500)
		return
	}
	rawPath = utils.ParsePath(rawPath)
	log.Debugf("down: %s", rawPath)
	account, path, driver, err := ParsePath(rawPath)
	if err != nil {
		ErrorResp(c, err, 500)
		return
	}
	if account.Type == "GoogleDrive" {
		Proxy(c)
		return
	}
	link, err := driver.Link(path, account)
	if err != nil {
		ErrorResp(c, err, 500)
		return
	}
	if account.Type == "Native" {
		c.File(link)
		return
	} else {
		c.Redirect(302, link)
		return
	}
}

func Proxy(c *gin.Context) {
	rawPath, err := url.PathUnescape(c.Param("path"))
	if err != nil {
		ErrorResp(c, err, 500)
		return
	}
	rawPath = utils.ParsePath(rawPath)
	log.Debugf("proxy: %s", rawPath)
	account, path, driver, err := ParsePath(rawPath)
	if err != nil {
		ErrorResp(c, err, 500)
		return
	}
	if !account.Proxy && utils.GetFileType(filepath.Ext(rawPath)) != conf.TEXT {
		ErrorResp(c, fmt.Errorf("[%s] not allowed proxy", account.Name), 403)
		return
	}
	link, err := driver.Link(path, account)
	if err != nil {
		ErrorResp(c, err, 500)
		return
	}
	if account.Type == "Native" {
		c.File(link)
		return
	} else {
		if utils.GetFileType(filepath.Ext(rawPath)) == conf.TEXT {
			Text(c, link)
			return
		}
		driver.Proxy(c,account)
		r := c.Request
		w := c.Writer
		target, err := url.Parse(link)
		if err != nil {
			ErrorResp(c, err, 500)
			return
		}
		protocol := "http://"
		if strings.HasPrefix(link, "https://") {
			protocol = "https://"
		}
		targetHost, err := url.Parse(fmt.Sprintf("%s%s", protocol, target.Host))
		proxy := httputil.NewSingleHostReverseProxy(targetHost)
		r.URL = target
		r.Host = target.Host
		proxy.ServeHTTP(w, r)
	}
}

var client *resty.Client

func init() {
	client = resty.New()
	client.SetRetryCount(3)
}

func Text(c *gin.Context, link string) {
	res, err := client.R().Get(link)
	if err != nil {
		ErrorResp(c, err, 500)
		return
	}
	text := res.String()
	t := utils.GetStrCoding(res.Body())
	log.Debugf("text type: %s", t)
	if t != utils.UTF8 {
		body, err := utils.GbkToUtf8(res.Body())
		if err != nil {
			ErrorResp(c, err, 500)
			return
		}
		text = string(body)
	}
	c.String(200, text)
}
