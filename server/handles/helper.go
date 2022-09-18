package handles

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/alist-org/alist/v3/internal/conf"
	"github.com/alist-org/alist/v3/internal/setting"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/alist-org/alist/v3/server/common"
	"github.com/gin-gonic/gin"
)

func Favicon(c *gin.Context) {
	c.Redirect(302, setting.GetStr(conf.Favicon))
}

func Plist(c *gin.Context) {
	link := c.Param("link")
	u, err := utils.SafeAtob(link)
	if err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	uUrl, err := url.Parse(u)
	if err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	fullName := c.Param("name")
	Url := uUrl.String()
	fullName = strings.TrimSuffix(fullName, ".plist")
	fullName, err = utils.SafeAtob(fullName)
	if err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	name := fullName
	identifier := fmt.Sprintf("ci.nn.%s", url.PathEscape(fullName))
	sep := "@"
	if strings.Contains(fullName, sep) {
		ss := strings.Split(fullName, sep)
		name = strings.Join(ss[:len(ss)-1], sep)
		identifier = ss[len(ss)-1]
	}

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
                        <string><![CDATA[%s]]></string>
                    </dict>
                </array>
                <key>metadata</key>
                <dict>
                    <key>bundle-identifier</key>
					<string>%s</string>
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
</plist>`, Url, identifier, name)
	c.Header("Content-Type", "application/xml;charset=utf-8")
	c.Status(200)
	_, _ = c.Writer.WriteString(plist)
}
