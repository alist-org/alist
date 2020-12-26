package bootstrap

import (
	"github.com/Xhofe/alist/alidrive"
	"github.com/Xhofe/alist/conf"
	log "github.com/sirupsen/logrus"
)

func InitAliDrive() bool {
	log.Infof("初始化阿里云盘...")
	//首先token_login
	if conf.Conf.AliDrive.RefreshToken == "" {
		tokenLogin,err:=alidrive.TokenLogin()
		if err!=nil {
			log.Errorf("登录失败:%s",err.Error())
			return false
		}
		//然后get_token
		token,err:=alidrive.GetToken(tokenLogin)
		if err!=nil {
			return false
		}
		conf.Authorization=token.TokenType+" "+token.AccessToken
	}
	conf.Authorization=conf.Bearer+conf.Conf.AliDrive.AccessToken
	log.Infof("token:%s",conf.Authorization)
	user,err:=alidrive.GetUserInfo()
	if err != nil {
		log.Errorf("初始化用户失败:%s",err.Error())
		return false
	}
	log.Infof("当前用户信息:%v",user)
	alidrive.User=user
	return true
}
