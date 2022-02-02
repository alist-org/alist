package controllers

import (
	"fmt"
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/drivers/base"
	"github.com/Xhofe/alist/drivers/operate"
	"github.com/Xhofe/alist/server/common"
	"github.com/Xhofe/alist/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
)

func Proxy(c *gin.Context) {
	rawPath := c.Param("path")
	rawPath = utils.ParsePath(rawPath)
	log.Debugf("proxy: %s", rawPath)
	account, path, driver, err := common.ParsePath(rawPath)
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	// 只有以下几种情况允许中转：
	// 1. 账号开启中转
	// 2. driver只能中转
	// 3. 是文本类型文件
	// 4. 开启webdav中转（需要验证sign）
	if !account.Proxy && !driver.Config().OnlyProxy && utils.GetFileType(filepath.Ext(rawPath)) != conf.TEXT {
		// 只开启了webdav中转，验证sign
		ok := false
		if account.WebdavProxy {
			_, ok = c.Get("sign")
		}
		if !ok {
			common.ErrorStrResp(c, fmt.Sprintf("[%s] not allowed proxy", account.Name), 403)
			return
		}
	}
	// 中转时有中转机器使用中转机器，若携带标志位则表明不能再走中转机器了
	if account.DownProxyUrl != "" && c.Query("d") != "1" {
		name := utils.Base(rawPath)
		link := fmt.Sprintf("%s%s?sign=%s", account.DownProxyUrl, rawPath, utils.SignWithToken(name, conf.Token))
		c.Redirect(302, link)
		return
	}
	// 检查文件
	file, err := operate.File(driver, account, path)
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	// 对于中转，不需要重设IP
	link, err := driver.Link(base.Args{Path: path}, account)
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	// 本机读取数据
	if link.Data != nil {
		//c.Data(http.StatusOK, "application/octet-stream", link.Data)
		defer func() {
			_ = link.Data.Close()
		}()
		c.Status(http.StatusOK)
		c.Header("Content-Type", "application/octet-stream")
		c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename=%s`, url.QueryEscape(file.Name)))
		c.Header("Content-Length", strconv.FormatInt(file.Size, 10))
		_, err = io.Copy(c.Writer, link.Data)
		if err != nil {
			_, _ = c.Writer.WriteString(err.Error())
		}
		return
	}
	// 本机文件直接返回文件
	if account.Type == "Native" {
		// 对于名称为index.html的文件需要特殊处理
		if utils.Base(rawPath) == "index.html" {
			file, err := os.Open(link.Url)
			if err != nil {
				common.ErrorResp(c, err, 500)
				return
			}
			defer func() {
				_ = file.Close()
			}()
			fileStat, err := os.Stat(link.Url)
			if err != nil {
				common.ErrorResp(c, err, 500)
				return
			}
			http.ServeContent(c.Writer, c.Request, utils.Base(rawPath), fileStat.ModTime(), file)
			return
		}
		c.File(link.Url)
		return
	} else {
		//if utils.GetFileType(filepath.Ext(rawPath)) == conf.TEXT {
		//	Text(c, link)
		//	return
		//}
		r := c.Request
		w := c.Writer
		//target, err := url.Parse(link.Url)
		//if err != nil {
		//	common.ErrorResp(c, err, 500)
		//	return
		//}
		req, err := http.NewRequest("GET", link.Url, nil)
		if err != nil {
			common.ErrorResp(c, err, 500)
			return
		}
		for h, val := range r.Header {
			req.Header[h] = val
		}
		for _, header := range link.Headers {
			req.Header.Set(header.Name, header.Value)
		}
		res, err := HttpClient.Do(req)
		if err != nil {
			common.ErrorResp(c, err, 500)
			return
		}
		defer func() {
			_ = res.Body.Close()
		}()
		log.Debugf("proxy status: %d", res.StatusCode)
		w.WriteHeader(res.StatusCode)
		for h, v := range res.Header {
			w.Header()[h] = v
		}
		_, err = io.Copy(w, res.Body)
		if err != nil {
			common.ErrorResp(c, err, 500)
			return
		}
	}
}

var client *resty.Client
var HttpClient = &http.Client{}

func init() {
	client = resty.New()
	client.SetRetryCount(3)
}

func Text(c *gin.Context, link *base.Link) {
	req := client.R()
	if link.Headers != nil {
		for _, header := range link.Headers {
			req.SetHeader(header.Name, header.Value)
		}
	}
	res, err := req.Get(link.Url)
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	text := res.String()
	t := utils.GetStrCoding(res.Body())
	log.Debugf("text type: %s", t)
	if t != utils.UTF8 {
		body, err := utils.GbkToUtf8(res.Body())
		if err != nil {
			common.ErrorResp(c, err, 500)
			return
		}
		text = string(body)
	}
	c.String(200, text)
}
