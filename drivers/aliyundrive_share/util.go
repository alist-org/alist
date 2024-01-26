package aliyundrive_share

import (
	"errors"
	"fmt"

	"github.com/alist-org/alist/v3/drivers/base"
	"github.com/alist-org/alist/v3/internal/op"
	log "github.com/sirupsen/logrus"
)

const (
	// CanaryHeaderKey CanaryHeaderValue for lifting rate limit restrictions
	CanaryHeaderKey   = "X-Canary"
	CanaryHeaderValue = "client=web,app=share,version=v2.3.1"
)

func (d *AliyundriveShare) refreshToken() error {
	url := "https://auth.alipan.com/v2/account/token"
	var resp base.TokenResp
	var e ErrorResp
	_, err := base.RestyClient.R().
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

// do others that not defined in Driver interface
func (d *AliyundriveShare) getShareToken() error {
	data := base.Json{
		"share_id": d.ShareId,
	}
	if d.SharePwd != "" {
		data["share_pwd"] = d.SharePwd
	}
	var e ErrorResp
	var resp ShareTokenResp
	_, err := base.RestyClient.R().
		SetResult(&resp).SetError(&e).SetBody(data).
		Post("https://api.alipan.com/v2/share_link/get_share_token")
	if err != nil {
		return err
	}
	if e.Code != "" {
		return errors.New(e.Message)
	}
	d.ShareToken = resp.ShareToken
	return nil
}

func (d *AliyundriveShare) request(url, method string, callback base.ReqCallback) ([]byte, error) {
	var e ErrorResp
	req := base.RestyClient.R().
		SetError(&e).
		SetHeader("content-type", "application/json").
		SetHeader("Authorization", "Bearer\t"+d.AccessToken).
		SetHeader(CanaryHeaderKey, CanaryHeaderValue).
		SetHeader("x-share-token", d.ShareToken)
	if callback != nil {
		callback(req)
	} else {
		req.SetBody("{}")
	}
	resp, err := req.Execute(method, url)
	if err != nil {
		return nil, err
	}
	if e.Code != "" {
		if e.Code == "AccessTokenInvalid" || e.Code == "ShareLinkTokenInvalid" {
			if e.Code == "AccessTokenInvalid" {
				err = d.refreshToken()
			} else {
				err = d.getShareToken()
			}
			if err != nil {
				return nil, err
			}
			return d.request(url, method, callback)
		} else {
			return nil, errors.New(e.Code + ": " + e.Message)
		}
	}
	return resp.Body(), nil
}

func (d *AliyundriveShare) getFiles(fileId string) ([]File, error) {
	files := make([]File, 0)
	data := base.Json{
		"image_thumbnail_process": "image/resize,w_160/format,jpeg",
		"image_url_process":       "image/resize,w_1920/format,jpeg",
		"limit":                   200,
		"order_by":                d.OrderBy,
		"order_direction":         d.OrderDirection,
		"parent_file_id":          fileId,
		"share_id":                d.ShareId,
		"video_thumbnail_process": "video/snapshot,t_1000,f_jpg,ar_auto,w_300",
		"marker":                  "first",
	}
	for data["marker"] != "" {
		if data["marker"] == "first" {
			data["marker"] = ""
		}
		var e ErrorResp
		var resp ListResp
		res, err := base.RestyClient.R().
			SetHeader("x-share-token", d.ShareToken).
			SetHeader(CanaryHeaderKey, CanaryHeaderValue).
			SetResult(&resp).SetError(&e).SetBody(data).
			Post("https://api.alipan.com/adrive/v3/file/list")
		if err != nil {
			return nil, err
		}
		log.Debugf("aliyundrive share get files: %s", res.String())
		if e.Code != "" {
			if e.Code == "AccessTokenInvalid" || e.Code == "ShareLinkTokenInvalid" {
				err = d.getShareToken()
				if err != nil {
					return nil, err
				}
				return d.getFiles(fileId)
			}
			return nil, errors.New(e.Message)
		}
		data["marker"] = resp.NextMarker
		files = append(files, resp.Items...)
	}
	if len(files) > 0 && d.DriveId == "" {
		d.DriveId = files[0].DriveId
	}
	return files, nil
}
