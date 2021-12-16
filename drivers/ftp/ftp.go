package ftp

import (
	"github.com/Xhofe/alist/drivers/base"
	"github.com/Xhofe/alist/model"
	"github.com/jlaffaye/ftp"
)

func (driver FTP) Login(account *model.Account) (*ftp.ServerConn, error) {
	conn, err := ftp.Connect(account.SiteUrl)
	if err != nil {
		return nil, err
	}
	err = conn.Login(account.Username, account.Password)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func init() {
	base.RegisterDriver(&FTP{})
}