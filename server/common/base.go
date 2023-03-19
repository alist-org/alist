package common

import (
	"fmt"
	"net/http"
	stdpath "path"
	"strings"

	"github.com/alist-org/alist/v3/internal/conf"
)

func GetApiUrl(r *http.Request) string {
	api := conf.Conf.SiteURL
	if strings.HasPrefix(api, "http") {
		return api
	}
	if r != nil {
		protocol := "http"
		if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
			protocol = "https"
		}
		host := r.Host
		if r.Header.Get("X-Forwarded-Host") != "" {
			host = r.Header.Get("X-Forwarded-Host")
		}
		api = fmt.Sprintf("%s://%s", protocol, stdpath.Join(host, api))
	}
	api = strings.TrimSuffix(api, "/")
	return api
}
