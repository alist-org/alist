package utils

import (
	"bytes"
	"encoding/json"
	"github.com/Xhofe/alist/conf"
	log "github.com/sirupsen/logrus"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"io/ioutil"
	"os"
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
	ext = strings.ToLower(strings.TrimLeft(ext,"."))
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

func IsGBK(data []byte) bool {
	length := len(data)
	var i int = 0
	for i < length {
		//fmt.Printf("for %x\n", data[i])
		if data[i] <= 0xff {
			//编码小于等于127,只有一个字节的编码，兼容ASCII吗
			i++
			continue
		} else {
			//大于127的使用双字节编码
			if  data[i] >= 0x81 &&
				data[i] <= 0xfe &&
				data[i + 1] >= 0x40 &&
				data[i + 1] <= 0xfe &&
				data[i + 1] != 0xf7 {
				i += 2
				continue
			} else {
				return false
			}
		}
	}
	return true
}

func GbkToUtf8(s []byte) ([]byte, error) {
	reader := transform.NewReader(bytes.NewReader(s), simplifiedchinese.GBK.NewDecoder())
	d, e := ioutil.ReadAll(reader)
	if e != nil {
		return nil, e
	}
	return d, nil
}
