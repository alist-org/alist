package server

import (
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/public"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"io/fs"
	"io/ioutil"
	"net/http"
)

func InitIndex() {
	var index fs.File
	var err error
	if conf.Conf.Local {
		index, err = public.Public.Open("local.html")
	} else {
		index, err = public.Public.Open("index.html")
	}
	if err != nil {
		log.Fatalf(err.Error())
		return
	}
	data, _ := ioutil.ReadAll(index)
	conf.RawIndexHtml = string(data)
}

func Static(r *gin.Engine) {
	//InitIndex()
	assets, err := fs.Sub(public.Public, "assets")
	if err != nil {
		log.Fatalf("can't find assets folder")
	}
	r.StaticFS("/assets/", http.FS(assets))
	r.NoRoute(func(c *gin.Context) {
		c.Status(200)
		c.Header("Content-Type", "text/html")
		_, _ = c.Writer.WriteString(conf.IndexHtml)
		c.Writer.Flush()
		c.Writer.WriteHeaderNow()
	})
}
