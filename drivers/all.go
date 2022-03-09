package drivers

import (
	_ "github.com/Xhofe/alist/drivers/123"
	_ "github.com/Xhofe/alist/drivers/139"
	_ "github.com/Xhofe/alist/drivers/189"
	_ "github.com/Xhofe/alist/drivers/189pc"
	_ "github.com/Xhofe/alist/drivers/alidrive"
	_ "github.com/Xhofe/alist/drivers/alist"
	_ "github.com/Xhofe/alist/drivers/baidu"
	"github.com/Xhofe/alist/drivers/base"
	_ "github.com/Xhofe/alist/drivers/ftp"
	_ "github.com/Xhofe/alist/drivers/google"
	_ "github.com/Xhofe/alist/drivers/lanzou"
	_ "github.com/Xhofe/alist/drivers/mediatrack"
	_ "github.com/Xhofe/alist/drivers/native"
	_ "github.com/Xhofe/alist/drivers/onedrive"
	_ "github.com/Xhofe/alist/drivers/pikpak"
	_ "github.com/Xhofe/alist/drivers/quark"
	_ "github.com/Xhofe/alist/drivers/s3"
	_ "github.com/Xhofe/alist/drivers/shandian"
	_ "github.com/Xhofe/alist/drivers/teambition"
	_ "github.com/Xhofe/alist/drivers/uss"
	_ "github.com/Xhofe/alist/drivers/webdav"
	_ "github.com/Xhofe/alist/drivers/xunlei"
	_ "github.com/Xhofe/alist/drivers/yandex"
	log "github.com/sirupsen/logrus"
	"strings"
)

var NoCors string
var NoUpload string

func GetConfig() {
	for k, v := range base.GetDriversMap() {
		if v.Config().NoCors {
			NoCors += k + ","
		}
		if v.Upload(nil, nil) != base.ErrEmptyFile {
			NoUpload += k + ","
		}
	}
	NoCors = strings.Trim(NoCors, ",")
	NoUpload += "root"
}

func init() {
	log.Debug("all init")
	GetConfig()
}
