package _115

import (
	"github.com/SheltonZhu/115driver/pkg/driver"
	"github.com/alist-org/alist/v3/drivers/base"
)

func (d *Pan115) login() error {
	opts := []driver.Option{
		driver.WithRestyClient(base.RestyClient),
		driver.UA("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/106.0.0.0 Safari/537.36 115Browser/23.9.3.2 115disk/30.1.0"),
	}

	d.client = driver.New(opts...)

	cr := &driver.Credential{}
	if err := cr.FromCookie(d.Addition.Cookie); err != nil {
		return err
	}
	d.client.ImportCredential(cr)
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
