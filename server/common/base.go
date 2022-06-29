package common

import (
	"fmt"
	"github.com/alist-org/alist/v3/internal/conf"
	"github.com/alist-org/alist/v3/internal/setting"
	"net/http"
	"strings"
)

func GetBaseUrl(r *http.Request) string {
	baseUrl := setting.GetByKey(conf.BaseUrl)
	if baseUrl == "" {
		baseUrl = fmt.Sprintf("//%s", r.Host)
	}
	strings.TrimSuffix(baseUrl, "/")
	return baseUrl
}
