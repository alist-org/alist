package controllers

import (
	"encoding/base64"
	"fmt"
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/server/common"
	"github.com/Xhofe/alist/utils"
	"github.com/gin-gonic/gin"
	"net/url"
	"strings"
)

func Favicon(c *gin.Context) {
	c.Redirect(302, conf.GetStr("favicon"))
}

func Plist(c *gin.Context) {
	data := c.Param("data")
	data = strings.ReplaceAll(data, "_", "/")
	data = strings.ReplaceAll(data, "-", "=")
	bytes, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	u := string(bytes)
	uUrl, err := url.Parse(u)
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	name := utils.Base(u)
	u = uUrl.String()
	ipaIndex := strings.Index(name, ".ipa")
	decodeName := name
	if ipaIndex != -1 {
		name = name[:ipaIndex]
		decodeName = name
		tmp, err := url.PathUnescape(name)
		if err == nil {
			decodeName = tmp
		}
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
</plist>`, u, name, decodeName)
	c.Header("Content-Type", "application/xml;charset=utf-8")
	c.Status(200)
	_, _ = c.Writer.WriteString(plist)
}
