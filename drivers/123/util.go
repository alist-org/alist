package _123

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/alist-org/alist/v3/drivers/base"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/go-resty/resty/v2"
	jsoniter "github.com/json-iterator/go"
)

// do others that not defined in Driver interface

const (
	API             = "https://www.123pan.com/b/api"
	SignIn          = API + "/user/sign_in"
	UserInfo        = API + "/user/info"
	FileList        = API + "/file/list/new"
	DownloadInfo    = "https://www.123pan.com/a/api/file/download_info"
	Mkdir           = API + "/file/upload_request"
	Move            = API + "/file/mod_pid"
	Rename          = API + "/file/rename"
	Trash           = API + "/file/trash"
	UploadRequest   = API + "/file/upload_request"
	UploadComplete  = API + "/file/upload_complete"
	S3PreSignedUrls = API + "/file/s3_repare_upload_parts_batch"
	S3Complete      = API + "/file/s3_complete_multipart_upload"
)

func (d *Pan123) login() error {
	var body base.Json
	if utils.IsEmailFormat(d.Username) {
		body = base.Json{
			"mail":     d.Username,
			"password": d.Password,
			"type":     2,
		}
	} else {
		body = base.Json{
			"passport": d.Username,
			"password": d.Password,
		}
	}
	res, err := base.RestyClient.R().
		SetBody(body).Post(SignIn)
	if err != nil {
		return err
	}
	if utils.Json.Get(res.Body(), "code").ToInt() != 200 {
		err = fmt.Errorf(utils.Json.Get(res.Body(), "message").ToString())
	} else {
		d.AccessToken = utils.Json.Get(res.Body(), "data", "token").ToString()
	}
	return err
}

func (d *Pan123) request(url string, method string, callback base.ReqCallback, resp interface{}) ([]byte, error) {
	req := base.RestyClient.R()
	req.SetHeaders(map[string]string{
		"origin":        "https://www.123pan.com",
		"authorization": "Bearer " + d.AccessToken,
		"platform":      "web",
		"app-version":   "1.2",
	})
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
	code := utils.Json.Get(body, "code").ToInt()
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
	page := 1
	res := make([]File, 0)
	for {
		var resp Files
		query := map[string]string{
			"driveId":        "0",
			"limit":          "100",
			"next":           "0",
			"orderBy":        d.OrderBy,
			"orderDirection": d.OrderDirection,
			"parentFileId":   parentId,
			"trashed":        "false",
			"Page":           strconv.Itoa(page),
		}
		_, err := d.request(FileList, http.MethodGet, func(req *resty.Request) {
			req.SetQueryParams(query)
		}, &resp)
		if err != nil {
			return nil, err
		}
		page++
		res = append(res, resp.Data.InfoList...)
		if len(resp.Data.InfoList) == 0 || resp.Data.Next == "-1" {
			break
		}
	}
	return res, nil
}
