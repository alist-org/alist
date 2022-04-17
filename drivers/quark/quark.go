package quark

import (
	"crypto/md5"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/Xhofe/alist/drivers/base"
	"github.com/Xhofe/alist/model"
	"github.com/Xhofe/alist/utils"
	"github.com/Xhofe/alist/utils/cookie"
	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func (driver Quark) Request(pathname string, method int, headers, query, form map[string]string, data interface{}, resp interface{}, account *model.Account) ([]byte, error) {
	u := "https://drive.quark.cn/1/clouddrive" + pathname
	req := base.RestyClient.R()
	req.SetHeaders(map[string]string{
		"Cookie":  account.AccessToken,
		"Accept":  "application/json, text/plain, */*",
		"Referer": "https://pan.quark.cn/",
	})
	req.SetQueryParam("pr", "ucpro")
	req.SetQueryParam("fr", "pc")
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
	if resp != nil {
		req.SetResult(resp)
	}
	var e Resp
	var err error
	var res *resty.Response
	req.SetError(&e)
	switch method {
	case base.Get:
		res, err = req.Get(u)
	case base.Post:
		res, err = req.Post(u)
	case base.Delete:
		res, err = req.Delete(u)
	case base.Patch:
		res, err = req.Patch(u)
	case base.Put:
		res, err = req.Put(u)
	default:
		return nil, base.ErrNotSupport
	}
	if err != nil {
		return nil, err
	}
	__puus := cookie.GetCookie(res.Cookies(), "__puus")
	if __puus != nil {
		account.AccessToken = cookie.SetStr(account.AccessToken, "__puus", __puus.Value)
		_ = model.SaveAccount(account)
	}
	//log.Debugf("%s response: %s", pathname, res.String())
	if e.Status >= 400 || e.Code != 0 {
		return nil, errors.New(e.Message)
	}
	return res.Body(), nil
}

func (driver Quark) Get(pathname string, query map[string]string, resp interface{}, account *model.Account) ([]byte, error) {
	return driver.Request(pathname, base.Get, nil, query, nil, nil, resp, account)
}

func (driver Quark) Post(pathname string, data interface{}, resp interface{}, account *model.Account) ([]byte, error) {
	return driver.Request(pathname, base.Post, nil, nil, nil, data, resp, account)
}

func (driver Quark) GetFiles(parent string, account *model.Account) ([]model.File, error) {
	files := make([]model.File, 0)
	page := 1
	size := 100
	query := map[string]string{
		"pdir_fid":     parent,
		"_size":        strconv.Itoa(size),
		"_fetch_total": "1",
		"_sort":        "file_type:asc," + account.OrderBy + ":" + account.OrderDirection,
	}
	for {
		query["_page"] = strconv.Itoa(page)
		var resp SortResp
		_, err := driver.Get("/file/sort", query, &resp, account)
		if err != nil {
			return nil, err
		}
		for _, f := range resp.Data.List {
			files = append(files, *driver.formatFile(&f))
		}
		if page*size >= resp.Metadata.Total {
			break
		}
		page++
	}
	return files, nil
}

func (driver Quark) UpPre(file *model.FileStream, parentId string, account *model.Account) (UpPreResp, error) {
	now := time.Now()
	data := base.Json{
		"ccp_hash_update": true,
		"dir_name":        "",
		"file_name":       file.Name,
		"format_type":     file.MIMEType,
		"l_created_at":    now.UnixMilli(),
		"l_updated_at":    now.UnixMilli(),
		"pdir_fid":        parentId,
		"size":            file.Size,
	}
	log.Debugf("uppre data: %+v", data)
	var resp UpPreResp
	_, err := driver.Post("/file/upload/pre", data, &resp, account)
	return resp, err
}

func (driver Quark) UpHash(md5, sha1, taskId string, account *model.Account) (bool, error) {
	data := base.Json{
		"md5":     md5,
		"sha1":    sha1,
		"task_id": taskId,
	}
	log.Debugf("hash: %+v", data)
	var resp HashResp
	_, err := driver.Post("/file/update/hash", data, &resp, account)
	return resp.Data.Finish, err
}

func (driver Quark) UpPart(pre UpPreResp, mineType string, partNumber int, bytes []byte, account *model.Account) (string, error) {
	//func (driver Quark) UpPart(pre UpPreResp, mineType string, partNumber int, bytes []byte, account *model.Account, md5Str, sha1Str string) (string, error) {
	timeStr := time.Now().UTC().Format(http.TimeFormat)
	data := base.Json{
		"auth_info": pre.Data.AuthInfo,
		"auth_meta": fmt.Sprintf(`PUT

%s
%s
x-oss-date:%s
x-oss-user-agent:aliyun-sdk-js/6.6.1 Chrome 98.0.4758.80 on Windows 10 64-bit
/%s/%s?partNumber=%d&uploadId=%s`,
			mineType, timeStr, timeStr, pre.Data.Bucket, pre.Data.ObjKey, partNumber, pre.Data.UploadId),
		"task_id": pre.Data.TaskId,
	}
	var resp UpAuthResp
	_, err := driver.Post("/file/upload/auth", data, &resp, account)
	if err != nil {
		return "", err
	}
	//if partNumber == 1 {
	//	finish, err := driver.UpHash(md5Str, sha1Str, pre.Data.TaskId, account)
	//	if err != nil {
	//		return "", err
	//	}
	//	if finish {
	//		return "finish", nil
	//	}
	//}
	u := fmt.Sprintf("https://%s.%s/%s", pre.Data.Bucket, pre.Data.UploadUrl[7:], pre.Data.ObjKey)
	res, err := base.RestyClient.R().
		SetHeaders(map[string]string{
			"Authorization":    resp.Data.AuthKey,
			"Content-Type":     mineType,
			"Referer":          "https://pan.quark.cn/",
			"x-oss-date":       timeStr,
			"x-oss-user-agent": "aliyun-sdk-js/6.6.1 Chrome 98.0.4758.80 on Windows 10 64-bit",
		}).
		SetQueryParams(map[string]string{
			"partNumber": strconv.Itoa(partNumber),
			"uploadId":   pre.Data.UploadId,
		}).SetBody(bytes).Put(u)
	if res.StatusCode() != 200 {
		return "", fmt.Errorf("up status: %d, error: %s", res.StatusCode(), res.String())
	}
	return res.Header().Get("ETag"), nil
}

func (driver Quark) UpCommit(pre UpPreResp, md5s []string, account *model.Account) error {
	timeStr := time.Now().UTC().Format(http.TimeFormat)
	log.Debugf("md5s: %+v", md5s)
	bodyBuilder := strings.Builder{}
	bodyBuilder.WriteString(`<?xml version="1.0" encoding="UTF-8"?>
<CompleteMultipartUpload>
`)
	for i, m := range md5s {
		bodyBuilder.WriteString(fmt.Sprintf(`<Part>
<PartNumber>%d</PartNumber>
<ETag>%s</ETag>
</Part>
`, i+1, m))
	}
	bodyBuilder.WriteString("</CompleteMultipartUpload>")
	body := bodyBuilder.String()
	m := md5.New()
	m.Write([]byte(body))
	contentMd5 := base64.StdEncoding.EncodeToString(m.Sum(nil))
	callbackBytes, err := utils.Json.Marshal(pre.Data.Callback)
	if err != nil {
		return err
	}
	callbackBase64 := base64.StdEncoding.EncodeToString(callbackBytes)
	data := base.Json{
		"auth_info": pre.Data.AuthInfo,
		"auth_meta": fmt.Sprintf(`POST
%s
application/xml
%s
x-oss-callback:%s
x-oss-date:%s
x-oss-user-agent:aliyun-sdk-js/6.6.1 Chrome 98.0.4758.80 on Windows 10 64-bit
/%s/%s?uploadId=%s`,
			contentMd5, timeStr, callbackBase64, timeStr,
			pre.Data.Bucket, pre.Data.ObjKey, pre.Data.UploadId),
		"task_id": pre.Data.TaskId,
	}
	log.Debugf("xml: %s", body)
	log.Debugf("auth data: %+v", data)
	var resp UpAuthResp
	_, err = driver.Post("/file/upload/auth", data, &resp, account)
	if err != nil {
		return err
	}
	u := fmt.Sprintf("https://%s.%s/%s", pre.Data.Bucket, pre.Data.UploadUrl[7:], pre.Data.ObjKey)
	res, err := base.RestyClient.R().
		SetHeaders(map[string]string{
			"Authorization":    resp.Data.AuthKey,
			"Content-MD5":      contentMd5,
			"Content-Type":     "application/xml",
			"Referer":          "https://pan.quark.cn/",
			"x-oss-callback":   callbackBase64,
			"x-oss-date":       timeStr,
			"x-oss-user-agent": "aliyun-sdk-js/6.6.1 Chrome 98.0.4758.80 on Windows 10 64-bit",
		}).
		SetQueryParams(map[string]string{
			"uploadId": pre.Data.UploadId,
		}).SetBody(body).Post(u)
	if res.StatusCode() != 200 {
		return fmt.Errorf("up status: %d, error: %s", res.StatusCode(), res.String())
	}
	return nil
}

func (driver Quark) UpFinish(pre UpPreResp, account *model.Account) error {
	data := base.Json{
		"obj_key": pre.Data.ObjKey,
		"task_id": pre.Data.TaskId,
	}
	_, err := driver.Post("/file/upload/finish", data, nil, account)
	if err != nil {
		return err
	}
	time.Sleep(time.Second)
	return nil
}

func init() {
	base.RegisterDriver(&Quark{})
}
