package _115

import (
	"fmt"

	"github.com/SheltonZhu/115driver/pkg/driver"
	"github.com/pkg/errors"
)

var UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/106.0.0.0 Safari/537.36 115Browser/23.9.3.2 115disk/30.1.0"

func (d *Pan115) login() error {
	var err error
	opts := []driver.Option{
		driver.UA(UserAgent),
	}
	d.client = driver.New(opts...)
	cr := &driver.Credential{}
	if d.Addition.QRCodeToken != "" {
		s := &driver.QRCodeSession{
			UID: d.Addition.QRCodeToken,
		}
		if cr, err = d.client.QRCodeLogin(s); err != nil {
			return errors.Wrap(err, "failed to login by qrcode")
		}
		d.Addition.Cookie = fmt.Sprintf("UID=%s;CID=%s;SEID=%s", cr.UID, cr.CID, cr.SEID)
		d.Addition.QRCodeToken = ""
	} else if d.Addition.Cookie != "" {
		if err = cr.FromCookie(d.Addition.Cookie); err != nil {
			return errors.Wrap(err, "failed to login by cookies")
		}
		d.client.ImportCredential(cr)
	} else {
		return errors.New("missing cookie or qrcode account")
	}
	return d.client.LoginCheck()
}

func (d *Pan115) getFiles(fileId string) ([]driver.File, error) {
	res := make([]driver.File, 0)
	files, err := d.client.List(fileId)
	if err != nil {
		return nil, err
	}
	for _, file := range *files {
		res = append(res, file)
	}
	return res, nil
}
