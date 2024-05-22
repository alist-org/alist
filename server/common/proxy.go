package common

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/net"
	"github.com/alist-org/alist/v3/internal/stream"
	"github.com/alist-org/alist/v3/pkg/http_range"
	"github.com/alist-org/alist/v3/pkg/utils"
	log "github.com/sirupsen/logrus"
)

func Proxy(w http.ResponseWriter, r *http.Request, link *model.Link, file model.Obj) error {
	if link.MFile != nil {
		defer link.MFile.Close()
		attachFileName(w, file)
		contentType := link.Header.Get("Content-Type")
		if contentType != "" {
			w.Header().Set("Content-Type", contentType)
		}
		http.ServeContent(w, r, file.GetName(), file.ModTime(), link.MFile)
		return nil
	} else if link.RangeReadCloser != nil {
		attachFileName(w, file)
		net.ServeHTTP(w, r, file.GetName(), file.ModTime(), file.GetSize(), link.RangeReadCloser.RangeRead)
		defer func() {
			_ = link.RangeReadCloser.Close()
		}()
		return nil
	} else if link.Concurrency != 0 || link.PartSize != 0 {
		attachFileName(w, file)
		size := file.GetSize()
		//var finalClosers model.Closers
		finalClosers := utils.EmptyClosers()
		header := net.ProcessHeader(r.Header, link.Header)
		rangeReader := func(ctx context.Context, httpRange http_range.Range) (io.ReadCloser, error) {
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
			rc, err := down.Download(ctx, req)
			finalClosers.Add(rc)
			return rc, err
		}
		net.ServeHTTP(w, r, file.GetName(), file.ModTime(), file.GetSize(), rangeReader)
		defer finalClosers.Close()
		return nil
	} else {
		//transparent proxy
		header := net.ProcessHeader(r.Header, link.Header)
		res, err := net.RequestHttp(context.Background(), r.Method, header, link.URL)
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
	w.Header().Set("Content-Type", utils.GetMimeType(fileName))
}

var NoProxyRange = &model.RangeReadCloser{}

func ProxyRange(link *model.Link, size int64) {
	if link.MFile != nil {
		return
	}
	if link.RangeReadCloser == nil {
		var rrc, err = stream.GetRangeReadCloserFromLink(size, link)
		if err != nil {
			log.Warnf("ProxyRange error: %s", err)
			return
		}
		link.RangeReadCloser = rrc
	} else if link.RangeReadCloser == NoProxyRange {
		link.RangeReadCloser = nil
	}
}
