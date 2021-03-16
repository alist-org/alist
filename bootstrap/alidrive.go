package bootstrap

import (
	"github.com/Xhofe/alist/alidrive"
	"github.com/Xhofe/alist/conf"
	log "github.com/sirupsen/logrus"
)

// init aliyun drive
func InitAliDrive() bool {
	log.Infof("初始化阿里云盘...")
	//首先token_login
	res := alidrive.RefreshTokenAll()
	if res != "" {
		log.Errorf("盘[%s]refresh_token失效,请检查", res)
	}
	log.Debugf("config:%+v", conf.Conf)
	for i, _ := range conf.Conf.AliDrive.Drives {
		InitDriveId(&conf.Conf.AliDrive.Drives[i])
	}
	return true
}

func InitDriveId(drive *conf.Drive) bool {
	user, err := alidrive.GetUserInfo(drive)
	if err != nil {
		log.Errorf("初始化盘[%s]失败:%s", drive.Name, err.Error())
		return false
	}
	drive.DefaultDriveId = user.DefaultDriveId
	log.Infof("初始化盘[%s]成功:%+v", drive.Name, user)
	return true
}
