package model

import (
	"github.com/Xhofe/alist/conf"
)

type Account struct {
	Name           string `json:"name" gorm:"primaryKey" validate:"required"`
	Type           string `json:"type"`
	Username       string `json:"username"`
	Password       string `json:"password"`
	RefreshToken   string `json:"refresh_token"`
	AccessToken    string `json:"access_token"`
	RootFolder     string `json:"root_folder"`
	Status         string
	CronId         int
	DriveId        string
	Limit          int    `json:"limit"`
	OrderBy        string `json:"order_by"`
	OrderDirection string `json:"order_direction"`
}

var accountsMap = map[string]Account{}

// SaveAccount save account to database
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
