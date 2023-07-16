package common

import (
	"fmt"
	"github.com/alist-org/alist/v3/internal/net"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"sync"

	"github.com/alist-org/alist/v3/drivers/base"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/pkg/errors"
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
		if r.Method == http.MethodHead {
			return nil
		}
		_, err = io.Copy(w, link.Data)
		if err != nil {
			return err
		}
		return nil
	}
	if link.ReadSeeker != nil {
		filename := file.GetName()
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"; filename*=UTF-8''%s`, filename, url.PathEscape(filename)))
		http.ServeContent(w, r, file.GetName(), file.ModTime(), link.ReadSeeker)
		return nil
	} else if link.RangeReader != nil {
		net.ServeHTTP(w, r, file.GetName(), file.ModTime(), file.GetSize(), link.RangeReader)
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
		if r.Method == http.MethodHead {
			return nil
		}
		return link.Writer(w)
	} else {
		//transparent proxy
		res, err := net.RequestHttp(r, link)
		if err != nil {
			return err
		}
		defer func() {
			_ = res.Body.Close()
		}()

		for h, v := range res.Header {
			w.Header()[h] = v
		}
		w.WriteHeader(res.StatusCode)
		if r.Method == http.MethodHead {
			return nil
		}
		_, err = io.Copy(w, res.Body)
		if err != nil {
			return err
		}
		return nil
	}
}
