package common

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/alist-org/alist/v3/internal/model"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

var HttpClient = &http.Client{}

func Proxy(w http.ResponseWriter, r *http.Request, link *model.Link, file model.Obj) error {
	// read data with native
	var err error
	if link.Data != nil {
		defer func() {
			_ = link.Data.Close()
		}()
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"; filename*=UTF-8''%s`, file.GetName(), url.QueryEscape(file.GetName())))
		w.Header().Set("Content-Length", strconv.FormatInt(file.GetSize(), 10))
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
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"; filename*=UTF-8''%s`, file.GetName(), url.QueryEscape(file.GetName())))
		http.ServeContent(w, r, file.GetName(), fileStat.ModTime(), f)
		return nil
	} else {
		req, err := http.NewRequest(r.Method, link.URL, nil)
		if err != nil {
			return err
		}
		for h, val := range r.Header {
			if strings.ToLower(h) == "authorization" {
				continue
			}
			req.Header[h] = val
		}
		for h, val := range link.Header {
			req.Header[h] = val
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
