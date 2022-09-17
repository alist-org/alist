package handles

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"

	"github.com/alist-org/alist/v3/internal/conf"
	"github.com/alist-org/alist/v3/internal/setting"
	"github.com/alist-org/alist/v3/server/common"
	"github.com/gin-gonic/gin"
)

func Favicon(c *gin.Context) {
	c.Redirect(302, setting.GetStr(conf.Favicon))
}

var DEC = map[string]string{
	"-": "+",
	"_": "/",
	".": "=",
}

func Plist(c *gin.Context) {
	link := c.Param("link")
	for k, v := range DEC {
		link = strings.ReplaceAll(link, k, v)
	}
	u, err := base64.StdEncoding.DecodeString(link)
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	uUrl, err := url.Parse(string(u))
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	fullName := c.Param("name")
	Url := uUrl.String()
	fullName = strings.TrimSuffix(fullName, ".plist")
	name := fullName
	identifier := fmt.Sprintf("ci.nn.%s", url.PathEscape(fullName))

	if strings.Contains(fullName, "_") {
		ss := strings.Split(fullName, "_")
		name = strings.Join(ss[:len(ss)-1], "_")
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
