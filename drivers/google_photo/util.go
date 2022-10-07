package google_photo

import (
	"fmt"
	"net/http"

	"github.com/alist-org/alist/v3/drivers/base"
	"github.com/go-resty/resty/v2"
)

// do others that not defined in Driver interface

func (d *GooglePhoto) refreshToken() error {
	url := "https://www.googleapis.com/oauth2/v4/token"
	var resp base.TokenResp
	var e TokenError
	_, err := base.RestyClient.R().SetResult(&resp).SetError(&e).
		SetFormData(map[string]string{
			"client_id":     d.ClientID,
			"client_secret": d.ClientSecret,
			"refresh_token": d.RefreshToken,
			"grant_type":    "refresh_token",
		}).Post(url)
	if err != nil {
		return err
	}
	if e.Error != "" {
		return fmt.Errorf(e.Error)
	}
	d.AccessToken = resp.AccessToken
	return nil
}

func (d *GooglePhoto) request(url string, method string, callback base.ReqCallback, resp interface{}, headers map[string]string) ([]byte, error) {
	req := base.RestyClient.R()
	req.SetHeader("Authorization", "Bearer "+d.AccessToken)
	if headers != nil {
		req.SetHeaders(headers)
	}

	if callback != nil {
		callback(req)
	}
	if resp != nil {
		req.SetResult(resp)
	}
	var e Error
	req.SetError(&e)
	res, err := req.Execute(method, url)
	if err != nil {
		return nil, err
	}
	if e.Error.Code != 0 {
		if e.Error.Code == 401 {
			err = d.refreshToken()
			if err != nil {
				return nil, err
			}
			return d.request(url, method, callback, resp, headers)
		}
		return nil, fmt.Errorf("%s: %v", e.Error.Message, e.Error.Errors)
	}
	return res.Body(), nil
}

func (d *GooglePhoto) getFiles() ([]MediaItem, error) {
	pageToken := "first"
	res := make([]MediaItem, 0)
	for pageToken != "" {
		if pageToken == "first" {
			pageToken = ""
		}
		var resp Files
		query := map[string]string{
			"fields":    "mediaItems(id,baseUrl,mimeType,mediaMetadata,filename),nextPageToken",
			"pageSize":  "100",
			"pageToken": pageToken,
		}
		_, err := d.request("https://photoslibrary.googleapis.com/v1/mediaItems", http.MethodGet, func(req *resty.Request) {
			req.SetQueryParams(query)
		}, &resp, nil)
		if err != nil {
			return nil, err
		}
		pageToken = resp.NextPageToken
		res = append(res, resp.MediaItems...)
	}
	return res, nil
}

func (d *GooglePhoto) getFile(id string) (MediaItem, error) {
	var resp MediaItem

	query := map[string]string{
		"fields": "baseUrl,mimeType",
	}
	_, err := d.request(fmt.Sprintf("https://photoslibrary.googleapis.com/v1/mediaItems/%s", id), http.MethodGet, func(req *resty.Request) {
		req.SetQueryParams(query)
	}, &resp, nil)
	if err != nil {
		return resp, err
	}

	return resp, nil
}
