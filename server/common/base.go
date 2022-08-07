package common

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/alist-org/alist/v3/internal/conf"
	"github.com/alist-org/alist/v3/internal/setting"
)

func GetBaseUrl(r *http.Request) string {
	baseUrl := setting.GetByKey(conf.ApiUrl)
	protocol := "http"
	if r.TLS != nil {
		protocol = "https"
	}
	if baseUrl == "" {
		baseUrl = fmt.Sprintf("%s://%s", protocol, r.Host)
	}
	strings.TrimSuffix(baseUrl, "/")
	return baseUrl
}
