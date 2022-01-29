package ftp

import (
	"github.com/Xhofe/alist/drivers/base"
	"github.com/Xhofe/alist/model"
	"github.com/jlaffaye/ftp"
)

var connMap map[string]*ftp.ServerConn

func (driver FTP) Login(account *model.Account) (*ftp.ServerConn, error) {
	conn, ok := connMap[account.Name]
	if ok {
		return conn, nil
	}
	conn, err := ftp.Connect(account.SiteUrl)
	if err != nil {
		return nil, err
	}
	err = conn.Login(account.Username, account.Password)
	if err != nil {
		return nil, err
	}
	connMap[account.Name] = conn
	return conn, nil
}

func init() {
	base.RegisterDriver(&FTP{})
	connMap = make(map[string]*ftp.ServerConn)
}
