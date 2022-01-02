package alist

import (
	"errors"
	"github.com/Xhofe/alist/drivers/base"
	"github.com/Xhofe/alist/model"
)

type BaseResp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type PathResp struct {
	BaseResp
	Data struct {
		Type string `json:"type"`
		//Meta  Meta         `json:"meta"`
		Files []model.File `json:"files"`
	} `json:"data"`
}

type PreviewResp struct {
	BaseResp
	Data interface{} `json:"data"`
}

func (driver *Alist) Login(account *model.Account) error {
	var resp BaseResp
	_, err := base.RestyClient.R().SetResult(&resp).
		SetHeader("Authorization", account.AccessToken).
		Get(account.SiteUrl + "/api/admin/login")
	if err != nil {
		return err
	}
	if resp.Code != 200 {
		return errors.New(resp.Message)
	}
	return nil
}

func init() {
	base.RegisterDriver(&Alist{})
}
