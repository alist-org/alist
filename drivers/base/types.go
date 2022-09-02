package base

import "github.com/go-resty/resty/v2"

type Json map[string]interface{}

type TokenResp struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type ReqCallback func(req *resty.Request)
