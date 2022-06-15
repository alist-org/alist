package utils

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/alist-org/alist/v3/conf"
	log "github.com/sirupsen/logrus"
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

// CreateNestedFile create nested file
func CreateNestedFile(path string) (*os.File, error) {
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

// CreateTempFile create temp file from io.ReadCloser, and seek to 0
func CreateTempFile(r io.ReadCloser) (*os.File, error) {
	f, err := ioutil.TempFile(conf.Conf.TempDir, "file-*")
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(f, r)
	if err != nil {
		return nil, err
	}
	_, err = f.Seek(0, io.SeekStart)
	if err != nil {
		return nil, err
	}
	return f, nil
}
