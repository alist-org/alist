package cloudreve

import (
	"errors"
	"github.com/alist-org/alist/v3/drivers/base"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/cookie"
	"github.com/go-resty/resty/v2"
	json "github.com/json-iterator/go"
	"net/http"
)

// do others that not defined in Driver interface

const loginPath = "/user/session"

func (d *Cloudreve) request(method string, path string, callback base.ReqCallback, out interface{}) error {
	u := d.Address + "/api/v3" + path
	req := base.RestyClient.R()
	req.SetHeaders(map[string]string{
		"Cookie":     "cloudreve-session=" + d.Cookie,
		"Accept":     "application/json, text/plain, */*",
		"User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36",
	})

	var r Resp

	req.SetResult(&r)

	if callback != nil {
		callback(req)
	}

	resp, err := req.Execute(method, u)
	if err != nil {
		return err
	}
	if !resp.IsSuccess() {
		return errors.New(resp.String())
	}

	if r.Code != 0 {

		// 刷新 cookie
		if r.Code == http.StatusUnauthorized && path != loginPath {
			err = d.login()
			if err != nil {
				return err
			}
			return d.request(method, path, callback, out)
		}

		return errors.New(r.Msg)
	}
	sess := cookie.GetCookie(resp.Cookies(), "cloudreve-session")
	if sess != nil {
		d.Cookie = sess.Value
	}
	if out != nil && r.Data != nil {
		var marshal []byte
		marshal, err = json.Marshal(r.Data)
		if err != nil {
			return err
		}
		err = json.Unmarshal(marshal, out)
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *Cloudreve) login() error {
	return d.request(http.MethodPost, loginPath, func(req *resty.Request) {
		req.SetBody(base.Json{
			"username":    d.Addition.Username,
			"Password":    d.Addition.Password,
			"captchaCode": "",
		})
	}, nil)
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
