package quqi

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	stdpath "path"
	"strings"

	"github.com/alist-org/alist/v3/drivers/base"
	"github.com/alist-org/alist/v3/internal/errs"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/go-resty/resty/v2"
)

// do others that not defined in Driver interface
func (d *Quqi) request(host string, path string, method string, callback base.ReqCallback, resp interface{}) (*resty.Response, error) {
	var (
		reqUrl = url.URL{
			Scheme: "https",
			Host:   "quqi.com",
			Path:   path,
		}
		req    = base.RestyClient.R()
		result BaseRes
	)

	if host != "" {
		reqUrl.Host = host
	}
	req.SetHeaders(map[string]string{
		"Origin": "https://quqi.com",
		"Cookie": d.Cookie,
	})

	if d.GroupID != "" {
		req.SetQueryParam("quqiid", d.GroupID)
	}

	if callback != nil {
		callback(req)
	}

	res, err := req.Execute(method, reqUrl.String())
	if err != nil {
		return nil, err
	}
	// resty.Request.SetResult cannot parse result correctly sometimes
	err = utils.Json.Unmarshal(res.Body(), &result)
	if err != nil {
		return nil, err
	}
	if result.Code != 0 {
		return nil, errors.New(result.Message)
	}
	if resp != nil {
		err = utils.Json.Unmarshal(res.Body(), resp)
		if err != nil {
			return nil, err
		}
	}
	return res, nil
}

func (d *Quqi) login() error {
	if d.Cookie != "" && d.checkLogin() {
		return nil
	}

	if d.Cookie != "" {
		return errors.New("cookie is invalid")
	}
	if d.Phone == "" {
		return errors.New("phone number is empty")
	}
	if d.Password == "" {
		return errs.EmptyPassword
	}

	resp, err := d.request("", "/auth/person/v2/login/password", resty.MethodPost, func(req *resty.Request) {
		req.SetFormData(map[string]string{
			"phone":    d.Phone,
			"password": base64.StdEncoding.EncodeToString([]byte(d.Password)),
		})
	}, nil)
	if err != nil {
		return err
	}

	var cookies []string
	for _, cookie := range resp.RawResponse.Cookies() {
		cookies = append(cookies, fmt.Sprintf("%s=%s", cookie.Name, cookie.Value))
	}
	d.Cookie = strings.Join(cookies, ";")

	return nil
}

func (d *Quqi) checkLogin() bool {
	if _, err := d.request("", "/auth/account/baseInfo", resty.MethodGet, nil, nil); err != nil {
		return false
	}
	return true
}

// rawExt 保留扩展名大小写
func rawExt(name string) string {
	ext := stdpath.Ext(name)
	if strings.HasPrefix(ext, ".") {
		ext = ext[1:]
	}

	return ext
}
