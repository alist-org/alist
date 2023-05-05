package plugin_manage

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"

	"github.com/alist-org/alist/v3/internal/conf"
	"github.com/alist-org/alist/v3/internal/model"
)

func CheckPluginMode(path string) (mode string) {
	if fileInfo, _ := os.Stat(filepath.Join(conf.Conf.PluginDir, "src", path)); fileInfo != nil {
		if fileInfo.IsDir() {
			dirs, _ := os.ReadDir(fileInfo.Name())
			for _, dir := range dirs {
				if dir.Name() == "plugin.go" {
					return model.PLUGIN_MODE_YAEGI
				}
			}
		}
	}
	return model.PLUGIN_MDOE_UNKNOWN
}

// 解压zip文件到指定目录
func UnzipArchive(file *os.File, dstDir string) error {
	stat, err := file.Stat()
	if err != nil {
		return err
	}
	zf, err := zip.NewReader(file, stat.Size())
	if err != nil {
		return err
	}
	return UnzipFile(zf, dstDir)
}

func UnzipFile(r *zip.Reader, dstDir string) error {
	for _, file := range r.File {
		if err := unzipFile(file, dstDir); err != nil {
			return err
		}
	}
	return nil
}

// 解压zip文件到指定目录
func unzipFile(file *zip.File, dstDir string) error {
	// create the directory of file
	filePath := filepath.Join(dstDir, file.Name)
	if file.FileInfo().IsDir() {
		if err := os.MkdirAll(filePath, os.ModePerm); err != nil {
			return err
		}
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
		return err
	}

	// open
	rc, err := file.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	// create
	w, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer w.Close()

	// copy
	_, err = io.Copy(w, rc)
	return err
}
