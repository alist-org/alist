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

func Options(c *gin.Context) {
	c.Header("Accept-Ranges", "bytes")
	c.Header("allow", "OPTIONS, GET, POST, HEAD, PROPFIND")
	c.Header("MS-Author-Via", "DAV")
	c.Header("DAV", "1, 2, 3")
	c.Status(204)
}

func Propfind(c *gin.Context) {
	c.Header("Content-Type", "text/xml; charset=utf-8")
	c.Status(207)
	_, _ = c.Writer.WriteString(`<?xml version="1.0" encoding="UTF-8"?><D:multistatus xmlns:D="DAV:"><D:response><D:href>/</D:href><D:propstat><D:prop><D:resourcetype><D:collection xmlns:D="DAV:"/></D:resourcetype><D:getlastmodified/></D:prop><D:status>HTTP/1.1 200 OK</D:status></D:propstat></D:response><D:response><D:href>/dav/</D:href><D:propstat><D:prop><D:resourcetype><D:collection xmlns:D="DAV:"/></D:resourcetype><D:getlastmodified/><D:supportedlock><D:lockentry xmlns:D="DAV:"><D:lockscope><D:exclusive/></D:lockscope><D:locktype><D:write/></D:locktype></D:lockentry></D:supportedlock></D:prop><D:status>HTTP/1.1 200 OK</D:status></D:propstat></D:response></D:multistatus>`)
}
