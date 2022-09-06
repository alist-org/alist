package common

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/alist-org/alist/v3/internal/conf"
	"github.com/alist-org/alist/v3/internal/setting"
)

func GetApiUrl(r *http.Request) string {
	api := setting.GetStr(conf.ApiUrl)
	protocol := "http"
	if r != nil {
		if r.TLS != nil {
			protocol = "https"
		}
		if api == "" {
			api = fmt.Sprintf("%s://%s", protocol, r.Host)
		}
	}
	strings.TrimSuffix(api, "/")
	return api
}
