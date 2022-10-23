package _115

import (
	"fmt"

	"github.com/alist-org/alist/v3/internal/conf"
	"github.com/deadblue/elevengo"
	"github.com/deadblue/elevengo/option"
	"github.com/pkg/errors"
)


func (d *Pan115) login() error {
	ua := fmt.Sprintf("Mozilla/5.0 AList/%s", conf.Version)
	d.agent = elevengo.New(option.NameOption(ua))
	credential := &elevengo.Credential{
		UID:  d.UID,
		CID:  d.CID,
		SEID: d.SEID,
	}
	if err := d.agent.CredentialImport(credential); err != nil {
		return errors.Wrap(err, "Import credentail error")
	}
	return nil
}

func (d *Pan115) getFiles(fileId string) ([]File, error) {
	res := make([]File, 0)
	it, err := d.agent.FileIterate(fileId)
	for ; err == nil; err = it.Next() {
		file := elevengo.File{}
		if err = it.Get(&file); err == nil {
			res = append(res, File(file))
		}
	}
	if !elevengo.IsIteratorEnd(err) {
		return nil, errors.Wrap(err, "Iterate files error")
	}
	return res, nil
}
