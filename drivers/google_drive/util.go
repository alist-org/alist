package google_drive

import (
	"fmt"
	"net/http"

	"github.com/alist-org/alist/v3/drivers/base"
	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
)

// do others that not defined in Driver interface

func (d *GoogleDrive) refreshToken() error {
	url := "https://www.googleapis.com/oauth2/v4/token"
	var resp base.TokenResp
	var e TokenError
	res, err := base.RestyClient.R().SetResult(&resp).SetError(&e).
		SetFormData(map[string]string{
			"client_id":     d.ClientID,
			"client_secret": d.ClientSecret,
			"refresh_token": d.RefreshToken,
			"grant_type":    "refresh_token",
		}).Post(url)
	if err != nil {
		return err
	}
	log.Debug(res.String())
	if e.Error != "" {
		return fmt.Errorf(e.Error)
	}
	d.AccessToken = resp.AccessToken
	return nil
}

func (d *GoogleDrive) request(url string, method string, callback base.ReqCallback, resp interface{}) ([]byte, error) {
	req := base.RestyClient.R()
	req.SetHeader("Authorization", "Bearer "+d.AccessToken)
	req.SetQueryParam("includeItemsFromAllDrives", "true")
	req.SetQueryParam("supportsAllDrives", "true")
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
			return d.request(url, method, callback, resp)
		}
		return nil, fmt.Errorf("%s: %v", e.Error.Message, e.Error.Errors)
	}
	return res.Body(), nil
}

func (d *GoogleDrive) getFiles(id string) ([]File, error) {
	pageToken := "first"
	res := make([]File, 0)
	for pageToken != "" {
		if pageToken == "first" {
			pageToken = ""
		}
		var resp Files
		orderBy := "folder,name,modifiedTime desc"
		if d.OrderBy != "" {
			orderBy = d.OrderBy + " " + d.OrderDirection
		}
		query := map[string]string{
			"orderBy":  orderBy,
			"fields":   "files(id,name,mimeType,size,modifiedTime,thumbnailLink),nextPageToken",
			"pageSize": "1000",
			"q":        fmt.Sprintf("'%s' in parents and trashed = false", id),
			//"includeItemsFromAllDrives": "true",
			//"supportsAllDrives":         "true",
			"pageToken": pageToken,
		}
		_, err := d.request("https://www.googleapis.com/drive/v3/files", http.MethodGet, func(req *resty.Request) {
			req.SetQueryParams(query)
		}, &resp)
		if err != nil {
			return nil, err
		}
		pageToken = resp.NextPageToken
		res = append(res, resp.Files...)
	}
	return res, nil
}
