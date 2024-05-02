package baidu_netdisk

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
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

func (d *BaiduNetdisk) postForm(pathname string, params map[string]string, form map[string]string, resp interface{}) ([]byte, error) {
	return d.request("https://pan.baidu.com/rest/2.0"+pathname, http.MethodPost, func(req *resty.Request) {
		req.SetQueryParams(params)
		req.SetFormData(form)
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

	updateObjMd5(file, "pan.baidu.com", u)

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

	updateObjMd5(file, d.CustomCrackUA, resp.Info[0].Dlink)

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
	return d.postForm("/xpan/file", params, map[string]string{
		"async":    "0",
		"filelist": marshal,
		"ondup":    "fail",
	}, nil)
}

func (d *BaiduNetdisk) create(path string, size int64, isdir int, uploadid, block_list string, resp any, mtime, ctime int64) ([]byte, error) {
	params := map[string]string{
		"method": "create",
	}
	form := map[string]string{
		"path":  path,
		"size":  strconv.FormatInt(size, 10),
		"isdir": strconv.Itoa(isdir),
		"rtype": "3",
	}
	if mtime != 0 && ctime != 0 {
		joinTime(form, ctime, mtime)
	}

	if uploadid != "" {
		form["uploadid"] = uploadid
	}
	if block_list != "" {
		form["block_list"] = block_list
	}
	return d.postForm("/xpan/file", params, form, resp)
}

func joinTime(form map[string]string, ctime, mtime int64) {
	form["local_mtime"] = strconv.FormatInt(mtime, 10)
	form["local_ctime"] = strconv.FormatInt(ctime, 10)
}

func updateObjMd5(obj model.Obj, userAgent, u string) {
	object := model.GetRawObject(obj)
	if object != nil {
		req, _ := http.NewRequest(http.MethodHead, u, nil)
		req.Header.Add("User-Agent", userAgent)
		resp, _ := base.HttpClient.Do(req)
		if resp != nil {
			contentMd5 := resp.Header.Get("Content-Md5")
			object.HashInfo = utils.NewHashInfo(utils.MD5, contentMd5)
		}
	}
}

const (
	DefaultSliceSize int64 = 4 * utils.MB
	VipSliceSize           = 16 * utils.MB
	SVipSliceSize          = 32 * utils.MB
)

func (d *BaiduNetdisk) getSliceSize() int64 {
	if d.CustomUploadPartSize != 0 {
		return d.CustomUploadPartSize
	}
	switch d.vipType {
	case 1:
		return VipSliceSize
	case 2:
		return SVipSliceSize
	default:
		return DefaultSliceSize
	}
}

// func encodeURIComponent(str string) string {
// 	r := url.QueryEscape(str)
// 	r = strings.ReplaceAll(r, "+", "%20")
// 	return r
// }
