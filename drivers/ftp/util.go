package ftp

import "github.com/jlaffaye/ftp"

// do others that not defined in Driver interface

func (d *FTP) login() error {
	if d.conn != nil {
		_, err := d.conn.CurrentDir()
		if err == nil {
			return nil
		}
	}
	conn, err := ftp.Dial(d.Address)
	if err != nil {
		return err
	}
	err = conn.Login(d.Username, d.Password)
	if err != nil {
		return err
	}
	d.conn = conn
	return nil
}
