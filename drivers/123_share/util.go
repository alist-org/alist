package _123Share

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/alist-org/alist/v3/drivers/base"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/go-resty/resty/v2"
	jsoniter "github.com/json-iterator/go"
)

const (
	Api          = "https://www.123pan.com/api"
	AApi         = "https://www.123pan.com/a/api"
	BApi         = "https://www.123pan.com/b/api"
	MainApi      = Api
	FileList     = MainApi + "/share/get"
	DownloadInfo = MainApi + "/share/download/info"
	//AuthKeySalt      = "8-8D$sL8gPjom7bk#cY"
)

func (d *Pan123Share) request(url string, method string, callback base.ReqCallback, resp interface{}) ([]byte, error) {
	req := base.RestyClient.R()
	req.SetHeaders(map[string]string{
		"origin":      "https://www.123pan.com",
		"referer":     "https://www.123pan.com/",
		"user-agent":  "Dart/2.19(dart:io)",
		"platform":    "android",
		"app-version": "36",
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
		return nil, errors.New(jsoniter.Get(body, "message").ToString())
	}
	return body, nil
}

func (d *Pan123Share) getFiles(parentId string) ([]File, error) {
	page := 1
	res := make([]File, 0)
	for {
		var resp Files
		query := map[string]string{
			"limit":          "100",
			"next":           "0",
			"orderBy":        d.OrderBy,
			"orderDirection": d.OrderDirection,
			"parentFileId":   parentId,
			"Page":           strconv.Itoa(page),
			"shareKey":       d.ShareKey,
			"SharePwd":       d.SharePwd,
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

// do others that not defined in Driver interface
