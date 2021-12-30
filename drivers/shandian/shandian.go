package shandian

import (
	"errors"
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/drivers/base"
	"github.com/Xhofe/alist/model"
	"github.com/Xhofe/alist/utils"
	log "github.com/sirupsen/logrus"
	"strconv"
	"time"
)

type Resp struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

type LoginResp struct {
	Resp
	Data struct {
		Token string `json:"token"`
	} `json:"data"`
}

func (driver Shandian) Login(account *model.Account) error {
	var resp LoginResp
	_, err := base.RestyClient.R().SetResult(&resp).SetHeader("Accept", "application/json").SetBody(base.Json{
		"mobile":   account.Username,
		"password": account.Password,
		"smscode":  "",
	}).Post("https://shandianpan.com/api/login")
	if err != nil {
		return err
	}
	if resp.Code != 0 {
		account.Status = resp.Msg
		err = errors.New(resp.Msg)
	} else {
		account.Status = "work"
		account.AccessToken = resp.Data.Token
	}
	_ = model.SaveAccount(account)
	return err
}

type File struct {
	Id         int64  `json:"id"`
	Type       int    `json:"type"`
	Name       string `json:"name"`
	UpdateTime int64  `json:"update_time"`
	Size       int64  `json:"size"`
	Ext        string `json:"ext"`
}

func (driver Shandian) FormatFile(file *File) *model.File {
	t := time.Unix(file.UpdateTime, 0)
	f := &model.File{
		Id:        strconv.FormatInt(file.Id, 10),
		Name:      file.Name,
		Size:      file.Size,
		Driver:    driver.Config().Name,
		UpdatedAt: &t,
	}
	if file.Type == 1 {
		f.Type = conf.FOLDER
	} else {
		f.Type = utils.GetFileType(file.Ext)
		if file.Ext != "" {
			f.Name += "." + file.Ext
		}
	}
	return f
}

func (driver Shandian) Post(url string, data map[string]interface{}, resp interface{}, account *model.Account) ([]byte, error) {
	req := base.RestyClient.R()
	req.SetHeader("Accept", "application/json")
	data["token"] = account.AccessToken
	req.SetBody(data)
	var e Resp
	if resp != nil {
		req.SetResult(resp)
	} else {
		req.SetResult(&e)
	}
	req.SetError(&e)
	res, err := req.Post(url)
	if err != nil {
		return nil, err
	}
	log.Debug(res.String())
	if e.Code != 0 {
		if e.Code == 10 {
			err = driver.Login(account)
			if err != nil {
				return nil, err
			}
			return driver.Post(url, data, resp, account)
		}
		return nil, errors.New(e.Msg)
	}
	return res.Body(), nil
}

type FilesResp struct {
	Resp
	Data []File `json:"data"`
}

func (driver Shandian) GetFiles(id string, account *model.Account) ([]File, error) {
	// TODO page not wok
	//res := make([]File, 0)
	page := 1
	//for {
	data := map[string]interface{}{
		"id":        id,
		"page":      page,
		"page_size": 100,
	}
	var resp FilesResp
	_, err := driver.Post("https://shandianpan.com/api/pan", data, &resp, account)
	if err != nil {
		return nil, err
	}
	//res = append(res, resp.Data...)
	//	if len(resp.Data) == 0 {
	//		break
	//	}
	//}
	//return res, nil
	return resp.Data, nil
}

type UploadResp struct {
	Resp
	Data struct {
		Accessid  string `json:"accessid"`
		Policy    string `json:"policy"`
		Expire    int    `json:"expire"`
		Callback  string `json:"callback"`
		Key       string `json:"key"`
		Host      string `json:"host"`
		Signature string `json:"signature"`
	} `json:"data"`
}

func init() {
	base.RegisterDriver(&Shandian{})
}
