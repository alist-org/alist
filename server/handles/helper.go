package handles

import (
	"fmt"
	"github.com/alist-org/alist/v3/internal/conf"
	"github.com/alist-org/alist/v3/internal/setting"
	"github.com/alist-org/alist/v3/server/common"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"net/url"
	"strings"
)

func Favicon(c *gin.Context) {
	c.Redirect(302, setting.GetByKey(conf.Favicon))
}

func Plist(c *gin.Context) {
	link := c.Param("link")
	u, err := url.PathUnescape(link)
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	uUrl, err := url.Parse(u)
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	name := c.Param("name")
	log.Debug("name", name)
	u = uUrl.String()
	name = strings.TrimSuffix(name, ".plist")
	name = strings.ReplaceAll(name, "<", "[")
	name = strings.ReplaceAll(name, ">", "]")
	plist := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?><!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
    <dict>
        <key>items</key>
        <array>
            <dict>
                <key>assets</key>
                <array>
                    <dict>
                        <key>kind</key>
                        <string>software-package</string>
                        <key>url</key>
                        <string>%s</string>
                    </dict>
                </array>
                <key>metadata</key>
                <dict>
                    <key>bundle-identifier</key>
					<string>ci.nn.%s</string>
					<key>bundle-version</key>
                    <string>4.4</string>
                    <key>kind</key>
                    <string>software</string>
                    <key>title</key>
                    <string>%s</string>
                </dict>
            </dict>
        </array>
    </dict>
</plist>`, u, url.PathEscape(name), name)
	c.Header("Content-Type", "application/xml;charset=utf-8")
	c.Status(200)
	_, _ = c.Writer.WriteString(plist)
}
