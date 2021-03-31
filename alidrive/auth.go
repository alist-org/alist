package alidrive

import (
	"encoding/json"
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/utils"
	log "github.com/sirupsen/logrus"
)

// refresh access_token token by refresh_token
func RefreshToken(drive *conf.Drive) bool {
	log.Infof("刷新[%s]token...", drive.Name)
	url := "https://auth.aliyundrive.com/v2/account/token"
	req := RefreshTokenReq{RefreshToken: drive.RefreshToken , GrantType: "refresh_token"}
	var token TokenResp
	if body, err := DoPost(url, req, ""); err != nil {
		log.Errorf("tokenLogin-doPost出错:%s", err.Error())
		return false
	} else {
		if err = json.Unmarshal(body, &token); err != nil {
			log.Errorf("解析json[%s]出错:%s", string(body), err.Error())
			log.Errorf("此处json解析失败应该是[%s]refresh_token失效", drive.Name)
			return false
		}
	}
	if token.Code != "" {
		log.Errorf("盘[%s]刷新token出错：%s", drive.Name, token.Message)
		return false
	}
	//刷新成功 更新token
	drive.AccessToken = token.AccessToken
	drive.RefreshToken = token.RefreshToken
	return true
}

func RefreshTokenAll() string {
	log.Infof("刷新所有token...")
	res := ""
	for i, drive := range conf.Conf.AliDrive.Drives {
		if !RefreshToken(&conf.Conf.AliDrive.Drives[i]) {
			res = res + drive.Name + ","
		}
	}
	utils.WriteToYml(conf.ConfigFile, conf.Conf)
	if res != "" {
		return res[:len(res)-1]
	}
	return ""
}
