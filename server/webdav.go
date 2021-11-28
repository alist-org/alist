package server

import (
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/server/webdav"
	"github.com/gin-gonic/gin"
	"net/http"
)

var handler *webdav.Handler

func init() {
	handler = &webdav.Handler{
		Prefix:     "/dav",
		LockSystem: webdav.NewMemLS(),
	}
}

func WebDav(r *gin.Engine) {
	dav := r.Group("/dav")
	dav.Use(WebDAVAuth)
	dav.Any("/*path", ServeWebDAV)
	dav.Any("", ServeWebDAV)
	dav.Handle("PROPFIND", "/*path", ServeWebDAV)
	dav.Handle("PROPFIND", "", ServeWebDAV)
	dav.Handle("MKCOL", "/*path", ServeWebDAV)
	dav.Handle("LOCK", "/*path", ServeWebDAV)
	dav.Handle("UNLOCK", "/*path", ServeWebDAV)
	dav.Handle("PROPPATCH", "/*path", ServeWebDAV)
	dav.Handle("COPY", "/*path", ServeWebDAV)
	dav.Handle("MOVE", "/*path", ServeWebDAV)
}

func ServeWebDAV(c *gin.Context) {
	fs := webdav.FileSystem{}
	handler.ServeHTTP(c.Writer,c.Request,&fs)
}

func WebDAVAuth(c *gin.Context) {
	if c.Request.Method == "OPTIONS" {
		c.Next()
		return
	}
	username, password, ok := c.Request.BasicAuth()
	if !ok {
		c.Writer.Header()["WWW-Authenticate"] = []string{`Basic realm="alist"`}
		c.Status(http.StatusUnauthorized)
		c.Abort()
		return
	}
	if conf.DavUsername != "" && conf.DavUsername != username {
		c.Status(http.StatusUnauthorized)
		c.Abort()
	}
	if conf.DavPassword != "" && conf.DavPassword != password {
		c.Status(http.StatusUnauthorized)
		c.Abort()
	}
	c.Next()
}