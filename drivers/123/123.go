package _23

import (
	"errors"
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/Xhofe/alist/drivers/base"
	"github.com/Xhofe/alist/model"
	"github.com/Xhofe/alist/utils"
	"github.com/go-resty/resty/v2"
	jsoniter "github.com/json-iterator/go"
	log "github.com/sirupsen/logrus"
)

func (driver Pan123) Login(account *model.Account) error {
	url := "https://www.123pan.com/api/user/sign_in"
	if account.APIProxyUrl != "" {
		url = fmt.Sprintf("%s/%s", account.APIProxyUrl, url)
	}
	var resp Pan123TokenResp
	_, err := base.RestyClient.R().
		SetResult(&resp).
		SetBody(base.Json{
			"passport": account.Username,
			"password": account.Password,
		}).Post(url)
	if err != nil {
		return err
	}
	if resp.Code != 200 {
		err = fmt.Errorf(resp.Message)
		account.Status = resp.Message
	} else {
		account.Status = "work"
		account.AccessToken = resp.Data.Token
	}
	_ = model.SaveAccount(account)
	return err
}

func (driver Pan123) FormatFile(file *File) *model.File {
	f := &model.File{
		Id:        strconv.FormatInt(file.FileId, 10),
		Name:      file.FileName,
		Size:      file.Size,
		Driver:    driver.Config().Name,
		UpdatedAt: file.UpdateAt,
		Thumbnail: file.DownloadUrl,
	}
	f.Type = file.GetType()
	return f
}

func (driver Pan123) GetFiles(parentId string, account *model.Account) ([]File, error) {
	next := "0"
	res := make([]File, 0)
	for next != "-1" {
		var resp Pan123Files
		query := map[string]string{
			"driveId":        "0",
			"limit":          "100",
			"next":           next,
			"orderBy":        account.OrderBy,
			"orderDirection": account.OrderDirection,
			"parentFileId":   parentId,
			"trashed":        "false",
		}
		_, err := driver.Request("https://www.123pan.com/api/file/list/new",
			base.Get, nil, query, nil, &resp, false, account)
		if err != nil {
			return nil, err
		}
		next = resp.Data.Next
		res = append(res, resp.Data.InfoList...)
	}
	return res, nil
}

func (driver Pan123) Request(url string, method int, headers, query map[string]string, data *base.Json, resp interface{}, proxy bool, account *model.Account) ([]byte, error) {
	rawUrl := url
	if account.APIProxyUrl != "" && proxy {
		url = fmt.Sprintf("%s/%s", account.APIProxyUrl, url)
	}
	req := base.RestyClient.R()
	req.SetHeader("Authorization", "Bearer "+account.AccessToken)
	if headers != nil {
		req.SetHeaders(headers)
	}
	if query != nil {
		req.SetQueryParams(query)
	}
	if data != nil {
		req.SetBody(data)
	}
	if resp != nil {
		req.SetResult(resp)
	}
	var res *resty.Response
	var err error
	switch method {
	case base.Get:
		res, err = req.Get(url)
	case base.Post:
		res, err = req.Post(url)
	default:
		return nil, base.ErrNotSupport
	}
	if err != nil {
		return nil, err
	}
	log.Debug(res.String())
	body := res.Body()
	code := jsoniter.Get(body, "code").ToInt()
	if code != 0 {
		if code == 401 {
			err := driver.Login(account)
			if err != nil {
				return nil, err
			}
			return driver.Request(rawUrl, method, headers, query, data, resp, proxy, account)
		}
		return nil, errors.New(jsoniter.Get(body, "message").ToString())
	}
	return body, nil
}

//func (driver Pan123) Post(url string, data base.Json, account *model.Account) ([]byte, error) {
//	res, err := pan123Client.R().
//		SetHeader("authorization", "Bearer "+account.AccessToken).
//		SetBody(data).Post(url)
//	if err != nil {
//		return nil, err
//	}
//	body := res.Body()
//	if jsoniter.Get(body, "code").ToInt() != 0 {
//		return nil, errors.New(jsoniter.Get(body, "message").ToString())
//	}
//	return body, nil
//}

func (driver Pan123) GetFile(path string, account *model.Account) (*File, error) {
	dir, name := filepath.Split(path)
	dir = utils.ParsePath(dir)
	_, err := driver.Files(dir, account)
	if err != nil {
		return nil, err
	}
	parentFiles_, _ := base.GetCache(dir, account)
	parentFiles, _ := parentFiles_.([]File)
	for _, file := range parentFiles {
		if file.FileName == name {
			//if file.Type != conf.FOLDER {
			//	return &file, err
			//} else {
			//	return nil, base.ErrNotFile
			//}
			return &file, nil
		}
	}
	return nil, base.ErrPathNotFound
}

//func HMAC(message string, secret string) string {
//	key := []byte(secret)
//	h := hmac.New(sha256.New, key)
//	h.Write([]byte(message))
//	//	fmt.Println(h.Sum(nil))
//	//sha := hex.EncodeToString(h.Sum(nil))
//	//	fmt.Println(sha)
//	//return sha
//	return string(h.Sum(nil))
//}

func init() {
	base.RegisterDriver(&Pan123{})
}
