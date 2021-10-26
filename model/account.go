package model

import (
	"github.com/Xhofe/alist/conf"
	log "github.com/sirupsen/logrus"
)

type Account struct {
	Name         string `json:"name" gorm:"primaryKey" validate:"required"`
	Type         string `json:"type"`
	Username     string `json:"username"`
	Password     string `json:"password"`
	RefreshToken string `json:"refresh_token"`
	AccessToken  string `json:"access_token"`
	RootFolder   string `json:"root_folder"`
	Status       int    `json:"status"`
	CronId       int    `json:"cron_id"`
}

var accountsMap = map[string]Account{}

func SaveAccount(account Account) error {
	if err := conf.DB.Save(account).Error; err != nil {
		return err
	}
	RegisterAccount(account)
	return nil
}

func DeleteAccount(name string) error {
	account := Account{
		Name: name,
	}
	if err := conf.DB.Delete(&account).Error; err != nil {
		return err
	}
	delete(accountsMap, name)
	return nil
}

func AccountsCount() int {
	return len(accountsMap)
}

func RegisterAccount(account Account) {
	accountsMap[account.Name] = account
}

func GetAccount(name string) (Account, bool) {
	if len(accountsMap) == 1 {
		for _, v := range accountsMap {
			return v, true
		}
	}
	account, ok := accountsMap[name]
	return account, ok
}

func GetAccountFiles() []*File {
	files := make([]*File, 0)
	for _, v := range accountsMap {
		files = append(files, &File{
			Name:      v.Name,
			Size:      0,
			Type:      conf.FOLDER,
			UpdatedAt: nil,
		})
	}
	return files
}

func GetAccounts() []*Account {
	accounts := make([]*Account, 0)
	for _, v := range accountsMap {
		accounts = append(accounts, &v)
	}
	return accounts
}

func initAccounts() {
	var accounts []Account
	if err := conf.DB.Find(&accounts).Error; err != nil {
		log.Fatalf("failed sync init accounts")
	}
	for _, account := range accounts {
		RegisterAccount(account)
	}
	log.Debugf("accounts:%+v", accountsMap)
}
