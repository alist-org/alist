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

type Native struct {
}

func (n Native) Preview(path string, account *model.Account) (interface{}, error) {
	return nil, fmt.Errorf("no need")
}

func (n Native) Items() []Item {
	return []Item{
		{
			Name:     "root_folder",
			Label:    "root folder path",
			Type:     "string",
			Required: true,
		},
	}
}

func (n Native) Proxy(c *gin.Context, account *model.Account) {
	// unnecessary
}

func (n Native) Save(account *model.Account, old *model.Account) error {
	log.Debugf("save a account: [%s]", account.Name)
	if !utils.Exists(account.RootFolder) {
		account.Status = fmt.Sprintf("[%s] not exist", account.RootFolder)
		_ = model.SaveAccount(account)
		return fmt.Errorf("[%s] not exist", account.RootFolder)
	}
	account.Status = "work"
	err := model.SaveAccount(account)
	if err != nil {
		return err
	}
	return nil
}

// TODO sort files
func (n Native) Path(path string, account *model.Account) (*model.File, []*model.File, error) {
	fullPath := filepath.Join(account.RootFolder, path)
	log.Debugf("%s-%s-%s", account.RootFolder, path, fullPath)
	if !utils.Exists(fullPath) {
		return nil, nil, fmt.Errorf("path not found")
	}
	if utils.IsDir(fullPath) {
		result := make([]*model.File, 0)
		files, err := ioutil.ReadDir(fullPath)
		if err != nil {
			return nil, nil, err
		}
		for _, f := range files {
			if strings.HasPrefix(f.Name(), ".") {
				continue
			}
			time := f.ModTime()
			file := &model.File{
				Name:      f.Name(),
				Size:      f.Size(),
				Type:      0,
				UpdatedAt: &time,
				Driver:    "Native",
			}
			if f.IsDir() {
				file.Type = conf.FOLDER
			} else {
				file.Type = utils.GetFileType(filepath.Ext(f.Name()))
			}
			result = append(result, file)
		}
		return nil, result, nil
	}
	f, err := os.Stat(fullPath)
	if err != nil {
		return nil, nil, err
	}
	time := f.ModTime()
	file := &model.File{
		Name:      f.Name(),
		Size:      f.Size(),
		Type:      utils.GetFileType(filepath.Ext(f.Name())),
		UpdatedAt: &time,
		Driver:    "Native",
	}
	return file, nil, nil
}

func (n Native) Link(path string, account *model.Account) (string, error) {
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

var _ Driver = (*Native)(nil)

func init() {
	RegisterDriver("Native", &Native{})
}
