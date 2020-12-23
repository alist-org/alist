package alidrive

import (
	"encoding/json"
	"fmt"
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/utils"
	log "github.com/sirupsen/logrus"
)

func TokenLogin() (*TokenLoginResp, error) {
	log.Infof("尝试使用token登录...")
	url:="https://auth.aliyundrive.com/v2/oauth/token_login"
	req:=TokenLoginReq{Token:conf.Conf.AliDrive.LoginToken}
	log.Debugf("token_login_req:%v",req)
	var tokenLogin TokenLoginResp
	if body, err := DoPost(url, req,false); err != nil {
		log.Errorf("tokenLogin-doPost出错:%s",err.Error())
		return nil,err
	}else {
		if err = json.Unmarshal(body,&tokenLogin);err!=nil {
			log.Errorf("解析json[%s]出错:%s",string(body),err.Error())
			return nil,err
		}
	}
	if tokenLogin.IsAvailable() {
		return &tokenLogin,nil
	}
	return nil,fmt.Errorf("登录token失效,请更换:%s",tokenLogin.Message)
}

func GetToken(tokenLogin *TokenLoginResp) (*TokenResp,error) {
	log.Infof("获取API token...")
	url:="https://websv.aliyundrive.com/token/get"
	code:=utils.GetCode(tokenLogin.Goto)
	if code == "" {
		return nil,fmt.Errorf("获取code出错")
	}
	req:=GetTokenReq{Code:code}
	var token TokenResp
	if body, err := DoPost(url, req,false); err != nil {
		log.Errorf("tokenLogin-doPost出错:%s",err.Error())
		return nil,err
	}else {
		if err = json.Unmarshal(body,&token);err!=nil {
			log.Errorf("解析json[%s]出错:%s",string(body),err.Error())
			log.Errorf("此处json解析失败应该是code失效")
			return nil,fmt.Errorf("code失效")
		}
	}
	return &token,nil
}

func RefreshToken() bool {
	log.Infof("刷新token...")
	url:="https://websv.aliyundrive.com/token/refresh"
	req:=RefreshTokenReq{RefreshToken:conf.Conf.AliDrive.RefreshToken}
	var token TokenResp
	if body, err := DoPost(url, req,false); err != nil {
		log.Errorf("tokenLogin-doPost出错:%s",err.Error())
		return false
	}else {
		if err = json.Unmarshal(body,&token);err!=nil {
			log.Errorf("解析json[%s]出错:%s",string(body),err.Error())
			log.Errorf("此处json解析失败应该是refresh_token失效")
			return false
		}
	}
	//刷新成功 更新token
	conf.Conf.AliDrive.AccessToken=token.AccessToken
	conf.Conf.AliDrive.RefreshToken=token.RefreshToken
	conf.Authorization=token.TokenType+"\t"+token.AccessToken
	return true
}
