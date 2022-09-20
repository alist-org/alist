package lanzou

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/alist-org/alist/v3/drivers/base"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/go-resty/resty/v2"
)

func (d *LanZou) get(url string, callback base.ReqCallback, resp interface{}) ([]byte, error) {
	return d.request(url, http.MethodGet, callback, false)
}

func (d *LanZou) post(url string, callback base.ReqCallback, resp interface{}) ([]byte, error) {
	return d._post(url, callback, resp, false)
}

func (d *LanZou) _post(url string, callback base.ReqCallback, resp interface{}, up bool) ([]byte, error) {
	data, err := d.request(url, http.MethodPost, callback, up)
	if err != nil {
		return nil, err
	}
	switch utils.Json.Get(data, "zt").ToInt() {
	case 1, 2, 4:
		if resp != nil {
			// 返回类型不统一,忽略错误
			utils.Json.Unmarshal(data, resp)
		}
		return data, nil
	default:
		info := utils.Json.Get(data, "inf").ToString()
		if info == "" {
			info = utils.Json.Get(data, "info").ToString()
		}
		return nil, fmt.Errorf(info)
	}
}

func (d *LanZou) request(url string, method string, callback base.ReqCallback, up bool) ([]byte, error) {
	var req *resty.Request
	if up {
		req = upClient.R()
	} else {
		req = base.RestyClient.R()
	}

	req.SetHeaders(map[string]string{
		"Referer": "https://pc.woozooo.com",
	})

	if d.Cookie != "" {
		req.SetHeader("cookie", d.Cookie)
	}

	if callback != nil {
		callback(req)
	}

	res, err := req.Execute(method, url)
	if err != nil {
		return nil, err
	}

	return res.Body(), err
}

/*
通过cookie获取数据
*/

// 获取文件和文件夹,获取到的文件大小、更改时间不可信
func (d *LanZou) GetFiles(ctx context.Context, folderID string) ([]model.Obj, error) {
	folders, err := d.getFolders(ctx, folderID)
	if err != nil {
		return nil, err
	}
	files, err := d.getFiles(ctx, folderID)
	if err != nil {
		return nil, err
	}
	objs := make([]model.Obj, 0, len(folders)+len(files))
	for _, folder := range folders {
		objs = append(objs, folder.ToObj())
	}

	for _, file := range files {
		objs = append(objs, file.ToObj())
	}
	return objs, nil
}

// 通过ID获取文件夹
func (d *LanZou) getFolders(ctx context.Context, folderID string) ([]FileOrFolder, error) {
	var resp FilesOrFoldersResp
	_, err := d.post(d.BaseUrl+"/doupload.php", func(req *resty.Request) {
		req.SetContext(ctx)
		req.SetFormData(map[string]string{
			"task":      "47",
			"folder_id": folderID,
		})
	}, &resp)
	if err != nil {
		return nil, err
	}
	return resp.Text, nil
}

// 通过ID获取文件
func (d *LanZou) getFiles(ctx context.Context, folderID string) ([]FileOrFolder, error) {
	files := make([]FileOrFolder, 0)
	for pg := 1; ; pg++ {
		var resp FilesOrFoldersResp
		_, err := d.post(d.BaseUrl+"/doupload.php", func(req *resty.Request) {
			req.SetContext(ctx)
			req.SetFormData(map[string]string{
				"task":      "5",
				"folder_id": folderID,
				"pg":        strconv.Itoa(pg),
			})
		}, &resp)
		if err != nil {
			return nil, err
		}
		if len(resp.Text) == 0 {
			break
		}
		files = append(files, resp.Text...)
	}
	return files, nil
}

// 通过ID获取文件夹分享地址
func (d *LanZou) getFolderShareUrlByID(ctx context.Context, fileID string) (share FileShare, err error) {
	var resp FileShareResp
	_, err = d.post(d.BaseUrl+"/doupload.php", func(req *resty.Request) {
		req.SetContext(ctx)
		req.SetFormData(map[string]string{
			"task":    "18",
			"file_id": fileID,
		})
	}, &resp)
	if err != nil {
		return
	}
	share = resp.Info
	return
}

// 通过ID获取文件分享地址
func (d *LanZou) getFileShareUrlByID(ctx context.Context, fileID string) (share FileShare, err error) {
	var resp FileShareResp
	_, err = d.post(d.BaseUrl+"/doupload.php", func(req *resty.Request) {
		req.SetContext(ctx)
		req.SetFormData(map[string]string{
			"task":    "22",
			"file_id": fileID,
		})
	}, &resp)
	if err != nil {
		return
	}
	share = resp.Info
	return
}

/*
通过分享链接获取数据
*/

// 判断类容
var isFileReg = regexp.MustCompile(`class="fileinfo"|id="file"|文件描述`)
var isFolderReg = regexp.MustCompile(`id="infos"`)

// 获取文件文件夹基础信息
var nameFindReg = regexp.MustCompile(`<title>(.+?) - 蓝奏云</title>|id="filenajax">(.+?)</div>|var filename = '(.+?)';|<div style="font-size.+?>([^<>].+?)</div>|<div class="filethetext".+?>([^<>]+?)</div>`)
var sizeFindReg = regexp.MustCompile(`(?i)大小\W*([0-9.]+\s*[bkm]+)`)
var timeFindReg = regexp.MustCompile(`\d+\s*[秒天分小][钟时]?前|[昨前]天|\d{4}-\d{2}-\d{2}`)

var findSubFolaerReg = regexp.MustCompile(`(folderlink|mbxfolder).+href="/(.+?)"(.+filename")?>(.+?)<`) // 查找分享文件夹子文件夹ID和名称

// 获取关键数据
var findDownPageParamReg = regexp.MustCompile(`<iframe.*?src="(.+?)"`)

// 通过分享链接获取文件或文件夹,如果是文件则会返回下载链接
func (d *LanZou) GetFileOrFolderByShareUrl(ctx context.Context, downID, pwd string) ([]model.Obj, error) {
	pageData, err := d.get(fmt.Sprint(d.ShareUrl, "/", downID), func(req *resty.Request) { req.SetContext(ctx) }, nil)
	if err != nil {
		return nil, err
	}
	pageData = RemoveNotes(pageData)

	var objs []model.Obj
	if !isFileReg.Match(pageData) {
		files, err := d.getFolderByShareUrl(ctx, downID, pwd, pageData)
		if err != nil {
			return nil, err
		}
		objs = make([]model.Obj, 0, len(files))
		for _, file := range files {
			objs = append(objs, file.ToObj())
		}
	} else {
		file, err := d.getFilesByShareUrl(ctx, downID, pwd, pageData)
		if err != nil {
			return nil, err
		}
		objs = []model.Obj{file.ToObj()}
	}
	return objs, nil
}

// 通过分享链接获取文件(下载链接也使用此方法)
// 参考 https://github.com/zaxtyson/LanZouCloud-API/blob/ab2e9ec715d1919bf432210fc16b91c6775fbb99/lanzou/api/core.py#L440
func (d *LanZou) getFilesByShareUrl(ctx context.Context, downID, pwd string, firstPageData []byte) (file FileInfoAndUrlByShareUrl, err error) {
	if firstPageData == nil {
		firstPageData, err = d.get(fmt.Sprint(d.ShareUrl, "/", downID), func(req *resty.Request) { req.SetContext(ctx) }, nil)
		if err != nil {
			return
		}
		firstPageData = RemoveNotes(firstPageData)
	}
	firstPageDataStr := string(firstPageData)

	if strings.Contains(firstPageDataStr, "acw_sc__v2") {
		var vs string
		if vs, err = CalcAcwScV2(firstPageDataStr); err != nil {
			return
		}
		firstPageData, err = d.get(fmt.Sprint(d.ShareUrl, "/", downID), func(req *resty.Request) {
			req.SetCookie(&http.Cookie{
				Name:  "acw_sc__v2",
				Value: vs,
			})
			req.SetContext(ctx)
		}, nil)
		if err != nil {
			return
		}
		firstPageData = RemoveNotes(firstPageData)
		firstPageDataStr = string(firstPageData)
	}

	var (
		param       map[string]string
		downloadUrl string
		baseUrl     string
	)

	// 需要密码
	if strings.Contains(firstPageDataStr, "pwdload") || strings.Contains(firstPageDataStr, "passwddiv") {
		param, err = htmlFormToMap(firstPageDataStr)
		if err != nil {
			return
		}
		param["p"] = pwd
		var resp FileShareInfoAndUrlResp[string]
		_, err = d.post(d.ShareUrl+"/ajaxm.php", func(req *resty.Request) { req.SetFormData(param).SetContext(ctx) }, &resp)
		if err != nil {
			return
		}
		file.Name = resp.Inf
		baseUrl = resp.GetBaseUrl()
		downloadUrl = resp.GetDownloadUrl()
	} else {
		urlpaths := findDownPageParamReg.FindStringSubmatch(firstPageDataStr)
		if len(urlpaths) != 2 {
			err = fmt.Errorf("not find file page param")
			return
		}
		var nextPageData []byte
		nextPageData, err = d.get(fmt.Sprint(d.ShareUrl, urlpaths[1]), func(req *resty.Request) { req.SetContext(ctx) }, nil)
		if err != nil {
			return
		}
		nextPageData = RemoveNotes(nextPageData)
		nextPageDataStr := string(nextPageData)

		param, err = htmlJsonToMap(nextPageDataStr)
		if err != nil {
			return
		}

		var resp FileShareInfoAndUrlResp[int]
		_, err = d.post(d.ShareUrl+"/ajaxm.php", func(req *resty.Request) { req.SetFormData(param).SetContext(ctx) }, &resp)
		if err != nil {
			return
		}
		baseUrl = resp.GetBaseUrl()
		downloadUrl = resp.GetDownloadUrl()

		names := nameFindReg.FindStringSubmatch(firstPageDataStr)
		if len(names) > 1 {
			for _, name := range names[1:] {
				if name != "" {
					file.Name = name
					break
				}
			}
		}
	}

	sizes := sizeFindReg.FindStringSubmatch(firstPageDataStr)
	if len(sizes) == 2 {
		file.Size = sizes[1]
	}
	file.ID = downID
	file.Time = timeFindReg.FindString(firstPageDataStr)

	// 重定向获取真实链接
	res, err := base.NoRedirectClient.R().SetHeaders(map[string]string{
		"accept-language": "zh-CN,zh;q=0.9,en;q=0.8,en-GB;q=0.7,en-US;q=0.6",
	}).SetContext(ctx).Get(downloadUrl)
	if err != nil {
		return
	}

	file.Url = res.Header().Get("location")

	// 触发验证
	rPageDataStr := res.String()
	if res.StatusCode() != 302 && strings.Contains(rPageDataStr, "网络异常") {
		param, err = htmlJsonToMap(rPageDataStr)
		if err != nil {
			return
		}
		param["el"] = "2"
		time.Sleep(time.Second * 2)

		// 通过验证获取直连
		var rUrl struct {
			Url string `json:"url"`
		}
		_, err = d.post(fmt.Sprint(baseUrl, "/ajax.php"), func(req *resty.Request) { req.SetContext(ctx).SetFormData(param) }, &rUrl)
		if err != nil {
			return
		}
		file.Url = rUrl.Url
	}
	return
}

// 通过分享链接获取文件夹
// 参考 https://github.com/zaxtyson/LanZouCloud-API/blob/ab2e9ec715d1919bf432210fc16b91c6775fbb99/lanzou/api/core.py#L1089
func (d *LanZou) getFolderByShareUrl(ctx context.Context, downID, pwd string, firstPageData []byte) ([]FileOrFolderByShareUrl, error) {
	if firstPageData == nil {
		var err error
		firstPageData, err = d.get(fmt.Sprint(d.ShareUrl, "/", downID), func(req *resty.Request) { req.SetContext(ctx) }, nil)
		if err != nil {
			return nil, err
		}
		firstPageData = RemoveNotes(firstPageData)
	}
	firstPageDataStr := string(firstPageData)

	//
	if strings.Contains(firstPageDataStr, "acw_sc__v2") {
		vs, err := CalcAcwScV2(firstPageDataStr)
		if err != nil {
			return nil, err
		}
		firstPageData, err = d.get(fmt.Sprint(d.ShareUrl, "/", downID), func(req *resty.Request) {
			req.SetCookie(&http.Cookie{
				Name:  "acw_sc__v2",
				Value: vs,
			})
			req.SetContext(ctx)
		}, nil)
		if err != nil {
			return nil, err
		}
		firstPageData = RemoveNotes(firstPageData)
		firstPageDataStr = string(firstPageData)
	}

	from, err := htmlJsonToMap(firstPageDataStr)
	if err != nil {
		return nil, err
	}
	from["pwd"] = pwd

	files := make([]FileOrFolderByShareUrl, 0)
	// vip获取文件夹
	floders := findSubFolaerReg.FindAllStringSubmatch(firstPageDataStr, -1)
	for _, floder := range floders {
		if len(floder) == 5 {
			files = append(files, FileOrFolderByShareUrl{
				ID:       floder[2],
				NameAll:  floder[4],
				IsFloder: true,
			})
		}
	}

	for page := 1; ; page++ {
		from["pg"] = strconv.Itoa(page)
		var resp FileOrFolderByShareUrlResp
		_, err := d.post(d.ShareUrl+"/filemoreajax.php", func(req *resty.Request) { req.SetFormData(from).SetContext(ctx) }, &resp)
		if err != nil {
			return nil, err
		}
		files = append(files, resp.Text...)
		if len(resp.Text) == 0 {
			break
		}
		time.Sleep(time.Millisecond * 600)
	}
	return files, nil
}
