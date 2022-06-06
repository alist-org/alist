package common

import (
	"errors"
	"fmt"
	"github.com/Xhofe/alist/drivers/base"
	"github.com/Xhofe/alist/model"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
)

var HttpClient = &http.Client{}

func Proxy(w http.ResponseWriter, r *http.Request, link *base.Link, file *model.File) error {
	// 本机读取数据
	var err error
	if link.Data != nil {
		//c.Data(http.StatusOK, "application/octet-stream", link.Data)
		defer func() {
			_ = link.Data.Close()
		}()
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"; filename*=UTF-8''%s`, file.Name, url.QueryEscape(file.Name)))
		w.Header().Set("Content-Length", strconv.FormatInt(file.Size, 10))
		if link.Header != nil {
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
	// 本机文件直接返回文件
	if link.FilePath != "" {
		f, err := os.Open(link.FilePath)
		if err != nil {
			return err
		}
		defer func() {
			_ = f.Close()
		}()
		fileStat, err := os.Stat(link.FilePath)
		if err != nil {
			return err
		}
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"; filename*=UTF-8''%s`, file.Name, url.QueryEscape(file.Name)))
		http.ServeContent(w, r, file.Name, fileStat.ModTime(), f)
		return nil
	} else {
		req, err := http.NewRequest(r.Method, link.Url, nil)
		if err != nil {
			return err
		}
		for h, val := range r.Header {
			if strings.ToLower(h) == "authorization" {
				continue
			}
			req.Header[h] = val
		}
		log.Debugf("req headers: %+v", r.Header)
		for _, header := range link.Headers {
			req.Header.Set(header.Name, header.Value)
		}
		res, err := HttpClient.Do(req)
		if err != nil {
			return err
		}
		defer func() {
			_ = res.Body.Close()
		}()
		log.Debugf("proxy status: %d", res.StatusCode)
		for h, v := range res.Header {
			w.Header()[h] = v
		}
		w.WriteHeader(res.StatusCode)
		if res.StatusCode >= 400 {
			all, _ := ioutil.ReadAll(res.Body)
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
