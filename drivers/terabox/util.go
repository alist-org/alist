package terbox

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/alist-org/alist/v3/drivers/base"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/go-resty/resty/v2"
)

func (d *Terabox) request(furl string, method string, callback base.ReqCallback, resp interface{}) ([]byte, error) {
	req := base.RestyClient.R()
	req.SetHeaders(map[string]string{
		"Cookie":           d.Cookie,
		"Accept":           "application/json, text/plain, */*",
		"Referer":          "https://www.terabox.com/",
		"User-Agent":       base.UserAgent,
		"X-Requested-With": "XMLHttpRequest",
	})
	req.SetQueryParams(map[string]string{
		"app_id":     "250528",
		"web":        "1",
		"channel":    "dubox",
		"clienttype": "0",
		"jsToken":    d.JsToken,
	})
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
	return res.Body(), nil
}

func (d *Terabox) get(pathname string, params map[string]string, resp interface{}) ([]byte, error) {
	return d.request("https://www.terabox.com"+pathname, http.MethodGet, func(req *resty.Request) {
		if params != nil {
			req.SetQueryParams(params)
		}
	}, resp)
}

func (d *Terabox) post(pathname string, params map[string]string, data interface{}, resp interface{}) ([]byte, error) {
	return d.request("https://www.terabox.com"+pathname, http.MethodPost, func(req *resty.Request) {
		if params != nil {
			req.SetQueryParams(params)
		}
		req.SetBody(data)
	}, resp)
}

func (d *Terabox) getFiles(dir string) ([]File, error) {
	page := 1
	num := 100
	params := map[string]string{
		"dir": dir,
	}
	if d.OrderBy != "" {
		params["order"] = d.OrderBy
		if d.OrderDirection == "desc" {
			params["desc"] = "1"
		}
	}
	res := make([]File, 0)
	for {
		params["page"] = strconv.Itoa(page)
		params["num"] = strconv.Itoa(num)
		var resp ListResp
		_, err := d.get("/api/list", params, &resp)
		if err != nil {
			return nil, err
		}
		if resp.Errno == 9000 {
			return nil, fmt.Errorf("terabox is not yet available in this area")
		}
		if len(resp.List) == 0 {
			break
		}
		res = append(res, resp.List...)
		page++
	}
	return res, nil
}

func sign(s1, s2 string) string {
	var a = make([]int, 256)
	var p = make([]int, 256)
	var o []byte
	var v = len(s1)
	for q := 0; q < 256; q++ {
		a[q] = int(s1[(q % v) : (q%v)+1][0])
		p[q] = q
	}
	for u, q := 0, 0; q < 256; q++ {
		u = (u + p[q] + a[q]) % 256
		p[q], p[u] = p[u], p[q]
	}
	for i, u, q := 0, 0, 0; q < len(s2); q++ {
		i = (i + 1) % 256
		u = (u + p[i]) % 256
		p[i], p[u] = p[u], p[i]
		k := p[((p[i] + p[u]) % 256)]
		o = append(o, byte(int(s2[q])^k))
	}
	return base64.StdEncoding.EncodeToString(o)
}

func (d *Terabox) genSign() (string, error) {
	var resp HomeInfoResp
	_, err := d.get("/api/home/info", map[string]string{}, &resp)
	if err != nil {
		return "", err
	}
	return sign(resp.Data.Sign3, resp.Data.Sign1), nil
}

func (d *Terabox) linkOfficial(file model.Obj, args model.LinkArgs) (*model.Link, error) {
	var resp DownloadResp
	signString, err := d.genSign()
	if err != nil {
		return nil, err
	}
	params := map[string]string{
		"type":      "dlink",
		"fidlist":   fmt.Sprintf("[%s]", file.GetID()),
		"sign":      signString,
		"vip":       "2",
		"timestamp": strconv.FormatInt(time.Now().Unix(), 10),
	}
	_, err = d.get("/api/download", params, &resp)
	if err != nil {
		return nil, err
	}

	if len(resp.Dlink) == 0 {
		return nil, fmt.Errorf("fid %s no dlink found, errno: %d", file.GetID(), resp.Errno)
	}

	res, err := base.NoRedirectClient.R().SetHeader("Cookie", d.Cookie).SetHeader("User-Agent", base.UserAgent).Get(resp.Dlink[0].Dlink)
	if err != nil {
		return nil, err
	}
	u := res.Header().Get("location")
	return &model.Link{
		URL: u,
		Header: http.Header{
			"User-Agent": []string{base.UserAgent},
		},
	}, nil
}

func (d *Terabox) linkCrack(file model.Obj, args model.LinkArgs) (*model.Link, error) {
	var resp DownloadResp2
	param := map[string]string{
		"target": fmt.Sprintf("[\"%s\"]", file.GetPath()),
		"dlink":  "1",
		"origin": "dlna",
	}
	_, err := d.get("/api/filemetas", param, &resp)
	if err != nil {
		return nil, err
	}
	return &model.Link{
		URL: resp.Info[0].Dlink,
		Header: http.Header{
			"User-Agent": []string{base.UserAgent},
		},
	}, nil
}

func (d *Terabox) manage(opera string, filelist interface{}) ([]byte, error) {
	params := map[string]string{
		"onnest": "fail",
		"opera":  opera,
	}
	marshal, err := utils.Json.Marshal(filelist)
	if err != nil {
		return nil, err
	}
	data := fmt.Sprintf("async=0&filelist=%s&ondup=newcopy", string(marshal))
	return d.post("/api/filemanager", params, data, nil)
}

func (d *Terabox) create(path string, size int64, isdir int, uploadid, block_list string) ([]byte, error) {
	params := map[string]string{}
	data := fmt.Sprintf("path=%s&size=%d&isdir=%d", encodeURIComponent(path), size, isdir)
	if uploadid != "" {
		data += fmt.Sprintf("&uploadid=%s&block_list=%s", uploadid, block_list)
	}
	return d.post("/api/create", params, data, nil)
}

func encodeURIComponent(str string) string {
	r := url.QueryEscape(str)
	r = strings.ReplaceAll(r, "+", "%20")
	return r
}
