package template

import (
	"errors"
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/drivers/base"
	"github.com/Xhofe/alist/model"
	"github.com/Xhofe/alist/utils"
	"github.com/go-resty/resty/v2"
	"path"
)

func LoginOrRefreshToken(account *model.Account) error {
	return nil
}

func Request(u string, method int, headers, query, form map[string]string, data interface{}, resp interface{}, account *model.Account) ([]byte, error) {
	req := base.RestyClient.R()
	req.SetHeaders(map[string]string{
		"Authorization": "Bearer" + account.AccessToken,
		"Accept":        "application/json, text/plain, */*",
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
	if e.Code >= 400 {
		if e.Code == 401 {
			err = LoginOrRefreshToken(account)
			if err != nil {
				return nil, err
			}
			return Request(u, method, headers, query, form, data, resp, account)
		}
		return nil, errors.New(e.Message)
	}
	return res.Body(), nil
}

func (driver ChaoxingDrive) formatFile(f *File) *model.File {
	file := model.File{
		Id:        f.Id,
		Name:      f.FileName,
		Size:      f.Size,
		Driver:    driver.Config().Name,
		UpdatedAt: f.UpdatedAt,
	}
	if f.File {
		file.Type = utils.GetFileType(path.Ext(f.FileName))
	} else {
		file.Type = conf.FOLDER
	}
	return &file
}

func init() {
	base.RegisterDriver(&ChaoxingDrive{})
}
