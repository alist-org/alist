package bootstrap

import (
	"github.com/Xhofe/alist/conf"
	"github.com/casdoor/casdoor-go-sdk/casdoorsdk"
	log "github.com/sirupsen/logrus"
	"strings"
)

type AuthConfig struct {
	OrganizationName string `json:"organization_name"`
	ApplicationName  string `json:"application_name"`
	Endpoint         string `json:"endpoint"`
	ClientId         string `json:"client_id"`
	ClientSecret     string `json:"client_secret"`
	Certificate      string `json:"certificate"`
}

// InitAuth init auth
func InitAuth() {
	log.Infof("init auth...")
	enableCasdoor := conf.GetBool("Enable Casdoor")
	if !enableCasdoor {
		return
	}
	auth := readAuthConfig()
	if !CheckAuthConfig(auth) {
		panic("invalid auth config")
	}
	casdoorsdk.InitConfig(strings.TrimRight(auth.Endpoint, "/"), auth.ClientId, auth.ClientSecret, auth.Certificate, auth.OrganizationName, auth.ApplicationName)
}

func readAuthConfig() AuthConfig {
	return AuthConfig{
		OrganizationName: conf.GetStr("Casdoor Organization name"),
		ApplicationName:  conf.GetStr("Casdoor Application name"),
		Endpoint:         conf.GetStr("Casdoor Endpoint"),
		ClientId:         conf.GetStr("Casdoor Client id"),
		ClientSecret:     conf.GetStr("Casdoor Client secret"),
		Certificate:      conf.GetStr("Casdoor Certificate"),
	}
}

func CheckAuthConfig(authConfig AuthConfig) bool {
	return authConfig.Endpoint != "" &&
		authConfig.ClientId != "" &&
		authConfig.ClientSecret != "" &&
		authConfig.OrganizationName != "" &&
		authConfig.ApplicationName != "" &&
		authConfig.Certificate != ""
}
