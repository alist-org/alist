package alidrive

import (
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"strings"
)

// check password
func HasPassword(files *Files) string {
	fileList := files.Items
	for i, file := range fileList {
		if strings.HasPrefix(file.Name, ".password-") {
			files.Items = fileList[:i+copy(fileList[i:], fileList[i+1:])]
			return file.Name[10:]
		}
	}
	return ""
}

// Deprecated: check readme, implemented by the front end now
func HasReadme(files *Files) string {
	fileList := files.Items
	for _, file := range fileList {
		if file.Name == "Readme.md" {
			resp, err := http.Get(file.Url)
			if err != nil {
				log.Errorf("Get Readme出错:%s", err.Error())
				return ""
			}
			data, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Errorf("读取 Readme出错:%s", err.Error())
				return ""
			}
			return string(data)
		}
	}
	return ""
}
