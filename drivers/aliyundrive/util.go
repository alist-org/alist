package aliyundrive

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/alist-org/alist/v3/drivers/base"
	"github.com/alist-org/alist/v3/internal/op"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/go-resty/resty/v2"
)

// do others that not defined in Driver interface

func (d *AliDrive) refreshToken() error {
	url := "https://auth.aliyundrive.com/v2/account/token"
	var resp base.TokenResp
	var e RespErr
	_, err := base.RestyClient.R().
		//ForceContentType("application/json").
		SetBody(base.Json{"refresh_token": d.RefreshToken, "grant_type": "refresh_token"}).
		SetResult(&resp).
		SetError(&e).
		Post(url)
	if err != nil {
		return err
	}
	if e.Code != "" {
		return fmt.Errorf("failed to refresh token: %s", e.Message)
	}
	d.RefreshToken, d.AccessToken = resp.RefreshToken, resp.AccessToken
	op.MustSaveDriverStorage(d)
	return nil
}

func (d *AliDrive) request(url, method string, callback base.ReqCallback, resp interface{}) ([]byte, error, RespErr) {
	req := base.RestyClient.R()
	req.SetHeader("Authorization", "Bearer\t"+d.AccessToken)
	req.SetHeader("content-type", "application/json")
	req.SetHeader("origin", "https://www.aliyundrive.com")
	if callback != nil {
		callback(req)
	} else {
		req.SetBody("{}")
	}
	if resp != nil {
		req.SetResult(resp)
	}
	var e RespErr
	req.SetError(&e)
	res, err := req.Execute(method, url)
	if err != nil {
		return nil, err, e
	}
	if e.Code != "" {
		if e.Code == "AccessTokenInvalid" {
			err = d.refreshToken()
			if err != nil {
				return nil, err, e
			}
			return d.request(url, method, callback, resp)
		}
		return nil, errors.New(e.Message), e
	}
	return res.Body(), nil, e
}

func (d *AliDrive) getFiles(fileId string) ([]File, error) {
	marker := "first"
	res := make([]File, 0)
	for marker != "" {
		if marker == "first" {
			marker = ""
		}
		var resp Files
		data := base.Json{
			"drive_id":                d.DriveId,
			"fields":                  "*",
			"image_thumbnail_process": "image/resize,w_400/format,jpeg",
			"image_url_process":       "image/resize,w_1920/format,jpeg",
			"limit":                   200,
			"marker":                  marker,
			"order_by":                d.OrderBy,
			"order_direction":         d.OrderDirection,
			"parent_file_id":          fileId,
			"video_thumbnail_process": "video/snapshot,t_0,f_jpg,ar_auto,w_300",
			"url_expire_sec":          14400,
		}
		_, err, _ := d.request("https://api.aliyundrive.com/v2/file/list", http.MethodPost, func(req *resty.Request) {
			req.SetBody(data)
		}, &resp)

		if err != nil {
			return nil, err
		}
		marker = resp.NextMarker
		res = append(res, resp.Items...)
	}
	return res, nil
}

func (d *AliDrive) batch(srcId, dstId string, url string) error {
	res, err, _ := d.request("https://api.aliyundrive.com/v3/batch", http.MethodPost, func(req *resty.Request) {
		req.SetBody(base.Json{
			"requests": []base.Json{
				{
					"headers": base.Json{
						"Content-Type": "application/json",
					},
					"method": "POST",
					"id":     srcId,
					"body": base.Json{
						"drive_id":          d.DriveId,
						"file_id":           srcId,
						"to_drive_id":       d.DriveId,
						"to_parent_file_id": dstId,
					},
					"url": url,
				},
			},
			"resource": "file",
		})
	}, nil)
	if err != nil {
		return err
	}
	status := utils.Json.Get(res, "responses", 0, "status").ToInt()
	if status < 400 && status >= 100 {
		return nil
	}
	return errors.New(string(res))
}
