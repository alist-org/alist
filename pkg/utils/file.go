package utils

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/alist-org/alist/v3/internal/conf"
	log "github.com/sirupsen/logrus"
)

// CopyFile File copies a single file from src to dst
func CopyFile(src, dst string) error {
	var err error
	var srcfd *os.File
	var dstfd *os.File
	var srcinfo os.FileInfo

	if srcfd, err = os.Open(src); err != nil {
		return err
	}
	defer srcfd.Close()

	if dstfd, err = CreateNestedFile(dst); err != nil {
		return err
	}
	defer dstfd.Close()

	if _, err = io.Copy(dstfd, srcfd); err != nil {
		return err
	}
	if srcinfo, err = os.Stat(src); err != nil {
		return err
	}
	return os.Chmod(dst, srcinfo.Mode())
}

// CopyDir Dir copies a whole directory recursively
func CopyDir(src string, dst string) error {
	var err error
	var fds []os.FileInfo
	var srcinfo os.FileInfo

	if srcinfo, err = os.Stat(src); err != nil {
		return err
	}
	if err = os.MkdirAll(dst, srcinfo.Mode()); err != nil {
		return err
	}
	if fds, err = ioutil.ReadDir(src); err != nil {
		return err
	}
	for _, fd := range fds {
		srcfp := path.Join(src, fd.Name())
		dstfp := path.Join(dst, fd.Name())

		if fd.IsDir() {
			if err = CopyDir(srcfp, dstfp); err != nil {
				fmt.Println(err)
			}
		} else {
			if err = CopyFile(srcfp, dstfp); err != nil {
				fmt.Println(err)
			}
		}
	}
	return nil
}

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
	if f, ok := r.(*os.File); ok {
		return f, nil
	}
	f, err := os.CreateTemp(conf.Conf.TempDir, "file-*")
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(f, r)
	if err != nil {
		_ = os.Remove(f.Name())
		return nil, err
	}
	_, err = f.Seek(0, io.SeekStart)
	if err != nil {
		_ = os.Remove(f.Name())
		return nil, err
	}
	return f, nil
}

// GetFileType get file type
func GetFileType(filename string) int {
	ext := strings.ToLower(Ext(filename))
	//if SliceContains(conf.TypesMap[conf.OfficeTypes], ext) {
	//	return conf.OFFICE
	//}
	if SliceContains(conf.TypesMap[conf.AudioTypes], ext) {
		return conf.AUDIO
	}
	if SliceContains(conf.TypesMap[conf.VideoTypes], ext) {
		return conf.VIDEO
	}
	if SliceContains(conf.TypesMap[conf.ImageTypes], ext) {
		return conf.IMAGE
	}
	if SliceContains(conf.TypesMap[conf.TextTypes], ext) {
		return conf.TEXT
	}
	return conf.UNKNOWN
}
