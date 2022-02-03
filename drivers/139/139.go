package _39

import (
	"errors"
	"fmt"
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/drivers/base"
	"github.com/Xhofe/alist/model"
	"github.com/Xhofe/alist/utils"
	"github.com/go-resty/resty/v2"
	jsoniter "github.com/json-iterator/go"
	log "github.com/sirupsen/logrus"
	"path"
	"time"
)

func (driver Cloud139) Request(pathname string, method int, headers, query, form map[string]string, data interface{}, resp interface{}, account *model.Account) ([]byte, error) {
	url := "https://yun.139.com" + pathname
	req := base.RestyClient.R()
	randStr := utils.RandomStr(16)
	ts := time.Now().Format("2006-01-02 15:04:05")
	log.Debugf("%+v", data)
	body, err := utils.Json.Marshal(data)
	if err != nil {
		return nil, err
	}
	sign := calSign(string(body), ts, randStr)
	svcType := "1"
	if isFamily(account) {
		svcType = "2"
	}
	req.SetHeaders(map[string]string{
		"Accept":         "application/json, text/plain, */*",
		"CMS-DEVICE":     "default",
		"Cookie":         account.AccessToken,
		"mcloud-channel": "1000101",
		"mcloud-client":  "10701",
		//"mcloud-route": "001",
		"mcloud-sign": fmt.Sprintf("%s,%s,%s", ts, randStr, sign),
		//"mcloud-skey":"",
		"mcloud-version":      "6.6.0",
		"Origin":              "https://yun.139.com",
		"Referer":             "https://yun.139.com/w/",
		"x-DeviceInfo":        "||9|6.6.0|chrome|95.0.4638.69|uwIy75obnsRPIwlJSd7D9GhUvFwG96ce||macos 10.15.2||zh-CN|||",
		"x-huawei-channelSrc": "10000034",
		"x-inner-ntwk":        "2",
		"x-m4c-caller":        "PC",
		"x-m4c-src":           "10002",
		"x-SvcType":           svcType,
	})
	if headers != nil {
		req.SetHeaders(headers)
	}
	if query != nil {
		req.SetQueryParams(query)
	}
	if form != nil {
		req.SetFormData(form)
	}
	if data != nil {
		req.SetBody(data)
	}
	var e BaseResp
	//var err error
	var res *resty.Response
	req.SetResult(&e)
	switch method {
	case base.Get:
		res, err = req.Get(url)
	case base.Post:
		res, err = req.Post(url)
	case base.Delete:
		res, err = req.Delete(url)
	case base.Patch:
		res, err = req.Patch(url)
	case base.Put:
		res, err = req.Put(url)
	default:
		return nil, base.ErrNotSupport
	}
	if err != nil {
		return nil, err
	}
	log.Debugln(res.String())
	if !e.Success {
		return nil, errors.New(e.Message)
	}
	if resp != nil {
		err = utils.Json.Unmarshal(res.Body(), resp)
		if err != nil {
			return nil, err
		}
	}
	return res.Body(), nil
}

func (driver Cloud139) Post(pathname string, data interface{}, resp interface{}, account *model.Account) ([]byte, error) {
	return driver.Request(pathname, base.Post, nil, nil, nil, data, resp, account)
}

func (driver Cloud139) GetFiles(catalogID string, account *model.Account) ([]model.File, error) {
	start := 0
	limit := 100
	files := make([]model.File, 0)
	for {
		data := base.Json{
			"catalogID":       catalogID,
			"sortDirection":   1,
			"startNumber":     start + 1,
			"endNumber":       start + limit,
			"filterType":      0,
			"catalogSortType": 0,
			"contentSortType": 0,
			"commonAccountInfo": base.Json{
				"account":     account.Username,
				"accountType": 1,
			},
		}
		var resp GetDiskResp
		_, err := driver.Post("/orchestration/personalCloud/catalog/v1.0/getDisk", data, &resp, account)
		if err != nil {
			return nil, err
		}
		for _, catalog := range resp.Data.GetDiskResult.CatalogList {
			f := model.File{
				Id:        catalog.CatalogID,
				Name:      catalog.CatalogName,
				Size:      0,
				Type:      conf.FOLDER,
				Driver:    driver.Config().Name,
				UpdatedAt: getTime(catalog.UpdateTime),
			}
			files = append(files, f)
		}
		for _, content := range resp.Data.GetDiskResult.ContentList {
			f := model.File{
				Id:        content.ContentID,
				Name:      content.ContentName,
				Size:      content.ContentSize,
				Type:      utils.GetFileType(path.Ext(content.ContentName)),
				Driver:    driver.Config().Name,
				UpdatedAt: getTime(content.UpdateTime),
				Thumbnail: content.ThumbnailURL,
				//Thumbnail: content.BigthumbnailURL,
			}
			files = append(files, f)
		}
		if start+limit >= resp.Data.GetDiskResult.NodeCount {
			break
		}
		start += limit
	}
	return files, nil
}

func (driver Cloud139) GetLink(contentId string, account *model.Account) (string, error) {
	data := base.Json{
		"appName":   "",
		"contentID": contentId,
		"commonAccountInfo": base.Json{
			"account":     account.Username,
			"accountType": 1,
		},
	}
	res, err := driver.Post("/orchestration/personalCloud/uploadAndDownload/v1.0/downloadRequest",
		data, nil, account)
	if err != nil {
		return "", err
	}
	return jsoniter.Get(res, "data", "downloadURL").ToString(), nil
}

func init() {
	base.RegisterDriver(&Cloud139{})
}
