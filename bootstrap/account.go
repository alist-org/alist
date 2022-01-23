package bootstrap

import (
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/drivers/base"
	"github.com/Xhofe/alist/model"
	log "github.com/sirupsen/logrus"
)

func InitAccounts() {
	log.Infof("init accounts...")
	var accounts []model.Account
	if err := conf.DB.Find(&accounts).Error; err != nil {
		log.Fatalf("failed sync init accounts")
	}
	for i, account := range accounts {
		model.RegisterAccount(account)
		driver, ok := base.GetDriver(account.Type)
		if !ok {
			log.Errorf("no [%s] driver", account.Type)
		} else {
			log.Infof("start init account: [%s], type: [%s]", account.Name, account.Type)
			err := driver.Save(&accounts[i], nil)
			if err != nil {
				log.Errorf("init account [%s] error:[%s]", account.Name, err.Error())
			} else {
				log.Infof("success init account: %s, type: %s", account.Name, account.Type)
			}
		}
	}
}
