package _115

import (
	"fmt"

	"github.com/SheltonZhu/115driver/pkg/driver"
	"github.com/alist-org/alist/v3/drivers/base"
	"github.com/pkg/errors"
)

var UserAgent = driver.UA115Desktop

func (d *Pan115) login() error {
	var err error
	opts := []driver.Option{
		driver.UA(UserAgent),
	}
	d.client = driver.New(opts...)
	d.client.SetHttpClient(base.HttpClient)
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
	if d.PageSize <= 0 {
		d.PageSize = driver.FileListLimit
	}
	files, err := d.client.ListWithLimit(fileId, d.PageSize)
	if err != nil {
		return nil, err
	}
	for _, file := range *files {
		res = append(res, file)
	}
	return res, nil
}
