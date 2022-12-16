package pikpak_share

import (
	"errors"
	"net/http"

	"github.com/alist-org/alist/v3/drivers/base"
	"github.com/alist-org/alist/v3/internal/op"
	"github.com/go-resty/resty/v2"
	jsoniter "github.com/json-iterator/go"
)

// do others that not defined in Driver interface

func (d *PikPakShare) login() error {
	url := "https://user.mypikpak.com/v1/auth/signin"
	var e RespErr
	res, err := base.RestyClient.R().SetError(&e).SetBody(base.Json{
		"captcha_token": "",
		"client_id":     "YNxT9w7GMdWvEOKa",
		"client_secret": "dbw2OtmVEeuUvIptb1Coyg",
		"username":      d.Username,
		"password":      d.Password,
	}).Post(url)
	if err != nil {
		return err
	}
	if e.ErrorCode != 0 {
		return errors.New(e.Error)
	}
	data := res.Body()
	d.RefreshToken = jsoniter.Get(data, "refresh_token").ToString()
	d.AccessToken = jsoniter.Get(data, "access_token").ToString()
	return nil
}

func (d *PikPakShare) refreshToken() error {
	url := "https://user.mypikpak.com/v1/auth/token"
	var e RespErr
	res, err := base.RestyClient.R().SetError(&e).
		SetHeader("user-agent", "").SetBody(base.Json{
		"client_id":     "YNxT9w7GMdWvEOKa",
		"client_secret": "dbw2OtmVEeuUvIptb1Coyg",
		"grant_type":    "refresh_token",
		"refresh_token": d.RefreshToken,
	}).Post(url)
	if err != nil {
		d.Status = err.Error()
		op.MustSaveDriverStorage(d)
		return err
	}
	if e.ErrorCode != 0 {
		if e.ErrorCode == 4126 {
			// refresh_token invalid, re-login
			return d.login()
		}
		d.Status = e.Error
		op.MustSaveDriverStorage(d)
		return errors.New(e.Error)
	}
	data := res.Body()
	d.Status = "work"
	d.RefreshToken = jsoniter.Get(data, "refresh_token").ToString()
	d.AccessToken = jsoniter.Get(data, "access_token").ToString()
	op.MustSaveDriverStorage(d)
	return nil
}

func (d *PikPakShare) request(url string, method string, callback base.ReqCallback, resp interface{}) ([]byte, error) {
	req := base.RestyClient.R()
	req.SetHeader("Authorization", "Bearer "+d.AccessToken)
	if callback != nil {
		callback(req)
	}
	if resp != nil {
		req.SetResult(resp)
	}
	var e RespErr
	req.SetError(&e)
	res, err := req.Execute(method, url)
	if err != nil {
		return nil, err
	}
	if e.ErrorCode != 0 {
		if e.ErrorCode == 16 {
			// login / refresh token
			err = d.refreshToken()
			if err != nil {
				return nil, err
			}
			return d.request(url, method, callback, resp)
		}
		return nil, errors.New(e.Error)
	}
	return res.Body(), nil
}

func (d *PikPakShare) getSharePassToken() error {
	query := map[string]string{
		"share_id":       d.ShareId,
		"pass_code":      d.SharePwd,
		"thumbnail_size": "SIZE_LARGE",
		"limit":          "100",
	}
	var resp ShareResp
	_, err := d.request("https://api-drive.mypikpak.com/drive/v1/share", http.MethodGet, func(req *resty.Request) {
		req.SetQueryParams(query)
	}, &resp)
	if err != nil {
		return err
	}
	d.PassCodeToken = resp.PassCodeToken
	return nil
}

func (d *PikPakShare) getFiles(id string) ([]File, error) {
	res := make([]File, 0)
	pageToken := "first"
	for pageToken != "" {
		if pageToken == "first" {
			pageToken = ""
		}
		query := map[string]string{
			"parent_id":       id,
			"share_id":        d.ShareId,
			"thumbnail_size":  "SIZE_LARGE",
			"with_audit":      "true",
			"limit":           "100",
			"filters":         `{"phase":{"eq":"PHASE_TYPE_COMPLETE"},"trashed":{"eq":false}}`,
			"page_token":      pageToken,
			"pass_code_token": d.PassCodeToken,
		}
		var resp ShareResp
		_, err := d.request("https://api-drive.mypikpak.com/drive/v1/share/detail", http.MethodGet, func(req *resty.Request) {
			req.SetQueryParams(query)
		}, &resp)
		if err != nil {
			return nil, err
		}
		if resp.ShareStatus != "OK" {
			if resp.ShareStatus == "PASS_CODE_EMPTY" || resp.ShareStatus == "PASS_CODE_ERROR" {
				err = d.getSharePassToken()
				if err != nil {
					return nil, err
				}
				return d.getFiles(id)
			}
			return nil, errors.New(resp.ShareStatusText)
		}
		pageToken = resp.NextPageToken
		res = append(res, resp.Files...)
	}
	return res, nil
}
