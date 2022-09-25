package static

import (
	stdpath "path"
	"strings"

	"github.com/alist-org/alist/v3/internal/conf"
	"github.com/alist-org/alist/v3/internal/setting"
	"github.com/alist-org/alist/v3/pkg/utils"
)

type SiteConfig struct {
	ApiURL   string
	BasePath string
	Cdn      string
}

func getSiteConfig() SiteConfig {
	siteConfig := SiteConfig{
		ApiURL:   conf.Conf.SiteURL,
		BasePath: stdpath.Base(conf.Conf.SiteURL),
		Cdn:      strings.ReplaceAll(strings.TrimSuffix(conf.Conf.Cdn, "/"), "$version", conf.WebVersion),
	}
	// try to get old config
	if siteConfig.ApiURL == "" {
		siteConfig.ApiURL = setting.GetStr(conf.ApiUrl)
		siteConfig.BasePath = setting.GetStr(conf.BasePath)
	}
	if siteConfig.BasePath != "" {
		siteConfig.BasePath = utils.StandardizePath(siteConfig.BasePath)
	}
	if siteConfig.Cdn == "" {
		siteConfig.Cdn = siteConfig.BasePath
	}
	return siteConfig
}
