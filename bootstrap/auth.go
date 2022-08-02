package bootstrap

import (
	"fmt"
	"github.com/Xhofe/alist/conf"
	"github.com/casdoor/casdoor-go-sdk/casdoorsdk"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var JwtPublicKey string

// InitAuth init auth
func InitAuth() {
	log.Infof("init auth...")
	auth := conf.Conf.Auth
	if !CheckAuthConfig(auth) {
		panic("invalid auth config")
	}
	if key, err := readJwtPublicKey(auth.JwtPublicKeyPemFile); err != nil {
		panic(err.Error())
	} else {
		JwtPublicKey = key
	}
	casdoorsdk.InitConfig(auth.Endpoint, auth.ClientId, auth.ClientSecret, JwtPublicKey, auth.OrganizationName, auth.ApplicationName)
}

func readJwtPublicKey(location string) (string, error) {
	var path string

	if location != "" {
		path = strings.TrimRight(location, string(filepath.Separator))
	} else {
		currentPath, err := exec.LookPath(os.Args[0])
		if err != nil {
			return "", fmt.Errorf("failed to get public key file path: %s", err.Error())
		}

		currentPath = strings.TrimRight(currentPath, string(filepath.Separator))
		path = filepath.Join(currentPath, "token_jwt_key.pem")
	}

	stat, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("public key file is not exist: %s", path)
		} else {
			return "", fmt.Errorf("failed to get the status of the public key file: %s", err.Error())
		}
	}

	if stat.IsDir() {
		return "", fmt.Errorf("public key file is not a file")
	}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read public key: %s", err.Error())
	}

	return string(data), nil
}

func CheckAuthConfig(authConfig conf.AuthConfig) bool {
	return authConfig.Endpoint != "" &&
		authConfig.ClientId != "" &&
		authConfig.ClientSecret != "" &&
		authConfig.OrganizationName != "" &&
		authConfig.ApplicationName != ""
}
