package yandex

import (
	"errors"
	"fmt"
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/drivers/base"
	"github.com/Xhofe/alist/model"
	"github.com/Xhofe/alist/utils"
	"github.com/go-resty/resty/v2"
	"path"
	"strconv"
)

func (driver Yandex) RefreshToken(account *model.Account) error {
	err := driver.refreshToken(account)
	if err != nil && err == base.ErrEmptyToken {
		err = driver.refreshToken(account)
	}
	if err != nil {
		account.Status = err.Error()
	}
	_ = model.SaveAccount(account)
	return err
}

func (driver Yandex) refreshToken(account *model.Account) error {
	u := "https://oauth.yandex.com/token"
	var resp base.TokenResp
	var e TokenErrResp
	_, err := base.RestyClient.R().SetResult(&resp).SetError(&e).SetFormData(map[string]string{
		"grant_type":    "refresh_token",
		"refresh_token": account.RefreshToken,
		"client_id":     account.ClientId,
		"client_secret": account.ClientSecret,
	}).Post(u)
	if err != nil {
		return err
	}
	if e.Error != "" {
		return fmt.Errorf("%s : %s", e.Error, e.ErrorDescription)
	}
	if resp.RefreshToken == "" {
		return base.ErrEmptyToken
	}
	account.Status = "work"
	account.AccessToken, account.RefreshToken = resp.AccessToken, resp.RefreshToken
	return nil
}

func (driver Yandex) Request(pathname string, method int, headers, query, form map[string]string, data interface{}, resp interface{}, account *model.Account) ([]byte, error) {
	u := "https://cloud-api.yandex.net/v1/disk/resources" + pathname
	req := base.RestyClient.R()
	req.SetHeader("Authorization", "OAuth "+account.AccessToken)
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
	var res *resty.Response
	var err error
	var e ErrResp
	req.SetError(&e)
	switch method {
	case base.Get:
		res, err = req.Get(u)
	case base.Post:
		res, err = req.Post(u)
	case base.Patch:
		res, err = req.Patch(u)
	case base.Delete:
		res, err = req.Delete(u)
	case base.Put:
		res, err = req.Put(u)
	default:
		return nil, base.ErrNotSupport
	}
	if err != nil {
		return nil, err
	}
	//log.Debug(res.String())
	if e.Error != "" {
		if e.Error == "UnauthorizedError" {
			err = driver.RefreshToken(account)
			if err != nil {
				return nil, err
			}
			return driver.Request(pathname, method, headers, query, form, data, resp, account)
		}
		return nil, errors.New(e.Description)
	}
	return res.Body(), nil
}

func (driver Yandex) GetFiles(rawPath string, account *model.Account) ([]model.File, error) {
	path_ := utils.Join(account.RootFolder, rawPath)
	limit := 100
	page := 1
	res := make([]model.File, 0)
	for {
		offset := (page - 1) * limit
		query := map[string]string{
			"path":   path_,
			"limit":  strconv.Itoa(limit),
			"offset": strconv.Itoa(offset),
		}
		if account.OrderBy != "" {
			if account.OrderDirection == "desc" {
				query["sort"] = "-" + account.OrderBy
			} else {
				query["sort"] = account.OrderBy
			}
		}
		var resp FilesResp
		_, err := driver.Request("", base.Get, nil, query, nil, nil, &resp, account)
		if err != nil {
			return nil, err
		}
		for _, file := range resp.Embedded.Items {
			f := model.File{
				Name:      file.Name,
				Size:      file.Size,
				Driver:    driver.Config().Name,
				UpdatedAt: &file.Modified,
				Thumbnail: file.Preview,
				Url:       file.File,
			}
			if file.Type == "dir" {
				f.Type = conf.FOLDER
			} else {
				f.Type = utils.GetFileType(path.Ext(file.Name))
			}
			res = append(res, f)
		}
		if resp.Embedded.Total <= offset+limit {
			break
		}
	}
	return res, nil
}

func init() {
	base.RegisterDriver(&Yandex{})
}
