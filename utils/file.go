package utils

import (
	"encoding/json"
	"github.com/Xhofe/alist/conf"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// Exists determine whether the file exists
func Exists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

// IsDir determine whether the file is dir
func IsDir(path string) bool {
	s, err := os.Stat(path)
	if err != nil {
		return false
	}
	return s.IsDir()
}

// GetFileType get file type
func GetFileType(ext string) int {
	if ext == "" {
		return conf.UNKNOWN
	}
	ext = strings.ToLower(strings.TrimLeft(ext, "."))
	if IsContain(conf.OfficeTypes, ext) {
		return conf.OFFICE
	}
	if IsContain(conf.AudioTypes, ext) {
		return conf.AUDIO
	}
	if IsContain(conf.VideoTypes, ext) {
		return conf.VIDEO
	}
	if IsContain(conf.TextTypes, ext) {
		return conf.TEXT
	}
	if IsContain(conf.ImageTypes, ext) {
		return conf.IMAGE
	}
	return conf.UNKNOWN
}

// CreatNestedFile create nested file
func CreatNestedFile(path string) (*os.File, error) {
	basePath := filepath.Dir(path)
	if !Exists(basePath) {
		err := os.MkdirAll(basePath, 0700)
		if err != nil {
			log.Errorf("can't create foler，%s", err)
			return nil, err
		}
	}
	return os.Create(path)
}

// WriteToJson write struct to json file
func WriteToJson(src string, conf interface{}) bool {
	data, err := json.MarshalIndent(conf, "", "  ")
	if err != nil {
		log.Errorf("failed convert Conf to []byte:%s", err.Error())
		return false
	}
	err = ioutil.WriteFile(src, data, 0777)
	if err != nil {
		log.Errorf("failed to write json file:%s", err.Error())
		return false
	}
	return true
}

func ParsePath(path string) string {
	path = strings.TrimRight(path, "/")
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return path
}

func RemoveLastSlash(path string) string {
	if len(path) > 1 {
		return strings.TrimSuffix(path, "/")
	}
	return path
}

func Dir(path string) string {
	idx := strings.LastIndex(path, "/")
	if idx == 0 {
		return "/"
	}
	if idx == -1 {
		return path
	}
	return path[:idx]
}

func Base(path string) string {
	idx := strings.LastIndex(path, "/")
	if idx == -1 {
		return path
	}
	return path[idx+1:]
}

func Join(elem ...string) string {
	res := path.Join(elem...)
	if res == "\\" {
		res = "/"
	}
	return res
}

func Split(p string) (string, string) {
	return path.Split(p)
}
