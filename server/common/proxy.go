package common

import (
	"context"
	"fmt"
	"github.com/alist-org/alist/v3/drivers/base"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/net"
	"github.com/alist-org/alist/v3/pkg/http_range"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/pkg/errors"
	"io"
	"net/http"
	"net/url"
	"sync"
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
	if link.ReadSeekCloser != nil {
		attachFileName(w, file)
		http.ServeContent(w, r, file.GetName(), file.ModTime(), link.ReadSeekCloser)
		defer link.ReadSeekCloser.Close()
		return nil
	} else if link.RangeReadCloser.RangeReader != nil {
		attachFileName(w, file)
		net.ServeHTTP(w, r, file.GetName(), file.ModTime(), file.GetSize(), link.RangeReadCloser.RangeReader)
		defer func() {
			if link.RangeReadCloser.Closers != nil {
				link.RangeReadCloser.Closers.Close()
			}
		}()
		return nil
	} else if link.Concurrency != 0 || link.PartSize != 0 {
		attachFileName(w, file)
		size := file.GetSize()
		//var finalClosers model.Closers
		finalClosers := utils.NewClosers()
		header := net.ProcessHeader(&r.Header, &link.Header)
		rangeReader := func(httpRange http_range.Range) (io.ReadCloser, error) {
			down := net.NewDownloader(func(d *net.Downloader) {
				d.Concurrency = link.Concurrency
				d.PartSize = link.PartSize
			})
			req := &net.HttpRequestParams{
				URL:       link.URL,
				Range:     httpRange,
				Size:      size,
				HeaderRef: header,
			}
			rc, err := down.Download(context.Background(), req)
			finalClosers.Add(*rc)
			return *rc, err
		}
		net.ServeHTTP(w, r, file.GetName(), file.ModTime(), file.GetSize(), rangeReader)
		defer finalClosers.Close()
		return nil
	} else {
		//transparent proxy
		header := net.ProcessHeader(&r.Header, &link.Header)
		res, err := net.RequestHttp(r.Method, header, link.URL)
		if err != nil {
			return err
		}
		defer res.Body.Close()

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
func attachFileName(w http.ResponseWriter, file model.Obj) {
	fileName := file.GetName()
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"; filename*=UTF-8''%s`, fileName, url.PathEscape(fileName)))
}
