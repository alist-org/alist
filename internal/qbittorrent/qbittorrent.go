package qbittorrent

import (
	"github.com/alist-org/alist/v3/internal/conf"
	"github.com/alist-org/alist/v3/internal/setting"
)

var qbclient Client

func InitClient() error {
	var err error
	qbclient = nil

	url := setting.GetStr(conf.QbittorrentUrl)
	qbclient, err = New(url)
	return err
}
