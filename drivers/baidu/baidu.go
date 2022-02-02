package baidu

import (
	"fmt"
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/drivers/base"
	"github.com/Xhofe/alist/model"
	"github.com/Xhofe/alist/utils"
	"github.com/go-resty/resty/v2"
	jsoniter "github.com/json-iterator/go"
	"path"
	"strconv"
)

func (driver Baidu) RefreshToken(account *model.Account) error {
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

func (driver Baidu) refreshToken(account *model.Account) error {
	u := "https://openapi.baidu.com/oauth/2.0/token"
	var resp base.TokenResp
	var e TokenErrResp
	_, err := base.RestyClient.R().SetResult(&resp).SetError(&e).SetQueryParams(map[string]string{
		"grant_type":    "refresh_token",
		"refresh_token": account.RefreshToken,
		"client_id":     account.ClientId,
		"client_secret": account.ClientSecret,
	}).Get(u)
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

func (driver Baidu) Request(pathname string, method int, headers, query, form map[string]string, data interface{}, resp interface{}, account *model.Account) ([]byte, error) {
	u := "https://pan.baidu.com/rest/2.0" + pathname
	req := base.RestyClient.R()
	req.SetQueryParam("access_token", account.AccessToken)
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
	errno := jsoniter.Get(res.Body(), "errno").ToInt()
	if errno != 0 {
		if errno == -6 {
			err = driver.RefreshToken(account)
			if err != nil {
				return nil, err
			}
			return driver.Request(pathname, method, headers, query, form, data, resp, account)
		}
		return nil, fmt.Errorf("errno: %d, refer to https://pan.baidu.com/union/doc/", errno)
	}
	return res.Body(), nil
}

func (driver Baidu) Get(pathname string, params map[string]string, resp interface{}, account *model.Account) ([]byte, error) {
	return driver.Request(pathname, base.Get, nil, params, nil, nil, resp, account)
}

func (driver Baidu) Post(pathname string, params map[string]string, data interface{}, resp interface{}, account *model.Account) ([]byte, error) {
	return driver.Request(pathname, base.Post, nil, params, nil, data, resp, account)
}

func (driver Baidu) manage(opera string, filelist interface{}, account *model.Account) ([]byte, error) {
	params := map[string]string{
		"method": "filemanager",
		"opera":  opera,
	}
	marshal, err := utils.Json.Marshal(filelist)
	if err != nil {
		return nil, err
	}
	data := fmt.Sprintf("async=0&filelist=%s&ondup=newcopy", string(marshal))
	return driver.Post("/xpan/file", params, data, nil, account)
}

func (driver Baidu) GetFiles(dir string, account *model.Account) ([]model.File, error) {
	dir = utils.Join(account.RootFolder, dir)
	start := 0
	limit := 200
	params := map[string]string{
		"method": "list",
		"dir":    dir,
		"web":    "web",
	}
	if account.OrderBy != "" {
		params["order"] = account.OrderBy
		if account.OrderDirection == "desc" {
			params["desc"] = "1"
		}
	}
	res := make([]model.File, 0)
	for {
		params["start"] = strconv.Itoa(start)
		params["limit"] = strconv.Itoa(limit)
		start += limit
		var resp ListResp
		_, err := driver.Get("/xpan/file", params, &resp, account)
		if err != nil {
			return nil, err
		}
		if len(resp.List) == 0 {
			break
		}
		for _, f := range resp.List {
			file := model.File{
				Id:        strconv.FormatInt(f.FsId, 10),
				Name:      f.ServerFilename,
				Size:      f.Size,
				Driver:    driver.Config().Name,
				UpdatedAt: getTime(f.ServerMtime),
				Thumbnail: f.Thumbs.Url3,
			}
			if f.Isdir == 1 {
				file.Type = conf.FOLDER
			} else {
				file.Type = utils.GetFileType(path.Ext(f.ServerFilename))
			}
			res = append(res, file)
		}
	}
	return res, nil
}

func (driver Baidu) create(path string, size uint64, isdir int, uploadid, block_list string, account *model.Account) ([]byte, error) {
	params := map[string]string{
		"method": "create",
	}
	data := fmt.Sprintf("path=%s&size=%d&isdir=%d", path, size, isdir)
	if uploadid != "" {
		data += fmt.Sprintf("&uploadid=%s&block_list=%s", uploadid, block_list)
	}
	return driver.Post("/xpan/file", params, data, nil, account)
}

func init() {
	base.RegisterDriver(&Baidu{})
}
