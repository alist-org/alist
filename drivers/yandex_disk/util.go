package yandex_disk

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/alist-org/alist/v3/drivers/base"
	"github.com/alist-org/alist/v3/internal/op"
	"github.com/go-resty/resty/v2"
)

// do others that not defined in Driver interface

func (d *YandexDisk) refreshToken() error {
	u := "https://oauth.yandex.com/token"
	var resp base.TokenResp
	var e TokenErrResp
	_, err := base.RestyClient.R().SetResult(&resp).SetError(&e).SetFormData(map[string]string{
		"grant_type":    "refresh_token",
		"refresh_token": d.RefreshToken,
		"client_id":     d.ClientID,
		"client_secret": d.ClientSecret,
	}).Post(u)
	if err != nil {
		return err
	}
	if e.Error != "" {
		return fmt.Errorf("%s : %s", e.Error, e.ErrorDescription)
	}
	d.AccessToken, d.RefreshToken = resp.AccessToken, resp.RefreshToken
	op.MustSaveDriverStorage(d)
	return nil
}

func (d *YandexDisk) request(pathname string, method string, callback base.ReqCallback, resp interface{}) ([]byte, error) {
	u := "https://cloud-api.yandex.net/v1/disk/resources" + pathname
	req := base.RestyClient.R()
	req.SetHeader("Authorization", "OAuth "+d.AccessToken)
	if callback != nil {
		callback(req)
	}
	if resp != nil {
		req.SetResult(resp)
	}
	var e ErrResp
	req.SetError(&e)
	res, err := req.Execute(method, u)
	if err != nil {
		return nil, err
	}
	//log.Debug(res.String())
	if e.Error != "" {
		if e.Error == "UnauthorizedError" {
			err = d.refreshToken()
			if err != nil {
				return nil, err
			}
			return d.request(pathname, method, callback, resp)
		}
		return nil, errors.New(e.Description)
	}
	return res.Body(), nil
}

func (d *YandexDisk) getFiles(path string) ([]File, error) {
	limit := 100
	page := 1
	res := make([]File, 0)
	for {
		offset := (page - 1) * limit
		query := map[string]string{
			"path":   path,
			"limit":  strconv.Itoa(limit),
			"offset": strconv.Itoa(offset),
		}
		if d.OrderBy != "" {
			if d.OrderDirection == "desc" {
				query["sort"] = "-" + d.OrderBy
			} else {
				query["sort"] = d.OrderBy
			}
		}
		var resp FilesResp
		_, err := d.request("", http.MethodGet, func(req *resty.Request) {
			req.SetQueryParams(query)
		}, &resp)
		if err != nil {
			return nil, err
		}
		res = append(res, resp.Embedded.Items...)
		if resp.Embedded.Total <= offset+limit {
			break
		}
	}
	return res, nil
}
