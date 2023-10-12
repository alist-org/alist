package cloudreve

import (
	"encoding/base64"
	"errors"
	"net/http"
	"strings"

	"github.com/alist-org/alist/v3/drivers/base"
	"github.com/alist-org/alist/v3/internal/conf"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/setting"
	"github.com/alist-org/alist/v3/pkg/cookie"
	"github.com/go-resty/resty/v2"
	json "github.com/json-iterator/go"
	jsoniter "github.com/json-iterator/go"
)

// do others that not defined in Driver interface

const loginPath = "/user/session"

func (d *Cloudreve) request(method string, path string, callback base.ReqCallback, out interface{}) error {
	u := d.Address + "/api/v3" + path
	ua := d.CustomUA
	if ua == "" {
		ua = base.UserAgent
	}
	req := base.RestyClient.R()
	req.SetHeaders(map[string]string{
		"Cookie":     "cloudreve-session=" + d.Cookie,
		"Accept":     "application/json, text/plain, */*",
		"User-Agent": ua,
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
			if d.Username != "" && d.Password != "" {
				err = d.login()
				if err != nil {
					return err
				}
				return d.request(method, path, callback, out)
			}
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
	var siteConfig Config
	err := d.request(http.MethodGet, "/site/config", nil, &siteConfig)
	if err != nil {
		return err
	}
	for i := 0; i < 5; i++ {
		err = d.doLogin(siteConfig.LoginCaptcha)
		if err == nil {
			break
		}
		if err != nil && err.Error() != "CAPTCHA not match." {
			break
		}
	}
	return err
}

func (d *Cloudreve) doLogin(needCaptcha bool) error {
	var captchaCode string
	var err error
	if needCaptcha {
		var captcha string
		err = d.request(http.MethodGet, "/site/captcha", nil, &captcha)
		if err != nil {
			return err
		}
		if len(captcha) == 0 {
			return errors.New("can not get captcha")
		}
		i := strings.Index(captcha, ",")
		dec := base64.NewDecoder(base64.StdEncoding, strings.NewReader(captcha[i+1:]))
		vRes, err := base.RestyClient.R().SetMultipartField(
			"image", "validateCode.png", "image/png", dec).
			Post(setting.GetStr(conf.OcrApi))
		if err != nil {
			return err
		}
		if jsoniter.Get(vRes.Body(), "status").ToInt() != 200 {
			return errors.New("ocr error:" + jsoniter.Get(vRes.Body(), "msg").ToString())
		}
		captchaCode = jsoniter.Get(vRes.Body(), "result").ToString()
	}
	var resp Resp
	err = d.request(http.MethodPost, loginPath, func(req *resty.Request) {
		req.SetBody(base.Json{
			"username":    d.Addition.Username,
			"Password":    d.Addition.Password,
			"captchaCode": captchaCode,
		})
	}, &resp)
	return err
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

func (d *Cloudreve) GetThumb(file Object) (model.Thumbnail, error) {
	ua := d.CustomUA
	if ua == "" {
		ua = base.UserAgent
	}
	req := base.NoRedirectClient.R()
	req.SetHeaders(map[string]string{
		"Cookie":     "cloudreve-session=" + d.Cookie,
		"Accept":     "image/webp,image/apng,image/svg+xml,image/*,*/*;q=0.8",
		"User-Agent": ua,
	})
	resp, err := req.Execute(http.MethodGet, d.Address+"/api/v3/file/thumb/"+file.Id)
	if err != nil {
		return model.Thumbnail{}, err
	}
	return model.Thumbnail{
		Thumbnail: resp.Header().Get("Location"),
	}, nil
}
