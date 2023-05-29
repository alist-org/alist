package common

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/alist-org/alist/v3/drivers/base"
	"github.com/alist-org/alist/v3/internal/conf"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func HttpClient() *http.Client {
	once.Do(func() {
		httpClient = base.NewHttpClient()
		httpClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return errors.New("stopped after 10 redirects")
			}
			req.Header.Del("Referer")
			return nil
		}
	})
	return httpClient
}

var once sync.Once
var httpClient *http.Client

func Proxy(w http.ResponseWriter, r *http.Request, link *model.Link, file model.Obj) error {
	// read data with native
	var err error
	if link.Data != nil {
		defer func() {
			_ = link.Data.Close()
		}()
		filename := file.GetName()
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"; filename*=UTF-8''%s`, filename, url.PathEscape(filename)))
		w.Header().Set("Content-Length", strconv.FormatInt(file.GetSize(), 10))
		if link.Header != nil {
			// TODO clean header with blacklist or whitelist
			link.Header.Del("set-cookie")
			for h, val := range link.Header {
				w.Header()[h] = val
			}
		}
		if link.Status == 0 {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(link.Status)
		}
		_, err = io.Copy(w, link.Data)
		if err != nil {
			return err
		}
		return nil
	}
	// local file
	if link.FilePath != nil && *link.FilePath != "" {
		f, err := os.Open(*link.FilePath)
		if err != nil {
			return err
		}
		defer func() {
			_ = f.Close()
		}()
		fileStat, err := os.Stat(*link.FilePath)
		if err != nil {
			return err
		}
		filename := file.GetName()
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"; filename*=UTF-8''%s`, filename, url.PathEscape(filename)))
		http.ServeContent(w, r, file.GetName(), fileStat.ModTime(), f)
		return nil
	} else if link.Writer != nil {
		if link.Header != nil {
			for h, v := range link.Header {
				w.Header()[h] = v
			}
		}
		if cd := w.Header().Get("Content-Disposition"); cd == "" {
			w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"; filename*=UTF-8''%s`, file.GetName(), url.PathEscape(file.GetName())))
		}
		if link.Status == 0 {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(link.Status)
		}
		return link.Writer(w)
	} else {
		req, err := http.NewRequest(r.Method, link.URL, nil)
		if err != nil {
			return err
		}
		// client header
		for h, val := range r.Header {
			if utils.SliceContains(conf.SlicesMap[conf.ProxyIgnoreHeaders], strings.ToLower(h)) {
				continue
			}
			req.Header[h] = val
		}
		// needed header
		for h, val := range link.Header {
			req.Header[h] = val
		}
		res, err := HttpClient().Do(req)
		if err != nil {
			return err
		}
		defer func() {
			_ = res.Body.Close()
		}()
		log.Debugf("proxy status: %d", res.StatusCode)
		// TODO clean header with blacklist or whitelist
		res.Header.Del("set-cookie")
		for h, v := range res.Header {
			w.Header()[h] = v
		}
		w.WriteHeader(res.StatusCode)
		if res.StatusCode >= 400 {
			all, _ := io.ReadAll(res.Body)
			msg := string(all)
			log.Debugln(msg)
			return errors.New(msg)
		}
		_, err = io.Copy(w, res.Body)
		if err != nil {
			return err
		}
		return nil
	}
}
