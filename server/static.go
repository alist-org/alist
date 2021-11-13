package server

import (
	"github.com/Xhofe/alist/public"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"io/fs"
	"io/ioutil"
	"net/http"
)

var data []byte

func init() {
	index, _ := public.Public.Open("index.html")
	data, _ = ioutil.ReadAll(index)
}

func Static(r *gin.Engine) {
	assets, err := fs.Sub(public.Public, "assets")
	if err != nil {
		log.Fatalf("can't find assets folder")
	}
	r.StaticFS("/assets/", http.FS(assets))
	r.NoRoute(func(c *gin.Context) {
		c.Status(200)
		c.Header("Content-Type", "text/html")
		_, _ = c.Writer.Write(data)
		c.Writer.Flush()
	})
}
