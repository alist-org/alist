package fastwebdav

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/alist-org/alist/v3/drivers/base"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
)

// do others that not defined in Driver interface

const loginPath = "/user/session"

func (d *FastWebdav) request(method string, path string, callback base.ReqCallback, resp interface{}) error {
	d.Address = strings.TrimSuffix(d.Address, "/")
	u := d.Address + "/" + path

	req := base.RestyClient.R()
	req.SetHeaders(map[string]string{
		"X-Space-App-Key": d.APIKey,
		"Accept":          "application/json, text/plain, */*",
		"Content-Type":    "application/json",
	})

	if callback != nil {
		callback(req)
	}

	if resp != nil {
		req.SetResult(resp)
	}

	res, err := req.Execute(method, u)
	if err != nil {
		return err
	}
	if !res.IsSuccess() {
		return errors.New(res.String())
	}

	return nil
}

func convertSrc(obj model.Obj) map[string]interface{} {
	m := make(map[string]interface{})
	var dirs []string
	var items []string
	if obj.IsDir() {
		dirs = append(dirs, obj.GetID())
	} else {
		items = append(items, obj.GetID())
	}
	m["dirs"] = dirs
	m["items"] = items
	return m
}

func (d *FastWebdav) getFiles(path string, id string) ([]File, error) {
	url := ""
	body := base.Json{}
	httpMethod := http.MethodGet

	if path != "/" {
		provider := getProvider(path)
		url = provider + "/list"
		log.Debug(url)
		httpMethod = http.MethodPost
		b, _ := base64.StdEncoding.DecodeString(id)
		var f File
		_ = json.Unmarshal(b, &f)
		body = base.Json{
			"path_str":       path,
			"parent_file_id": f.Id,
		}
	}

	res := make([]File, 0)
	var resp []File
	err := d.request(httpMethod, url, func(req *resty.Request) {
		req.SetBody(body)
	}, &resp)
	if err != nil {
		return nil, err
	}
	res = append(res, resp...)
	return res, nil
}

func getProvider(s string) string {
	if strings.Count(s, "/") >= 2 {
		start := strings.Index(s, "/")
		end := strings.Index(s[start+1:], "/") + start + 1
		return s[start+1 : end]
	}
	return s[1:]
}
