package server

import (
	"io/fs"
	"os"
	"io/ioutil"
	"net/http"
	"net/http/pprof"
	"strings"
	"path/filepath"

	"github.com/Xhofe/alist/utils"
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/public"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func InitIndex() {
	var index fs.File
	var err error
	if !strings.Contains(conf.Conf.Assets, "/") {
		conf.Conf.Assets = conf.DefaultConfig().Assets
	}
	// if LocalAssets is local path, read local index.html.
	if (utils.IsDir(filepath.Dir(conf.Conf.LocalAssets))) && utils.Exists(filepath.Join(conf.Conf.LocalAssets, "index.html")) {
		index, err = os.Open(filepath.Join(conf.Conf.LocalAssets, "index.html"))
		defer index.Close()
		log.Infof("used local index.html")
	} else {
		index, err = public.Public.Open("index.html")
	}
	if err != nil {
		log.Fatal(err.Error())
	}
    data, _ := ioutil.ReadAll(index)
	conf.RawIndexHtml = string(data)
	// if exist SUB_FOLDER, replace it by config: SubFolder
	subfolder := strings.Trim(conf.Conf.SubFolder, "/")
	if strings.Contains(conf.RawIndexHtml, "SUB_FOLDER") {
		conf.RawIndexHtml = strings.ReplaceAll(conf.RawIndexHtml, "SUB_FOLDER", subfolder)
	}
	cdnUrl := strings.ReplaceAll(conf.Conf.Assets, "$version", conf.WebTag)
	cdnUrl = strings.TrimRight(cdnUrl, "/")
	if strings.Contains(conf.RawIndexHtml, "CDN_URL") {
		if (cdnUrl == "") && (subfolder != "") {
			conf.RawIndexHtml = strings.ReplaceAll(conf.RawIndexHtml, "CDN_URL", subfolder)
			conf.RawIndexHtml = strings.ReplaceAll(conf.RawIndexHtml, "assets/", "/" + subfolder+"/assets/")	
		} else {
			conf.RawIndexHtml = strings.ReplaceAll(conf.RawIndexHtml, "/CDN_URL", cdnUrl)
			conf.RawIndexHtml = strings.ReplaceAll(conf.RawIndexHtml, "assets/", cdnUrl+"/assets/")	
		}
	}
}

func Static(r *gin.Engine) {
	var assets fs.FS
	var pub fs.FS
	var err error
	var fsys fs.FS
	//InitIndex()
	// if LocalAssets is local path, read local assets.
	fsys = os.DirFS(conf.Conf.LocalAssets)
	if (utils.IsDir(filepath.Dir(conf.Conf.LocalAssets))) && utils.Exists(filepath.Join(conf.Conf.LocalAssets, "assets")) {
		assets, err = fs.Sub(fsys, "assets")
		log.Infof("used local assets")
	} else {
		assets, err = fs.Sub(public.Public, "assets")
	}
	if err != nil {
		log.Fatalf("can't find assets folder")
	}
	r.StaticFS("/assets/", http.FS(assets))
    // if LocalAssets is local path, read local assets.
	if (utils.IsDir(filepath.Dir(conf.Conf.LocalAssets))) && utils.Exists(filepath.Join(conf.Conf.LocalAssets, "public")) {
		pub, err = fs.Sub(fsys, "public")
		log.Infof("used local public")
	} else {
		pub, err = fs.Sub(public.Public, "public")
	}
	if err != nil {
		log.Fatalf("can't find public folder")
	}
	r.StaticFS("/public/", http.FS(pub))
	r.NoRoute(func(c *gin.Context) {
		c.Header("Content-Type", "text/html")
		c.Status(200)
		if strings.HasPrefix(c.Request.URL.Path, "/@manage") {
			_, _ = c.Writer.WriteString(conf.ManageHtml)
		} else if strings.HasPrefix(c.Request.URL.Path, "/debug/pprof") && conf.Debug {
			pprof.Index(c.Writer, c.Request)
		} else {
			_, _ = c.Writer.WriteString(conf.IndexHtml)
		}
		c.Writer.Flush()
		c.Writer.WriteHeaderNow()
	})
}
