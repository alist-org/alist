package utils

import (
	"fmt"
	"io"
	"mime"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/alist-org/alist/v3/internal/errs"

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

	if _, err = CopyWithBuffer(dstfd, srcfd); err != nil {
		return err
	}
	if srcinfo, err = os.Stat(src); err != nil {
		return err
	}
	return os.Chmod(dst, srcinfo.Mode())
}

// CopyDir Dir copies a whole directory recursively
func CopyDir(src, dst string) error {
	var err error
	var fds []os.DirEntry
	var srcinfo os.FileInfo

	if srcinfo, err = os.Stat(src); err != nil {
		return err
	}
	if err = os.MkdirAll(dst, srcinfo.Mode()); err != nil {
		return err
	}
	if fds, err = os.ReadDir(src); err != nil {
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

// SymlinkOrCopyFile symlinks a file or copy if symlink failed
func SymlinkOrCopyFile(src, dst string) error {
	if err := CreateNestedDirectory(filepath.Dir(dst)); err != nil {
		return err
	}
	if err := os.Symlink(src, dst); err != nil {
		return CopyFile(src, dst)
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

// CreateNestedDirectory create nested directory
func CreateNestedDirectory(path string) error {
	err := os.MkdirAll(path, 0700)
	if err != nil {
		log.Errorf("can't create folder, %s", err)
	}
	return err
}

// CreateNestedFile create nested file
func CreateNestedFile(path string) (*os.File, error) {
	basePath := filepath.Dir(path)
	if err := CreateNestedDirectory(basePath); err != nil {
		return nil, err
	}
	return os.Create(path)
}

// CreateTempFile create temp file from io.ReadCloser, and seek to 0
func CreateTempFile(r io.Reader, size int64) (*os.File, error) {
	if f, ok := r.(*os.File); ok {
		return f, nil
	}
	f, err := os.CreateTemp(conf.Conf.TempDir, "file-*")
	if err != nil {
		return nil, err
	}
	readBytes, err := CopyWithBuffer(f, r)
	if err != nil {
		_ = os.Remove(f.Name())
		return nil, errs.NewErr(err, "CreateTempFile failed")
	}
	if size > 0 && readBytes != size {
		_ = os.Remove(f.Name())
		return nil, errs.NewErr(err, "CreateTempFile failed, incoming stream actual size= %d, expect = %d ", readBytes, size)
	}
	_, err = f.Seek(0, io.SeekStart)
	if err != nil {
		_ = os.Remove(f.Name())
		return nil, errs.NewErr(err, "CreateTempFile failed, can't seek to 0 ")
	}
	return f, nil
}

// GetFileType get file type
func GetFileType(filename string) int {
	ext := strings.ToLower(Ext(filename))
	if SliceContains(conf.SlicesMap[conf.AudioTypes], ext) {
		return conf.AUDIO
	}
	if SliceContains(conf.SlicesMap[conf.VideoTypes], ext) {
		return conf.VIDEO
	}
	if SliceContains(conf.SlicesMap[conf.ImageTypes], ext) {
		return conf.IMAGE
	}
	if SliceContains(conf.SlicesMap[conf.TextTypes], ext) {
		return conf.TEXT
	}
	return conf.UNKNOWN
}

func GetObjType(filename string, isDir bool) int {
	if isDir {
		return conf.FOLDER
	}
	return GetFileType(filename)
}

var extraMimeTypes = map[string]string{
	".apk": "application/vnd.android.package-archive",
}

func GetMimeType(name string) string {
	ext := path.Ext(name)
	if m, ok := extraMimeTypes[ext]; ok {
		return m
	}
	m := mime.TypeByExtension(ext)
	if m != "" {
		return m
	}
	return "application/octet-stream"
}

const (
	KB = 1 << (10 * (iota + 1))
	MB
	GB
	TB
)
