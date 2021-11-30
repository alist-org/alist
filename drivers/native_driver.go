package drivers

import (
	"fmt"
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/model"
	"github.com/Xhofe/alist/utils"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type Native struct{}

func (driver Native) Config() DriverConfig {
	return DriverConfig{
		Name: "Native",
		OnlyProxy: true,
	}
}

func (driver Native) Items() []Item {
	return []Item{
		{
			Name:     "root_folder",
			Label:    "root folder path",
			Type:     "string",
			Required: true,
		},
		{
			Name:     "order_by",
			Label:    "order_by",
			Type:     "select",
			Values:   "name,size,updated_at",
			Required: false,
		},
		{
			Name:     "order_direction",
			Label:    "order_direction",
			Type:     "select",
			Values:   "ASC,DESC",
			Required: false,
		},
	}
}

func (driver Native) Save(account *model.Account, old *model.Account) error {
	log.Debugf("save a account: [%s]", account.Name)
	if !utils.Exists(account.RootFolder) {
		account.Status = fmt.Sprintf("[%s] not exist", account.RootFolder)
		_ = model.SaveAccount(account)
		return fmt.Errorf("[%s] not exist", account.RootFolder)
	}
	account.Status = "work"
	account.Proxy = true
	err := model.SaveAccount(account)
	if err != nil {
		return err
	}
	return nil
}

func (driver Native) File(path string, account *model.Account) (*model.File, error) {
	fullPath := filepath.Join(account.RootFolder, path)
	if !utils.Exists(fullPath) {
		return nil, PathNotFound
	}
	f, err := os.Stat(fullPath)
	if err != nil {
		return nil, err
	}
	time := f.ModTime()
	file := &model.File{
		Name:      f.Name(),
		Size:      f.Size(),
		UpdatedAt: &time,
		Driver:    driver.Config().Name,
	}
	if f.IsDir() {
		file.Type = conf.FOLDER
	} else {
		file.Type = utils.GetFileType(filepath.Ext(f.Name()))
	}
	return file, nil
}

func (driver Native) Files(path string, account *model.Account) ([]model.File, error) {
	fullPath := filepath.Join(account.RootFolder, path)
	if !utils.Exists(fullPath) {
		return nil, PathNotFound
	}
	files := make([]model.File, 0)
	rawFiles, err := ioutil.ReadDir(fullPath)
	if err != nil {
		return nil, err
	}
	for _, f := range rawFiles {
		if strings.HasPrefix(f.Name(), ".") {
			continue
		}
		time := f.ModTime()
		file := model.File{
			Name:      f.Name(),
			Size:      f.Size(),
			Type:      0,
			UpdatedAt: &time,
			Driver:    driver.Config().Name,
		}
		if f.IsDir() {
			file.Type = conf.FOLDER
		} else {
			file.Type = utils.GetFileType(filepath.Ext(f.Name()))
		}
		files = append(files, file)
	}
	model.SortFiles(files, account)
	return files, nil
}

func (driver Native) Link(path string, account *model.Account) (string, error) {
	fullPath := filepath.Join(account.RootFolder, path)
	s, err := os.Stat(fullPath)
	if err != nil {
		return "", err
	}
	if s.IsDir() {
		return "", fmt.Errorf("can't down folder")
	}
	return fullPath, nil
}

func (driver Native) Path(path string, account *model.Account) (*model.File, []model.File, error) {
	log.Debugf("native path: %s", path)
	file, err := driver.File(path, account)
	if err != nil {
		return nil, nil, err
	}
	if file.Type != conf.FOLDER {
		//file.Url, _ = driver.Link(path, account)
		return file, nil, nil
	}
	files, err := driver.Files(path, account)
	if err != nil {
		return nil, nil, err
	}
	model.SortFiles(files, account)
	return nil, files, nil
}

func (driver Native) Proxy(c *gin.Context, account *model.Account) {
	// unnecessary
}

func (driver Native) Preview(path string, account *model.Account) (interface{}, error) {
	return nil, fmt.Errorf("no need")
}

var _ Driver = (*Native)(nil)
