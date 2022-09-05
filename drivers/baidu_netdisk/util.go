package baidu_netdisk

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/alist-org/alist/v3/drivers/base"
	"github.com/alist-org/alist/v3/internal/errs"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/op"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/go-resty/resty/v2"
)

// do others that not defined in Driver interface

func (d *BaiduNetdisk) refreshToken() error {
	err := d._refreshToken()
	if err != nil && err == errs.EmptyToken {
		err = d._refreshToken()
	}
	return err
}

func (d *BaiduNetdisk) _refreshToken() error {
	u := "https://openapi.baidu.com/oauth/2.0/token"
	var resp base.TokenResp
	var e TokenErrResp
	_, err := base.RestyClient.R().SetResult(&resp).SetError(&e).SetQueryParams(map[string]string{
		"grant_type":    "refresh_token",
		"refresh_token": d.RefreshToken,
		"client_id":     d.ClientID,
		"client_secret": d.ClientSecret,
	}).Get(u)
	if err != nil {
		return err
	}
	if e.Error != "" {
		return fmt.Errorf("%s : %s", e.Error, e.ErrorDescription)
	}
	if resp.RefreshToken == "" {
		return errs.EmptyToken
	}
	d.AccessToken, d.RefreshToken = resp.AccessToken, resp.RefreshToken
	op.MustSaveDriverStorage(d)
	return nil
}

func (d *BaiduNetdisk) request(furl string, method string, callback base.ReqCallback, resp interface{}) ([]byte, error) {
	req := base.RestyClient.R()
	req.SetQueryParam("access_token", d.AccessToken)
	if callback != nil {
		callback(req)
	}
	if resp != nil {
		req.SetResult(resp)
	}
	res, err := req.Execute(method, furl)
	if err != nil {
		return nil, err
	}
	errno := utils.Json.Get(res.Body(), "errno").ToInt()
	if errno != 0 {
		if errno == -6 {
			err = d.refreshToken()
			if err != nil {
				return nil, err
			}
			return d.request(furl, method, callback, resp)
		}
		return nil, fmt.Errorf("errno: %d, refer to https://pan.baidu.com/union/doc/", errno)
	}
	return res.Body(), nil
}

func (d *BaiduNetdisk) get(pathname string, params map[string]string, resp interface{}) ([]byte, error) {
	return d.request("https://pan.baidu.com/rest/2.0"+pathname, http.MethodGet, func(req *resty.Request) {
		req.SetQueryParams(params)
	}, resp)
}

func (d *BaiduNetdisk) post(pathname string, params map[string]string, data interface{}, resp interface{}) ([]byte, error) {
	return d.request("https://pan.baidu.com/rest/2.0"+pathname, http.MethodPost, func(req *resty.Request) {
		req.SetQueryParams(params)
		req.SetBody(data)
	}, resp)
}

func (d *BaiduNetdisk) getFiles(dir string) ([]File, error) {
	start := 0
	limit := 200
	params := map[string]string{
		"method": "list",
		"dir":    dir,
		"web":    "web",
	}
	if d.OrderBy != "" {
		params["order"] = d.OrderBy
		if d.OrderDirection == "desc" {
			params["desc"] = "1"
		}
	}
	res := make([]File, 0)
	for {
		params["start"] = strconv.Itoa(start)
		params["limit"] = strconv.Itoa(limit)
		start += limit
		var resp ListResp
		_, err := d.get("/xpan/file", params, &resp)
		if err != nil {
			return nil, err
		}
		if len(resp.List) == 0 {
			break
		}
		res = append(res, resp.List...)
	}
	return res, nil
}

func (d *BaiduNetdisk) linkOfficial(file model.Obj, args model.LinkArgs) (*model.Link, error) {
	var resp DownloadResp
	params := map[string]string{
		"method": "filemetas",
		"fsids":  fmt.Sprintf("[%s]", file.GetID()),
		"dlink":  "1",
	}
	_, err := d.get("/xpan/multimedia", params, &resp)
	if err != nil {
		return nil, err
	}
	u := fmt.Sprintf("%s&access_token=%s", resp.List[0].Dlink, d.AccessToken)
	res, err := base.NoRedirectClient.R().SetHeader("User-Agent", "pan.baidu.com").Head(u)
	if err != nil {
		return nil, err
	}
	//if res.StatusCode() == 302 {
	u = res.Header().Get("location")
	//}
	return &model.Link{
		URL: u,
		Header: http.Header{
			"User-Agent": []string{"pan.baidu.com"},
		},
	}, nil
}

func (d *BaiduNetdisk) linkCrack(file model.Obj, args model.LinkArgs) (*model.Link, error) {
	var resp DownloadResp2
	param := map[string]string{
		"target": fmt.Sprintf("[\"%s\"]", file.GetPath()),
		"dlink":  "1",
		"web":    "5",
		"origin": "dlna",
	}
	_, err := d.request("https://pan.baidu.com/api/filemetas", http.MethodGet, func(req *resty.Request) {
		req.SetQueryParams(param)
	}, &resp)
	if err != nil {
		return nil, err
	}
	return &model.Link{
		URL: resp.Info[0].Dlink,
		Header: http.Header{
			"User-Agent": []string{"pan.baidu.com"},
		},
	}, nil
}

func (d *BaiduNetdisk) manage(opera string, filelist interface{}) ([]byte, error) {
	params := map[string]string{
		"method": "filemanager",
		"opera":  opera,
	}
	marshal, err := utils.Json.Marshal(filelist)
	if err != nil {
		return nil, err
	}
	data := fmt.Sprintf("async=0&filelist=%s&ondup=newcopy", string(marshal))
	return d.post("/xpan/file", params, data, nil)
}

func (d *BaiduNetdisk) create(path string, size int64, isdir int, uploadid, block_list string) ([]byte, error) {
	params := map[string]string{
		"method": "create",
	}
	data := fmt.Sprintf("path=%s&size=%d&isdir=%d", path, size, isdir)
	if uploadid != "" {
		data += fmt.Sprintf("&uploadid=%s&block_list=%s", uploadid, block_list)
	}
	return d.post("/xpan/file", params, data, nil)
}

func encodeURIComponent(str string) string {
	r := url.QueryEscape(str)
	r = strings.ReplaceAll(r, "+", "%20")
	return r
}
