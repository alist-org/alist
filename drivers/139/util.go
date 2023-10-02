package _139

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/alist-org/alist/v3/drivers/base"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/alist-org/alist/v3/pkg/utils/random"
	"github.com/go-resty/resty/v2"
	jsoniter "github.com/json-iterator/go"
	log "github.com/sirupsen/logrus"
)

// do others that not defined in Driver interface
func (d *Yun139) isFamily() bool {
	return d.Type == "family"
}

func encodeURIComponent(str string) string {
	r := url.QueryEscape(str)
	r = strings.Replace(r, "+", "%20", -1)
	r = strings.Replace(r, "%21", "!", -1)
	r = strings.Replace(r, "%27", "'", -1)
	r = strings.Replace(r, "%28", "(", -1)
	r = strings.Replace(r, "%29", ")", -1)
	r = strings.Replace(r, "%2A", "*", -1)
	return r
}

func calSign(body, ts, randStr string) string {
	body = encodeURIComponent(body)
	strs := strings.Split(body, "")
	sort.Strings(strs)
	body = strings.Join(strs, "")
	body = base64.StdEncoding.EncodeToString([]byte(body))
	res := utils.GetMD5EncodeStr(body) + utils.GetMD5EncodeStr(ts+":"+randStr)
	res = strings.ToUpper(utils.GetMD5EncodeStr(res))
	return res
}

func getTime(t string) time.Time {
	stamp, _ := time.ParseInLocation("20060102150405", t, utils.CNLoc)
	return stamp
}

func (d *Yun139) request(pathname string, method string, callback base.ReqCallback, resp interface{}) ([]byte, error) {
	url := "https://yun.139.com" + pathname
	req := base.RestyClient.R()
	randStr := random.String(16)
	ts := time.Now().Format("2006-01-02 15:04:05")
	if callback != nil {
		callback(req)
	}
	body, err := utils.Json.Marshal(req.Body)
	if err != nil {
		return nil, err
	}
	sign := calSign(string(body), ts, randStr)
	svcType := "1"
	if d.isFamily() {
		svcType = "2"
	}
	req.SetHeaders(map[string]string{
		"Accept":         "application/json, text/plain, */*",
		"CMS-DEVICE":     "default",
		"Authorization":  "Basic " + d.Authorization,
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

	var e BaseResp
	req.SetResult(&e)
	res, err := req.Execute(method, url)
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
func (d *Yun139) post(pathname string, data interface{}, resp interface{}) ([]byte, error) {
	return d.request(pathname, http.MethodPost, func(req *resty.Request) {
		req.SetBody(data)
	}, resp)
}

func (d *Yun139) getFiles(catalogID string) ([]model.Obj, error) {
	start := 0
	limit := 100
	files := make([]model.Obj, 0)
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
				"account":     d.Account,
				"accountType": 1,
			},
		}
		var resp GetDiskResp
		_, err := d.post("/orchestration/personalCloud/catalog/v1.0/getDisk", data, &resp)
		if err != nil {
			return nil, err
		}
		for _, catalog := range resp.Data.GetDiskResult.CatalogList {
			f := model.Object{
				ID:       catalog.CatalogID,
				Name:     catalog.CatalogName,
				Size:     0,
				Modified: getTime(catalog.UpdateTime),
				Ctime:    getTime(catalog.CreateTime),
				IsFolder: true,
			}
			files = append(files, &f)
		}
		for _, content := range resp.Data.GetDiskResult.ContentList {
			f := model.ObjThumb{
				Object: model.Object{
					ID:       content.ContentID,
					Name:     content.ContentName,
					Size:     content.ContentSize,
					Modified: getTime(content.UpdateTime),
					HashInfo: utils.NewHashInfo(utils.MD5, content.Digest),
				},
				Thumbnail: model.Thumbnail{Thumbnail: content.ThumbnailURL},
				//Thumbnail: content.BigthumbnailURL,
			}
			files = append(files, &f)
		}
		if start+limit >= resp.Data.GetDiskResult.NodeCount {
			break
		}
		start += limit
	}
	return files, nil
}

func (d *Yun139) newJson(data map[string]interface{}) base.Json {
	common := map[string]interface{}{
		"catalogType": 3,
		"cloudID":     d.CloudID,
		"cloudType":   1,
		"commonAccountInfo": base.Json{
			"account":     d.Account,
			"accountType": 1,
		},
	}
	return utils.MergeMap(data, common)
}

func (d *Yun139) familyGetFiles(catalogID string) ([]model.Obj, error) {
	pageNum := 1
	files := make([]model.Obj, 0)
	for {
		data := d.newJson(base.Json{
			"catalogID":       catalogID,
			"contentSortType": 0,
			"pageInfo": base.Json{
				"pageNum":  pageNum,
				"pageSize": 100,
			},
			"sortDirection": 1,
		})
		var resp QueryContentListResp
		_, err := d.post("/orchestration/familyCloud/content/v1.0/queryContentList", data, &resp)
		if err != nil {
			return nil, err
		}
		for _, catalog := range resp.Data.CloudCatalogList {
			f := model.Object{
				ID:       catalog.CatalogID,
				Name:     catalog.CatalogName,
				Size:     0,
				IsFolder: true,
				Modified: getTime(catalog.LastUpdateTime),
				Ctime:    getTime(catalog.CreateTime),
			}
			files = append(files, &f)
		}
		for _, content := range resp.Data.CloudContentList {
			f := model.ObjThumb{
				Object: model.Object{
					ID:       content.ContentID,
					Name:     content.ContentName,
					Size:     content.ContentSize,
					Modified: getTime(content.LastUpdateTime),
					Ctime:    getTime(content.CreateTime),
				},
				Thumbnail: model.Thumbnail{Thumbnail: content.ThumbnailURL},
				//Thumbnail: content.BigthumbnailURL,
			}
			files = append(files, &f)
		}
		if 100*pageNum > resp.Data.TotalCount {
			break
		}
		pageNum++
	}
	return files, nil
}

func (d *Yun139) getLink(contentId string) (string, error) {
	data := base.Json{
		"appName":   "",
		"contentID": contentId,
		"commonAccountInfo": base.Json{
			"account":     d.Account,
			"accountType": 1,
		},
	}
	res, err := d.post("/orchestration/personalCloud/uploadAndDownload/v1.0/downloadRequest",
		data, nil)
	if err != nil {
		return "", err
	}
	return jsoniter.Get(res, "data", "downloadURL").ToString(), nil
}

func unicode(str string) string {
	textQuoted := strconv.QuoteToASCII(str)
	textUnquoted := textQuoted[1 : len(textQuoted)-1]
	return textUnquoted
}
