package common

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/alist-org/alist/v3/internal/conf"
	"github.com/alist-org/alist/v3/internal/setting"
)

func GetApiUrl(r *http.Request) string {
	api := conf.Conf.SiteURL
	if api == "" {
		api = setting.GetStr(conf.ApiUrl)
	}
	if r != nil && api == "" {
		protocol := "http"
		if r.TLS != nil {
			protocol = "https"
		}
		api = fmt.Sprintf("%s://%s", protocol, r.Host)

	}
	strings.TrimSuffix(api, "/")
	return api
}
