package bootstrap

import (
	"github.com/Xhofe/alist/conf"
	"net/http"
)

func InitClient()  {
	conf.Client=&http.Client{}
}