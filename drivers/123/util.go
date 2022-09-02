package _123

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/alist-org/alist/v3/drivers/base"
	"github.com/go-resty/resty/v2"
	jsoniter "github.com/json-iterator/go"
)

// do others that not defined in Driver interface

func (d *Pan123) login() error {
	url := "https://www.123pan.com/api/user/sign_in"
	var resp TokenResp
	_, err := base.RestyClient.R().
		SetResult(&resp).
		SetBody(base.Json{
			"passport": d.Username,
			"password": d.Password,
		}).Post(url)
	if err != nil {
		return err
	}
	if resp.Code != 200 {
		err = fmt.Errorf(resp.Message)
	} else {
		d.AccessToken = resp.Data.Token
	}
	return err
}

func (d *Pan123) request(url string, method string, callback base.ReqCallback, resp interface{}) ([]byte, error) {
	req := base.RestyClient.R()
	req.SetHeader("Authorization", "Bearer "+d.AccessToken)
	if callback != nil {
		callback(req)
	}
	if resp != nil {
		req.SetResult(resp)
	}
	res, err := req.Execute(method, url)
	if err != nil {
		return nil, err
	}
	body := res.Body()
	code := jsoniter.Get(body, "code").ToInt()
	if code != 0 {
		if code == 401 {
			err := d.login()
			if err != nil {
				return nil, err
			}
			return d.request(url, method, callback, resp)
		}
		return nil, errors.New(jsoniter.Get(body, "message").ToString())
	}
	return body, nil
}

func (d *Pan123) getFiles(parentId string) ([]File, error) {
	next := "0"
	res := make([]File, 0)
	for next != "-1" {
		var resp Files
		query := map[string]string{
			"driveId":        "0",
			"limit":          "100",
			"next":           next,
			"orderBy":        d.OrderBy,
			"orderDirection": d.OrderDirection,
			"parentFileId":   parentId,
			"trashed":        "false",
		}
		_, err := d.request("https://www.123pan.com/api/file/list/new", http.MethodGet, func(req *resty.Request) {
			req.SetQueryParams(query)
		}, &resp)
		if err != nil {
			return nil, err
		}
		next = resp.Data.Next
		res = append(res, resp.Data.InfoList...)
	}
	return res, nil
}
