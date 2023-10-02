package baidu_netdisk

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/alist-org/alist/v3/drivers/base"
	"github.com/alist-org/alist/v3/internal/errs"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/op"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/avast/retry-go"
	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
)

// do others that not defined in Driver interface

func (d *BaiduNetdisk) refreshToken() error {
	err := d._refreshToken()
	if err != nil && errors.Is(err, errs.EmptyToken) {
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
	var result []byte
	err := retry.Do(func() error {
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
			return err
		}
		log.Debugf("[baidu_netdisk] req: %s, resp: %s", furl, res.String())
		errno := utils.Json.Get(res.Body(), "errno").ToInt()
		if errno != 0 {
			if utils.SliceContains([]int{111, -6}, errno) {
				log.Info("refreshing baidu_netdisk token.")
				err2 := d.refreshToken()
				if err2 != nil {
					return retry.Unrecoverable(err2)
				}
			}
			return fmt.Errorf("req: [%s] ,errno: %d, refer to https://pan.baidu.com/union/doc/", furl, errno)
		}
		result = res.Body()
		return nil
	},
		retry.LastErrorOnly(true),
		retry.Attempts(3),
		retry.Delay(time.Second),
		retry.DelayType(retry.BackOffDelay))
	return result, err
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
			"User-Agent": []string{d.CustomCrackUA},
		},
	}, nil
}

func (d *BaiduNetdisk) manage(opera string, filelist any) ([]byte, error) {
	params := map[string]string{
		"method": "filemanager",
		"opera":  opera,
	}
	marshal, _ := utils.Json.MarshalToString(filelist)
	data := fmt.Sprintf("async=0&filelist=%s&ondup=fail", marshal)
	return d.post("/xpan/file", params, data, nil)
}

func (d *BaiduNetdisk) create(path string, size int64, isdir int, uploadid, block_list string, resp any, mtime, ctime int64) ([]byte, error) {
	params := map[string]string{
		"method": "create",
	}
	data := ""
	if mtime == 0 || ctime == 0 {
		data = fmt.Sprintf("path=%s&size=%d&isdir=%d&rtype=3", encodeURIComponent(path), size, isdir)
	} else {
		data = fmt.Sprintf("path=%s&size=%d&isdir=%d&rtype=3&local_mtime=%d&local_ctime=%d", encodeURIComponent(path), size, isdir, mtime, ctime)
	}

	if uploadid != "" {
		data += fmt.Sprintf("&uploadid=%s&block_list=%s", uploadid, block_list)
	}
	return d.post("/xpan/file", params, data, resp)
}

func encodeURIComponent(str string) string {
	r := url.QueryEscape(str)
	r = strings.ReplaceAll(r, "+", "%20")
	return r
}
